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

type gormAuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) (AdminRepository, RoleRepository, OTPRepository, BranchRepository) {
	r := &gormAuthRepository{db: db}
	return r, r, r, r
}

// AdminRepository
func (r *gormAuthRepository) Create(ctx context.Context, admin *domain.Admin) error {
	if admin.ID == uuid.Nil {
		admin.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *gormAuthRepository) Update(ctx context.Context, admin *domain.Admin) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

func (r *gormAuthRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Admin{}, "id = ?", id).Error
}

func (r *gormAuthRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error) {
	var a domain.Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").First(&a, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAuthRepository) FindByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	var a domain.Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("LOWER(email) = ?", strings.ToLower(email)).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAuthRepository) FindByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	var a domain.Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("LOWER(username) = ?", strings.ToLower(username)).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAuthRepository) FindByPhone(ctx context.Context, phone string) (*domain.Admin, error) {
	var a domain.Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("phone = ?", phone).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAuthRepository) FindByGoogleID(ctx context.Context, googleID string) (*domain.Admin, error) {
	var a domain.Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("google_id = ?", googleID).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *gormAuthRepository) FindAll(ctx context.Context) ([]domain.Admin, error) {
	var list []domain.Admin
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Order("created_at DESC").Find(&list).Error
	return list, err
}

// RoleRepository
func (r *gormAuthRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *gormAuthRepository) UpdateRole(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *gormAuthRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Role{}, "id = ?", id).Error
}

func (r *gormAuthRepository) FindRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *gormAuthRepository) FindRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *gormAuthRepository) FindAllRoles(ctx context.Context) ([]domain.Role, error) {
	var list []domain.Role
	err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error
	return list, err
}

func (r *gormAuthRepository) CountUsersByRoleID(ctx context.Context, roleID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Admin{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}

// RoleRepository alias wrappers
func (r *gormAuthRepository) CreateRoleDirect(ctx context.Context, role *domain.Role) error {
	return r.CreateRole(ctx, role)
}

func (r *gormAuthRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	return r.FindRoleByName(ctx, name)
}

// OTPRepository
func (r *gormAuthRepository) CreateOTP(ctx context.Context, otp *domain.OTPRequest) error {
	if otp.ID == uuid.Nil {
		otp.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(otp).Error
}

func (r *gormAuthRepository) FindOTPByID(ctx context.Context, id uuid.UUID) (*domain.OTPRequest, error) {
	var otp domain.OTPRequest
	err := r.db.WithContext(ctx).First(&otp, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *gormAuthRepository) FindByNationalID(ctx context.Context, nationalID string) (*domain.OTPRequest, error) {
	var otp domain.OTPRequest
	err := r.db.WithContext(ctx).Where("national_id = ?", nationalID).Order("created_at DESC").First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *gormAuthRepository) FindByCode(ctx context.Context, nationalID, code string) (*domain.OTPRequest, error) {
	var otp domain.OTPRequest
	err := r.db.WithContext(ctx).Where("national_id = ? AND otp_code = ? AND status = ? AND expires_at > ?", nationalID, code, "pending", time.Now()).Order("created_at DESC").First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *gormAuthRepository) UpdateOTP(ctx context.Context, otp *domain.OTPRequest) error {
	return r.db.WithContext(ctx).Save(otp).Error
}

func (r *gormAuthRepository) FindAllOTPs(ctx context.Context, query dto.OTPListQuery) ([]domain.OTPRequest, int64, error) {
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

func (r *gormAuthRepository) InvalidatePrevious(ctx context.Context, nationalID string) error {
	return r.db.WithContext(ctx).Model(&domain.OTPRequest{}).Where("national_id = ? AND status = ?", nationalID, "pending").Update("status", "expired").Error
}

// BranchRepository
func (r *gormAuthRepository) CreateBranch(ctx context.Context, branch *domain.Branch) error {
	if branch.ID == uuid.Nil {
		branch.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(branch).Error
}

func (r *gormAuthRepository) UpdateBranch(ctx context.Context, branch *domain.Branch) error {
	return r.db.WithContext(ctx).Save(branch).Error
}

func (r *gormAuthRepository) DeleteBranch(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Branch{}, "id = ?", id).Error
}

func (r *gormAuthRepository) FindBranchByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error) {
	var b domain.Branch
	err := r.db.WithContext(ctx).First(&b, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormAuthRepository) FindBranchByName(ctx context.Context, name string) (*domain.Branch, error) {
	var b domain.Branch
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormAuthRepository) FindAllBranches(ctx context.Context) ([]domain.Branch, error) {
	var list []domain.Branch
	err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error
	return list, err
}

func (r *gormAuthRepository) CountEmployeesByBranchID(ctx context.Context, branchID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("employees").Where("branch_id = ? AND deleted_at IS NULL", branchID).Count(&count).Error
	return count, err
}
