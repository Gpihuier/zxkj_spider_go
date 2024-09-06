package collect

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gookit/slog"
	"github.com/sourcegraph/conc/pool"
	"zxkj.com/zxkj_spider_go/internal/service"
	"zxkj.com/zxkj_spider_go/pkg/helper"
)

type Item struct {
	Name      string
	Url       string
	DomainUrl string
	Data      map[string]any
	Template  Template
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
	softUrls, err := c.Parse.ParseListUrl(task.Soft.Urls)
	if err != nil {
		slog.Error(err)
		return
	}
	for _, url := range softUrls {
		if err = c.fetchList(ctx, &Item{
			Name:      task.Name,
			Url:       url,
			DomainUrl: task.DomainUrl,
			Template:  task.Soft,
		}); err != nil {
			slog.Error(err)
			continue
		}
	}

	if err = c.pool.Wait(); err != nil {
		slog.Error(err)
	}

}

// fetchList
func (c *Collect) fetchList(ctx context.Context, item *Item) error {
	resp, err := c.Req.Get(ctx, item.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	list, err := c.Parse.ParseList(resp.String(), item.Template)
	if err != nil {
		return err
	}

	for _, ls := range list {
		href, err := helper.CompleteURL(item.DomainUrl, ls)
		if err != nil {
			return err
		}
		// 解析内容页
		c.pool.Go(func() error {
			err = c.fetchContent(ctx, &Item{
				Name:      item.Name,
				Url:       href,
				DomainUrl: item.DomainUrl,
				Template:  item.Template,
			})
			if err != nil {
				slog.Error(err)
				return err
			}
			return nil
		})
	}

	return nil
}

func (c *Collect) fetchContent(ctx context.Context, item *Item) error {
	for _, m := range c.Middleware {
		if err := m.Before(ctx, item.Url); err != nil {
			return errors.Join(err, errors.New(item.Url))
		}
	}

	if c.RateLimit != nil {
		select {
		case <-c.RateLimit.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	resp, err := c.Req.Get(ctx, item.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := c.Parse.ParseContext(resp.String(), item.Template)
	if err != nil {
		return err
	}

	// TODO how to add data in cache if err != nil
	item.Data, err = c.Parse.ParseProcess(data, item.DomainUrl, item.Template)
	if err != nil {
		return fmt.Errorf("url: %s, err: %v", item.Url, err)
	}

	// 写入管道pipeline
	if err = c.Pipeline.Process(item); err != nil {
		return err
	}

	for _, m := range c.Middleware {
		if err = m.After(ctx, item); err != nil {
			return err
		}
	}

	return nil
}
