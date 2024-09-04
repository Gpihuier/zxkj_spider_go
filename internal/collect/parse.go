package collect

import (
	"errors"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
)

type Parse struct {
}

func NewParse() *Parse {
	return &Parse{}
}

func (p *Parse) ParseListUrl(url string) ([]string, error) {
	if url == "" {
		return nil, nil
	}

	// 使用正则表达式提取出括号内的内容
	re, err := regexp.Compile(`\{([\S|]+)}`)
	if err != nil {
		return nil, err
	}
	matched := re.FindStringSubmatch(url)
	if len(matched) <= 1 {
		return nil, errors.New("no match found")
	}

	// 分割字符串
	parts := strings.Split(matched[1], "|")
	urls := make([]string, 0, len(parts))

	// 替换大括号内容，并生成每个URL
	for _, s := range parts {
		fr := strings.Replace(url, matched[0], s, 1)
		urls = append(urls, fr)
	}

	return urls, nil
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
		if strings.HasPrefix(xpath, "array:") {
			arrays := htmlquery.Find(doc, strings.TrimPrefix(xpath, "array:"))
			array := make([]string, 0, len(arrays))
			for _, node := range arrays {
				if node != nil {
					array = append(array, htmlquery.InnerText(node))
				}
			}
			data[key] = array
		} else {
			node := htmlquery.FindOne(doc, xpath)
			if node != nil {
				data[key] = htmlquery.InnerText(node)
			}
		}

	}

	return data, nil
}

func (p *Parse) ParseProcess(data map[string]any, template Template) map[string]any {

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
	}

	return data
}
