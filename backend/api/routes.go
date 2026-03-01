package api

import (
	"backend/config"
	"backend/service"
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

		// 验证参数
		if index == "" || query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "index and q are required"})
			return
		}

		// 执行搜索
		result, err := searchService.Search(c.Request.Context(), index, query, fields, page, size)
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
}
