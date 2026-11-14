package repository

import (
	"context"

	"delivery-backend/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gormRoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &gormRoleRepository{db: db}
}

func (r *gormRoleRepository) FindAll(ctx context.Context) ([]domain.Role, error) {
	var roles []domain.Role
	if err := r.db.WithContext(ctx).Order("is_system DESC, created_at ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *gormRoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *gormRoleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.WithContext(ctx).Where("LOWER(name) = LOWER(?)", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *gormRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *gormRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *gormRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Role{}, "id = ?", id).Error
}

func (r *gormRoleRepository) CountUsersByRoleID(ctx context.Context, roleID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Admin{}).Where("role_id = ?", roleID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

type gormAdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &gormAdminRepository{db: db}
}

func (r *gormAdminRepository) FindByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("LOWER(email) = LOWER(?)", email).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").First(&admin, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("employee_id = ?", employeeID).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) FindByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("LOWER(username) = LOWER(?)", username).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) FindByPhone(ctx context.Context, phone string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("phone = ?", phone).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) FindByLogin(ctx context.Context, login string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where(
		"LOWER(email) = LOWER(?) OR LOWER(username) = LOWER(?) OR phone = ?",
		login, login, login,
	).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) Create(ctx context.Context, admin *domain.Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *gormAdminRepository) Update(ctx context.Context, admin *domain.Admin) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

func (r *gormAdminRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Admin{}, "id = ?", id).Error
}

func (r *gormAdminRepository) FindByGoogleEmail(ctx context.Context, email string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("LOWER(google_email) = LOWER(?)", email).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) FindByGoogleID(ctx context.Context, googleID string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("google_id = ?", googleID).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *gormAdminRepository) UpdateGoogleLink(ctx context.Context, adminID uuid.UUID, googleEmail, googleID, googleAvatar string) error {
	return r.db.WithContext(ctx).Model(&domain.Admin{}).Where("id = ?", adminID).Updates(map[string]interface{}{
		"google_email":  googleEmail,
		"google_id":     googleID,
		"google_avatar": googleAvatar,
	}).Error
}

func (r *gormAdminRepository) UnlinkGoogle(ctx context.Context, adminID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Admin{}).Where("id = ?", adminID).Updates(map[string]interface{}{
		"google_email":  "",
		"google_id":     "",
		"google_avatar": "",
	}).Error
}

func (r *gormAdminRepository) FindAll(ctx context.Context) ([]domain.Admin, error) {
	var admins []domain.Admin
	if err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Order("created_at DESC").Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

// GORM Employee Repository