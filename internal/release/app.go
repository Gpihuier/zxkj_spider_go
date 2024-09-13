package release

import (
	"context"

	"github.com/gookit/config/v2/yaml"
	"github.com/gookit/slog"

	"github.com/sourcegraph/conc/pool"
	"zxkj.com/zxkj_spider_go/internal/pkg/config"
	"zxkj.com/zxkj_spider_go/internal/service"
)

func Run(ctx context.Context, app *service.App) {
	tasks, err := config.New[Config](yaml.Driver, "../../config/release/tasks.yaml")
	if err != nil {
		slog.Error(err)
		return
	}

	release := NewRelease(app)

	ps := pool.New().WithMaxGoroutines(app.Cfg.Server.MaxThreads)

	for key, val := range tasks.List {
		ps.Go(func() {
			release.Release(ctx, key, val)
		})
	}

	ps.Wait()

	if err = release.Wait(); err != nil {
		slog.Error(err)
	}

}
