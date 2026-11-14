package system

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------------- Audit Logs ----------------

func (r *gormSystemRepository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *gormSystemRepository) FindAuditLogs(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&AuditLog{})
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (r *gormSystemRepository) GetLatestAuditLogs(ctx context.Context, limit int, branchID *uuid.UUID) ([]AuditLog, error) {
	var logs []AuditLog
	query := r.db.WithContext(ctx).Model(&AuditLog{})
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if limit <= 0 {
		limit = 10
	}
	err := query.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *gormSystemRepository) DeleteAuditLog(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&AuditLog{}, "id = ?", id).Error
}

func (r *gormSystemRepository) BulkDeleteAuditLogs(ctx context.Context, ids []uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&AuditLog{}, "id IN ?", ids).Error
}

func (r *gormSystemRepository) ClearAuditLogs(ctx context.Context, branchID *uuid.UUID) error {
	query := r.db.WithContext(ctx)
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	} else {
		query = query.Where("1 = 1")
	}
	return query.Delete(&AuditLog{}).Error
}

// ---------------- Settings ----------------

func (r *gormSystemRepository) GetSettings(ctx context.Context) ([]AppSetting, error) {
	var settings []AppSetting
	err := r.db.WithContext(ctx).Find(&settings).Error
	return settings, err
}

func (r *gormSystemRepository) GetSettingByKey(ctx context.Context, key string) (*AppSetting, error) {
	var setting AppSetting
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *gormSystemRepository) UpsertSetting(ctx context.Context, setting *AppSetting) error {
	var existing AppSetting
	err := r.db.WithContext(ctx).Where("key = ?", setting.Key).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.db.WithContext(ctx).Create(setting).Error
		}
		return err
	}
	existing.Value = setting.Value
	return r.db.WithContext(ctx).Save(&existing).Error
}
