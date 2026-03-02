package main

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

func Test1(t *testing.T) {
	url := "http://localhost/api/search?index=web_text_zh&q=1&fields=title,content"
	concurrency := 100
	requests := 10000

	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalTime time.Duration
	var completed int

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{}
			for j := 0; j < requests/concurrency; j++ {
				req, _ := http.NewRequest("GET", url, nil)
				resp, err := client.Do(req)
				if err == nil {
					resp.Body.Close()
				}
				mu.Lock()
				completed++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	totalTime = time.Since(start)
	qps := float64(requests) / totalTime.Seconds()

	fmt.Printf("Total requests: %d\n", requests)
	fmt.Printf("Total time: %v\n", totalTime)
	fmt.Printf("QPS: %.2f\n", qps)
}
