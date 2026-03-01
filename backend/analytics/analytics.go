package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Analytics 定义了分析服务接口
type Analytics interface {
	RecordSearch(ctx context.Context, query string, fields []string, page, size int, sort, filter string, duration int64, results int)
	GetTopQueries(ctx context.Context, limit int) ([]QueryStats, error)
	GetSearchTrends(ctx context.Context, days int) ([]TrendData, error)
}

// QueryStats 定义了查询统计信息
type QueryStats struct {
	Query string `json:"query"`
	Count int    `json:"count"`
}

// TrendData 定义了趋势数据
type TrendData struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// redisAnalytics 实现了 Analytics 接口
type redisAnalytics struct {
	client *redis.Client
}

// NewRedisAnalytics 创建 Redis 分析服务实例
func NewRedisAnalytics(addr, password string, db int) Analytics {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &redisAnalytics{
		client: client,
	}
}

// RecordSearch 记录搜索行为
func (a *redisAnalytics) RecordSearch(ctx context.Context, query string, fields []string, page, size int, sort, filter string, duration int64, results int) {
	// 记录搜索次数
	searchKey := fmt.Sprintf("analytics:search:count")
	a.client.Incr(ctx, searchKey)

	// 记录热门查询
	queryKey := fmt.Sprintf("analytics:query:%s", query)
	a.client.Incr(ctx, queryKey)

	// 记录每日搜索趋势
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("analytics:daily:%s", today)
	a.client.Incr(ctx, dailyKey)

	// 记录搜索详情（可选，用于更详细的分析）
	searchDetail := map[string]interface{}{
		"query":    query,
		"fields":   fields,
		"page":     page,
		"size":     size,
		"sort":     sort,
		"filter":   filter,
		"duration": duration,
		"results":  results,
		"timestamp": time.Now().Unix(),
	}
	detailData, _ := json.Marshal(searchDetail)
	detailKey := fmt.Sprintf("analytics:detail:%d", time.Now().UnixNano())
	a.client.Set(ctx, detailKey, detailData, 24*time.Hour)
}

// GetTopQueries 获取热门查询
func (a *redisAnalytics) GetTopQueries(ctx context.Context, limit int) ([]QueryStats, error) {
	// 使用 Redis 的 ZREVRANGE 命令获取热门查询
	// 这里简化实现，实际项目中可能需要更复杂的逻辑
	var stats []QueryStats

	// 扫描所有查询键
	iter := a.client.Scan(ctx, 0, "analytics:query:*", 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		count, err := a.client.Get(ctx, key).Int()
		if err != nil {
			continue
		}

		query := key[len("analytics:query:"):]
		stats = append(stats, QueryStats{
			Query: query,
			Count: count,
		})
	}

	// 简单排序（实际项目中可以使用 Redis 的有序集合）
	for i := 0; i < len(stats); i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[i].Count < stats[j].Count {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	// 限制返回数量
	if len(stats) > limit {
		stats = stats[:limit]
	}

	return stats, nil
}

// GetSearchTrends 获取搜索趋势
func (a *redisAnalytics) GetSearchTrends(ctx context.Context, days int) ([]TrendData, error) {
	var trends []TrendData

	// 获取最近几天的搜索数据
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		dailyKey := fmt.Sprintf("analytics:daily:%s", date)
		count, err := a.client.Get(ctx, dailyKey).Int()
		if err != nil {
			count = 0
		}

		trends = append(trends, TrendData{
			Date:  date,
			Count: count,
		})
	}

	// 反转顺序，使日期从早到晚
	for i, j := 0, len(trends)-1; i < j; i, j = i+1, j-1 {
		trends[i], trends[j] = trends[j], trends[i]
	}

	return trends, nil
}
