package main

import (
	"fmt"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	dbPath := "../delivery_local.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}

	rawDB, _ := db.DB()
	rawDB.Exec("PRAGMA foreign_keys=OFF")

	tablesToCheck := []string{
		"work_sessions",
		"audit_logs",
		"inventory_transactions",
		"inventory_items",
		"maintenance_logs",
		"investigations",
		"attendances",
		"custody_expenses",
		"custody_days",
		"custody_logs",
	}

	fmt.Println("=== COUNTS BEFORE RESET ===")
	var count int64
	for _, table := range tablesToCheck {
		db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		fmt.Printf("  %s: %d\n", table, count)
	}

	tablesInOrder := []string{
		"custody_expenses",
		"custody_logs",
		"custody_days",
		"attendances",
		"work_sessions",
		"audit_logs",
		"inventory_transactions",
		"inventory_items",
		"maintenance_logs",
		"investigations",
	}

	fmt.Println("\n=== CLEARING TABLES ===")
	for _, table := range tablesInOrder {
		result := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if result.Error != nil {
			fmt.Printf("  ERROR %s: %v\n", table, result.Error)
		} else {
			fmt.Printf("  Cleared %s: %d rows\n", table, result.RowsAffected)
		}
	}

	fmt.Println("\n=== KEPT TABLES (DATA PRESERVED) ===")
	db.Raw("SELECT COUNT(*) FROM admins").Scan(&count)
	fmt.Printf("  admins: %d\n", count)
	db.Raw("SELECT COUNT(*) FROM employees").Scan(&count)
	fmt.Printf("  employees: %d\n", count)
	db.Raw("SELECT COUNT(*) FROM branches").Scan(&count)
	fmt.Printf("  branches: %d\n", count)

	rawDB.Exec("PRAGMA foreign_keys=ON")

	fmt.Println("\n=== VACUUM ===")
	db.Exec("VACUUM")
	fmt.Println("  Done - space reclaimed")

	fmt.Println("\n=== RESET COMPLETE ===")
	fmt.Println("Kept data: branches, admins (including supervisors), employees")
	fmt.Println("Cleared data: work_sessions, audit_logs, inventory, maintenance, investigations, attendances, custody")
}
