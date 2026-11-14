package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

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
	"delivery-backend/pkg/config"
	"delivery-backend/pkg/jwt"
	"delivery-backend/pkg/middleware"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}
	// SQLite ":memory:" creates an isolated database per connection unless restricted to 1 connection.
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	// Auto migrate models across modules
	err = db.AutoMigrate(
		&auth.Admin{},
		&auth.Role{},
		&auth.Branch{},
		&auth.OTPRequest{},
		&employee.Employee{},
		&attendance.Attendance{},
		&fleet.Vehicle{},
		&inventory.InventoryItem{},
		&inventory.InventoryTransaction{},
		&inventory.PurchaseInvoice{},
		&inventory.PurchaseInvoiceItem{},
		&maintenance.MaintenanceLog{},
		&maintenance.MaintenanceRequest{},
		&custody.CustodyDay{},
		&custody.CustodyExpense{},
		&custody.CustodyLog{},
		&hr_legal.Investigation{},
		&hr_legal.LeaveRequest{},
		&hr_legal.TrafficViolation{},
		&hr_legal.EmployeeDocument{},
		&hr_legal.EmployeeBankAccount{},
		&hr_legal.FuelLog{},
		&ticket.SupportTicket{},
		&target.Identifier{},
		&target.Driver{},
		&target.IdentifierDriver{},
		&target.DailyOrder{},
		&target.TargetAlert{},
		&target.TargetSetting{},
		&target.ImportBatch{},
		&system.AppSetting{},
		&system.Notification{},
		&system.BroadcastNotification{},
		&system.BroadcastRead{},
		&system.BroadcastVote{},
		&work.WorkSession{},
		&operation_video.OperationVideo{},
	)
	if err != nil {
		t.Fatalf("failed to auto-migrate models: %v", err)
	}

	return db
}

func buildVerificationRouter(t *testing.T) (*gin.Engine, *gorm.DB, *config.Config, uuid.UUID, uuid.UUID) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	cfg := &config.Config{
		JWTSecret:        "test-secret-at-least-32-chars-long!",
		JWTRefreshSecret: "test-refresh-secret-at-least-32-chars!",
		Port:             "8081",
	}

	adminID := uuid.New()
	admin := auth.Admin{
		ID:          adminID,
		Email:       "admin@test.com",
		Username:    "admin",
		Name:        "Super Admin",
		Role:        "SUPER_ADMIN",
		Permissions: "[\"*\"]",
	}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("failed to seed admin: %v", err)
	}

	empID := uuid.New()
	emp := employee.Employee{
		ID:         empID,
		Name:       "Test Driver",
		NationalID: "1000000001",
		Phone:      "0500000001",
		JobRole:    "DRIVER",
		RoleName:   "DRIVER",
		IsWorking:  false,
	}
	if err := db.Create(&emp).Error; err != nil {
		t.Fatalf("failed to seed driver employee: %v", err)
	}

	adminRepo := repository.NewAdminRepository(db)
	empRepo := repository.NewEmployeeRepository(db)
	storageService := service.NewStorageService(cfg)
	auditRepo := repository.NewAuditRepository(db)
	auditService := service.NewAuditService(auditRepo)

	// Initialize all 13 modules
	authMod := auth.NewModule(db, cfg)
	systemMod := system.NewModule(db)
	ticketMod := ticket.NewModule(db, nil, nil)
	inventoryMod := inventory.NewModule(db)
	fleetMod := fleet.NewModule(db, storageService, auditService, inventoryMod.Contract)
	maintMod := maintenance.NewModule(db)
	custodyMod := custody.NewModule(db, storageService)
	hrLegalMod := hr_legal.NewModule(db)
	attMod := attendance.NewModule(db)
	workMod := work.NewModule(db, storageService, auditService)
	targetMod := target.NewModule(db)
	empMod := employee.NewModule(db, storageService)
	opVideoMod := operation_video.NewModule(db, t.TempDir())

	r := gin.New()
	configureStaticUploads(r)
	registerCoreHealthRoutes(r, db, time.Now())

	// Public routes
	authMod.RegisterPublicRoutes(r)
	systemMod.RegisterPublicRoutes(r)
	hrLegalMod.RegisterPublicRoutes(r)

	// Protected routes
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret, adminRepo, empRepo))
	{
		authMod.RegisterRoutes(isolatedModuleGroup(protected, "auth", 5*time.Second))
		empMod.RegisterRoutes(isolatedModuleGroup(protected, "employee", 5*time.Second))
		workMod.RegisterRoutes(isolatedModuleGroup(protected, "work", 5*time.Second))
		fleetMod.RegisterRoutes(isolatedModuleGroup(protected, "fleet", 5*time.Second))
		inventoryMod.RegisterRoutes(isolatedModuleGroup(protected, "inventory", 5*time.Second))
		maintMod.RegisterRoutes(isolatedModuleGroup(protected, "maintenance", 5*time.Second))
		attMod.RegisterRoutes(isolatedModuleGroup(protected, "attendance", 5*time.Second))
		custodyMod.RegisterRoutes(isolatedModuleGroup(protected, "custody", 5*time.Second))
		hrLegalMod.RegisterRoutes(isolatedModuleGroup(protected, "hr_legal", 5*time.Second))
		ticketMod.RegisterRoutes(isolatedModuleGroup(protected, "ticket", 5*time.Second))
		targetMod.RegisterRoutes(isolatedModuleGroup(protected, "target", 5*time.Second))
		systemMod.RegisterRoutes(isolatedModuleGroup(protected, "system", 5*time.Second))
		opVideoMod.RegisterRoutes(isolatedModuleGroup(protected, "operation_video", 5*time.Second))
	}

	return r, db, cfg, adminID, empID
}

