package database

import (
	"fmt"
	"log"
	"time"

	"delivery-backend/internal/domain"
	"delivery-backend/pkg/config"
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
	sqlDB.SetMaxOpenConns(120)
	sqlDB.SetMaxIdleConns(40)
	sqlDB.SetConnMaxLifetime(60 * time.Minute)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

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
		&domain.BroadcastNotification{},
		&domain.BroadcastRead{},
		&domain.BroadcastVote{},
		&domain.OTPRequest{},
		&domain.Identifier{},
		&domain.Driver{},
		&domain.IdentifierDriver{},
		&domain.ImportBatch{},
		&domain.DailyOrder{},
		&domain.TargetAlert{},
		&domain.TargetSetting{},
		&domain.OperationVideo{},
	)
	if err != nil {
		return nil, fmt.Errorf("auto migration failed: %w", err)
	}

	// Performance indexes and schema upgrades
	if rawDB, err := db.DB(); err == nil {
		rawDB.Exec("ALTER TABLE work_sessions ADD COLUMN IF NOT EXISTS review_video_url text")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_work_sessions_start_time ON work_sessions (start_time)")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_work_sessions_start_status ON work_sessions (start_time, status)")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_work_sessions_status_reviewed_start ON work_sessions (status, is_reviewed, start_time DESC)")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_work_sessions_emp_status ON work_sessions (employee_id, status)")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_employees_branch_working ON employees (branch_id, is_working) WHERE deleted_at IS NULL")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_employees_phone ON employees (phone) WHERE deleted_at IS NULL")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_employees_national_id ON employees (national_id) WHERE deleted_at IS NULL")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_admins_phone ON admins (phone) WHERE deleted_at IS NULL")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_admins_username_lower ON admins (LOWER(username)) WHERE deleted_at IS NULL")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_custody_days_branch_date ON custody_days (branch_id, date DESC)")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_custody_expenses_day_id ON custody_expenses (custody_day_id)")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_custody_logs_branch_date ON custody_logs (branch_id, date)")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_operation_videos_expires_at ON operation_videos (expires_at) WHERE is_deleted = false")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_operation_videos_employee_name ON operation_videos (employee_name)")
		rawDB.Exec("CREATE INDEX IF NOT EXISTS idx_operation_videos_order_number ON operation_videos (order_number)")
		rawDB.Exec("ALTER TABLE investigations ADD COLUMN IF NOT EXISTS approved_by char(36)")
		rawDB.Exec("ALTER TABLE investigations ADD COLUMN IF NOT EXISTS rejected_by char(36)")
		rawDB.Exec("ALTER TABLE investigations ADD COLUMN IF NOT EXISTS approved_by_name varchar(100)")
		rawDB.Exec("ALTER TABLE investigations ADD COLUMN IF NOT EXISTS approved_by_username varchar(50)")
		rawDB.Exec("ALTER TABLE investigations ADD COLUMN IF NOT EXISTS rejected_by_name varchar(100)")
		rawDB.Exec("ALTER TABLE investigations ADD COLUMN IF NOT EXISTS rejected_by_username varchar(50)")
		rawDB.Exec("ALTER TABLE investigations ADD COLUMN IF NOT EXISTS approved_at timestamp with time zone")
		rawDB.Exec("ALTER TABLE investigations ADD COLUMN IF NOT EXISTS rejected_at timestamp with time zone")
		rawDB.Exec("ALTER TABLE investigations ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone")
		rawDB.Exec("ALTER TABLE investigations ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone")
		rawDB.Exec("ALTER TABLE investigations ADD COLUMN IF NOT EXISTS employee_name varchar(255)")
		rawDB.Exec("ALTER TABLE investigations ADD COLUMN IF NOT EXISTS supervisor_name varchar(100)")
		rawDB.Exec("ALTER TABLE investigations ADD COLUMN IF NOT EXISTS national_id varchar(50)")

		// Backfill existing investigations with employee and supervisor names if missing
		rawDB.Exec(`UPDATE investigations SET 
			employee_name = employees.name, 
			national_id = COALESCE(NULLIF(investigations.national_id, ''), employees.national_id)
			FROM employees WHERE investigations.employee_id = employees.id AND (investigations.employee_name IS NULL OR investigations.employee_name = '')`)

		rawDB.Exec(`UPDATE investigations SET 
			supervisor_name = admins.name
			FROM admins WHERE investigations.supervisor_id = admins.id AND (investigations.supervisor_name IS NULL OR investigations.supervisor_name = '')`)

		rawDB.Exec(`UPDATE investigations SET 
			supervisor_name = employees.name
			FROM employees WHERE investigations.supervisor_id = employees.id AND (investigations.supervisor_name IS NULL OR investigations.supervisor_name = '')`)

		rawDB.Exec("ALTER TABLE otp_requests ADD COLUMN IF NOT EXISTS branch_id char(36)")
		rawDB.Exec("ALTER TABLE otp_requests ADD COLUMN IF NOT EXISTS branch_name varchar(100)")
		rawDB.Exec("ALTER TABLE otp_requests ADD COLUMN IF NOT EXISTS employee_name varchar(150)")
		rawDB.Exec("ALTER TABLE otp_requests ALTER COLUMN employee_id DROP NOT NULL")
		rawDB.Exec("ALTER TABLE otp_requests ALTER COLUMN employee_name DROP NOT NULL")
		rawDB.Exec("ALTER TABLE otp_requests DROP CONSTRAINT IF EXISTS fk_otp_requests_employee")

		// Safe DDL for User Linking, Password Reset, Explicit Permissions, and User Avatar
		rawDB.Exec("ALTER TABLE admins ADD COLUMN IF NOT EXISTS avatar text")
		rawDB.Exec("UPDATE admins SET avatar = google_avatar WHERE (avatar IS NULL OR avatar = '') AND google_avatar IS NOT NULL AND google_avatar != ''")
		rawDB.Exec("ALTER TABLE admins ADD COLUMN IF NOT EXISTS employee_id char(36)")
		rawDB.Exec("ALTER TABLE admins ADD COLUMN IF NOT EXISTS must_change_password boolean DEFAULT false")
		rawDB.Exec("ALTER TABLE admins ADD COLUMN IF NOT EXISTS permissions_configured boolean DEFAULT false")
		rawDB.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS must_change_password boolean DEFAULT false")
		rawDB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_admins_employee_id ON admins (employee_id) WHERE deleted_at IS NULL AND employee_id IS NOT NULL")

		rawDB.Exec("ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS registration_image text")
		rawDB.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS push_token text")
		rawDB.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS device_uuid varchar(100)")
		rawDB.Exec("ALTER TABLE notifications ADD COLUMN IF NOT EXISTS image_url text")
		rawDB.Exec("ALTER TABLE broadcast_notifications ADD COLUMN IF NOT EXISTS title_ar text")
		rawDB.Exec("ALTER TABLE broadcast_notifications ADD COLUMN IF NOT EXISTS title_en text")
		rawDB.Exec("ALTER TABLE broadcast_notifications ADD COLUMN IF NOT EXISTS title_bn text")
		rawDB.Exec("ALTER TABLE broadcast_notifications ADD COLUMN IF NOT EXISTS body_ar text")
		rawDB.Exec("ALTER TABLE broadcast_notifications ADD COLUMN IF NOT EXISTS body_en text")
		rawDB.Exec("ALTER TABLE broadcast_notifications ADD COLUMN IF NOT EXISTS body_bn text")
		rawDB.Exec("ALTER TABLE broadcast_notifications ADD COLUMN IF NOT EXISTS poll_question_ar text")
		rawDB.Exec("ALTER TABLE broadcast_notifications ADD COLUMN IF NOT EXISTS poll_question_en text")
		rawDB.Exec("ALTER TABLE broadcast_notifications ADD COLUMN IF NOT EXISTS poll_question_bn text")
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

		// ── Vehicle plate_number partial-index migration ──────────────────────────
		// The old full unique index on plate_number includes soft-deleted rows,
		// which prevents renaming a plate to one that was previously used by a
		// deleted vehicle. We convert it to a partial index (WHERE deleted_at IS NULL)
		// so only active vehicles are constrained, freeing up plate names of
		// soft-deleted vehicles.
		rawDB.Exec("DELETE FROM vehicles WHERE deleted_at IS NOT NULL")
		rawDB.Exec("DROP INDEX IF EXISTS idx_vehicles_plate_number")
		rawDB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_vehicles_plate_number ON vehicles (plate_number) WHERE deleted_at IS NULL`)
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
	// seedTargetAccounts(db) - disabled to keep only requested primary admin
	seedTargetSettings(db)

	// Clean any legacy Base64 stored in work_sessions to disk files in background
	go sanitizeWorkSessionLegacyImages(db)

	return db, nil
}

