package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gookit/config/v2/yaml"
	"zxkj.com/zxkj_spider_go/internal/pkg/cache"
	"zxkj.com/zxkj_spider_go/internal/pkg/config"
	"zxkj.com/zxkj_spider_go/internal/pkg/db"
	"zxkj.com/zxkj_spider_go/internal/pkg/logger"
	"zxkj.com/zxkj_spider_go/internal/release"
	"zxkj.com/zxkj_spider_go/internal/service"
)

var cfg string

func initApp(c config.Config) (*service.App, func(), error) {
	r, err := cache.NewRedis(&c)
	if err != nil {
		return nil, nil, err
	}
	g, clean, err := db.NewGorm(&c)
	if err != nil {
		return nil, nil, err
	}
	l := logger.New(c.Log.File)

	return service.NewApp(&c, g, l, r), func() {
		clean()
	}, nil
}

func main() {
	flag.StringVar(&cfg, "config", "../../config/config.yaml", "path to config.yaml file")
	flag.Parse()

	c, err := config.New[config.Config](yaml.Driver, cfg)
	if err != nil {
		panic(err)
	}

	app, clean, err := initApp(c)
	if err != nil {
		panic(err)
	}

	defer func() {
		clean()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals
		fmt.Println("发布服务停止")
		cancel()
	}()

	Run(ctx, app)
}

func Run(ctx context.Context, app *service.App) {
	for {
		now := time.Now()
		hour := now.Hour()

		// 检查是否在 08:00 到 20:00 之间
		if hour >= 8 && hour < 20 {
			fmt.Println("开始发布:", now)
			release.Run(ctx, app)
		} else {
			fmt.Printf("当前时间 %s，不在 08:00 到 20:00 之间  ，跳过执行\n", now.Format("15:04"))
		}

		// 每次任务执行完成后，等待一小时再执行下一次任务
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Hour):
			// 执行下一个任务
		}
	}
}
