package auth

import (
	"context"

	"github.com/google/uuid"
)

func (r *gormRoleRepository) Create(ctx context.Context, role *Role) error {
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *gormRoleRepository) Update(ctx context.Context, role *Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *gormRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Role{}, "id = ?", id).Error
}

func (r *gormRoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*Role, error) {
	var role Role
	err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *gormRoleRepository) FindByName(ctx context.Context, name string) (*Role, error) {
	var role Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *gormRoleRepository) FindAll(ctx context.Context) ([]Role, error) {
	var list []Role
	err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error
	return list, err
}

func (r *gormRoleRepository) CountUsersByRoleID(ctx context.Context, roleID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Admin{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}

func (r *gormBranchRepository) Create(ctx context.Context, branch *Branch) error {
	if branch.ID == uuid.Nil {
		branch.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(branch).Error
}

func (r *gormBranchRepository) Update(ctx context.Context, branch *Branch) error {
	return r.db.WithContext(ctx).Save(branch).Error
}

func (r *gormBranchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Branch{}, "id = ?", id).Error
}

func (r *gormBranchRepository) FindByID(ctx context.Context, id uuid.UUID) (*Branch, error) {
	var b Branch
	err := r.db.WithContext(ctx).First(&b, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormBranchRepository) FindByName(ctx context.Context, name string) (*Branch, error) {
	var b Branch
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormBranchRepository) FindAll(ctx context.Context) ([]Branch, error) {
	var list []Branch
	err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error
	return list, err
}

func (r *gormBranchRepository) CountEmployeesByBranchID(ctx context.Context, branchID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("employees").Where("branch_id = ? AND deleted_at IS NULL", branchID).Count(&count).Error
	return count, err
}
