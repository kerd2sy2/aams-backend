package operation_video

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OperationVideoRepository interface {
	Create(ctx context.Context, video *OperationVideo) error
	FindByID(ctx context.Context, id uuid.UUID) (*OperationVideo, error)
	List(ctx context.Context, filter OperationVideoFilter) ([]OperationVideo, int64, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindExpiredVideos(ctx context.Context, now time.Time, limit int) ([]OperationVideo, error)
	MarkDeleted(ctx context.Context, id uuid.UUID) error
}

type gormOperationVideoRepository struct {
	db *gorm.DB
}

func NewOperationVideoRepository(db *gorm.DB) OperationVideoRepository {
	return &gormOperationVideoRepository{db: db}
}

func (r *gormOperationVideoRepository) likeOp() string {
	if r.db.Dialector.Name() == "postgres" {
		return "ILIKE"
	}
	return "LIKE"
}

func (r *gormOperationVideoRepository) Create(ctx context.Context, video *OperationVideo) error {
	return r.db.WithContext(ctx).Create(video).Error
}

func (r *gormOperationVideoRepository) FindByID(ctx context.Context, id uuid.UUID) (*OperationVideo, error) {
	var video OperationVideo
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

func (r *gormOperationVideoRepository) List(ctx context.Context, filter OperationVideoFilter) ([]OperationVideo, int64, error) {
	q := r.db.WithContext(ctx).Model(&OperationVideo{}).Where("is_deleted = false")
	op := r.likeOp()

	if filter.Date != "" {
		q = q.Where("DATE(recorded_at) = ?", filter.Date)
	}

	if strings.TrimSpace(filter.EmployeeName) != "" {
		search := "%" + strings.TrimSpace(filter.EmployeeName) + "%"
		q = q.Where(fmt.Sprintf("employee_name %s ?", op), search)
	}

	if strings.TrimSpace(filter.OrderNumber) != "" {
		orderSearch := "%" + strings.TrimSpace(filter.OrderNumber) + "%"
		q = q.Where(fmt.Sprintf("order_number %s ?", op), orderSearch)
	}

	if filter.BranchID != "" {
		if bID, err := uuid.Parse(filter.BranchID); err == nil && bID != uuid.Nil {
			q = q.Where("branch_id = ?", bID)
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	var videos []OperationVideo
	err := q.Order("recorded_at DESC").Limit(limit).Offset(offset).Find(&videos).Error
	if err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

func (r *gormOperationVideoRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&OperationVideo{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(updates).Error
}

func (r *gormOperationVideoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&OperationVideo{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}

func (r *gormOperationVideoRepository) FindExpiredVideos(ctx context.Context, now time.Time, limit int) ([]OperationVideo, error) {
	var list []OperationVideo
	if limit <= 0 {
		limit = 200
	}
	err := r.db.WithContext(ctx).
		Where("is_deleted = false AND expires_at <= ?", now).
		Limit(limit).
		Find(&list).Error
	return list, err
}

func (r *gormOperationVideoRepository) MarkDeleted(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&OperationVideo{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}
