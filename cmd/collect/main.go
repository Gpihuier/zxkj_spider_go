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
	"zxkj.com/zxkj_spider_go/internal/collect"
	"zxkj.com/zxkj_spider_go/internal/pkg/cache"
	"zxkj.com/zxkj_spider_go/internal/pkg/config"
	"zxkj.com/zxkj_spider_go/internal/pkg/db"
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

	return service.NewApp(&c, g, r), func() {
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(c.Server.MaxTimeout))
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals
		fmt.Println("Received termination signal")
		cancel()
	}()

	// 开始采集
	collect.Run(ctx, app)
}
