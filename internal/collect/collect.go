package collect

import (
	"context"
	"fmt"
	"time"

	"github.com/gookit/slog"
	"github.com/sourcegraph/conc/pool"
	"zxkj.com/zxkj_spider_go/internal/service"
	"zxkj.com/zxkj_spider_go/pkg/helper"
)

type Data struct {
	Url  string
	Html string
	Data map[string]any
	Task *Config
}

type Collect struct {
	App        *service.App
	Req        *Request
	Parse      *Parse
	Pipeline   Pipeline
	Middleware []Middleware
	RateLimit  *time.Ticker    // 爬取速率控制
	pool       *pool.ErrorPool // Goroutine pool
	queue      chan string
}

type Option func(*Collect)

func WithReq(req *Request) Option {
	return func(c *Collect) {
		c.Req = req
	}
}

func WithParse(p *Parse) Option {
	return func(c *Collect) {
		c.Parse = p
	}
}

func WithPipeline(p Pipeline) Option {
	return func(c *Collect) {
		c.Pipeline = p
	}
}

func WithMiddleware(ms ...Middleware) Option {
	return func(c *Collect) {
		c.Middleware = append(c.Middleware, ms...)
	}
}

func WithRateLimiter(rate time.Duration) Option {
	return func(c *Collect) {
		c.RateLimit = time.NewTicker(rate)
	}
}

func NewCollect(app *service.App, opts ...Option) *Collect {

	c := &Collect{
		App:        app,
		Middleware: make([]Middleware, 0),
		pool:       pool.New().WithMaxGoroutines(app.Cfg.Server.MaxThreads).WithErrors(),
		queue:      make(chan string, 100),
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

func (c *Collect) Crawler(ctx context.Context, task *Config) {

	// 解析Url
	softUrls, err := c.Parse.ParseListUrl(task.Soft.Url)
	if err != nil {
		slog.Error(err)
		return
	}
	for _, url := range softUrls {
		if err = c.fetchList(ctx, url, task.DomainUrl, task.Soft); err != nil {
			slog.Error(err)
			continue
		}
	}

	if err = c.pool.Wait(); err != nil {
		slog.Error(err)
	}

}

// fetchList
func (c *Collect) fetchList(ctx context.Context, url, domainUrl string, template Template) error {
	resp, err := c.Req.Get(ctx, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	list, err := c.Parse.ParseList(resp.String(), template)
	if err != nil {
		return err
	}

	for _, item := range list {
		href, err := helper.CompleteURL(domainUrl, item)
		if err != nil {
			return err
		}
		// 解析内容页
		c.pool.Go(func() error {
			return c.fetchContent(ctx, href, domainUrl, template)
		})
	}

	return nil
}

func (c *Collect) fetchContent(ctx context.Context, url, domainUrl string, template Template) error {

	if c.RateLimit != nil {
		select {
		case <-c.RateLimit.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	for _, m := range c.Middleware {
		if err := m.Before(ctx, url); err != nil {
			return err
		}
	}

	resp, err := c.Req.Get(ctx, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := c.Parse.ParseContext(resp.String(), template)
	if err != nil {
		return err
	}

	data = c.Parse.ParseProcess(data, template)

	for _, m := range c.Middleware {
		if err := m.After(ctx, url, resp.String()); err != nil {
			return err
		}
	}

	fmt.Println(data, 777)

	return nil
}
