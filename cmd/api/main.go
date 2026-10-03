package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/modules/attendance"
	"delivery-backend/internal/modules/auth"
	"delivery-backend/internal/modules/custody"
	"delivery-backend/internal/modules/employee"
	"delivery-backend/internal/modules/fleet"
	"delivery-backend/internal/modules/hr_legal"
	"delivery-backend/internal/modules/inventory"
	"delivery-backend/internal/modules/maintenance"
	"delivery-backend/internal/modules/system"
	"delivery-backend/internal/modules/target"
	"delivery-backend/internal/modules/ticket"
	"delivery-backend/internal/modules/work"
	"delivery-backend/internal/repository"
	"delivery-backend/internal/service"
	"delivery-backend/pkg/backup"
	"delivery-backend/pkg/config"
	"delivery-backend/pkg/database"
	"delivery-backend/pkg/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	// Set server timezone to Saudi Arabia (Asia/Riyadh, UTC+3)
	riyadhLoc, err := time.LoadLocation("Asia/Riyadh")
	if err != nil {
		log.Fatalf("Failed to load Asia/Riyadh timezone: %v", err)
	}
	time.Local = riyadhLoc
	log.Println("Timezone set to Asia/Riyadh (UTC+3)")

	cfg := config.LoadConfig()

	// Initialize Database with GORM Postgres / SQLite Fallback
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start automatic database backup service (checks every 600s)
	backupDir := os.Getenv("BACKUP_DIR")
	if backupDir == "" {
		backupDir = "./backups"
	}
	backupSvc := backup.NewService(cfg.PGDumpPath, cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, backupDir, 600*time.Second, 50)
	go backupSvc.Start()

	// Start background cron jobs
	startIqamaExpirationChecker(db)

	// Repositories needed for core Auth Middleware
	adminRepo := repository.NewAdminRepository(db)
	empRepo := repository.NewEmployeeRepository(db)

	// Shared services
	storageService := service.NewStorageService(cfg)
	auditRepo := repository.NewAuditRepository(db)
	auditService := service.NewAuditService(auditRepo)

	// Initialize Firebase Cloud Messaging (FCM V1) for background push notifications
	service.InitFCM("")

	// ---------------------------------------------------------
	// Initialize Bounded Context Modules (Modular Monolith Architecture)
	// ---------------------------------------------------------
	authModule := auth.NewModule(db, cfg)
	systemModule := system.NewModule(db)
	ticketModule := ticket.NewModule(db, nil, nil)
	fleetModule := fleet.NewModule(db, storageService, auditService)
	inventoryModule := inventory.NewModule(db)
	maintenanceModule := maintenance.NewModule(db)
	custodyModule := custody.NewModule(db)
	hrLegalModule := hr_legal.NewModule(db)
	attendanceModule := attendance.NewModule(db)
	workModule := work.NewModule(db, storageService)
	targetModule := target.NewModule(db)
	employeeModule := employee.NewModule(db, nil)

	// Set Gin to release mode in production
	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup Gin Router
	r := gin.New()

	// CORS — restrict to allowed origins (configurable)
	corsConfig := cors.DefaultConfig()
	rawOrigins := strings.Split(cfg.AllowedOrigins, ",")
	allowedOrigins := make([]string, 0, len(rawOrigins))
	for _, origin := range rawOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowedOrigins = append(allowedOrigins, trimmed)
		}
	}
	hasWildcard := false
	for _, origin := range allowedOrigins {
		if origin == "*" {
			hasWildcard = true
			break
		}
	}
	if hasWildcard {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowCredentials = false
	} else {
		corsConfig.AllowOrigins = allowedOrigins
		corsConfig.AllowCredentials = true
		corsConfig.AllowOriginFunc = func(origin string) bool {
			for _, o := range allowedOrigins {
				if strings.TrimSpace(o) == origin {
					return true
				}
			}
			if strings.HasPrefix(origin, "http://localhost:") ||
				strings.HasPrefix(origin, "http://127.0.0.1:") ||
				strings.HasPrefix(origin, "http://192.168.") ||
				strings.HasPrefix(origin, "http://10.") ||
				strings.HasSuffix(origin, "kerd2sy.com") ||
				strings.HasSuffix(origin, "aams-logistic.com") ||
				strings.HasSuffix(origin, "onrender.com") ||
				strings.HasSuffix(origin, "vercel.app") {
				return true
			}
			return false
		}
	}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept", "X-Requested-With"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	// Global security middleware
	r.Use(middleware.SecurityHeaders())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(corsConfig))
	r.Use(middleware.RateLimiter(5000, time.Minute)) // 5000 req/min per IP

	// Serve uploaded images statically
	r.Static("/uploads", "./uploads")
	r.Static("/api/v1/uploads", "./uploads")

	// Health check & root route
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
	serverStartTime := time.Now()

	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"app":       "Delivery Employee Management System API",
			"version":   "1.0.0",
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
			dbStatus = "error: " + err.Error()
		} else {
			if pingErr := sqlDB.Ping(); pingErr != nil {
				dbStatus = "disconnected: " + pingErr.Error()
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
		})
	})

	// ---------------------------------------------------------
	// 1. Unauthenticated Public Routes
	// ---------------------------------------------------------
	authModule.RegisterPublicRoutes(r)
	systemModule.RegisterPublicRoutes(r)
	hrLegalModule.RegisterPublicRoutes(r)

	// ---------------------------------------------------------
	// 2. Protected Authenticated Routes (JWT Required)
	// ---------------------------------------------------------
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret, adminRepo, empRepo))
	{
		authModule.RegisterRoutes(protected)
		employeeModule.RegisterRoutes(protected)
		workModule.RegisterRoutes(protected)
		fleetModule.RegisterRoutes(protected)
		inventoryModule.RegisterRoutes(protected)
		maintenanceModule.RegisterRoutes(protected)
		attendanceModule.RegisterRoutes(protected)
		custodyModule.RegisterRoutes(protected)
		hrLegalModule.RegisterRoutes(protected)
		ticketModule.RegisterRoutes(protected)
		targetModule.RegisterRoutes(protected)
		systemModule.RegisterRoutes(protected)
	}

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s...", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Stop the backup service
	backupSvc.Stop()

	// Give outstanding requests 15 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close database connection
	if sqlDB, err := db.DB(); err == nil {
		if closeErr := sqlDB.Close(); closeErr != nil {
			log.Printf("Error closing database: %v", closeErr)
		} else {
			log.Println("Database connection closed")
		}
	}

	log.Println("Server exited gracefully")
}

