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
	Search(ctx context.Context, index string, query interface{}) (*SearchResult, error)
}

// SearchResult 定义了搜索结果结构
type SearchResult struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source json.RawMessage `json:"_source"`
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
