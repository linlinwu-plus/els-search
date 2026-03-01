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
}
