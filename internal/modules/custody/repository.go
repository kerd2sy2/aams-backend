package custody

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustodyRepository interface {
	CreateDay(ctx context.Context, day *CustodyDay) error
	UpdateDay(ctx context.Context, day *CustodyDay) error
	FindDayByID(ctx context.Context, id uuid.UUID) (*CustodyDay, error)
	FindDayByDate(ctx context.Context, branchID *uuid.UUID, date string) (*CustodyDay, error)
	FindLastDay(ctx context.Context, branchID *uuid.UUID) (*CustodyDay, error)
	FindAll(ctx context.Context, branchID *uuid.UUID) ([]CustodyDay, error)
	CreateExpense(ctx context.Context, expense *CustodyExpense) error
	UpdateExpense(ctx context.Context, expense *CustodyExpense) error
	DeleteExpense(ctx context.Context, id uuid.UUID) error
	FindExpenseByID(ctx context.Context, id uuid.UUID) (*CustodyExpense, error)
	CreateLog(ctx context.Context, log *CustodyLog) error
	FindLogs(ctx context.Context, filter CustodyLogFilter) ([]CustodyLog, int64, error)
	FindLogByID(ctx context.Context, id uuid.UUID) (*CustodyLog, error)
	DeleteLog(ctx context.Context, id uuid.UUID) error
	FindDaysFrom(ctx context.Context, branchID *uuid.UUID, startDate string) ([]CustodyDay, error)
	FindPreviousDay(ctx context.Context, branchID *uuid.UUID, beforeDate string) (*CustodyDay, error)
}

type gormCustodyRepository struct {
	db *gorm.DB
}

func NewCustodyRepository(db *gorm.DB) CustodyRepository {
	return &gormCustodyRepository{db: db}
}

func (r *gormCustodyRepository) CreateDay(ctx context.Context, day *CustodyDay) error {
	return r.db.WithContext(ctx).Create(day).Error
}

func (r *gormCustodyRepository) UpdateDay(ctx context.Context, day *CustodyDay) error {
	return r.db.WithContext(ctx).Model(&CustodyDay{}).
		Where("id = ?", day.ID).
		Updates(map[string]interface{}{
			"opening_balance":  day.OpeningBalance,
			"added_amount":     day.AddedAmount,
			"closing_balance":  day.ClosingBalance,
			"ledger_image_url": day.LedgerImageURL,
		}).Error
}

func (r *gormCustodyRepository) FindDayByID(ctx context.Context, id uuid.UUID) (*CustodyDay, error) {
	var day CustodyDay
	err := r.db.WithContext(ctx).
		Preload("Expenses").
		First(&day, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &day, nil
}

func (r *gormCustodyRepository) FindDayByDate(ctx context.Context, branchID *uuid.UUID, date string) (*CustodyDay, error) {
	var day CustodyDay
	query := r.db.WithContext(ctx).Preload("Expenses").Where("date = ?", date)
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.First(&day).Error; err != nil {
		return nil, err
	}
	return &day, nil
}

func (r *gormCustodyRepository) FindLastDay(ctx context.Context, branchID *uuid.UUID) (*CustodyDay, error) {
	var day CustodyDay
	query := r.db.WithContext(ctx).Preload("Expenses")
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.Order("date DESC").First(&day).Error; err != nil {
		return nil, err
	}
	return &day, nil
}

func (r *gormCustodyRepository) FindAll(ctx context.Context, branchID *uuid.UUID) ([]CustodyDay, error) {
	var days []CustodyDay
	query := r.db.WithContext(ctx).Preload("Expenses")
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.Order("date DESC").Find(&days).Error; err != nil {
		return nil, err
	}
	return days, nil
}

func (r *gormCustodyRepository) CreateExpense(ctx context.Context, expense *CustodyExpense) error {
	return r.db.WithContext(ctx).Create(expense).Error
}

func (r *gormCustodyRepository) UpdateExpense(ctx context.Context, expense *CustodyExpense) error {
	return r.db.WithContext(ctx).Save(expense).Error
}

func (r *gormCustodyRepository) DeleteExpense(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&CustodyExpense{}, "id = ?", id).Error
}

func (r *gormCustodyRepository) FindExpenseByID(ctx context.Context, id uuid.UUID) (*CustodyExpense, error) {
	var expense CustodyExpense
	if err := r.db.WithContext(ctx).First(&expense, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *gormCustodyRepository) CreateLog(ctx context.Context, log *CustodyLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *gormCustodyRepository) FindLogs(ctx context.Context, filter CustodyLogFilter) ([]CustodyLog, int64, error) {
	var logs []CustodyLog
	var total int64

	query := r.db.WithContext(ctx).Model(&CustodyLog{})

	if filter.BranchID != "" {
		if bid, err := uuid.Parse(filter.BranchID); err == nil {
			query = query.Where("(branch_id = ? OR branch_id IS NULL)", bid)
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

func (r *gormCustodyRepository) FindLogByID(ctx context.Context, id uuid.UUID) (*CustodyLog, error) {
	var log CustodyLog
	if err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *gormCustodyRepository) DeleteLog(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&CustodyLog{}, "id = ?", id).Error
}

func (r *gormCustodyRepository) FindDaysFrom(ctx context.Context, branchID *uuid.UUID, startDate string) ([]CustodyDay, error) {
	var days []CustodyDay
	query := r.db.WithContext(ctx).Preload("Expenses").Where("date >= ?", startDate)
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.Order("date ASC").Find(&days).Error; err != nil {
		return nil, err
	}
	return days, nil
}

func (r *gormCustodyRepository) FindPreviousDay(ctx context.Context, branchID *uuid.UUID, beforeDate string) (*CustodyDay, error) {
	var day CustodyDay
	query := r.db.WithContext(ctx).Preload("Expenses").Where("date < ?", beforeDate)
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.Order("date DESC").First(&day).Error; err != nil {
		return nil, err
	}
	return &day, nil
}
