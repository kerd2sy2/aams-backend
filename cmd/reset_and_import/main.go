package main

import (
	"fmt"
	"log"
	"os"

	"delivery-backend/pkg/config"
	"delivery-backend/pkg/database"
)

func main() {
	cfg := config.LoadConfig()
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	sqlBytes, err := os.ReadFile("backups/reset_and_import_2026.sql")
	if err != nil {
		// Fallback path
		sqlBytes, err = os.ReadFile("d:/AAMS/backend/backups/reset_and_import_2026.sql")
		if err != nil {
			log.Fatalf("Failed to read migration sql: %v", err)
		}
	}

	sqlStr := string(sqlBytes)
	log.Println("Executing reset and import migration...")

	if err := db.Exec(sqlStr).Error; err != nil {
		log.Fatalf("Migration execution failed: %v", err)
	}

	log.Println("Migration completed successfully! All employees imported and Admin 0500626432 preserved.")
	fmt.Println("Done.")
}
