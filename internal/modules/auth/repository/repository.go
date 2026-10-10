package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/auth/domain"
	"delivery-backend/internal/modules/auth/dto"
)

type AdminRepository interface {
	Create(ctx context.Context, admin *domain.Admin) error
	Update(ctx context.Context, admin *domain.Admin) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error)
	FindByEmail(ctx context.Context, email string) (*domain.Admin, error)
	FindByUsername(ctx context.Context, username string) (*domain.Admin, error)
	FindByPhone(ctx context.Context, phone string) (*domain.Admin, error)
	FindByGoogleID(ctx context.Context, googleID string) (*domain.Admin, error)
	FindAll(ctx context.Context) ([]domain.Admin, error)
}

type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	FindByName(ctx context.Context, name string) (*domain.Role, error)
	FindAll(ctx context.Context) ([]domain.Role, error)
	CountUsersByRoleID(ctx context.Context, roleID uuid.UUID) (int64, error)
}

type OTPRepository interface {
	Create(ctx context.Context, otp *domain.OTPRequest) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.OTPRequest, error)
	FindByNationalID(ctx context.Context, nationalID string) (*domain.OTPRequest, error)
	FindByCode(ctx context.Context, nationalID, code string) (*domain.OTPRequest, error)
	Update(ctx context.Context, otp *domain.OTPRequest) error
	FindAll(ctx context.Context, query dto.OTPListQuery) ([]domain.OTPRequest, int64, error)
	InvalidatePrevious(ctx context.Context, nationalID string) error
}

type BranchRepository interface {
	Create(ctx context.Context, branch *domain.Branch) error
	Update(ctx context.Context, branch *domain.Branch) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error)
	FindByName(ctx context.Context, name string) (*domain.Branch, error)
	FindAll(ctx context.Context) ([]domain.Branch, error)
	CountEmployeesByBranchID(ctx context.Context, branchID uuid.UUID) (int64, error)
}

type gormAdminRepository struct{ db *gorm.DB }
type gormRoleRepository struct{ db *gorm.DB }
type gormOTPRepository struct{ db *gorm.DB }
type gormBranchRepository struct{ db *gorm.DB }

func NewAuthRepository(db *gorm.DB) (AdminRepository, RoleRepository, OTPRepository, BranchRepository) {
	return &gormAdminRepository{db: db}, &gormRoleRepository{db: db}, &gormOTPRepository{db: db}, &gormBranchRepository{db: db}
}

// ------------------------------------------------------------------
// AdminRepository Implementation
// ------------------------------------------------------------------

func (r *gormAdminRepository) Create(ctx context.Context, admin *domain.Admin) error {
	if admin.ID == uuid.Nil {
		admin.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *gormAdminRepository) Update(ctx context.Context, admin *domain.Admin) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

func (r *gormAdminRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Admin{}, "id = ?", id).Error
}

func (r *gormAdminRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error) {
	var a domain.Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").First(&a, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAdminRepository) FindByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	var a domain.Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("LOWER(email) = ?", strings.ToLower(email)).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAdminRepository) FindByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	var a domain.Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("LOWER(username) = ?", strings.ToLower(username)).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAdminRepository) FindByPhone(ctx context.Context, phone string) (*domain.Admin, error) {
	var a domain.Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("phone = ?", phone).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAdminRepository) FindByGoogleID(ctx context.Context, googleID string) (*domain.Admin, error) {
	var a domain.Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("google_id = ?", googleID).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAdminRepository) FindAll(ctx context.Context) ([]domain.Admin, error) {
	var list []domain.Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Order("created_at DESC").Find(&list).Error
	return list, err
}

// ------------------------------------------------------------------
// RoleRepository Implementation
// ------------------------------------------------------------------

func (r *gormRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *gormRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *gormRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Role{}, "id = ?", id).Error
}

func (r *gormRoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *gormRoleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *gormRoleRepository) FindAll(ctx context.Context) ([]domain.Role, error) {
	var list []domain.Role
	err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error
	return list, err
}

