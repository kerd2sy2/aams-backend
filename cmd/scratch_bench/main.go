package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

// Generate valid JWT token for testing
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

type TestResult struct {
	TotalRequests int64
	SuccessCount  int64
	ErrorCount    int64
	TotalDuration time.Duration
	Latencies     []time.Duration
	Mu            sync.Mutex
}

func runConcurrencyTest(concurrency int, requestsPerUser int, testType string) {
	fmt.Printf("\n=================================================================\n")
	fmt.Printf("🚀 Starting Load Test: %d Concurrent Delegates | %s\n", concurrency, testType)
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
		Timeout:   15 * time.Second,
	}

	var totalReqs int64
	var successReqs int64
	var errorReqs int64

	latencies := make([]time.Duration, 0, concurrency*requestsPerUser)
	var latMu sync.Mutex

	sampleImage := "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP//////////////////////////////////////////////////////////////////////////////////////wgALCAABAAEBAREA/8QAFBABAAAAAAAAAAAAAAAAAAAAAP/aAAgBAQABPxA="

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

				var req *http.Request
				var err error

				switch testType {
				case "profile_and_active_session":
					req, _ = http.NewRequest("GET", baseURL+"/work/active?employee_id="+empID, nil)
					req.Header.Set("Authorization", "Bearer "+token)

				case "last_km_lookup":
					req, _ = http.NewRequest("GET", baseURL+"/work/last-km?employee_id="+empID+"&motorcycle_number=7572", nil)
					req.Header.Set("Authorization", "Bearer "+token)

				case "history_reports":
					req, _ = http.NewRequest("GET", baseURL+"/reports?employee_id="+empID+"&limit=20", nil)
					req.Header.Set("Authorization", "Bearer "+token)

				case "full_shift_lifecycle":
					// Read profile, lookup last KM, fetch active session
					req, _ = http.NewRequest("GET", baseURL+"/work/last-km?employee_id="+empID+"&motorcycle_number=7572", nil)
					req.Header.Set("Authorization", "Bearer "+token)

				case "image_upload_stress":
					bodyData, _ := json.Marshal(map[string]interface{}{
						"employee_id":       empID,
						"motorcycle_number": "7572",
						"start_km":          15000 + userIndex,
						"start_km_image":    sampleImage,
						"notes":             "Stress test odometer upload",
					})
					req, _ = http.NewRequest("POST", baseURL+"/work/start", bytes.NewBuffer(bodyData))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+token)
				}

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
					if resp.StatusCode >= 200 && resp.StatusCode < 500 {
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

	// Calculate Stats
	rps := float64(totalReqs) / totalElapsed.Seconds()
	
	// Sort latencies
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
	fmt.Printf("📊 Total Requests: %d | Successful: %d | Errors: %d (%.2f%%)\n", totalReqs, successReqs, errorReqs, float64(errorReqs)/float64(totalReqs)*100)
	fmt.Printf("⚡ Throughput (RPS): %.2f Requests / Second\n", rps)
	fmt.Printf("🟢 Latency: Min: %v | Avg: %v | Max: %v\n", minLat, avgLat, maxLat)
}

func main() {
	fmt.Println("=================================================================")
	fmt.Println("🚀 AAMS ENTERPRISE LOAD & CONCURRENCY BENCHMARK (1,000+ DELEGATES)")
	fmt.Println("=================================================================")

	// Test 1: 50 Concurrent Users
	runConcurrencyTest(50, 20, "last_km_lookup")

	// Test 2: 200 Concurrent Users
	runConcurrencyTest(200, 10, "profile_and_active_session")

	// Test 3: 500 Concurrent Users
	runConcurrencyTest(500, 5, "last_km_lookup")

	// Test 4: 1,000 Concurrent Users (Peak Morning Shift Rush)
	runConcurrencyTest(1000, 5, "full_shift_lifecycle")

	// Test 5: 1,500 Concurrent Users (Extreme Surge Stress Test)
	runConcurrencyTest(1500, 3, "profile_and_active_session")

	// Test 6: 2,000 Concurrent Users (Maximum Capacity Test)
	runConcurrencyTest(2000, 2, "last_km_lookup")
}
