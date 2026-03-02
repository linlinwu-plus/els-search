package api

import (
	"backend/config"
	"backend/service"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置 API 路由
func SetupRoutes(router *gin.Engine, searchService service.SearchService, rateLimit config.RateLimitConfig) {
	// CORS 中间件
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 全局流量控制
	router.Use(RateLimiter(rateLimit.Global.RPS))

	// 搜索接口流量控制
	router.GET("/api/search", BurstRateLimiter(rateLimit.Search.RPS, rateLimit.Search.Burst), func(c *gin.Context) {
		// 获取查询参数
		index := c.Query("index")
		query := c.Query("q")
		fields := strings.Split(c.Query("fields"), ",")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
		sort := c.Query("sort")
		filter := c.Query("filter")
		highlight := c.DefaultQuery("highlight", "true") == "true"

		// 验证参数
		if index == "" || query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "index and q are required"})
			return
		}

		// 执行搜索
		result, err := searchService.Search(c.Request.Context(), index, query, fields, page, size, sort, filter, highlight)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 返回结果
		c.JSON(http.StatusOK, result)
	})

	// 健康检查接口
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 分析接口
	router.GET("/api/analytics/top-queries", func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		// 这里需要实现获取热门查询的逻辑
		c.JSON(http.StatusOK, gin.H{"queries": limit})
	})

	router.GET("/api/analytics/trends", func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
		// 这里需要实现获取搜索趋势的逻辑
		c.JSON(http.StatusOK, gin.H{"trends": days})
	})

	// 性能监控接口
	router.POST("/api/analytics/performance", func(c *gin.Context) {
		var performanceData map[string]interface{}
		if err := c.ShouldBindJSON(&performanceData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 这里可以将性能数据存储到数据库或日志中
		fmt.Printf("Performance data: %v\n", performanceData)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 错误监控接口
	router.POST("/api/analytics/error", func(c *gin.Context) {
		var errorData map[string]interface{}
		if err := c.ShouldBindJSON(&errorData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 这里可以将错误数据存储到数据库或日志中
		fmt.Printf("Error data: %v\n", errorData)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 文档管理接口
	router.POST("/api/documents", func(c *gin.Context) {
		var request struct {
			Index    string                 `json:"index" binding:"required"`
			ID       string                 `json:"id"`
			Document map[string]interface{} `json:"document" binding:"required"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := searchService.AddDocument(c.Request.Context(), request.Index, request.ID, request.Document); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/api/documents/:index/:id", func(c *gin.Context) {
		index := c.Param("index")
		id := c.Param("id")
		document, err := searchService.GetDocument(c.Request.Context(), index, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, document)
	})

	router.PUT("/api/documents/:index/:id", func(c *gin.Context) {
		index := c.Param("index")
		id := c.Param("id")
		var document map[string]interface{}
		if err := c.ShouldBindJSON(&document); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := searchService.UpdateDocument(c.Request.Context(), index, id, document); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.DELETE("/api/documents/:index/:id", func(c *gin.Context) {
		index := c.Param("index")
		id := c.Param("id")
		if err := searchService.DeleteDocument(c.Request.Context(), index, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 批量操作接口
	router.POST("/api/documents/bulk", func(c *gin.Context) {
		var request struct {
			Index     string                   `json:"index" binding:"required"`
			Documents []map[string]interface{} `json:"documents"`
			Operation string                   `json:"operation" binding:"required,oneof=index update delete"`
			IDs       []string                 `json:"ids"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var err error
		switch request.Operation {
		case "index":
			err = searchService.BulkAddDocuments(c.Request.Context(), request.Index, request.Documents)
		case "update":
			err = searchService.BulkUpdateDocuments(c.Request.Context(), request.Index, request.Documents)
		case "delete":
			err = searchService.BulkDeleteDocuments(c.Request.Context(), request.Index, request.IDs)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 索引管理接口
	router.POST("/api/indexes", func(c *gin.Context) {
		var request struct {
			Index    string                 `json:"index" binding:"required"`
			Mapping  map[string]interface{} `json:"mapping"`
			Settings map[string]interface{} `json:"settings"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := searchService.CreateIndex(c.Request.Context(), request.Index, request.Mapping, request.Settings); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.DELETE("/api/indexes/:index", func(c *gin.Context) {
		index := c.Param("index")
		if err := searchService.DeleteIndex(c.Request.Context(), index); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/api/indexes/:index/exists", func(c *gin.Context) {
		index := c.Param("index")
		exists, err := searchService.IndexExists(c.Request.Context(), index)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"exists": exists})
	})

	// 分析接口
	router.POST("/api/analyze", func(c *gin.Context) {
		var request struct {
			Index    string `json:"index" binding:"required"`
			Text     string `json:"text" binding:"required"`
			Analyzer string `json:"analyzer" binding:"required"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		tokens, err := searchService.AnalyzeText(c.Request.Context(), request.Index, request.Text, request.Analyzer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"tokens": tokens})
	})

	// 统计接口
	router.GET("/api/count", func(c *gin.Context) {
		index := c.Query("index")
		query := c.Query("q")
		fields := strings.Split(c.Query("fields"), ",")
		if index == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "index is required"})
			return
		}
		count, err := searchService.CountDocuments(c.Request.Context(), index, query, fields)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"count": count})
	})

	router.GET("/api/indexes/:index/stats", func(c *gin.Context) {
		index := c.Param("index")
		stats, err := searchService.GetIndexStats(c.Request.Context(), index)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, stats)
	})
}
