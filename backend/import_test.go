package main

import (
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
)

func TestImportData(t *testing.T) {
	// 初始化 Elasticsearch 客户端
	hosts := []string{
		"http://elasticsearch1:9200",
		"http://elasticsearch2:9200",
		"http://elasticsearch3:9200",
	}

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: hosts,
	})
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	// 测试集群健康状态
	resp, err := client.Cluster.Health()
	if err != nil {
		t.Fatalf("Failed to check cluster health: %v", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		t.Fatalf("Cluster health check failed: %s", resp.Status())
	}

	// 这里可以添加数据导入的测试代码
	t.Logf("Cluster health check passed")
}
