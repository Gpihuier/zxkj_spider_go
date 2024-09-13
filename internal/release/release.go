package release

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/gookit/slog"
	"github.com/imroc/req/v3"
	"github.com/sourcegraph/conc/pool"
	"gorm.io/gorm"
	"zxkj.com/zxkj_spider_go/internal/collect/model"
	"zxkj.com/zxkj_spider_go/internal/service"
	"zxkj.com/zxkj_spider_go/pkg/helper"
)

type Release struct {
	App       *service.App
	pool      *pool.ErrorPool
	req       *req.Client
	rateLimit *time.Ticker
}

func NewRelease(app *service.App) *Release {
	return &Release{
		App:  app,
		pool: pool.New().WithMaxGoroutines(app.Cfg.Server.MaxThreads).WithErrors(),
		req: req.C().SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36").
			SetCommonRetryCount(3).
			OnAfterResponse(func(client *req.Client, resp *req.Response) error {
				if resp.Err != nil { // Ignore when there is an underlying error, e.g. network error.
					return nil
				}
				// Treat non-successful responses as errors, record raw dump content in error message.
				if !resp.IsSuccessState() { // Status code is not between 200 and 299.
					resp.Err = fmt.Errorf("bad response, raw content:\n%d", resp.StatusCode)
				}
				return nil
			}),
		rateLimit: time.NewTicker(time.Second * 1),
	}
}

func (r *Release) Release(ctx context.Context, tb string, us []string) {

	// 查询三天前的日期的数据
	threeDayAgo := helper.NowAddDay(-3)
	data := make([]map[string]any, 0)
	err := r.App.DB.Table(tb).Select("`id`, `url`,`content`").Where("`create_time` >= ?", threeDayAgo).Find(&data).Error
	if err != nil {
		slog.Error(err)
		return
	}

	// 将数据发送到目标站点
	for _, u := range us {
		r.pool.Go(func() error {
			return r.PushSite(ctx, tb, u, data)
		})
	}

}

func (r *Release) PushSite(ctx context.Context, tb, site string, data []map[string]any) error {

	parse, err := url.Parse(site)
	if err != nil {
		return err
	}

	// 提取域名
	key := parse.Hostname()

	for _, item := range data {
		if err = r.handlePushSite(ctx, tb, key, item, site); err != nil {
			slog.Error(err)
		}

	}

	return nil
}

func (r *Release) handlePushSite(ctx context.Context, tb, key string, item map[string]any, siteUrl string) error {

	if r.rateLimit != nil {
		select {
		case <-r.rateLimit.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// 查询推送记录
	var pushSite model.PushSite
	err := r.App.DB.Where("data_id = ? AND data_table = ? AND domain = ?", item["id"].(int64), tb, key).First(&pushSite).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 没有找到记录，执行推送并创建记录
		return r.pushAndCreateSite(tb, key, item, siteUrl)
	} else if err != nil {
		return err
	}

	// 找到记录
	if pushSite.State == 1 {
		return nil
	} else if pushSite.Retry >= 10 {
		// 如果重试次数超过10次，返回错误
		return fmt.Errorf("推送失败，重试次数超过10次")
	}

	// 推送并更新记录
	return r.pushAndUpdateSite(key, item, siteUrl, &pushSite)
}

func (r *Release) pushAndCreateSite(tb, key string, item map[string]any, siteUrl string) error {
	dat := r.buildForm(key, item["content"].(string), item["url"].(string))
	resp, err := r.request(siteUrl, dat)
	if err != nil {
		return err
	}
	// 创建推送记录
	return r.createPushSite(resp, item["id"].(int64), tb, key)
}

func (r *Release) pushAndUpdateSite(key string, item map[string]any, siteUrl string, pushSite *model.PushSite) error {
	dat := r.buildForm(key, item["content"].(string), item["url"].(string))
	resp, err := r.request(siteUrl, dat)
	if err != nil {
		return err
	}
	// 更新推送记录
	return r.updatePushSite(resp, pushSite)
}

func (r *Release) createPushSite(resp map[string]any, id int64, tb, domain string) error {
	pushSite := model.PushSite{
		Retry:     1,
		DataId:    id,
		Domain:    domain,
		DataTable: tb,
	}
	if resp["code"].(float64) == 200 {
		pushSite.State = 1
	}
	return r.App.DB.Create(&pushSite).Error
}

func (r *Release) updatePushSite(resp map[string]any, pushSite *model.PushSite) error {
	if resp["code"].(float64) == 200 {
		pushSite.State = 1
	} else {
		pushSite.Retry++
	}
	return r.App.DB.Save(pushSite).Error
}

func (r *Release) request(url string, data map[string]string) (map[string]any, error) {
	resp, err := r.req.R().
		SetHeader("Content-Type", "application/json").
		SetFormData(data).
		Post(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	slog.Infof("推送站点：%s, 返回数据：%s", url, string(body))

	res := make(map[string]any)
	err = json.Unmarshal(body, &res)

	return res, err
}

// 构建表单数据
func (r *Release) buildForm(key string, content, url string) map[string]string {
	now := fmt.Sprintf("%d", time.Now().Unix())
	raw := content + key + now + url
	md5h := md5.Sum([]byte(raw))
	sign := hex.EncodeToString(md5h[:])
	return map[string]string{"time": now, "data": content, "hash": sign, "page": url}
}

func (r *Release) Wait() error {
	return r.pool.Wait()
}
