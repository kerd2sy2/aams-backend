package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"delivery-backend/internal/domain"
	"delivery-backend/pkg/config"

	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Riyadh",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure the connection pool for concurrency safety and stability.
	// This prevents the "database is locked"/connection exhaustion that caused
	// the previous random disconnects when running on a single SQLite connection.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to access underlying database: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	log.Printf("[PostgreSQL] Connected to database %q on %s:%s", cfg.DBName, cfg.DBHost, cfg.DBPort)

	// Auto Migration
	err = db.AutoMigrate(
		&domain.Branch{},
		&domain.Role{},
		&domain.Admin{},
		&domain.Employee{},
		&domain.WorkSession{},
		&domain.AuditLog{},
		&domain.InventoryItem{},
		&domain.InventoryTransaction{},
		&domain.PurchaseInvoice{},
		&domain.PurchaseInvoiceItem{},
		&domain.MaintenanceLog{},
		&domain.Investigation{},
		&domain.Attendance{},
		&domain.CustodyDay{},
		&domain.CustodyExpense{},
		&domain.CustodyLog{},
		&domain.AppSetting{},
		&domain.Vehicle{},
		&domain.FuelLog{},
		&domain.TrafficViolation{},
		&domain.MaintenanceRequest{},
		&domain.EmployeeDocument{},
		&domain.EmployeeBankAccount{},
		&domain.LeaveRequest{},
		&domain.SupportTicket{},
		&domain.Notification{},
		&domain.OTPRequest{},
	)
	if err != nil {
		return nil, fmt.Errorf("auto migration failed: %w", err)
	}

	// One-time cleanup: barcode no longer needs a unique index.
	// The DDL statements below are PostgreSQL-compatible and safe to re-run.
	if rawDB, dropErr := db.DB(); dropErr == nil {
		rawDB.Exec("UPDATE inventory_items SET barcode = '' WHERE barcode IS NULL")
		rawDB.Exec("DROP INDEX IF EXISTS idx_inventory_items_barcode")
		rawDB.Exec("DROP INDEX IF EXISTS uni_inventory_items_barcode")
	}

	// Seed default data
	seedBranches(db)
	seedRoles(db)
	seedAdmin(db)
	seedAppSettings(db)

	return db, nil
}

func seedRoles(db *gorm.DB) {
	defaultRoles := []domain.Role{
		{
			Name:        "SUPER_ADMIN",
			DisplayName: "مدير عام (مسؤول النظام)",
			Description: "كامل الصلاحيات للتحكم في كافة أقسام وإعدادات ومستخدمي النظام",
			Permissions: `["*"]`,
			IsSystem:    true,
		},
		{
			Name:        "SUPERVISOR",
			DisplayName: "مشرف وردية",
			Description: "متابعة المناديب، تسجيل الدوام، العهدة اليومية، فحص الزيت ومتابعة العمليات الميدانية",
			Permissions: `["employees.view","employees.create","employees.edit","employees.cards","work.view","work.start","work.end","custody.view","custody.add","vehicles.view","vehicles.oil","fuel.view","fuel.manage","maintenance.view","maintenance.manage","attendance.view","investigations.view","investigations.create","inventory.view","inventory.dispense","tickets.view","tickets.manage"]`,
			IsSystem:    true,
		},
		{
			Name:        "ACCOUNTANT",
			DisplayName: "إدارة مالية ومحاسبة",
			Description: "إدارة العهدة، المصروفات، الحسابات البنكية، الوقود، المخالفات والتقارير المالية",
			Permissions: `["custody.view","custody.add","custody.delete","bank_accounts.view","bank_accounts.manage","fuel.view","fuel.manage","violations.view","violations.manage","reports.view","reports.export","inventory.view","investigations.view"]`,
			IsSystem:    true,
		},
		{
			Name:        "HR",
			DisplayName: "مسؤول الموارد البشرية (HR)",
			Description: "إدارة المناديب، المستندات والرخص، الحسابات البنكية، طلبات الإجازات، الحضور والغياب، والتحقيقات",
			Permissions: `["employees.view","employees.create","employees.edit","employees.cards","documents.view","documents.manage","bank_accounts.view","bank_accounts.manage","leaves.view","leaves.manage","attendance.view","investigations.view","investigations.create","tickets.view","tickets.manage","reports.view"]`,
			IsSystem:    true,
		},
		{
			Name:        "FLEET_MANAGER",
			DisplayName: "مسؤول الأسطول والصيانة",
			Description: "متابعة وصيانة المركبات والدبابات، غيار الزيت، الوقود، المخالفات المرورية والمخزون",
			Permissions: `["vehicles.view","vehicles.manage","vehicles.oil","maintenance.view","maintenance.manage","fuel.view","fuel.manage","violations.view","violations.manage","inventory.view","inventory.dispense","reports.view"]`,
			IsSystem:    true,
		},
	}

	for _, role := range defaultRoles {
		var existing domain.Role
		if err := db.Where("name = ?", role.Name).First(&existing).Error; err != nil {
			if err := db.Create(&role).Error; err != nil {
				log.Printf("Failed to seed role %s: %v", role.Name, err)
			} else {
				log.Printf("[SEED SUCCESS] Role created: %s (%s)", role.Name, role.DisplayName)
			}
		}
	}
}

