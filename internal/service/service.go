package service

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"zxkj.com/zxkj_spider_go/internal/pkg/config"
	"zxkj.com/zxkj_spider_go/internal/pkg/logger"
)

type App struct {
	Cfg   *config.Config
	Log   *logger.Logger
	DB    *gorm.DB
	Cache *redis.Client
}

func NewApp(cfg *config.Config, db *gorm.DB, log *logger.Logger, cache *redis.Client) *App {
	return &App{
		Cfg:   cfg,
		Log:   log,
		DB:    db,
		Cache: cache,
	}
}
