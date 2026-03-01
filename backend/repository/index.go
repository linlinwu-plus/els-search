package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// CreateIndex 创建索引并设置映射
func CreateIndex(ctx context.Context, client *elasticsearch.Client, index string, mapping map[string]interface{}) error {
	// 检查索引是否存在
	exists, err := IndexExists(ctx, client, index)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if exists {
		return nil // 索引已存在，跳过创建
	}

	// 构建创建索引请求
	req := esapi.IndicesCreateRequest{
		Index: index,
	}

	// 如果提供了映射，则添加到请求中
	if mapping != nil {
		data, err := json.Marshal(map[string]interface{}{
			"mappings": mapping,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal mapping: %w", err)
		}
		req.Body = bytes.NewReader(data)
	}

	// 执行请求
	res, err := req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	// 检查响应状态
	if res.IsError() {
		return fmt.Errorf("create index error: %s", res.Status())
	}

	return nil
}

// IndexExists 检查索引是否存在
func IndexExists(ctx context.Context, client *elasticsearch.Client, index string) (bool, error) {
	req := esapi.IndicesExistsRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// UpdateSettings 更新索引设置
func UpdateSettings(ctx context.Context, client *elasticsearch.Client, index string, settings map[string]interface{}) error {
	data, err := json.Marshal(map[string]interface{}{
		"settings": settings,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	req := esapi.IndicesPutSettingsRequest{
		Index: []string{index},
		Body:  bytes.NewReader(data),
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("update settings error: %s", res.Status())
	}

	return nil
}
