package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/handler"
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

	// Initialize Repositories (Data Layer)
	roleRepo := repository.NewRoleRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	empRepo := repository.NewEmployeeRepository(db)
	workRepo := repository.NewWorkRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	branchRepo := repository.NewBranchRepository(db)
	invRepo := repository.NewInventoryRepository(db)
	maintenanceRepo := repository.NewMaintenanceRepository(db)
	investigationRepo := repository.NewInvestigationRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	custodyRepo := repository.NewCustodyRepository(db)
	settingRepo := repository.NewSettingRepository(db)
	vehicleRepo := repository.NewVehicleRepository(db)
	fuelLogRepo := repository.NewFuelLogRepository(db)
	violationRepo := repository.NewTrafficViolationRepository(db)
	maintRequestRepo := repository.NewMaintenanceRequestRepository(db)
	docRepo := repository.NewEmployeeDocumentRepository(db)
	bankRepo := repository.NewEmployeeBankAccountRepository(db)
	leaveRepo := repository.NewLeaveRequestRepository(db)
	ticketRepo := repository.NewSupportTicketRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	archiveRepo := repository.NewArchiveRepository(db)

	// Initialize Services (Business Layer)
	storageService := service.NewStorageService(cfg)
	auditService := service.NewAuditService(auditRepo)
	authService := service.NewAuthService(adminRepo, branchRepo, cfg)
	roleService := service.NewRoleService(roleRepo)
	adminService := service.NewAdminService(adminRepo)
	branchService := service.NewBranchService(branchRepo, empRepo)
	empService := service.NewEmployeeService(empRepo)
	vehicleService := service.NewVehicleService(vehicleRepo)
	workService := service.NewWorkService(workRepo, empRepo, maintenanceRepo, vehicleRepo)
	dashService := service.NewDashboardService(workRepo, auditRepo, empRepo)
	reportService := service.NewReportService(workRepo)
	invService := service.NewInventoryService(invRepo, empRepo, maintenanceRepo)
	maintService := service.NewMaintenanceService(maintenanceRepo)
	investigationService := service.NewInvestigationService(investigationRepo, empRepo, adminRepo)
	attendanceService := service.NewAttendanceService(attendanceRepo, empRepo, workRepo)
	custodyService := service.NewCustodyService(custodyRepo)
	settingService := service.NewSettingService(settingRepo)
	fuelLogService := service.NewFuelLogService(fuelLogRepo)
	violationService := service.NewTrafficViolationService(violationRepo)
	maintRequestService := service.NewMaintenanceRequestService(maintRequestRepo)
	docService := service.NewEmployeeDocumentService(docRepo)
	bankService := service.NewEmployeeBankAccountService(bankRepo)
	leaveService := service.NewLeaveRequestService(leaveRepo)
	ticketService := service.NewSupportTicketService(ticketRepo)
	notifService := service.NewNotificationService(notifRepo, adminRepo)
	archiveService := service.NewArchiveService(archiveRepo)

	// Initialize Handlers (Presentation Layer)
	authHandler := handler.NewAuthHandler(authService, auditService)
	roleHandler := handler.NewRoleHandler(roleService, auditService)
	adminHandler := handler.NewAdminHandler(adminService, auditService, adminRepo)

	branchHandler := handler.NewBranchHandler(branchService, auditService)
	empHandler := handler.NewEmployeeHandler(empService, storageService, auditService, workRepo)
	vehicleHandler := handler.NewVehicleHandler(vehicleService, auditService)
	workHandler := handler.NewWorkHandler(workService, auditService, attendanceService, empRepo)
	dashHandler := handler.NewDashboardHandler(dashService)
	reportHandler := handler.NewReportHandler(reportService)
	auditHandler := handler.NewAuditHandler(auditService)
	invHandler := handler.NewInventoryHandler(invService, auditService)
	maintHandler := handler.NewMaintenanceHandler(maintService)
	investigationHandler := handler.NewInvestigationHandler(investigationService, empRepo)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService, auditService, empRepo)
	custodyHandler := handler.NewCustodyHandler(custodyService)
	settingHandler := handler.NewSettingHandler(settingService, auditService)
	fuelLogHandler := handler.NewFuelLogHandler(fuelLogService, auditService)
	violationHandler := handler.NewTrafficViolationHandler(violationService, auditService)
	maintRequestHandler := handler.NewMaintenanceRequestHandler(maintRequestService, auditService)
	docHandler := handler.NewEmployeeDocumentHandler(docService)
	bankHandler := handler.NewEmployeeBankAccountHandler(bankService)
	leaveHandler := handler.NewLeaveRequestHandler(leaveService, auditService)
	ticketHandler := handler.NewSupportTicketHandler(ticketService, auditService)
	notifHandler := handler.NewNotificationHandler(notifService)
	archiveHandler := handler.NewArchiveHandler(archiveService, auditService)

	// Set Gin to release mode in production
	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup Gin Router
	r := gin.New()

	// CORS â€” restrict to allowed origins (configurable)
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
				strings.HasSuffix(origin, "aams-logistic.com") {
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

	// Health check & root route
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "online",
			"app":     "Delivery Employee Management System API",
			"version": "1.0.0",
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

	// Auth routes
	authRoutes := r.Group("/api/v1")
	{
		authRoutes.POST("/login", middleware.StrictLoginLimiter(), authHandler.Login)
		authRoutes.POST("/refresh", middleware.StrictLoginLimiter(), authHandler.RefreshToken)
		authRoutes.POST("/auth/google/login", middleware.StrictLoginLimiter(), authHandler.GoogleLogin)
	}

	// Public settings (no auth required) — used by login page
	r.GET("/api/v1/settings/public", settingHandler.GetPublicSettings)

	// Public document & investigation access for QR code scanner (no auth required)
	r.GET("/api/v1/public/doc/:id", investigationHandler.GetPublicByID)
	r.GET("/api/v1/public/investigations/:id", investigationHandler.GetPublicByID)

	// Protected Routes (JWT Required)
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret, adminRepo))
	{
		// Current authenticated user (real-time branch from DB)
		protected.GET("/me", authHandler.Me)
		protected.POST("/auth/google/link", authHandler.LinkGoogle)
		protected.POST("/auth/google/unlink", authHandler.UnlinkGoogle)

		// Employee Routes + file upload (now secured)
		protected.POST("/employees", empHandler.Create)
		protected.GET("/employees", empHandler.GetAll)
		protected.GET("/employees/search", empHandler.Search)
		protected.GET("/employees/working", empHandler.GetWorking)
		protected.GET("/employees/:id", empHandler.GetByID)
		protected.PUT("/employees/:id", empHandler.Update)
		protected.DELETE("/employees/:id", empHandler.Delete)
		protected.POST("/employees/batch-oil-setup", empHandler.BatchSetOilChange)
		protected.GET("/employees/:id/barcode", empHandler.GetBarcode)
		protected.GET("/employees/:id/qrcode", empHandler.GetQRCode)
		protected.GET("/employees/:id/print-card", empHandler.GetPrintCard)
		protected.POST("/upload", empHandler.UploadImage)
		protected.POST("/upload-file", empHandler.UploadFile)

		// Work Sessions
		protected.POST("/work/start", workHandler.StartWork)
		protected.POST("/work/end", workHandler.EndWork)
		protected.PUT("/work/:id", workHandler.UpdateWorkSession)
		protected.GET("/work/active", workHandler.GetActiveSession)
		protected.GET("/work/last-km", workHandler.GetLastKM)
		protected.GET("/work/today-count", workHandler.TodayCount)
		protected.GET("/work/check-oil", workHandler.CheckOilChange)

		// Vehicles Management (ط§ظ„ط¯ط¨ط§ط¨ط§طھ ظˆط§ظ„ظ…ط±ظƒط¨ط§طھ)
		protected.GET("/vehicles", vehicleHandler.GetAll)
		protected.POST("/vehicles", vehicleHandler.Create)
		protected.GET("/vehicles/check-km", vehicleHandler.CheckKM)
		protected.GET("/vehicles/:id", vehicleHandler.GetByID)
		protected.PUT("/vehicles/:id", vehicleHandler.Update)
		protected.DELETE("/vehicles/:id", vehicleHandler.Delete)
		protected.POST("/vehicles/:id/oil-change", vehicleHandler.RecordOilChange)

		// Inventory Management
		protected.GET("/inventory/items", invHandler.GetItems)
		protected.GET("/inventory/items/:id", invHandler.GetItemByID)
		protected.GET("/inventory/barcode", invHandler.FindByBarcode)
		protected.POST("/inventory/items", invHandler.CreateItem)
		protected.PUT("/inventory/items/:id", invHandler.UpdateItem)
		protected.DELETE("/inventory/items/:id", invHandler.DeleteItem)
		protected.POST("/inventory/add-stock", invHandler.AddStock)
		protected.POST("/inventory/remove-stock", invHandler.RemoveStock)
		protected.POST("/inventory/dispense-oil", invHandler.DispenseOil)
		protected.GET("/inventory/transactions", invHandler.GetTransactions)
		protected.DELETE("/inventory/transactions", invHandler.DeleteAllTransactions)

		// Maintenance
		protected.GET("/maintenance/logs", maintHandler.GetAllLogs)
		protected.GET("/maintenance/employee-logs", maintHandler.GetEmployeeLogs)

		// Investigation
		protected.POST("/investigations", investigationHandler.Create)
		protected.GET("/investigations", investigationHandler.GetAll)
		protected.GET("/investigations/pending-count", investigationHandler.PendingCount)
		protected.GET("/investigations/:id", investigationHandler.GetByID)
		protected.PUT("/investigations/:id", investigationHandler.Update)
		protected.POST("/investigations/:id/approve", investigationHandler.Approve)

		// Notifications
		protected.GET("/notifications", notifHandler.GetMyNotifications)
		protected.PUT("/notifications/read-all", notifHandler.MarkAllAsRead)
		protected.PUT("/notifications/:id/read", notifHandler.MarkAsRead)
		protected.POST("/investigations/:id/reject", investigationHandler.Reject)

		// Attendance
		protected.GET("/attendance", attendanceHandler.GetAttendance)
		protected.POST("/attendance/:employee_id", attendanceHandler.ToggleAttendance)

		// Custody (ط§ظ„ط¹ظ‡ط¯ط©)
		protected.GET("/custody", custodyHandler.List)
		protected.POST("/custody", custodyHandler.Create)
		protected.POST("/custody/add-amount", custodyHandler.AddAmount)
		protected.GET("/custody/logs", custodyHandler.GetLogs)
		protected.DELETE("/custody/logs/:id", custodyHandler.DeleteLog)
		protected.POST("/custody/:id/expenses", custodyHandler.AddExpense)
		protected.DELETE("/custody/expenses/:id", custodyHandler.DeleteExpense)

		// Analytics & Reports
		protected.GET("/dashboard", dashHandler.GetStats)
		protected.GET("/reports", reportHandler.GetReports)
		protected.GET("/reports/export", reportHandler.ExportReports)
		protected.GET("/reports/daily", reportHandler.GetDailyReport)
		protected.GET("/reports/daily/export", reportHandler.ExportDailyReport)

		// Audit Logs
		protected.GET("/audit-logs", auditHandler.GetLogs)
		protected.DELETE("/audit-logs/clear", auditHandler.ClearLogs)
		protected.DELETE("/audit-logs/bulk", auditHandler.BulkDeleteLogs)
		protected.DELETE("/audit-logs/:id", auditHandler.DeleteLog)

		// User Management (Admins)
		protected.GET("/users", adminHandler.GetAll)
		protected.POST("/users", adminHandler.Create)
		protected.PUT("/users/:id", adminHandler.Update)
		protected.DELETE("/users/:id", adminHandler.Delete)
		protected.POST("/users/change-password", adminHandler.ChangePassword)

		// Roles & Permissions Management 
		protected.GET("/roles", roleHandler.GetAll)
		protected.POST("/roles", roleHandler.Create)
		protected.GET("/roles/:id", roleHandler.GetByID)
		protected.PUT("/roles/:id", roleHandler.Update)
		protected.DELETE("/roles/:id", roleHandler.Delete)
		protected.GET("/permissions", roleHandler.GetPermissions)


		// Branch Management
		protected.GET("/branches", branchHandler.GetAll)
		protected.GET("/branches/:id", branchHandler.GetByID)
		protected.POST("/branches", branchHandler.Create)
		protected.PUT("/branches/:id", branchHandler.Update)
		protected.DELETE("/branches/:id", branchHandler.Delete)

		// Settings Management
		protected.GET("/settings", settingHandler.GetSettings)
		protected.PUT("/settings", settingHandler.UpdateSettings)

		// 1. Fuel Logs (ط³ط¬ظ„ط§طھ ط§ظ„ظˆظ‚ظˆط¯)
		protected.GET("/fuel-logs", fuelLogHandler.GetAll)
		protected.POST("/fuel-logs", fuelLogHandler.Create)
		protected.PUT("/fuel-logs/:id", fuelLogHandler.Update)
		protected.DELETE("/fuel-logs/:id", fuelLogHandler.Delete)

		// 2. Traffic Violations (ط§ظ„ظ…ط®ط§ظ„ظپط§طھ ط§ظ„ظ…ط±ظˆط±ظٹط©)
		protected.GET("/violations", violationHandler.GetAll)
		protected.POST("/violations", violationHandler.Create)
		protected.PUT("/violations/:id", violationHandler.Update)
		protected.DELETE("/violations/:id", violationHandler.Delete)

		// 3. Maintenance Requests (ط·ظ„ط¨ط§طھ طµظٹط§ظ†ط© ط§ظ„ظ…ط±ظƒط¨ط§طھ)
		protected.GET("/maintenance-requests", maintRequestHandler.GetAll)
		protected.POST("/maintenance-requests", maintRequestHandler.Create)
		protected.PUT("/maintenance-requests/:id", maintRequestHandler.Update)
		protected.DELETE("/maintenance-requests/:id", maintRequestHandler.Delete)

		// 4. Employee Documents (المستندات والرخص)
		protected.GET("/documents", docHandler.GetAll)
		protected.GET("/documents/expiring", docHandler.GetExpiringSoon)
		protected.GET("/documents/:id", docHandler.GetByID)
		protected.POST("/documents", docHandler.Create)
		protected.PUT("/documents/:id", docHandler.Update)
		protected.DELETE("/documents/:id", docHandler.Delete)

		// 5. Employee Bank Accounts (ط§ظ„ط­ط³ط§ط¨ط§طھ ط§ظ„ط¨ظ†ظƒظٹط©)
		protected.GET("/bank-accounts", bankHandler.GetAll)
		protected.POST("/bank-accounts", bankHandler.Create)
		protected.PUT("/bank-accounts/:id", bankHandler.Update)
		protected.DELETE("/bank-accounts/:id", bankHandler.Delete)

		// 6. Leave Requests (ط·ظ„ط¨ط§طھ ط§ظ„ط¥ط¬ط§ط²ط§طھ)
		protected.GET("/leaves", leaveHandler.GetAll)
		protected.POST("/leaves", leaveHandler.Create)
		protected.PUT("/leaves/:id/status", leaveHandler.UpdateStatus)
		protected.DELETE("/leaves/:id", leaveHandler.Delete)

		// 7. Support Tickets (تذاكر الدعم والشكاوى)
		protected.GET("/tickets", ticketHandler.GetAll)
		protected.POST("/tickets", ticketHandler.Create)
		protected.PUT("/tickets/:id", ticketHandler.Update)
		protected.DELETE("/tickets/:id", ticketHandler.Delete)

		// 8. Archive & Trash Management (سجل الأرشيف والمحذوفات)
		protected.GET("/archive", archiveHandler.GetArchived)
		protected.POST("/archive/restore", archiveHandler.Restore)
		protected.DELETE("/archive/permanent", archiveHandler.PermanentDelete)
		protected.POST("/archive/restore-bulk", archiveHandler.BulkRestore)
		protected.DELETE("/archive/permanent-bulk", archiveHandler.BulkPermanentDelete)
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

