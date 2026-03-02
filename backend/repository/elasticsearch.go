package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ESRepository 定义了 Elasticsearch 仓库接口
type ESRepository interface {
	// 搜索相关
	Search(ctx context.Context, index string, query interface{}) (*SearchResult, error)
	SearchWithHighlight(ctx context.Context, index string, query interface{}, highlightFields []string) (*SearchResult, error)
	SearchWithAggregation(ctx context.Context, index string, query interface{}, aggregation map[string]interface{}) (*SearchResult, error)
	
	// 文档相关
	IndexDocument(ctx context.Context, index string, id string, document interface{}) error
	GetDocument(ctx context.Context, index string, id string, result interface{}) error
	UpdateDocument(ctx context.Context, index string, id string, document interface{}) error
	DeleteDocument(ctx context.Context, index string, id string) error
	
	// 批量操作
	BulkIndex(ctx context.Context, index string, documents []map[string]interface{}) error
	BulkUpdate(ctx context.Context, index string, documents []map[string]interface{}) error
	BulkDelete(ctx context.Context, index string, ids []string) error
	
	// 索引管理
	CreateIndex(ctx context.Context, index string, mapping map[string]interface{}, settings map[string]interface{}) error
	DeleteIndex(ctx context.Context, index string) error
	IndexExists(ctx context.Context, index string) (bool, error)
	
	// 分析功能
	AnalyzeText(ctx context.Context, index string, text string, analyzer string) ([]string, error)
	
	// 统计功能
	CountDocuments(ctx context.Context, index string, query interface{}) (int64, error)
	GetIndexStats(ctx context.Context, index string) (map[string]interface{}, error)
}

// SearchResult 定义了搜索结果结构
type SearchResult struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source    json.RawMessage          `json:"_source"`
			Highlight map[string][]string       `json:"highlight,omitempty"`
		} `json:"hits"`
	} `json:"hits"`
}

// esRepository 实现了 ESRepository 接口
type esRepository struct {
	client *elasticsearch.Client
}

// NewESClient 创建 Elasticsearch 客户端
func NewESClient(hosts []string) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: hosts,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
			ResponseHeaderTimeout: time.Second * 30,
		},
	}
	return elasticsearch.NewClient(cfg)
}

// NewESRepository 创建 ESRepository 实例
func NewESRepository(client *elasticsearch.Client) ESRepository {
	return &esRepository{client: client}
}

// Search 执行搜索操作
func (r *esRepository) Search(ctx context.Context, index string, query interface{}) (*SearchResult, error) {
	// 构建搜索请求
	req := esapi.SearchRequest{
		Index: []string{index},
	}

	// 序列化查询
	data, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}
	req.Body = bytes.NewReader(data)

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.Status())
	}

	// 解析响应
	var result SearchResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// SearchWithHighlight 执行带高亮的搜索操作
func (r *esRepository) SearchWithHighlight(ctx context.Context, index string, query interface{}, highlightFields []string) (*SearchResult, error) {
	// 构建带高亮的查询
	highlightQuery := map[string]interface{}{
		"query": query,
		"highlight": map[string]interface{}{
			"fields": make(map[string]interface{}),
		},
	}

	// 添加高亮字段
	for _, field := range highlightFields {
		highlightQuery["highlight"].(map[string]interface{})["fields"].(map[string]interface{})[field] = map[string]interface{}{}
	}

	// 执行搜索
	return r.Search(ctx, index, highlightQuery)
}

// SearchWithAggregation 执行带聚合的搜索操作
func (r *esRepository) SearchWithAggregation(ctx context.Context, index string, query interface{}, aggregation map[string]interface{}) (*SearchResult, error) {
	// 构建带聚合的查询
	aggregationQuery := map[string]interface{}{
		"query":      query,
		"aggregations": aggregation,
	}

	// 执行搜索
	return r.Search(ctx, index, aggregationQuery)
}

