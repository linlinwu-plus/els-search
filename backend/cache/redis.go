package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/repository"

	"github.com/go-redis/redis/v8"
)

// redisCache 实现了 Cache 接口
type redisCache struct {
	client *redis.Client
	ttl    int
}

// NewRedisCache 创建 Redis 缓存实例
func NewRedisCache(addr, password string, db, ttl int) Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &redisCache{
		client: client,
		ttl:    ttl,
	}
}

// Get 从缓存中获取数据
func (c *redisCache) Get(ctx context.Context, key string) (*repository.SearchResult, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // 缓存未命中
	} else if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var result repository.SearchResult
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return &result, nil
}

// Set 将数据存入缓存
func (c *redisCache) Set(ctx context.Context, key string, value *repository.SearchResult, ttl int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if ttl <= 0 {
		ttl = c.ttl
	}

	if err := c.client.Set(ctx, key, data, time.Duration(ttl)*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete 从缓存中删除数据
func (c *redisCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}
