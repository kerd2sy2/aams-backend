package auth

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (r *gormAdminRepository) Create(ctx context.Context, admin *Admin) error {
	if admin.ID == uuid.Nil {
		admin.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *gormAdminRepository) Update(ctx context.Context, admin *Admin) error {
	updates := map[string]interface{}{
		"name":                   admin.Name,
		"email":                  admin.Email,
		"username":               admin.Username,
		"phone":                  admin.Phone,
		"password":               admin.Password,
		"role":                   admin.Role,
		"role_id":                admin.RoleID,
		"employee_id":            admin.EmployeeID,
		"permissions":            admin.Permissions,
		"permissions_configured": admin.PermissionsConfigured,
		"must_change_password":    admin.MustChangePassword,
		"branch_id":              admin.BranchID,
		"avatar":                 admin.Avatar,
		"google_avatar":          admin.GoogleAvatar,
		"updated_at":             time.Now(),
	}
	if err := r.db.WithContext(ctx).Model(&Admin{}).Where("id = ?", admin.ID).Updates(updates).Error; err != nil {
		return err
	}
	if admin.EmployeeID != nil {
		empUpdates := map[string]interface{}{
			"branch_id": admin.BranchID,
		}
		if admin.Avatar != "" {
			empUpdates["personal_image"] = admin.Avatar
		}
		_ = r.db.WithContext(ctx).Table("employees").Where("id = ?", *admin.EmployeeID).Updates(empUpdates).Error
	}
	return nil
}

func (r *gormAdminRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Admin{}, "id = ?", id).Error
}

func (r *gormAdminRepository) FindByID(ctx context.Context, id uuid.UUID) (*Admin, error) {
	var a Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").First(&a, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAdminRepository) FindByEmail(ctx context.Context, email string) (*Admin, error) {
	var a Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("LOWER(email) = ?", strings.ToLower(email)).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAdminRepository) FindByUsername(ctx context.Context, username string) (*Admin, error) {
	var a Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("LOWER(username) = ?", strings.ToLower(username)).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}


func (r *gormAdminRepository) FindByPhone(ctx context.Context, phone string) (*Admin, error) {
	var a Admin
	variations := NormalizeSaudiPhoneVariations(phone)
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").
		Where("phone IN (?) OR employee_id IN (SELECT id FROM employees WHERE phone IN (?) AND deleted_at IS NULL)", variations, variations).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// FindByLogin resolves an admin account in a single query across username, email, phone variations, or linked employee phone/national_id.
func (r *gormAdminRepository) FindByLogin(ctx context.Context, userKey string, phoneVariations []string) (*Admin, error) {
	var a Admin
	trimmedKey := strings.TrimSpace(userKey)
	lowerKey := strings.ToLower(trimmedKey)
	if len(phoneVariations) == 0 {
		phoneVariations = NormalizeSaudiPhoneVariations(trimmedKey)
	}
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").
		Where("LOWER(username) = ? OR LOWER(email) = ? OR phone IN (?) OR employee_id IN (SELECT id FROM employees WHERE deleted_at IS NULL AND (phone IN (?) OR national_id = ?))",
			lowerKey, lowerKey, phoneVariations, phoneVariations, trimmedKey).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAdminRepository) FindByGoogleID(ctx context.Context, googleID string) (*Admin, error) {
	var a Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("google_id = ?", googleID).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAdminRepository) FindAll(ctx context.Context) ([]Admin, error) {
	var list []Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *gormAdminRepository) FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*Admin, error) {
	var a Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("employee_id = ?", employeeID).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAdminRepository) FindEmployeeByLogin(ctx context.Context, login string) (map[string]interface{}, error) {
	var emp struct {
		ID                 uuid.UUID  `gorm:"column:id"`
		Name               string     `gorm:"column:name"`
		NationalID         string     `gorm:"column:national_id"`
		EmployeeNumber     string     `gorm:"column:employee_number"`
		JobRole            string     `gorm:"column:job_role"`
		Phone              string     `gorm:"column:phone"`
		PersonalImage      string     `gorm:"column:personal_image"`
		MotorcycleNumber   string     `gorm:"column:motorcycle_number"`
		KeyNumber          string     `gorm:"column:key_number"`
		VehicleType        string     `gorm:"column:vehicle_type"`
		Shift              string     `gorm:"column:shift"`
		PasswordHash       string     `gorm:"column:password_hash"`
		MustChangePassword bool       `gorm:"column:must_change_password"`
		BranchID           *uuid.UUID `gorm:"column:branch_id"`
		BranchName         string     `gorm:"column:branch_name"`
		Permissions        string     `gorm:"column:permissions"`
	}

	variations := NormalizeSaudiPhoneVariations(login)

	query := r.db.WithContext(ctx).Table("employees").
		Select("employees.id, employees.name, employees.national_id, employees.employee_number, employees.job_role, employees.phone, employees.personal_image, employees.motorcycle_number, employees.key_number, employees.vehicle_type, employees.shift, employees.password_hash, employees.must_change_password, employees.branch_id, branches.name as branch_name, employees.permissions").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id")

	if parsedID, err := uuid.Parse(strings.TrimSpace(login)); err == nil && parsedID != uuid.Nil {
		query = query.Where("(employees.id = ? OR employees.national_id = ? OR employees.phone IN (?) OR employees.employee_number = ?) AND employees.deleted_at IS NULL", parsedID, login, variations, login)
	} else {
		query = query.Where("(employees.national_id = ? OR employees.phone IN (?) OR employees.employee_number = ?) AND employees.deleted_at IS NULL", login, variations, login)
	}

	err := query.First(&emp).Error
	if err != nil {
		return nil, err
	}

	// Never inline oversized raw Base64 blobs into login/refresh payloads (migrated to /uploads/... URLs in background)
	personalImg := emp.PersonalImage
	if strings.HasPrefix(personalImg, "data:") && len(personalImg) > 4096 {
		personalImg = ""
	}

	res := map[string]interface{}{
		"id":                   emp.ID.String(),
		"name":                 emp.Name,
		"national_id":          emp.NationalID,
		"employee_number":      emp.EmployeeNumber,
		"job_role":             emp.JobRole,
		"phone":                emp.Phone,
		"personal_image":       personalImg,
		"motorcycle_number":    emp.MotorcycleNumber,
		"key_number":           emp.KeyNumber,
		"vehicle_type":         emp.VehicleType,
		"shift":                emp.Shift,
		"branch_name":          emp.BranchName,
		"is_employee":          true,
		"role":                 emp.JobRole,
		"password_hash":        emp.PasswordHash,
		"must_change_password": emp.MustChangePassword,
		"permissions":          emp.Permissions,
	}
	if emp.BranchID != nil {
		res["branch_id"] = emp.BranchID.String()
	}
	return res, nil
}
