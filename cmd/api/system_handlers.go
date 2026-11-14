package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"delivery-backend/pkg/middleware"
)

func registerCoreHealthRoutes(r *gin.Engine, db *gorm.DB, serverStartTime time.Time) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "online",
			"app":     "Delivery Employee Management System API",
			"version": "1.0.0",
		})
	})

	r.GET("/api/v1", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "running",
			"app":       "AAMS Delivery & Target Tracking System API (v1)",
			"version":   "1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
			"health":    "/api/v1/health",
		})
	})

	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"app":       "Delivery Employee Management System API",
			"version":   "1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	r.GET("/api/v1/health/modules", func(c *gin.Context) {
		modules := middleware.GetModulesHealth()
		overall := "healthy"
		for _, m := range modules {
			if m.Status != "healthy" {
				overall = "degraded"
				break
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"status":    overall,
			"modules":   modules,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Server performance & health metrics endpoint (used by Admin Dashboard)
	r.GET("/api/v1/system/performance", func(c *gin.Context) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		sqlDB, err := db.DB()
		dbStatus := "connected"
		dbOpen := 0
		dbInUse := 0
		dbIdle := 0
		dbWaitCount := int64(0)
		if err != nil {
			log.Printf("[HEALTH] Failed to get database handle: %v", err)
			dbStatus = "error"
		} else {
			if pingErr := sqlDB.Ping(); pingErr != nil {
				log.Printf("[HEALTH] Database ping failed: %v", pingErr)
				dbStatus = "disconnected"
			}
			stats := sqlDB.Stats()
			dbOpen = stats.OpenConnections
			dbInUse = stats.InUse
			dbIdle = stats.Idle
			dbWaitCount = stats.WaitCount
		}

		uptimeDuration := time.Since(serverStartTime)
		uptimeDays := int(uptimeDuration.Hours()) / 24
		uptimeHours := int(uptimeDuration.Hours()) % 24
		uptimeMins := int(uptimeDuration.Minutes()) % 60
		uptimeSecs := int(uptimeDuration.Seconds()) % 60

		uptimeStr := fmt.Sprintf("%d يوم و %d ساعة و %d دقيقة و %d ثانية", uptimeDays, uptimeHours, uptimeMins, uptimeSecs)
		if uptimeDays == 0 && uptimeHours == 0 {
			uptimeStr = fmt.Sprintf("%d دقيقة و %d ثانية", uptimeMins, uptimeSecs)
		}

		c.JSON(http.StatusOK, gin.H{
			"status":         "healthy",
			"app":            "AAMS Backend Server",
			"server_time":    time.Now().Format("2006-01-02 15:04:05"),
			"uptime_seconds": int64(uptimeDuration.Seconds()),
			"uptime_string":  uptimeStr,
			"started_at":     serverStartTime.Format("2006-01-02 15:04:05"),
			"goroutines":     runtime.NumGoroutine(),
			"cpus":           runtime.NumCPU(),
			"go_version":     runtime.Version(),
			"os":             runtime.GOOS,
			"arch":           runtime.GOARCH,
			"memory": gin.H{
				"alloc_mb":       math.Round(float64(mem.Alloc)/1024/1024*100) / 100,
				"total_alloc_mb": math.Round(float64(mem.TotalAlloc)/1024/1024*100) / 100,
				"sys_mb":         math.Round(float64(mem.Sys)/1024/1024*100) / 100,
				"heap_alloc_mb":  math.Round(float64(mem.HeapAlloc)/1024/1024*100) / 100,
				"heap_inuse_mb":  math.Round(float64(mem.HeapInuse)/1024/1024*100) / 100,
				"num_gc":         mem.NumGC,
			},
			"database": gin.H{
				"status":           dbStatus,
				"open_connections": dbOpen,
				"in_use":           dbInUse,
				"idle":             dbIdle,
				"wait_count":       dbWaitCount,
			},
			"modules": middleware.GetModulesHealth(),
		})
	})
}
