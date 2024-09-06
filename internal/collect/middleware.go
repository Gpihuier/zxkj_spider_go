package collect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var cacheKeyPrefix = "zxkj.com/zxkj_spider_go:"

var (
	ErrNoCache     = errors.New("no cache")
	ErrCacheExists = errors.New("cache exists")
)

type Middleware interface {
	Before(ctx context.Context, url string) error
	After(ctx context.Context, item *Item) error
}

type RedisMiddleware struct {
	Client *redis.Client
}

func NewRedisMiddleware(client *redis.Client) *RedisMiddleware {
	return &RedisMiddleware{
		Client: client,
	}
}
func (r *RedisMiddleware) Before(ctx context.Context, url string) error {
	cacheKey := cacheKeyPrefix + url
	_, err := r.Client.Get(ctx, cacheKey).Result()

	if errors.Is(err, redis.Nil) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}

	return ErrCacheExists
}

func (r *RedisMiddleware) After(ctx context.Context, item *Item) error {
	dj, err := json.Marshal(item.Data)
	if err != nil {
		return err
	}

	return r.Client.Set(ctx, cacheKeyPrefix+item.Url, string(dj), 0).Err()
}

type ProcessMiddleware struct {
}

func NewProcessMiddleware() *ProcessMiddleware {
	return &ProcessMiddleware{}
}

func (r *ProcessMiddleware) Before(ctx context.Context, url string) error {
	return nil
}

func (r *ProcessMiddleware) After(ctx context.Context, item *Item) error {
	return nil
}