func startIqamaExpirationChecker(db *gorm.DB) {
	// Run every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		// Run once immediately on startup
		checkIqamaExpirations(db)
		for range ticker.C {
			checkIqamaExpirations(db)
		}
	}()
}

func checkIqamaExpirations(db *gorm.DB) {
	// Find all employees whose iqama expires in <= 60 days
	var emps []domain.Employee
	threshold := time.Now().AddDate(0, 0, 60)

	if err := db.Where("iqama_expiration_date IS NOT NULL AND iqama_expiration_date <= ?", threshold).Find(&emps).Error; err != nil {
		log.Printf("[Cron] Error checking iqama expirations: %v", err)
		return
	}

	todayStr := time.Now().Format("2006-01-02")
	count := 0

	for _, emp := range emps {
		// Check if a notification already exists for today
		var existing domain.Notification
		err := db.Where("employee_id = ? AND type = ? AND DATE(created_at) = ?", emp.ID, "iqama_expiry", todayStr).First(&existing).Error
		if err != nil {
			// Doesn't exist, create it
			var expTime time.Time
			if emp.IqamaExpirationDate != nil {
				parsed, err := time.Parse("2006-01-02", *emp.IqamaExpirationDate)
				if err == nil {
					expTime = parsed
				} else {
					expTime = time.Now()
				}
			} else {
				expTime = time.Now()
			}
			daysLeft := int(time.Until(expTime).Hours() / 24)

			title := "تنبيه اقتراب انتهاء إقامة"
			body := "إقامة الموظف " + emp.Name + " تنتهي بعد " + fmt.Sprintf("%d", daysLeft) + " يوم."
			if daysLeft <= 0 {
				title = "تنبيه انتهاء إقامة"
				body = "إقامة الموظف " + emp.Name + " منتهية!"
			}

			notif := domain.Notification{
				BranchID:   emp.BranchID,
				EmployeeID: &emp.ID,
				Title:      title,
				Body:       body,
				Type:       "iqama_expiry",
				Status:     "unread",
			}
			db.Create(&notif)
			count++
		}
	}

	if count > 0 {
		log.Printf("[Cron] Generated %d iqama expiration notifications", count)
	}
}
