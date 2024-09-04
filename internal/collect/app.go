package collect

import (
	"context"
	"time"

	"zxkj.com/zxkj_spider_go/internal/service"
)

func Run(ctx context.Context, app *service.App) {
	req := NewRequest()
	if app.Cfg.Server.Mode == "dev" {
		req = req.SetDevMode()
	}

	crawler := NewCollect(app,
		WithReq(req),
		WithParse(NewParse()),
		WithPipeline(NewMySQLPipeline()),
		WithMiddleware(NewRedisMiddleware(app.Cache)),
		WithRateLimiter(time.Second*time.Duration(app.Cfg.Server.RateLimit)),
	)

	// 获取任务
	for task := range LoadTasks() {
		crawler.Crawler(ctx, task)
	}

}
