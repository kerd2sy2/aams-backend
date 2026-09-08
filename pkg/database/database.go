package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"delivery-backend/internal/domain"
	"delivery-backend/pkg/config"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
		&domain.Identifier{},
		&domain.Driver{},
		&domain.IdentifierDriver{},
		&domain.ImportBatch{},
		&domain.DailyOrder{},
		&domain.TargetAlert{},
		&domain.TargetSetting{},
	)
	if err != nil {
		return nil, fmt.Errorf("auto migration failed: %w", err)
	}

	// Performance indexes for target orders
	if rawDB, err := db.DB(); err == nil {
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_daily_orders_date_id ON daily_orders (order_date, identifier_id)")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_daily_orders_driver ON daily_orders (driver_id, order_date)")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_daily_orders_dedup ON daily_orders (order_date, identifier_id, driver_id, app_name)")

		// Identifier per app migration:
		// Drop old single name unique index if exists
		rawDB.Exec("ALTER TABLE identifiers DROP CONSTRAINT IF EXISTS uni_identifiers_name")
		rawDB.Exec("DROP INDEX IF EXISTS uni_identifiers_name")
		rawDB.Exec("DROP INDEX IF EXISTS idx_identifiers_name")

		// Backfill app_name for existing identifiers from their daily orders
		rawDB.Exec(`
			UPDATE identifiers i 
			SET app_name = sub.app_name 
			FROM (
				SELECT DISTINCT ON (identifier_id) identifier_id, app_name 
				FROM daily_orders 
				WHERE app_name IS NOT NULL AND app_name != '' 
				ORDER BY identifier_id, created_at DESC
			) sub 
			WHERE i.id = sub.identifier_id AND (i.app_name IS NULL OR i.app_name = '')
		`)

		// Split any daily orders that belonged to a different app than their identifier's assigned app
		rawDB.Exec(`
		DO $$
		DECLARE
			r RECORD;
			target_ident_id CHAR(36);
		BEGIN
			FOR r IN 
				SELECT DISTINCT i.name, d.app_name 
				FROM daily_orders d 
				JOIN identifiers i ON d.identifier_id = i.id 
				WHERE d.app_name IS NOT NULL AND d.app_name != '' 
				  AND LOWER(TRIM(COALESCE(i.app_name, ''))) != LOWER(TRIM(d.app_name))
			LOOP
				SELECT id::text INTO target_ident_id 
				FROM identifiers 
				WHERE LOWER(TRIM(name)) = LOWER(TRIM(r.name)) 
				  AND LOWER(TRIM(COALESCE(app_name, ''))) = LOWER(TRIM(r.app_name)) 
				LIMIT 1;

				IF target_ident_id IS NULL THEN
					target_ident_id := gen_random_uuid()::text;
					INSERT INTO identifiers (id, name, app_name, monthly_target, daily_target, is_active, created_at, updated_at)
					VALUES (target_ident_id::uuid, r.name, r.app_name, 460, 18, true, NOW(), NOW());
				END IF;

				UPDATE daily_orders 
				SET identifier_id = target_ident_id::uuid 
				WHERE app_name = r.app_name 
				  AND identifier_id IN (
					  SELECT id FROM identifiers 
					  WHERE LOWER(TRIM(name)) = LOWER(TRIM(r.name)) 
					    AND id != target_ident_id::uuid
				  );
			END LOOP;
		END $$;
		`)

		rawDB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_identifiers_name_app ON identifiers (LOWER(TRIM(name)), LOWER(TRIM(COALESCE(app_name, ''))))")
	}

	// One-time cleanup: barcode no longer needs a unique index.
	// The DDL statements below are PostgreSQL-compatible and safe to re-run.
	if rawDB, dropErr := db.DB(); dropErr == nil {
		rawDB.Exec("UPDATE inventory_items SET barcode = '' WHERE barcode IS NULL")
		rawDB.Exec("DROP INDEX IF EXISTS idx_inventory_items_barcode")
		rawDB.Exec("DROP INDEX IF EXISTS uni_inventory_items_barcode")
		rawDB.Exec("UPDATE identifiers SET daily_target = 18 WHERE daily_target = 15 OR daily_target <= 0")
		rawDB.Exec("UPDATE target_settings SET setting_value = '18' WHERE setting_key = 'DEFAULT_DAILY_TARGET'")
		rawDB.Exec("UPDATE target_alerts SET target_orders = 18, deficit = 18 - actual_orders WHERE target_orders = 15")
	}

	// Seed default data
	seedBranches(db)
	seedRoles(db)
	seedAdmin(db)
	seedAppSettings(db)
	seedTargetAccounts(db)
	seedTargetSettings(db)

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

