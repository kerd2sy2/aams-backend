package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	jwtSecret = "hV6bI9UDaAw1zqlMdTL7tEB3orZ5gQpPfJCGnFm0N8vcixesHOkyRK4uWS2YXj"
	baseURL   = "http://127.0.0.1:8081/api/v1"
)

func generateToken(empID string) string {
	claims := jwt.MapClaims{
		"user_id":  empID,
		"role":     "DRIVER",
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"orig_iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwtSecret))
	return tokenString
}

func runImageUploadTest(concurrency int, requestsPerUser int) {
	fmt.Printf("\n=================================================================\n")
	fmt.Printf("📸 Stress Test: %d Concurrent Delegates Uploading Odometer Images\n", concurrency)
	fmt.Printf("=================================================================\n")

	tr := &http.Transport{
		MaxIdleConns:        concurrency * 2,
		MaxIdleConnsPerHost: concurrency * 2,
		MaxConnsPerHost:     concurrency * 2,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	var totalReqs int64
	var successReqs int64
	var errorReqs int64

	latencies := make([]time.Duration, 0, concurrency*requestsPerUser)
	var latMu sync.Mutex

	// Mock 50KB image payload
	dummyImageBytes := make([]byte, 50*1024)
	for i := range dummyImageBytes {
		dummyImageBytes[i] = byte(i % 256)
	}

	startTotal := time.Now()
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(userIndex int) {
			defer wg.Done()
			empID := "250db7b9-cf51-48ab-a05f-f4bcd95ce543"
			token := generateToken(empID)

			for r := 0; r < requestsPerUser; r++ {
				atomic.AddInt64(&totalReqs, 1)
				startReq := time.Now()

				var b bytes.Buffer
				w := multipart.NewWriter(&b)
				w.WriteField("category", "odometer")
				part, err := w.CreateFormFile("file", fmt.Sprintf("odometer_%d_%d.jpg", userIndex, r))
				if err == nil {
					part.Write(dummyImageBytes)
				}
				w.Close()

				req, _ := http.NewRequest("POST", baseURL+"/upload", &b)
				req.Header.Set("Content-Type", w.FormDataContentType())
				req.Header.Set("Authorization", "Bearer "+token)

				resp, err := client.Do(req)
				elapsed := time.Since(startReq)

				latMu.Lock()
				latencies = append(latencies, elapsed)
				latMu.Unlock()

				if err != nil {
					atomic.AddInt64(&errorReqs, 1)
				} else {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
					if resp.StatusCode >= 200 && resp.StatusCode < 400 {
						atomic.AddInt64(&successReqs, 1)
					} else {
						atomic.AddInt64(&errorReqs, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	totalElapsed := time.Since(startTotal)

	rps := float64(totalReqs) / totalElapsed.Seconds()

	var totalLat time.Duration
	var minLat = time.Hour
	var maxLat time.Duration
	for _, l := range latencies {
		totalLat += l
		if l < minLat {
			minLat = l
		}
		if l > maxLat {
			maxLat = l
		}
	}
	avgLat := time.Duration(0)
	if len(latencies) > 0 {
		avgLat = totalLat / time.Duration(len(latencies))
	}

	fmt.Printf("⏱️ Total Time: %.2f seconds\n", totalElapsed.Seconds())
	fmt.Printf("📊 Total Uploads: %d | Successful: %d | Errors: %d (%.2f%%)\n", totalReqs, successReqs, errorReqs, float64(errorReqs)/float64(totalReqs)*100)
	fmt.Printf("⚡ Throughput (RPS): %.2f Uploads / Second\n", rps)
	fmt.Printf("🟢 Latency: Min: %v | Avg: %v | Max: %v\n", minLat, avgLat, maxLat)
}

func main() {
	// 100 concurrent uploads
	runImageUploadTest(100, 2)

	// 500 concurrent uploads
	runImageUploadTest(500, 2)

	// 1,000 concurrent uploads (Peak Morning Rush)
	runImageUploadTest(1000, 2)
}
