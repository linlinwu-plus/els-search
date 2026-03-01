package main

import (
	"backend/config"
	"backend/repository"
	"fmt"
	"log"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化 Elasticsearch 客户端
	esClient, err := repository.NewESClient(cfg.Elasticsearch.Hosts)
	if err != nil {
		log.Fatalf("Failed to create ES client: %v", err)
	}

	// 导入数据
	filePath := "config/web_text_zh_valid.json"
	index := "web_text_zh"

	fmt.Printf("Importing data from %s to index %s...\n", filePath, index)
	if err := repository.ImportData(esClient, index, filePath); err != nil {
		log.Fatalf("Failed to import data: %v", err)
	}

	fmt.Println("Data imported successfully!")
}
