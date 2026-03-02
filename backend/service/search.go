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
	// 搜索相关
	Search(ctx context.Context, index, query string, fields []string, page, size int, sort, filter string, highlight bool) (*repository.SearchResult, error)
	SearchWithHighlight(ctx context.Context, index, query string, fields []string, highlightFields []string, page, size int) (*repository.SearchResult, error)
	SearchWithAggregation(ctx context.Context, index, query string, fields []string, aggregation map[string]interface{}, page, size int) (*repository.SearchResult, error)

	// 文档管理
	AddDocument(ctx context.Context, index string, id string, document map[string]interface{}) error
	GetDocument(ctx context.Context, index string, id string) (map[string]interface{}, error)
	UpdateDocument(ctx context.Context, index string, id string, document map[string]interface{}) error
	DeleteDocument(ctx context.Context, index string, id string) error

	// 批量操作
	BulkAddDocuments(ctx context.Context, index string, documents []map[string]interface{}) error
	BulkUpdateDocuments(ctx context.Context, index string, documents []map[string]interface{}) error
	BulkDeleteDocuments(ctx context.Context, index string, ids []string) error

	// 索引管理
	CreateIndex(ctx context.Context, index string, mapping map[string]interface{}, settings map[string]interface{}) error
	DeleteIndex(ctx context.Context, index string) error
	IndexExists(ctx context.Context, index string) (bool, error)

	// 分析功能
	AnalyzeText(ctx context.Context, index string, text string, analyzer string) ([]string, error)

	// 统计功能
	CountDocuments(ctx context.Context, index string, query string, fields []string) (int64, error)
	GetIndexStats(ctx context.Context, index string) (map[string]interface{}, error)
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
				go s.analytics.RecordSearch(ctx, query, fields, page, size, sort, filter, time.Since(startTime).Milliseconds(), cachedResult.Hits.Total.Value)
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
				"type":   "best_fields",
				"tie_breaker": 0.3,
			},
		},
		"from": (page - 1) * size,
		"size": size,
		"track_total_hits": 10000,
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
				"title": map[string]interface{}{},
				"content": map[string]interface{}{},
				"desc": map[string]interface{}{},
			},
			"pre_tags":  []string{"<em>"},
			"post_tags": []string{"</em>"},
			"fragment_size": 150,
			"number_of_fragments": 3,
		}
	}

	// 执行搜索
	result, err := s.esRepo.Search(ctx, index, esQuery)
	if err != nil {
		return nil, err
	}

	// 异步存入缓存
	if s.cache != nil {
		go func() {
			_ = s.cache.Set(ctx, cacheKey, result, 3600)
		}()
	}

	// 异步记录搜索行为
	if s.analytics != nil {
		go s.analytics.RecordSearch(ctx, query, fields, page, size, sort, filter, time.Since(startTime).Milliseconds(), result.Hits.Total.Value)
	}

	return result, nil
}

// SearchWithHighlight 执行带高亮的搜索操作
func (s *searchService) SearchWithHighlight(ctx context.Context, index, query string, fields []string, highlightFields []string, page, size int) (*repository.SearchResult, error) {
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

	// 执行带高亮的搜索
	return s.esRepo.SearchWithHighlight(ctx, index, esQuery, highlightFields)
}

// SearchWithAggregation 执行带聚合的搜索操作
func (s *searchService) SearchWithAggregation(ctx context.Context, index, query string, fields []string, aggregation map[string]interface{}, page, size int) (*repository.SearchResult, error) {
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

	// 执行带聚合的搜索
	return s.esRepo.SearchWithAggregation(ctx, index, esQuery, aggregation)
}

// AddDocument 添加文档
func (s *searchService) AddDocument(ctx context.Context, index string, id string, document map[string]interface{}) error {
	return s.esRepo.IndexDocument(ctx, index, id, document)
}

// GetDocument 获取文档
func (s *searchService) GetDocument(ctx context.Context, index string, id string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := s.esRepo.GetDocument(ctx, index, id, &result)
	return result, err
}

// UpdateDocument 更新文档
func (s *searchService) UpdateDocument(ctx context.Context, index string, id string, document map[string]interface{}) error {
	return s.esRepo.UpdateDocument(ctx, index, id, document)
}

// DeleteDocument 删除文档
func (s *searchService) DeleteDocument(ctx context.Context, index string, id string) error {
	return s.esRepo.DeleteDocument(ctx, index, id)
}

// BulkAddDocuments 批量添加文档
func (s *searchService) BulkAddDocuments(ctx context.Context, index string, documents []map[string]interface{}) error {
	return s.esRepo.BulkIndex(ctx, index, documents)
}

// BulkUpdateDocuments 批量更新文档
func (s *searchService) BulkUpdateDocuments(ctx context.Context, index string, documents []map[string]interface{}) error {
	return s.esRepo.BulkUpdate(ctx, index, documents)
}

// BulkDeleteDocuments 批量删除文档
func (s *searchService) BulkDeleteDocuments(ctx context.Context, index string, ids []string) error {
	return s.esRepo.BulkDelete(ctx, index, ids)
}

// CreateIndex 创建索引
func (s *searchService) CreateIndex(ctx context.Context, index string, mapping map[string]interface{}, settings map[string]interface{}) error {
	return s.esRepo.CreateIndex(ctx, index, mapping, settings)
}

// DeleteIndex 删除索引
func (s *searchService) DeleteIndex(ctx context.Context, index string) error {
	return s.esRepo.DeleteIndex(ctx, index)
}

// IndexExists 检查索引是否存在
func (s *searchService) IndexExists(ctx context.Context, index string) (bool, error) {
	return s.esRepo.IndexExists(ctx, index)
}

// AnalyzeText 分析文本
func (s *searchService) AnalyzeText(ctx context.Context, index string, text string, analyzer string) ([]string, error) {
	return s.esRepo.AnalyzeText(ctx, index, text, analyzer)
}

// CountDocuments 统计文档数量
func (s *searchService) CountDocuments(ctx context.Context, index string, query string, fields []string) (int64, error) {
	// 构建查询
	esQuery := map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":  query,
			"fields": fields,
		},
	}

	// 执行计数
	return s.esRepo.CountDocuments(ctx, index, esQuery)
}

// GetIndexStats 获取索引统计信息
func (s *searchService) GetIndexStats(ctx context.Context, index string) (map[string]interface{}, error) {
	return s.esRepo.GetIndexStats(ctx, index)
}
