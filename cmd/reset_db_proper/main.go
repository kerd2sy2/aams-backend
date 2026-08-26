package main

import (
	"fmt"
	"log"
	"os"

	"delivery-backend/internal/domain"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dbPath := "d:\\aams\\backend\\delivery_local.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	fmt.Println("=== AAMS DB RESET (Proper) ===")
	fmt.Printf("Target DB: %s\n\n", dbPath)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	rawDB, _ := db.DB()
	rawDB.Exec("PRAGMA journal_mode=WAL")
	rawDB.Exec("PRAGMA foreign_keys=OFF")
	rawDB.Exec("PRAGMA busy_timeout=5000")

	fmt.Println("[1/5] Running AutoMigrate to ensure all tables exist...")
	err = db.AutoMigrate(
		&domain.Branch{},
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
	)
	if err != nil {
		log.Fatalf("AutoMigrate FAILED: %v", err)
	}
	fmt.Println("  AutoMigrate OK - all tables exist now")

	ensureSeedData(db)

	fmt.Println("\n[2/5] Counts BEFORE reset (kept tables):")
	printCounts(db, true)

	fmt.Println("\n[3/5] Clearing operational tables (keeping branches/admins/employees)...")
	clearOrder := []struct {
		name  string
		model interface{}
	}{
		{"custody_expenses", &domain.CustodyExpense{}},
		{"custody_logs", &domain.CustodyLog{}},
		{"custody_days", &domain.CustodyDay{}},
		{"attendances", &domain.Attendance{}},
		{"work_sessions", &domain.WorkSession{}},
		{"audit_logs", &domain.AuditLog{}},
		{"purchase_invoice_items", &domain.PurchaseInvoiceItem{}},
		{"purchase_invoices", &domain.PurchaseInvoice{}},
		{"inventory_transactions", &domain.InventoryTransaction{}},
		{"inventory_items", &domain.InventoryItem{}},
		{"maintenance_logs", &domain.MaintenanceLog{}},
		{"investigations", &domain.Investigation{}},
	}

	var totalCleared int64
	for _, t := range clearOrder {
		res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(t.model)
		if res.Error != nil {
			fmt.Printf("  ERROR clearing %s: %v\n", t.name, res.Error)
		} else {
			fmt.Printf("  Cleared %s: %d rows\n", t.name, res.RowsAffected)
			totalCleared += res.RowsAffected
		}
	}
	fmt.Printf("  Total rows cleared: %d\n", totalCleared)

	fmt.Println("\n[4/5] Resetting employee counters (total_distance, last_oil_change, barcode/QR)...")
	res := db.Exec(`UPDATE employees SET 
		total_distance = 0, 
		last_oil_change_distance = 0,
		barcode = NULL,
		qr_code = NULL`)
	if res.Error != nil {
		fmt.Printf("  ERROR: %v\n", res.Error)
	} else {
		fmt.Printf("  Updated %d employee records\n", res.RowsAffected)
	}

	fmt.Println("\n[5/5] Running VACUUM to reclaim space...")
	rawDB.Exec("PRAGMA foreign_keys=ON")
	rawDB.Exec("VACUUM")
	fmt.Println("  VACUUM done")

	fmt.Println("\n=== FINAL STATE (kept tables) ===")
	printCounts(db, true)

	fmt.Println("\n=== FINAL STATE (cleared tables - should all be 0) ===")
	printCounts(db, false)

	fmt.Println("\n=== RESET COMPLETED SUCCESSFULLY ===")
	fmt.Println("Preserved: branches, admins (supervisors), employees")
	fmt.Println("Reset/cleared: work_sessions, audit_logs, inventory, maintenance, investigations, attendances, custody")
}

func printCounts(db *gorm.DB, kept bool) {
	type cntRow struct {
		name  string
		count int64
	}
	var tables []cntRow
	if kept {
		tables = append(tables, cntRow{"branches", countModel(db, &domain.Branch{})})
		tables = append(tables, cntRow{"admins (incl. supervisors)", countModel(db, &domain.Admin{})})
		tables = append(tables, cntRow{"employees", countModel(db, &domain.Employee{})})
	} else {
		tables = append(tables, cntRow{"work_sessions", countModel(db, &domain.WorkSession{})})
		tables = append(tables, cntRow{"audit_logs", countModel(db, &domain.AuditLog{})})
		tables = append(tables, cntRow{"inventory_items", countModel(db, &domain.InventoryItem{})})
		tables = append(tables, cntRow{"inventory_transactions", countModel(db, &domain.InventoryTransaction{})})
		tables = append(tables, cntRow{"maintenance_logs", countModel(db, &domain.MaintenanceLog{})})
		tables = append(tables, cntRow{"investigations", countModel(db, &domain.Investigation{})})
		tables = append(tables, cntRow{"attendances", countModel(db, &domain.Attendance{})})
		tables = append(tables, cntRow{"custody_days", countModel(db, &domain.CustodyDay{})})
		tables = append(tables, cntRow{"custody_expenses", countModel(db, &domain.CustodyExpense{})})
		tables = append(tables, cntRow{"custody_logs", countModel(db, &domain.CustodyLog{})})
	}
	for _, t := range tables {
		fmt.Printf("  %-30s : %d\n", t.name, t.count)
	}
}

func countModel(db *gorm.DB, m interface{}) int64 {
	var c int64
	db.Model(m).Count(&c)
	return c
}

func ensureSeedData(db *gorm.DB) {
	var bc int64
	db.Model(&domain.Branch{}).Count(&bc)
	if bc == 0 {
		branches := []domain.Branch{
			{Name: "الفرع الرئيسي"},
			{Name: "الفرع الثاني"},
		}
		for i := range branches {
			db.Create(&branches[i])
		}
		fmt.Println("  Seeded 2 default branches (tables were empty)")
	}

	var ac int64
	db.Model(&domain.Admin{}).Count(&ac)
	if ac == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("Admin@2026!"), bcrypt.DefaultCost)
		admin := domain.Admin{
			Email:    "hani@aams.com",
			Username: "hani",
			Password: string(hashed),
			Name:     "هاني",
			Role:     "ADMIN",
		}
		if err := db.Create(&admin).Error; err == nil {
			fmt.Println("  Seeded default admin: hani@aams.com / Admin@2026!")
		}
	}
}
