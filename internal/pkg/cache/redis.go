package cache

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"zxkj.com/zxkj_spider_go/internal/pkg/config"
)

func NewRedis(cfg *config.Config) (*redis.Client, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Db,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("cache drive init is panic: %s", err)
	}
	return client, nil
}
