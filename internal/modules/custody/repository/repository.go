package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/custody/domain"
	"delivery-backend/internal/modules/custody/dto"
)

type CustodyRepository interface {
	CreateDay(ctx context.Context, day *domain.CustodyDay) error
	UpdateDay(ctx context.Context, day *domain.CustodyDay) error
	FindDayByID(ctx context.Context, id uuid.UUID) (*domain.CustodyDay, error)
	FindDayByDate(ctx context.Context, branchID *uuid.UUID, date string) (*domain.CustodyDay, error)
	FindLastDay(ctx context.Context, branchID *uuid.UUID) (*domain.CustodyDay, error)
	FindAll(ctx context.Context, branchID *uuid.UUID) ([]domain.CustodyDay, error)
	CreateExpense(ctx context.Context, expense *domain.CustodyExpense) error
	DeleteExpense(ctx context.Context, id uuid.UUID) error
	FindExpenseByID(ctx context.Context, id uuid.UUID) (*domain.CustodyExpense, error)
	CreateLog(ctx context.Context, log *domain.CustodyLog) error
	FindLogs(ctx context.Context, filter dto.CustodyLogFilter) ([]domain.CustodyLog, int64, error)
	FindLogByID(ctx context.Context, id uuid.UUID) (*domain.CustodyLog, error)
	DeleteLog(ctx context.Context, id uuid.UUID) error
}

type gormCustodyRepository struct {
	db *gorm.DB
}

func NewCustodyRepository(db *gorm.DB) CustodyRepository {
	return &gormCustodyRepository{db: db}
}

func (r *gormCustodyRepository) CreateDay(ctx context.Context, day *domain.CustodyDay) error {
	return r.db.WithContext(ctx).Create(day).Error
}

func (r *gormCustodyRepository) UpdateDay(ctx context.Context, day *domain.CustodyDay) error {
	return r.db.WithContext(ctx).Model(&domain.CustodyDay{}).
		Where("id = ?", day.ID).
		Updates(map[string]interface{}{
			"opening_balance": day.OpeningBalance,
			"added_amount":    day.AddedAmount,
			"closing_balance": day.ClosingBalance,
		}).Error
}

func (r *gormCustodyRepository) FindDayByID(ctx context.Context, id uuid.UUID) (*domain.CustodyDay, error) {
	var day domain.CustodyDay
	err := r.db.WithContext(ctx).
		Preload("Expenses").
		First(&day, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &day, nil
}

func (r *gormCustodyRepository) FindDayByDate(ctx context.Context, branchID *uuid.UUID, date string) (*domain.CustodyDay, error) {
	var day domain.CustodyDay
	query := r.db.WithContext(ctx).Preload("Expenses").Where("date = ?", date)
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.First(&day).Error; err != nil {
		return nil, err
	}
	return &day, nil
}

func (r *gormCustodyRepository) FindLastDay(ctx context.Context, branchID *uuid.UUID) (*domain.CustodyDay, error) {
	var day domain.CustodyDay
	query := r.db.WithContext(ctx).Preload("Expenses")
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.Order("date DESC").First(&day).Error; err != nil {
		return nil, err
	}
	return &day, nil
}

func (r *gormCustodyRepository) FindAll(ctx context.Context, branchID *uuid.UUID) ([]domain.CustodyDay, error) {
	var days []domain.CustodyDay
	query := r.db.WithContext(ctx).Preload("Expenses")
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.Order("date DESC").Find(&days).Error; err != nil {
		return nil, err
	}
	return days, nil
}

func (r *gormCustodyRepository) CreateExpense(ctx context.Context, expense *domain.CustodyExpense) error {
	return r.db.WithContext(ctx).Create(expense).Error
}

func (r *gormCustodyRepository) DeleteExpense(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.CustodyExpense{}, "id = ?", id).Error
}

func (r *gormCustodyRepository) FindExpenseByID(ctx context.Context, id uuid.UUID) (*domain.CustodyExpense, error) {
	var expense domain.CustodyExpense
	if err := r.db.WithContext(ctx).First(&expense, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *gormCustodyRepository) CreateLog(ctx context.Context, log *domain.CustodyLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *gormCustodyRepository) FindLogs(ctx context.Context, filter dto.CustodyLogFilter) ([]domain.CustodyLog, int64, error) {
	var logs []domain.CustodyLog
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.CustodyLog{})

	if filter.BranchID != "" {
		if bid, err := uuid.Parse(filter.BranchID); err == nil {
			query = query.Where("branch_id = ?", bid)
		}
	}
	if filter.Date != "" {
		query = query.Where("date = ?", filter.Date)
	}
	if filter.StartDate != "" {
		query = query.Where("date >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("date <= ?", filter.EndDate)
	}
	if filter.ActionType != "" {
		query = query.Where("action_type = ?", filter.ActionType)
	}
	if filter.CreatedBy != "" {
		searchTerm := "%" + filter.CreatedBy + "%"
		query = query.Where("admin_name LIKE ? OR admin_username LIKE ?", searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.GetEffectivePage()
	limit := filter.GetEffectiveLimit()
	offset := (page - 1) * limit

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *gormCustodyRepository) FindLogByID(ctx context.Context, id uuid.UUID) (*domain.CustodyLog, error) {
	var log domain.CustodyLog
	if err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *gormCustodyRepository) DeleteLog(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.CustodyLog{}, "id = ?", id).Error
}
