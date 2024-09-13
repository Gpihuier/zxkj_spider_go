package collect

import (
	"context"
	"time"

	"github.com/gookit/slog"
	"github.com/sourcegraph/conc/pool"
	"zxkj.com/zxkj_spider_go/internal/collect/model"
	"zxkj.com/zxkj_spider_go/internal/service"
)

func Run(ctx context.Context, app *service.App) {
	if err := model.Migrator(app.DB); err != nil {
		slog.Error(err)
		return
	}

	req := NewRequest()
	if app.Cfg.Server.Mode == "dev" {
		req = req.SetDevMode()
	}

	crawler := NewCollect(app,
		WithReq(req),
		WithParse(NewParse()),
		WithPipeline(NewMySQLPipeline(app.DB)),
		WithMiddleware(NewRedisMiddleware(app.Cache)),
		WithRateLimiter(time.Second*time.Duration(app.Cfg.Server.RateLimit)),
	)

	ps := pool.New().WithMaxGoroutines(app.Cfg.Server.MaxThreads)

	// 获取任务
	for task := range LoadTasks() {
		ps.Go(func() {
			crawler.Crawler(ctx, task)
		})
	}

	ps.Wait()
}
