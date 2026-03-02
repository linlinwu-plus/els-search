package repository

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"
)

func TestImportData(t *testing.T) {
	// 初始化 Elasticsearch 客户端
	hosts := []string{
		"http://elasticsearch1:9200",
		"http://elasticsearch2:9200",
		"http://elasticsearch3:9200",
	}

	client, err := NewESClient(hosts)
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	// 导入数据
	index := "web_text_zh"
	filePath := "../config/web_text_zh_valid.json"

	t.Logf("Starting data import to index %s from file %s", index, filePath)

	// 直接导入数据，跳过创建索引
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 10*1024*1024) // 10MB
	scanner.Buffer(buf, cap(buf))
	batchSize := 100
	var batch []map[string]interface{}

	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var item map[string]interface{}
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			t.Logf("Failed to parse JSON: %v", err)
			continue
		}

		batch = append(batch, item)
		count++

		if len(batch) >= batchSize {
			if err := bulkIndex(client, index, batch); err != nil {
				t.Logf("Failed to bulk index: %v", err)
			}
			batch = []map[string]interface{}{}
			t.Logf("Imported %d documents", count)
		}
	}

	if len(batch) > 0 {
		if err := bulkIndex(client, index, batch); err != nil {
			t.Fatalf("Failed to bulk index: %v", err)
		}
		t.Logf("Imported %d documents", count)
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Failed to scan file: %v", err)
	}

	t.Logf("Data import completed successfully. Total documents: %d", count)

	t.Logf("Data import completed successfully")
}
