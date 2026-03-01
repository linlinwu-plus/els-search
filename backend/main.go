package main

import (
	"backend/api"
	"backend/config"
	"backend/repository"
	"backend/service"
	"log"

	"github.com/gin-gonic/gin"
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

	// 初始化仓库
	esRepo := repository.NewESRepository(esClient)

	// 初始化服务
	searchService := service.NewSearchService(esRepo)

	// 初始化 API 路由
	router := gin.Default()
	api.SetupRoutes(router, searchService, cfg.RateLimit)

	// 启动服务器
	log.Printf("Server starting on %s", cfg.Server.Addr)
	if err := router.Run(cfg.Server.Addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
