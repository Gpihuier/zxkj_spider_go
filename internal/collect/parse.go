package collect

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"zxkj.com/zxkj_spider_go/pkg/helper"
)

type Parse struct {
}

func NewParse() *Parse {
	return &Parse{}
}

func (p *Parse) ParseListUrl(urls []string) ([]string, error) {
	if len(urls) == 0 {
		return nil, errors.New("urls is empty")
	}

	res := make([]string, 0)

	for _, url := range urls {
		if url == "" {
			return nil, nil
		}

		// 使用正则表达式提取出括号内的内容
		re, err := regexp.Compile(`\{([\S|]+)}`)
		if err != nil {
			return nil, err
		}
		matched := re.FindStringSubmatch(url)
		// 如果没有匹配到大括号，直接输出原数据
		if len(matched) <= 1 {
			res = append(res, url)
			continue
		}

		// 分割字符串
		parts := strings.Split(matched[1], "|")

		// 替换大括号内容，并生成每个URL
		for _, s := range parts {
			fr := strings.Replace(url, matched[0], s, 1)
			res = append(res, fr)
		}
	}

	return res, nil
}
func (p *Parse) ParseList(html string, template Template) ([]string, error) {
	doc, err := htmlquery.Parse(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	nodes := htmlquery.Find(doc, template.List.Node)
	if nodes == nil {
		return nil, err
	}

	urls := make([]string, 0, len(nodes))

	for _, node := range nodes {
		href := htmlquery.InnerText(htmlquery.FindOne(node, template.List.Href))
		urls = append(urls, href)
	}

	return urls, nil
}

func (p *Parse) ParseContext(html string, template Template) (map[string]any, error) {
	doc, err := htmlquery.Parse(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	data := make(map[string]any, len(template.Content))
	for key, xpath := range template.Content {
		xpath = strings.TrimSpace(xpath)
		if xpath == "" {
			data[key] = ""
			continue
		}

		// 判断是否是数组类型
		if strings.HasPrefix(xpath, "array:") {
			arrays := htmlquery.Find(doc, strings.TrimPrefix(xpath, "array:"))
			array := make([]string, 0, len(arrays))
			for _, node := range arrays {
				if node != nil {
					array = append(array, htmlquery.InnerText(node))
				}
			}
			data[key] = array
		} else if strings.HasPrefix(xpath, "html:") {
			node := htmlquery.FindOne(doc, strings.TrimPrefix(xpath, "html:"))
			if node != nil {
				data[key] = htmlquery.OutputHTML(node, false)
			}
		} else if strings.HasPrefix(xpath, "text:") {
			data[key] = strings.TrimPrefix(strings.TrimSpace(xpath), "text:")
		} else {
			node := htmlquery.FindOne(doc, xpath)
			if node != nil {
				data[key] = htmlquery.InnerText(node)
			}
		}
	}

	return data, nil
}

func (p *Parse) ParseProcess(data map[string]any, domainUrl string, template Template) (map[string]any, error) {

	var err error

	fn := func(text string, fn Process) string {
		for _, trim := range fn.Trim {
			// text = strings.Trim(text, trim)
			text = strings.ReplaceAll(text, trim, "")
		}

		// 处理 replace 规则
		for o, n := range fn.Replace {
			text = strings.ReplaceAll(text, o, n)
		}

		return text
	}

	for key, val := range data {
		switch v := val.(type) {
		case string:
			if config, ok := template.Process[key]; ok {
				data[key] = fn(v, config)
			} else {
				data[key] = strings.TrimSpace(v)
			}
		default:
			data[key] = v
		}

		// 再次处理数据
		if key == "title" {
			if data[key].(string) == "暂未上线" {
				return nil, errors.New("暂未上线")
			}
		}

		if key == "keyword" && data[key] == "" {
			data[key] = strings.ReplaceAll(data["title"].(string), "-", ",")
		}

		if key == "filesize" {
			data[key], err = p.ParseFilesize(data[key].(string))
			if err != nil {
				return nil, err
			}
		}

		if key == "downfiles" {
			data[key], err = p.ParseDownfiles(data[key].(string))
			if err != nil {
				return nil, err
			}
		}

		if key == "content" {
			data[key], err = p.ParseContent(data[key].(string), domainUrl, data["title"].(string))
			if err != nil {
				return nil, err
			}
		}

		if key == "images" {
			data[key] = p.ParseImages(data[key].([]string), domainUrl)
		}

		if key == "image" {
			data[key] = helper.MustCompleteURL(domainUrl, data[key].(string))
		}

		if key == "release_at" && data[key] == "" {
			// 取当前时间
			data[key] = time.Now().Format("2006-01-02")
		}
	}

	return data, nil
}

func (p *Parse) ParseImages(images []string, domainUrl string) []string {
	for i, image := range images {
		images[i] = helper.MustCompleteURL(domainUrl, image)
	}
	return images
}

func (p *Parse) ParseContent(content string, domainUrl string, title string) (string, error) {

	soup, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return "", err
	}

	// Find and remove the tag of the first <div> element (if it exists)
	soup.Find("div").First().Remove()

	// Find and remove the tag of the first <iframe> element (if it exists)
	soup.Find("iframe").First().Remove()

	// Find and remove the tag of the first <script> element (if it exists)
	soup.Find("script").First().Remove()

	// Remove specified tags
	soup.Find("a, dd").Each(func(i int, s *goquery.Selection) {
		s.Unwrap()
	})

	// Process img tags and wrap them in <p> tags with center alignment
	soup.Find("img").Each(func(_ int, s *goquery.Selection) {
		src := s.AttrOr("data-src", s.AttrOr("src", ""))
		src = helper.MustCompleteURL(domainUrl, src)
		s.SetAttr("src", src)
		s.SetAttr("alt", title)
	})

	// Remove attributes for p, h3, span tags
	soup.Find("p, h3").Each(func(i int, s *goquery.Selection) {
		// 清空属性
		s.RemoveAttr("style")
		s.RemoveAttr("align")
		// 如果标签中包含 img，则添加 style="text-align:center"
		if s.Find("img").Length() > 0 {
			s.SetAttr("style", "text-align:center")
		}
	})

	// Style the first <p> tag
	firstP := soup.Find("p").First()
	if firstP.Length() > 0 {
		firstP.SetAttr("style", "text-indent: 2em;")
	}

	// 选择需要保留的标签
	var selectedTags []string
	soup.Find("div, p, h3").Each(func(i int, s *goquery.Selection) {
		// 获取当前标签的 HTML
		html, err := goquery.OuterHtml(s)
		if err == nil {
			// 去掉 \n 和 \t
			html = strings.ReplaceAll(html, "\n", "")
			html = strings.ReplaceAll(html, "\t", "")
			selectedTags = append(selectedTags, html)
		}
	})

	// 拼接保留的 HTML 内容
	html := strings.Join(selectedTags, "")

	return html, nil

}

func (p *Parse) ParseFilesize(filesize string) (string, error) {
	if filesize == "" {
		return "", errors.New("filesize is empty")
	}

	if filesize == "0KB" || filesize == "0" {
		return "", errors.New("filesize is empty")
	}

	// 使用正则表达式去除数字和小数点
	re := regexp.MustCompile(`[0-9.]`)
	size := re.ReplaceAllString(filesize, "")
	size = strings.TrimSpace(size)

	var formattedFilesize string
	if len(size) == 1 {
		formattedFilesize = strings.TrimSuffix(filesize, size) + " " + size + "B"
	} else {
		formattedFilesize = strings.TrimSuffix(filesize, size) + " " + size
	}
	return formattedFilesize, nil
}

func (p *Parse) ParseDownfiles(downfiles string) (string, error) {
	if downfiles == "#" || downfiles == "javascript:;" || downfiles == "javascript:void(0)" {
		return "", errors.New("downfiles is empty")
	}
	return downfiles, nil
}
