package cache

import (
	"context"
	"backend/repository"
)

// Cache 定义了缓存接口
type Cache interface {
	Get(ctx context.Context, key string) (*repository.SearchResult, error)
	Set(ctx context.Context, key string, value *repository.SearchResult, ttl int) error
	Delete(ctx context.Context, key string) error
}
