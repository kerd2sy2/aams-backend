package database

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"delivery-backend/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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
			Name:        "OWNER",
			DisplayName: "مالك المؤسسة",
			Description: "كامل الصلاحيات للإشراف العام والتقارير والعمليات المالية والإدارية",
			Permissions: `["*"]`,
			IsSystem:    true,
		},
		{
			Name:        "SUPERVISOR",
			DisplayName: "مشرف وردية",
			Description: "متابعة المناديب، تسجيل الدوام، العهدة اليومية، فحص الزيت ومتابعة العمليات الميدانية",
			Permissions: `["employees.view","employees.create","employees.edit","employees.cards","work.view","work.start","work.end","custody","custody.access","custody.view","custody.add","vehicles.view","vehicles.oil","fuel.view","fuel.manage","maintenance.view","maintenance.manage","attendance.view","investigations.view","investigations.create","inventory.view","inventory.dispense","tickets.view","tickets.manage"]`,
			IsSystem:    true,
		},
		{
			Name:        "ACCOUNTANT",
			DisplayName: "إدارة مالية ومحاسبة",
			Description: "إدارة العهدة، المصروفات، الحسابات البنكية، الوقود، المخالفات والتقارير المالية",
			Permissions: `["custody","custody.access","custody.view","custody.add","custody.delete","bank_accounts.view","bank_accounts.manage","fuel.view","fuel.manage","violations.view","violations.manage","reports.view","reports.export","inventory.view","investigations.view"]`,
			IsSystem:    true,
		},
		{
			Name:        "CUSTODY",
			DisplayName: "مسؤول العهدة والمصروفات",
			Description: "إدارة وصرف العهدة اليومية وتوثيق الفواتير والمصروفات",
			Permissions: `["custody","custody.access","custody.view","custody.add","custody.delete","bank_accounts.view","reports.view"]`,
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

	// Ensure primary system admin 0500626432 exists with SUPER_ADMIN role
	var admin0500 domain.Admin
	var superRole domain.Role
	var roleID *uuid.UUID
	if err := db.Where("name = ?", "SUPER_ADMIN").First(&superRole).Error; err == nil {
		roleID = &superRole.ID
	}

	if err := db.Unscoped().Where("phone = ? OR username = ?", "0500626432", "0500626432").First(&admin0500).Error; err == nil {
		db.Unscoped().Model(&domain.Admin{}).Where("id = ?", admin0500.ID).Updates(map[string]interface{}{
			"role":       "SUPER_ADMIN",
			"role_id":    roleID,
			"branch_id":  nil,
			"deleted_at": nil,
		})
	} else {
		defaultAdminPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
		if defaultAdminPassword == "" {
			defaultAdminPassword = "Admin@2026!"
		}
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
		newAdmin := domain.Admin{
			Phone:    "0500626432",
			Username: "0500626432",
			Email:    "admin0500626432@aams.com",
			Name:     "المدير العام",
			Password: string(hashedPassword),
			Role:     "SUPER_ADMIN",
			RoleID:   roleID,
		}
		if err := db.Create(&newAdmin).Error; err == nil {
			log.Println("[SEED SUCCESS] Admin 0500626432 created with SUPER_ADMIN role")
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

// sanitizeWorkSessionLegacyImages converts any large base64 image strings stored in work_sessions
// into local static files on disk (/uploads/...) to keep the database lightweight and fast.
func sanitizeWorkSessionLegacyImages(db *gorm.DB) {
	_ = os.MkdirAll("uploads/plates", 0755)
	_ = os.MkdirAll("uploads/odometer", 0755)
	_ = os.MkdirAll("uploads/custody", 0755)
	_ = os.MkdirAll("uploads/employees", 0755)

	type legacySession struct {
		ID              uuid.UUID `gorm:"column:id"`
		StartKMImage    string    `gorm:"column:start_km_image"`
		StartPlateImage string    `gorm:"column:start_plate_image"`
		EndKMImage      string    `gorm:"column:end_km_image"`
	}

	saveB64ToFile := func(b64, category string) string {
		if !strings.HasPrefix(b64, "data:image") {
			return b64
		}
		idx := strings.Index(b64, ",")
		if idx == -1 {
			return b64
		}
		prefix := b64[:idx]
		ext := ".jpg"
		if strings.Contains(prefix, "png") {
			ext = ".png"
		} else if strings.Contains(prefix, "webp") {
			ext = ".webp"
		}
		raw := b64[idx+1:]
		data, decErr := base64.StdEncoding.DecodeString(raw)
		if decErr != nil {
			return b64
		}
		filename := fmt.Sprintf("%s_%d%s", category, time.Now().UnixNano(), ext)
		dstPath := filepath.Join("uploads", category, filename)
		if writeErr := os.WriteFile(dstPath, data, 0644); writeErr == nil {
			return fmt.Sprintf("/uploads/%s/%s", category, filename)
		}
		return b64
	}

	var rows []legacySession
	err := db.Table("work_sessions").
		Where("start_plate_image LIKE 'data:image%' OR start_km_image LIKE 'data:image%' OR end_km_image LIKE 'data:image%'").
		Limit(250).
		Find(&rows).Error
	if err == nil && len(rows) > 0 {
		cleanedCount := 0
		for _, r := range rows {
			updates := make(map[string]interface{})
			if strings.HasPrefix(r.StartPlateImage, "data:image") {
				if newPath := saveB64ToFile(r.StartPlateImage, "plates"); newPath != r.StartPlateImage {
					updates["start_plate_image"] = newPath
				}
			}
			if strings.HasPrefix(r.StartKMImage, "data:image") {
				if newPath := saveB64ToFile(r.StartKMImage, "odometer"); newPath != r.StartKMImage {
					updates["start_km_image"] = newPath
				}
			}
			if strings.HasPrefix(r.EndKMImage, "data:image") {
				if newPath := saveB64ToFile(r.EndKMImage, "odometer"); newPath != r.EndKMImage {
					updates["end_km_image"] = newPath
				}
			}
			if len(updates) > 0 {
				if updateErr := db.Table("work_sessions").Where("id = ?", r.ID).Updates(updates).Error; updateErr == nil {
					cleanedCount++
				}
			}
		}
		if cleanedCount > 0 {
			log.Printf("[Sanitize] Cleaned %d work_sessions with legacy base64 images into disk files", cleanedCount)
		}
	}

	sanitizeColumn := func(table, col, category string) {
		type b64Row struct {
			ID  uuid.UUID `gorm:"column:id"`
			Val string    `gorm:"column:val"`
		}
		var targets []b64Row
		qErr := db.Table(table).
			Select("id, "+col+" as val").
			Where(col+" LIKE 'data:image%'").
			Limit(200).
			Find(&targets).Error
		if qErr != nil || len(targets) == 0 {
			return
		}
		fixed := 0
		for _, t := range targets {
			if newPath := saveB64ToFile(t.Val, category); newPath != t.Val {
				if uErr := db.Table(table).Where("id = ?", t.ID).Update(col, newPath).Error; uErr == nil {
					fixed++
				}
			}
		}
		if fixed > 0 {
			log.Printf("[Sanitize] Migrated %d base64 images in %s.%s to disk files", fixed, table, col)
		}
	}

	sanitizeColumn("custody_expenses", "image_url", "custody")
	sanitizeColumn("custody_days", "ledger_image_url", "custody")
	sanitizeColumn("custody_logs", "image_url", "custody")
	sanitizeColumn("employees", "personal_image", "employees")
}