// seedTargetAccounts ensures Admin (2642799148) and Supervisor (500500) exist with bcrypt password '3121'
func seedTargetAccounts(db *gorm.DB) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("3121"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash target account password: %v", err)
		return
	}

	// 1. Admin Account: 2642799148
	var adminRole domain.Role
	var adminRoleID *uuid.UUID
	if err := db.Where("name = ?", "SUPER_ADMIN").First(&adminRole).Error; err == nil {
		adminRoleID = &adminRole.ID
	}

	var admin domain.Admin
	if err := db.Where("username = ? OR phone = ? OR email = ?", "2642799148", "2642799148", "admin@aams-target.com").First(&admin).Error; err != nil {
		newAdmin := domain.Admin{
			Email:    "admin@aams-target.com",
			Username: "2642799148",
			Phone:    "2642799148",
			Password: string(hashedPassword),
			Name:     "مدير النظام (Admin)",
			Role:     "ADMIN",
			RoleID:   adminRoleID,
		}
		if err := db.Create(&newAdmin).Error; err != nil {
			log.Printf("[SEED NOTICE] Admin 2642799148 creation: %v", err)
		} else {
			log.Println("[SEED SUCCESS] Target Admin (2642799148) created successfully.")
		}
	} else {
		// Update password to 3121 and ensure correct role
		db.Model(&admin).Updates(map[string]interface{}{
			"password": string(hashedPassword),
			"role":     "ADMIN",
		})
	}

	// 2. Supervisor Account: 500500
	var supRole domain.Role
	var supRoleID *uuid.UUID
	if err := db.Where("name = ?", "SUPERVISOR").First(&supRole).Error; err == nil {
		supRoleID = &supRole.ID
	}

	var supervisor domain.Admin
	if err := db.Where("username = ? OR phone = ? OR email = ?", "500500", "500500", "supervisor@aams-target.com").First(&supervisor).Error; err != nil {
		newSupervisor := domain.Admin{
			Email:    "supervisor@aams-target.com",
			Username: "500500",
			Phone:    "500500",
			Password: string(hashedPassword),
			Name:     "مشرف المعرفين (Supervisor)",
			Role:     "SUPERVISOR",
			RoleID:   supRoleID,
		}
		if err := db.Create(&newSupervisor).Error; err != nil {
			log.Printf("[SEED NOTICE] Supervisor 500500 creation: %v", err)
		} else {
			log.Println("[SEED SUCCESS] Target Supervisor (500500) created successfully.")
		}
	} else {
		// Update password to 3121 and ensure correct role
		db.Model(&supervisor).Updates(map[string]interface{}{
			"password": string(hashedPassword),
			"role":     "SUPERVISOR",
		})
	}
}

// seedTargetSettings ensures default monthly target (460) and daily target (15) are present
func seedTargetSettings(db *gorm.DB) {
	defaultSettings := []domain.TargetSetting{
		{SettingKey: "DEFAULT_MONTHLY_TARGET", SettingValue: "460"},
		{SettingKey: "DEFAULT_DAILY_TARGET", SettingValue: "18"},
	}

	for _, s := range defaultSettings {
		var existing domain.TargetSetting
		if err := db.Where("setting_key = ?", s.SettingKey).First(&existing).Error; err != nil {
			if err := db.Create(&s).Error; err != nil {
				log.Printf("Failed to seed target setting %s: %v", s.SettingKey, err)
			} else {
				log.Printf("[SEED SUCCESS] Target Setting created: %s = %s", s.SettingKey, s.SettingValue)
			}
		}
	}
}