func TestAllModules_InitializationAndRouting(t *testing.T) {
	r, _, cfg, adminID, _ := buildVerificationRouter(t)

	// Generate valid access token for tests
	accessToken, _, err := jwt.GenerateTokens(adminID, "admin@test.com", "Super Admin", "SUPER_ADMIN", nil, cfg.JWTSecret, cfg.JWTRefreshSecret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		moduleName     string
		method         string
		url            string
		body           interface{}
		expectedStatus int
	}{
		// Public routes
		{"system (public settings)", "GET", "/api/v1/settings/public", nil, http.StatusOK},

		// Auth Module
		{"auth (me)", "GET", "/api/v1/me", nil, http.StatusOK},
		{"auth (branches list)", "GET", "/api/v1/branches", nil, http.StatusOK},
		{"auth (roles list)", "GET", "/api/v1/roles", nil, http.StatusOK},
		{"auth (users list)", "GET", "/api/v1/users", nil, http.StatusOK},

		// Employee Module
		{"employee (list)", "GET", "/api/v1/employees", nil, http.StatusOK},

		// Fleet Module
		{"fleet (list vehicles)", "GET", "/api/v1/vehicles", nil, http.StatusOK},

		// Inventory Module
		{"inventory (items list)", "GET", "/api/v1/inventory/items", nil, http.StatusOK},

		// Maintenance Module
		{"maintenance (requests)", "GET", "/api/v1/maintenance-requests", nil, http.StatusOK},

		// Custody Module
		{"custody (days list)", "GET", "/api/v1/custody", nil, http.StatusOK},

		// Attendance Module
		{"attendance (records)", "GET", "/api/v1/attendance", nil, http.StatusOK},

		// Ticket Module
		{"ticket (list tickets)", "GET", "/api/v1/tickets", nil, http.StatusOK},

		// Target Module
		{"target (dashboard summary)", "GET", "/api/v1/target/dashboard", nil, http.StatusOK},
		{"target (identifiers list)", "GET", "/api/v1/target/identifiers", nil, http.StatusOK},

		// System Module & Health Endpoints
		{"core (health)", "GET", "/api/v1/health", nil, http.StatusOK},
		{"core (modules health)", "GET", "/api/v1/health/modules", nil, http.StatusOK},
		{"core (system performance)", "GET", "/api/v1/system/performance", nil, http.StatusOK},
		{"system (dashboard stats)", "GET", "/api/v1/dashboard", nil, http.StatusOK},
		{"system (settings)", "GET", "/api/v1/settings", nil, http.StatusOK},
		{"system (reports)", "GET", "/api/v1/reports", nil, http.StatusOK},
		{"system (daily reports)", "GET", "/api/v1/reports/daily", nil, http.StatusOK},
		{"system (archive)", "GET", "/api/v1/archive", nil, http.StatusOK},

		// HR/Legal Module
		{"hr_legal (investigations)", "GET", "/api/v1/investigations", nil, http.StatusOK},
		{"hr_legal (leave requests)", "GET", "/api/v1/leaves", nil, http.StatusOK},
		{"hr_legal (violations)", "GET", "/api/v1/violations", nil, http.StatusOK},

		// Work Module
		{"work (last km)", "GET", "/api/v1/work/last-km", nil, http.StatusOK},

		// Operation Video Module
		{"operation_video (list)", "GET", "/api/v1/operation-videos", nil, http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.moduleName, func(t *testing.T) {
			var bodyBytes []byte
			if tc.body != nil {
				bodyBytes, _ = json.Marshal(tc.body)
			}

			req, _ := http.NewRequest(tc.method, tc.url, bytes.NewBuffer(bodyBytes))
			if tc.url != "/api/v1/settings/public" && tc.url != "/api/v1/health" && tc.url != "/api/v1/health/modules" && tc.url != "/api/v1/system/performance" {
				req.Header.Set("Authorization", "Bearer "+accessToken)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("[%s] Expected status %d, got %d. Body: %s", tc.moduleName, tc.expectedStatus, w.Code, w.Body.String())
			}

			if tc.url == "/api/v1/health/modules" && w.Code == http.StatusOK {
				var payload struct {
					Status  string                    `json:"status"`
					Modules []middleware.ModuleStatus `json:"modules"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
					t.Fatalf("failed to decode /api/v1/health/modules response: %v", err)
				}
				if payload.Status != "healthy" || len(payload.Modules) < 11 {
					t.Errorf("expected healthy status with >=11 modules, got status=%q count=%d", payload.Status, len(payload.Modules))
				}
			}
		})
	}
}

func TestConcurrent500UsersAndSecurity(t *testing.T) {
	r, db, cfg, adminID, empID := buildVerificationRouter(t)

	adminToken, _, err := jwt.GenerateTokens(adminID, "admin@test.com", "Super Admin", "SUPER_ADMIN", nil, cfg.JWTSecret, cfg.JWTRefreshSecret)
	if err != nil {
		t.Fatalf("failed to generate admin token: %v", err)
	}

	driverToken, _, err := jwt.GenerateTokens(empID, "1000000001", "Test Driver", "DRIVER", nil, cfg.JWTSecret, cfg.JWTRefreshSecret)
	if err != nil {
		t.Fatalf("failed to generate driver token: %v", err)
	}

	t.Run("500 concurrent users load", func(t *testing.T) {
		endpoints := []string{
			"/api/v1/me",
			"/api/v1/employees",
			"/api/v1/custody",
			"/api/v1/work/last-km",
			"/api/v1/target/dashboard",
			"/api/v1/health/modules",
		}

		const totalWorkers = 500
		startGate := make(chan struct{})
		var wg sync.WaitGroup
		var failed atomic.Int64
		var errSamples sync.Map

		started := time.Now()
		for i := 0; i < totalWorkers; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				<-startGate
				targetURL := endpoints[idx%len(endpoints)]
				req, _ := http.NewRequest(http.MethodGet, targetURL, nil)
				if targetURL != "/api/v1/health/modules" {
					req.Header.Set("Authorization", "Bearer "+adminToken)
				}
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				if w.Code != http.StatusOK {
					failed.Add(1)
					errSamples.Store(targetURL, fmt.Sprintf("status=%d body=%s", w.Code, w.Body.String()))
				}
			}(i)
		}

		close(startGate)
		wg.Wait()
		elapsed := time.Since(started)
		t.Logf("500 concurrent requests completed in %v", elapsed)

		if count := failed.Load(); count > 0 {
			errSamples.Range(func(k, v interface{}) bool {
				t.Logf("sample failure for %v: %v", k, v)
				return true
			})
			t.Fatalf("expected 0 failed requests under 500 concurrent users, got %d", count)
		}
	})

	t.Run("SQL injection guard in employees sort_by", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/employees?sort_by=name%3BDROP+TABLE+employees--&order=DESC%3BSELECT+1", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK with sanitized fallback sort, got %d: %s", w.Code, w.Body.String())
		}

		var empCount int64
		if err := db.Table("employees").Count(&empCount).Error; err != nil || empCount < 1 {
			t.Fatalf("employees table integrity check failed after sort_by injection attempt: count=%d err=%v", empCount, err)
		}
	})

	t.Run("RBAC guard on target alerts for driver", func(t *testing.T) {
		paths := []string{
			"/api/v1/target/alerts/resolve-all",
			fmt.Sprintf("/api/v1/target/alerts/%s/resolve", uuid.New().String()),
		}
		for _, p := range paths {
			req, _ := http.NewRequest(http.MethodPatch, p, nil)
			req.Header.Set("Authorization", "Bearer "+driverToken)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusForbidden {
				t.Fatalf("expected 403 Forbidden for driver calling %s, got %d: %s", p, w.Code, w.Body.String())
			}
		}
	})

	t.Run("Static operation_videos access guard", func(t *testing.T) {
		paths := []string{
			"/uploads/operation_videos/test.mp4",
			"/api/v1/uploads/operation_videos/test.mp4",
			"/uploads/personal/../operation_videos/test.mp4",
		}
		for _, p := range paths {
			req, _ := http.NewRequest(http.MethodGet, p, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusForbidden {
				t.Fatalf("expected 403 Forbidden for static operation video path %s, got %d: %s", p, w.Code, w.Body.String())
			}
		}
	})

	t.Run("CORS exact subdomain validation", func(t *testing.T) {
		corsCfg := buildCORSConfig("")
		if corsCfg.AllowOriginFunc("https://evilkerd2sy.com") {
			t.Fatal("expected https://evilkerd2sy.com to be rejected")
		}
		if corsCfg.AllowOriginFunc("http://app.kerd2sy.com") {
			t.Fatal("expected insecure http subdomain to be rejected")
		}
		if !corsCfg.AllowOriginFunc("https://kerd2sy.com") {
			t.Fatal("expected https://kerd2sy.com to be allowed")
		}
		if !corsCfg.AllowOriginFunc("https://dashboard.kerd2sy.com") {
			t.Fatal("expected https://dashboard.kerd2sy.com to be allowed")
		}
	})
}
