package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
)

// RateLimiter 流量控制中间件
func RateLimiter(rps int) gin.HandlerFunc {
	// 创建令牌桶，每秒生成 rps 个令牌
	tb := ratelimit.NewBucketWithRate(float64(rps), int64(rps))

	return func(c *gin.Context) {
		// 尝试获取一个令牌
		if tb.TakeAvailable(1) == 0 {
			// 没有可用令牌，返回 429 错误
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": 1,
			})
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next()
	}
}

// BurstRateLimiter 带突发流量的限流中间件
func BurstRateLimiter(rps, burst int) gin.HandlerFunc {
	// 创建令牌桶，每秒生成 rps 个令牌，最大容量为 burst
	tb := ratelimit.NewBucketWithRate(float64(rps), int64(burst))

	return func(c *gin.Context) {
		// 尝试获取一个令牌
		if tb.TakeAvailable(1) == 0 {
			// 没有可用令牌，返回 429 错误
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": 1,
			})
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next()
	}
}

// TimeWindowRateLimiter 时间窗口限流中间件
func TimeWindowRateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	var (
		requests  int
		startTime = time.Now()
	)

	return func(c *gin.Context) {
		// 检查时间窗口
		if time.Since(startTime) > window {
			// 重置计数器
			requests = 0
			startTime = time.Now()
		}

		// 检查请求数
		if requests >= maxRequests {
			// 超过限制，返回 429 错误
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": int(window.Seconds()),
			})
			c.Abort()
			return
		}

		// 增加请求计数
		requests++

		// 继续处理请求
		c.Next()
	}
}