func seedAppSettings(db *gorm.DB) {
	var count int64
	db.Model(&domain.AppSetting{}).Count(&count)
	if count == 0 {
		settings := []domain.AppSetting{
			{Key: "site_name", Value: "نظام إدارة التوصيل AAMS"},
			{Key: "logo_url", Value: ""},
		}
		for i := range settings {
			if err := db.Create(&settings[i]).Error; err != nil {
				log.Printf("Failed to seed setting %s: %v", settings[i].Key, err)
			} else {
				log.Printf("[SEED SUCCESS] App Setting created: %s = %s", settings[i].Key, settings[i].Value)
			}
		}
	}
}

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&domain.Admin{}).Count(&count)
	if count == 0 {
		defaultAdminPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
		if defaultAdminPassword == "" {
			defaultAdminPassword = "Admin@2026!"
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash default admin password: %v", err)
			return
		}

		// Find super admin role
		var superRole domain.Role
		var roleID *uuid.UUID
		if err := db.Where("name = ?", "SUPER_ADMIN").First(&superRole).Error; err == nil {
			roleID = &superRole.ID
		}

		admin := domain.Admin{
			Email:    "hani@aams.com",
			Username: "hani",
			Password: string(hashedPassword),
			Name:     "هاني",
			Role:     "ADMIN",
			RoleID:   roleID,
		}
		if err := db.Create(&admin).Error; err != nil {
			log.Printf("Failed to seed admin user: %v", err)
		} else {
			log.Println("[SEED SUCCESS] Default admin created. Change password immediately on first login!")
		}
	} else {
		// Link any unlinked ADMIN to super admin role if role_id is null
		var superRole domain.Role
		if err := db.Where("name = ?", "SUPER_ADMIN").First(&superRole).Error; err == nil {
			db.Model(&domain.Admin{}).Where("(role = 'ADMIN' OR role = 'SUPER_ADMIN') AND role_id IS NULL").Update("role_id", superRole.ID)
		}
		var supRole domain.Role
		if err := db.Where("name = ?", "SUPERVISOR").First(&supRole).Error; err == nil {
			db.Model(&domain.Admin{}).Where("role = 'SUPERVISOR' AND role_id IS NULL").Update("role_id", supRole.ID)
		}
	}
}

func seedBranches(db *gorm.DB) {
	var count int64
	db.Model(&domain.Branch{}).Count(&count)
	if count == 0 {
		branches := []domain.Branch{
			{Name: "الفرع الرئيسي"},
			{Name: "الفرع الثاني"},
		}
		for i := range branches {
			if err := db.Create(&branches[i]).Error; err != nil {
				log.Printf("Failed to seed branch: %v", err)
			} else {
				log.Printf("[SEED SUCCESS] Branch created: %s", branches[i].Name)
			}
		}
	}
}

