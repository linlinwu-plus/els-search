package service

import (
	"backend/analytics"
	"backend/cache"
	"backend/repository"
	"context"
	"fmt"
	"strings"
	"time"
)

// SearchService 定义了搜索服务接口
type SearchService interface {
	Search(ctx context.Context, index, query string, fields []string, page, size int, sort, filter string, highlight bool) (*repository.SearchResult, error)
}

// searchService 实现了 SearchService 接口
type searchService struct {
	esRepo    repository.ESRepository
	cache     cache.Cache
	analytics analytics.Analytics
}

// NewSearchService 创建 SearchService 实例
func NewSearchService(esRepo repository.ESRepository, cache cache.Cache, analytics analytics.Analytics) SearchService {
	return &searchService{esRepo: esRepo, cache: cache, analytics: analytics}
}

// Search 执行搜索操作
func (s *searchService) Search(ctx context.Context, index, query string, fields []string, page, size int, sort, filter string, highlight bool) (*repository.SearchResult, error) {
	startTime := time.Now()

	// 生成缓存键
	cacheKey := fmt.Sprintf("search:%s:%s:%v:%d:%d:%s:%s:%t", index, query, fields, page, size, sort, filter, highlight)

	// 尝试从缓存获取
	if s.cache != nil {
		cachedResult, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cachedResult != nil {
			// 记录搜索行为
			if s.analytics != nil {
				s.analytics.RecordSearch(ctx, query, fields, page, size, sort, filter, time.Since(startTime).Milliseconds(), cachedResult.Hits.Total.Value)
			}
			return cachedResult, nil
		}
	}

	// 构建 Elasticsearch 查询
	esQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": fields,
			},
		},
		"from": (page - 1) * size,
		"size": size,
	}

	// 添加排序
	if sort != "" {
		sortFields := strings.Split(sort, ",")
		sortArray := make([]map[string]interface{}, len(sortFields))
		for i, sortField := range sortFields {
			sortOrder := "desc"
			if strings.HasPrefix(sortField, "-") {
				sortOrder = "asc"
				sortField = strings.TrimPrefix(sortField, "-")
			}
			sortArray[i] = map[string]interface{}{
				sortField: map[string]interface{}{
					"order": sortOrder,
				},
			}
		}
		esQuery["sort"] = sortArray
	}

	// 添加过滤
	if filter != "" {
		esQuery["post_filter"] = map[string]interface{}{
			"term": map[string]interface{}{
				"category": filter,
			},
		}
	}

	// 添加高亮
	if highlight {
		esQuery["highlight"] = map[string]interface{}{
			"fields": map[string]interface{}{
				"*": map[string]interface{}{},
			},
			"pre_tags":  []string{"<em>"},
			"post_tags": []string{"</em>"},
		}
	}

	// 执行搜索
	result, err := s.esRepo.Search(ctx, index, esQuery)
	if err != nil {
		return nil, err
	}

	// 存入缓存
	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, result, 3600)
	}

	// 记录搜索行为
	if s.analytics != nil {
		s.analytics.RecordSearch(ctx, query, fields, page, size, sort, filter, time.Since(startTime).Milliseconds(), result.Hits.Total.Value)
	}

	return result, nil
}