// IndexDocument 索引文档
func (r *esRepository) IndexDocument(ctx context.Context, index string, id string, document interface{}) error {
	// 序列化文档
	data, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	// 构建索引请求
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to execute index: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return fmt.Errorf("index error: %s", res.Status())
	}

	return nil
}

// GetDocument 获取文档
func (r *esRepository) GetDocument(ctx context.Context, index string, id string, result interface{}) error {
	// 构建获取请求
	req := esapi.GetRequest{
		Index:      index,
		DocumentID: id,
	}

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to execute get: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return fmt.Errorf("get error: %s", res.Status())
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// 提取 _source 字段
	if source, ok := response["_source"]; ok {
		sourceData, err := json.Marshal(source)
		if err != nil {
			return fmt.Errorf("failed to marshal source: %w", err)
		}
		if err := json.Unmarshal(sourceData, result); err != nil {
			return fmt.Errorf("failed to unmarshal source: %w", err)
		}
	}

	return nil
}

// UpdateDocument 更新文档
func (r *esRepository) UpdateDocument(ctx context.Context, index string, id string, document interface{}) error {
	// 构建更新请求体
	updateBody := map[string]interface{}{
		"doc": document,
	}

	// 序列化更新体
	data, err := json.Marshal(updateBody)
	if err != nil {
		return fmt.Errorf("failed to marshal update body: %w", err)
	}

	// 构建更新请求
	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return fmt.Errorf("update error: %s", res.Status())
	}

	return nil
}

// DeleteDocument 删除文档
func (r *esRepository) DeleteDocument(ctx context.Context, index string, id string) error {
	// 构建删除请求
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: id,
		Refresh:    "true",
	}

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to execute delete: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return fmt.Errorf("delete error: %s", res.Status())
	}

	return nil
}

// BulkIndex 批量索引文档
func (r *esRepository) BulkIndex(ctx context.Context, index string, documents []map[string]interface{}) error {
	var bulkBuffer bytes.Buffer

	// 构建批量请求体
	for _, doc := range documents {
		// 写入操作元数据
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
			},
		}
		if id, ok := doc["id"].(string); ok {
			action["index"].(map[string]interface{})["_id"] = id
			delete(doc, "id")
		}

		if err := json.NewEncoder(&bulkBuffer).Encode(action); err != nil {
			return fmt.Errorf("failed to encode action: %w", err)
		}

		// 写入文档数据
		if err := json.NewEncoder(&bulkBuffer).Encode(doc); err != nil {
			return fmt.Errorf("failed to encode document: %w", err)
		}
	}

	// 构建批量请求
	req := esapi.BulkRequest{
		Body:   &bulkBuffer,
		Refresh: "true",
	}

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to execute bulk: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return fmt.Errorf("bulk error: %s", res.Status())
	}

	return nil
}

// BulkUpdate 批量更新文档
func (r *esRepository) BulkUpdate(ctx context.Context, index string, documents []map[string]interface{}) error {
	var bulkBuffer bytes.Buffer

	// 构建批量请求体
	for _, doc := range documents {
		// 获取文档 ID
		id, ok := doc["id"].(string)
		if !ok {
			return fmt.Errorf("document must have 'id' field")
		}
		delete(doc, "id")

		// 写入操作元数据
		action := map[string]interface{}{
			"update": map[string]interface{}{
				"_index": index,
				"_id":    id,
			},
		}

		if err := json.NewEncoder(&bulkBuffer).Encode(action); err != nil {
			return fmt.Errorf("failed to encode action: %w", err)
		}

		// 写入更新数据
		updateBody := map[string]interface{}{"doc": doc}
		if err := json.NewEncoder(&bulkBuffer).Encode(updateBody); err != nil {
			return fmt.Errorf("failed to encode update body: %w", err)
		}
	}

	// 构建批量请求
	req := esapi.BulkRequest{
		Body:   &bulkBuffer,
		Refresh: "true",
	}

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to execute bulk: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return fmt.Errorf("bulk error: %s", res.Status())
	}

	return nil
}

