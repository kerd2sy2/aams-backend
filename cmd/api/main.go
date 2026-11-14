package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strings"
	"syscall"
	"time"

	"delivery-backend/internal/modules/attendance"
	"delivery-backend/internal/modules/auth"
	"delivery-backend/internal/modules/custody"
	"delivery-backend/internal/modules/employee"
	"delivery-backend/internal/modules/fleet"
	"delivery-backend/internal/modules/hr_legal"
	"delivery-backend/internal/modules/inventory"
	"delivery-backend/internal/modules/maintenance"
	"delivery-backend/internal/modules/operation_video"
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
)

// isolatedModuleGroup wraps a module's route group in a Bulkhead panic-recovery shield
// and a strict context timeout so no single module failure can ever affect sibling modules.
func isolatedModuleGroup(parent *gin.RouterGroup, moduleName string, timeout time.Duration) *gin.RouterGroup {
	g := parent.Group("")
	g.Use(
		middleware.ModuleIsolationMiddleware(moduleName),
		middleware.ModuleTimeoutMiddleware(timeout),
	)
	return g
}

func main() {
	riyadhLoc, err := time.LoadLocation("Asia/Riyadh")
	if err != nil {
		log.Fatalf("Failed to load Asia/Riyadh timezone: %v", err)
	}
	time.Local = riyadhLoc
	log.Println("Timezone set to Asia/Riyadh (UTC+3)")

	cfg := config.LoadConfig()

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	backupDir := os.Getenv("BACKUP_DIR")
	if backupDir == "" {
		backupDir = "./backups"
	}
	backupSvc := backup.NewService(cfg.PGDumpPath, cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, backupDir, 600*time.Second, 50)
	go backupSvc.Start()

	seedDemoReviewer(db)
	startIqamaExpirationChecker(db)

	adminRepo := repository.NewAdminRepository(db)
	empRepo := repository.NewEmployeeRepository(db)

	storageService := service.NewStorageService(cfg)
	auditRepo := repository.NewAuditRepository(db)
	auditService := service.NewAuditService(auditRepo)

	service.InitFCM("")

	// Initialize Bounded Context Modules (Modular Monolith Architecture)
	authModule := auth.NewModule(db, cfg)
	systemModule := system.NewModule(db)
	ticketModule := ticket.NewModule(db, nil, nil)
	inventoryModule := inventory.NewModule(db)
	fleetModule := fleet.NewModule(db, storageService, auditService, inventoryModule.Contract)
	maintenanceModule := maintenance.NewModule(db)
	custodyModule := custody.NewModule(db, storageService)
	hrLegalModule := hr_legal.NewModule(db)
	attendanceModule := attendance.NewModule(db)
	workModule := work.NewModule(db, storageService, auditService)
	targetModule := target.NewModule(db)
	employeeModule := employee.NewModule(db, storageService)
	operationVideoModule := operation_video.NewModule(db, "./uploads")

	startOperationVideoCleanupWorker(operationVideoModule.Service, "./uploads")

	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.TrustedPlatform = gin.PlatformCloudflare

	r.Use(middleware.SecurityHeaders())
	tokenQueryRegex := regexp.MustCompile(`(?i)(token=)[^&]+`)
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		path := param.Path
		if strings.Contains(strings.ToLower(path), "token=") {
			path = tokenQueryRegex.ReplaceAllString(path, "${1}[REDACTED]")
		}
		return fmt.Sprintf("[GIN] %v | %3d | %13v | %15s | %-7s %#v\n%s",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode, param.Latency, param.ClientIP, param.Method, path, param.ErrorMessage,
		)
	}))
	r.Use(gin.Recovery())
	r.Use(cors.New(buildCORSConfig(cfg.AllowedOrigins)))
	r.Use(middleware.RateLimiter(25000, time.Minute))

	configureStaticUploads(r)

	registerPrivacyRoutes(r)
	registerCoreHealthRoutes(r, db, time.Now())

	// 1. Unauthenticated Public Routes (Bulkhead Protected)
	authModule.RegisterPublicRoutes(r)
	systemModule.RegisterPublicRoutes(r)
	hrLegalModule.RegisterPublicRoutes(r)

	// 2. Protected Authenticated Routes (JWT + Per-Module Bulkhead & Context Timeout)
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret, adminRepo, empRepo))
	{
		authModule.RegisterRoutes(isolatedModuleGroup(protected, "auth", 10*time.Second))
		employeeModule.RegisterRoutes(isolatedModuleGroup(protected, "employee", 12*time.Second))
		workModule.RegisterRoutes(isolatedModuleGroup(protected, "work", 20*time.Second))
		fleetModule.RegisterRoutes(isolatedModuleGroup(protected, "fleet", 12*time.Second))
		inventoryModule.RegisterRoutes(isolatedModuleGroup(protected, "inventory", 12*time.Second))
		maintenanceModule.RegisterRoutes(isolatedModuleGroup(protected, "maintenance", 12*time.Second))
		attendanceModule.RegisterRoutes(isolatedModuleGroup(protected, "attendance", 12*time.Second))
		custodyModule.RegisterRoutes(isolatedModuleGroup(protected, "custody", 15*time.Second))
		hrLegalModule.RegisterRoutes(isolatedModuleGroup(protected, "hr_legal", 15*time.Second))
		ticketModule.RegisterRoutes(isolatedModuleGroup(protected, "ticket", 12*time.Second))
		targetModule.RegisterRoutes(isolatedModuleGroup(protected, "target", 30*time.Second))
		systemModule.RegisterRoutes(isolatedModuleGroup(protected, "system", 15*time.Second))
		operationVideoModule.RegisterRoutes(isolatedModuleGroup(protected, "operation_video", 45*time.Second))
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s...", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	backupSvc.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	log.Println("Server exited gracefully")
}

// configureStaticUploads serves public uploaded assets while blocking direct static
// access to protected operation videos (which must flow through RBAC stream routes).
func configureStaticUploads(r *gin.Engine) {
	r.Use(func(c *gin.Context) {
		cleanPath := strings.ToLower(path.Clean(c.Request.URL.Path))
		if strings.HasPrefix(cleanPath, "/uploads/operation_videos") ||
			strings.HasPrefix(cleanPath, "/api/v1/uploads/operation_videos") ||
			strings.Contains(cleanPath, "/operation_videos/") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "غير مصرح بالوصول المباشر للفيديوهات"})
			return
		}
		c.Next()
	})
	r.Static("/uploads", "./uploads")
	r.Static("/api/v1/uploads", "./uploads")
}

func buildCORSConfig(rawAllowedOrigins string) cors.Config {
	corsConfig := cors.DefaultConfig()
	rawOrigins := strings.Split(rawAllowedOrigins, ",")
	allowedOrigins := make([]string, 0, len(rawOrigins))
	hasWildcard := false
	for _, origin := range rawOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowedOrigins = append(allowedOrigins, trimmed)
			if trimmed == "*" {
				hasWildcard = true
			}
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
			return origin == "https://kerd2sy.com" || (strings.HasPrefix(origin, "https://") && strings.HasSuffix(origin, ".kerd2sy.com"))
		}
	}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept", "X-Requested-With", "X-Branch-ID", "X-App-ID", "X-App-Name", "Cache-Control", "Pragma"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	return corsConfig
}
