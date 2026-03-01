package repository

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ImportData 从 JSON 文件导入数据到 Elasticsearch
func ImportData(client *elasticsearch.Client, index string, filePath string) error {
	// 打开 JSON 文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 创建索引（如果不存在）
	if err := createIndexIfNotExists(client, index); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// 读取并导入数据
	scanner := bufio.NewScanner(file)
	// 增加缓冲区大小，处理长行
	buf := make([]byte, 10*1024*1024) // 10MB
	scanner.Buffer(buf, cap(buf))
	var wg sync.WaitGroup
	batchSize := 100
	var batch []map[string]interface{}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// 解析 JSON 行
		var item map[string]interface{}
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			fmt.Printf("failed to parse JSON: %v\n", err)
			continue
		}

		// 添加到批量
		batch = append(batch, item)

		// 达到批量大小，执行导入
		if len(batch) >= batchSize {
			wg.Add(1)
			go func(b []map[string]interface{}) {
				defer wg.Done()
				if err := bulkIndex(client, index, b); err != nil {
					fmt.Printf("failed to bulk index: %v\n", err)
				}
			}(batch)
			batch = []map[string]interface{}{}
		}
	}

	// 导入剩余数据
	if len(batch) > 0 {
		if err := bulkIndex(client, index, batch); err != nil {
			return fmt.Errorf("failed to bulk index: %w", err)
		}
	}

	wg.Wait()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan file: %w", err)
	}

	return nil
}

// createIndexIfNotExists 创建索引（如果不存在）并设置优化的映射
func createIndexIfNotExists(client *elasticsearch.Client, index string) error {
	// 优化的映射结构
	mapping := map[string]interface{}{
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":     "text",
				"analyzer": "ik_max_word",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type": "keyword",
						"ignore_above": 256,
					},
				},
			},
			"content": map[string]interface{}{
				"type":     "text",
				"analyzer": "ik_max_word",
			},
			"category": map[string]interface{}{
				"type": "keyword",
			},
			"timestamp": map[string]interface{}{
				"type": "date",
			},
		},
	}

	// 优化的索引设置
	settings := map[string]interface{}{
		"index": map[string]interface{}{
			"number_of_shards":   3,
			"number_of_replicas": 1,
			"refresh_interval":   "1s",
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					"ik_max_word": map[string]interface{}{
						"type":      "custom",
						"tokenizer": "ik_max_word",
						"filter": []string{
							"lowercase",
						},
					},
				},
			},
		},
	}

	// 创建索引
	if err := CreateIndex(context.Background(), client, index, mapping); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// 更新索引设置
	if err := UpdateSettings(context.Background(), client, index, settings); err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	return nil
}

// bulkIndex 批量索引数据
func bulkIndex(client *elasticsearch.Client, index string, items []map[string]interface{}) error {
	// 构建批量请求
	var bulkData []byte
	for _, item := range items {
		// 添加索引操作
		indexOp := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
			},
		}
		indexOpData, err := json.Marshal(indexOp)
		if err != nil {
			return fmt.Errorf("failed to marshal index operation: %w", err)
		}
		bulkData = append(bulkData, indexOpData...)
		bulkData = append(bulkData, '\n')

		// 添加文档数据
		itemData, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal item: %w", err)
		}
		bulkData = append(bulkData, itemData...)
		bulkData = append(bulkData, '\n')
	}

	// 执行批量请求
	req := esapi.BulkRequest{
		Body: bytes.NewReader(bulkData),
	}

	res, err := req.Do(context.Background(), client)
	if err != nil {
		return fmt.Errorf("failed to execute bulk request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk request failed: %s", res.Status())
	}

	return nil
}
