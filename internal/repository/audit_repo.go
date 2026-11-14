package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/domain"
)

type gormAuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &gormAuditRepository{db: db}
}

func (r *gormAuditRepository) CreateLog(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *gormAuditRepository) GetLatestLogs(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	err := r.db.WithContext(ctx).Order("created_at desc").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *gormAuditRepository) GetLatestLogsByBranch(ctx context.Context, branchID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	err := r.db.WithContext(ctx).
		Where("branch_id = ?", branchID).
		Order("created_at desc").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *gormAuditRepository) FindAll(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AuditLog{})
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (r *gormAuditRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.AuditLog{}, "id = ?", id).Error
}

func (r *gormAuditRepository) DeleteBulk(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Delete(&domain.AuditLog{}, "id IN ?", ids).Error
}

func (r *gormAuditRepository) ClearAll(ctx context.Context, branchID *uuid.UUID) error {
	query := r.db.WithContext(ctx)
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	} else {
		query = query.Where("1 = 1")
	}
	return query.Delete(&domain.AuditLog{}).Error
}

// GORM Branch Repository