// BulkDelete 批量删除文档
func (r *esRepository) BulkDelete(ctx context.Context, index string, ids []string) error {
	var bulkBuffer bytes.Buffer

	// 构建批量请求体
	for _, id := range ids {
		// 写入操作元数据
		action := map[string]interface{}{
			"delete": map[string]interface{}{
				"_index": index,
				"_id":    id,
			},
		}

		if err := json.NewEncoder(&bulkBuffer).Encode(action); err != nil {
			return fmt.Errorf("failed to encode action: %w", err)
		}
	}

	// 构建批量请求
	req := esapi.BulkRequest{
		Body:   &bulkBuffer,
		Refresh: "true",
	}

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to execute bulk: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return fmt.Errorf("bulk error: %s", res.Status())
	}

	return nil
}

// CreateIndex 创建索引
func (r *esRepository) CreateIndex(ctx context.Context, index string, mapping map[string]interface{}, settings map[string]interface{}) error {
	// 构建索引配置
	config := make(map[string]interface{})
	if mapping != nil {
		config["mappings"] = mapping
	}
	if settings != nil {
		config["settings"] = settings
	}

	// 序列化配置
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal index config: %w", err)
	}

	// 构建创建索引请求
	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  bytes.NewReader(data),
	}

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to execute create index: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return fmt.Errorf("create index error: %s", res.Status())
	}

	return nil
}

// DeleteIndex 删除索引
func (r *esRepository) DeleteIndex(ctx context.Context, index string) error {
	// 构建删除索引请求
	req := esapi.IndicesDeleteRequest{
		Index: []string{index},
	}

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to execute delete index: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return fmt.Errorf("delete index error: %s", res.Status())
	}

	return nil
}

// IndexExists 检查索引是否存在
func (r *esRepository) IndexExists(ctx context.Context, index string) (bool, error) {
	// 构建索引存在请求
	req := esapi.IndicesExistsRequest{
		Index: []string{index},
	}

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return false, fmt.Errorf("failed to execute index exists: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	return res.StatusCode == http.StatusOK, nil
}

// AnalyzeText 分析文本
func (r *esRepository) AnalyzeText(ctx context.Context, index string, text string, analyzer string) ([]string, error) {
	// 构建分析请求体
	analyzeBody := map[string]interface{}{
		"text":     text,
		"analyzer": analyzer,
	}

	// 序列化请求体
	data, err := json.Marshal(analyzeBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal analyze body: %w", err)
	}

	// 构建分析请求
	req := esapi.IndicesAnalyzeRequest{
		Index: index,
		Body:  bytes.NewReader(data),
	}

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute analyze: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return nil, fmt.Errorf("analyze error: %s", res.Status())
	}

	// 解析响应
	var response struct {
		Tokens []struct {
			Token string `json:"token"`
		} `json:"tokens"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 提取 tokens
	tokens := make([]string, len(response.Tokens))
	for i, token := range response.Tokens {
		tokens[i] = token.Token
	}

	return tokens, nil
}

// CountDocuments 统计文档数量
func (r *esRepository) CountDocuments(ctx context.Context, index string, query interface{}) (int64, error) {
	// 序列化查询
	data, err := json.Marshal(query)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal query: %w", err)
	}

	// 构建计数请求
	req := esapi.CountRequest{
		Index: []string{index},
		Body:  bytes.NewReader(data),
	}

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return 0, fmt.Errorf("count error: %s", res.Status())
	}

	// 解析响应
	var response struct {
		Count int64 `json:"count"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Count, nil
}

// GetIndexStats 获取索引统计信息
func (r *esRepository) GetIndexStats(ctx context.Context, index string) (map[string]interface{}, error) {
	// 构建统计请求
	req := esapi.IndicesStatsRequest{
		Index: []string{index},
	}

	// 执行请求
	res, err := req.Do(ctx, r.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute index stats: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return nil, fmt.Errorf("index stats error: %s", res.Status())
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}