func (r *gormRoleRepository) CountUsersByRoleID(ctx context.Context, roleID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Admin{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}

// ------------------------------------------------------------------
// OTPRepository Implementation
// ------------------------------------------------------------------

func (r *gormOTPRepository) Create(ctx context.Context, otp *domain.OTPRequest) error {
	if otp.ID == uuid.Nil {
		otp.ID = uuid.New()
	}

	// Auto-lookup employee details by NationalID if not already set
	if otp.EmployeeID == nil || *otp.EmployeeID == uuid.Nil {
		var emp struct {
			ID         uuid.UUID
			Name       string
			BranchID   *uuid.UUID
			BranchName string
		}
		if err := r.db.WithContext(ctx).Table("employees").
			Select("employees.id, employees.name, employees.branch_id, branches.name as branch_name").
			Joins("LEFT JOIN branches ON branches.id = employees.branch_id").
			Where("employees.national_id = ?", otp.NationalID).
			First(&emp).Error; err == nil && emp.ID != uuid.Nil {
			otp.EmployeeID = &emp.ID
			if otp.EmployeeName == "" {
				otp.EmployeeName = emp.Name
			}
			if otp.BranchID == nil {
				otp.BranchID = emp.BranchID
			}
			if otp.BranchName == "" {
				otp.BranchName = emp.BranchName
			}
		}
	}

	return r.db.WithContext(ctx).Create(otp).Error
}

func (r *gormOTPRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.OTPRequest, error) {
	var otp domain.OTPRequest
	err := r.db.WithContext(ctx).First(&otp, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *gormOTPRepository) FindByNationalID(ctx context.Context, nationalID string) (*domain.OTPRequest, error) {
	var otp domain.OTPRequest
	err := r.db.WithContext(ctx).Where("national_id = ?", nationalID).Order("created_at DESC").First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *gormOTPRepository) FindByCode(ctx context.Context, nationalID, code string) (*domain.OTPRequest, error) {
	var otp domain.OTPRequest
	err := r.db.WithContext(ctx).Where("national_id = ? AND otp_code = ? AND status = ? AND expires_at > ?", nationalID, code, "pending", time.Now()).Order("created_at DESC").First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *gormOTPRepository) Update(ctx context.Context, otp *domain.OTPRequest) error {
	return r.db.WithContext(ctx).Save(otp).Error
}

func (r *gormOTPRepository) FindAll(ctx context.Context, query dto.OTPListQuery) ([]domain.OTPRequest, int64, error) {
	var list []domain.OTPRequest
	var total int64
	q := r.db.WithContext(ctx).Model(&domain.OTPRequest{})
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}
	if query.Search != "" {
		s := "%" + query.Search + "%"
		q = q.Where("national_id LIKE ? OR employee_name LIKE ?", s, s)
	}
	_ = q.Count(&total)
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	err := q.Order("created_at DESC").Offset(query.Offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *gormOTPRepository) InvalidatePrevious(ctx context.Context, nationalID string) error {
	return r.db.WithContext(ctx).Model(&domain.OTPRequest{}).Where("national_id = ? AND status = ?", nationalID, "pending").Update("status", "expired").Error
}

// ------------------------------------------------------------------
// BranchRepository Implementation
// ------------------------------------------------------------------

func (r *gormBranchRepository) Create(ctx context.Context, branch *domain.Branch) error {
	if branch.ID == uuid.Nil {
		branch.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(branch).Error
}

func (r *gormBranchRepository) Update(ctx context.Context, branch *domain.Branch) error {
	return r.db.WithContext(ctx).Save(branch).Error
}

func (r *gormBranchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Branch{}, "id = ?", id).Error
}

func (r *gormBranchRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error) {
	var b domain.Branch
	err := r.db.WithContext(ctx).First(&b, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormBranchRepository) FindByName(ctx context.Context, name string) (*domain.Branch, error) {
	var b domain.Branch
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormBranchRepository) FindAll(ctx context.Context) ([]domain.Branch, error) {
	var list []domain.Branch
	err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error
	return list, err
}

func (r *gormBranchRepository) CountEmployeesByBranchID(ctx context.Context, branchID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("employees").Where("branch_id = ? AND deleted_at IS NULL", branchID).Count(&count).Error
	return count, err
}
