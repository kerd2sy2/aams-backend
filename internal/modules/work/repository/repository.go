package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/work/domain"
)

type WorkRepository interface {
	Create(ctx context.Context, session *domain.WorkSession) error
	Update(ctx context.Context, session *domain.WorkSession) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.WorkSession, error)
	FindActiveSessionByEmployeeID(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error)
	FindLastCompletedSession(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error)
	FindLastSessionByMotorcycle(ctx context.Context, motorcycleNumber string) (*domain.WorkSession, error)
	CountTodaySessions(ctx context.Context, empID uuid.UUID) (int64, error)
	GetActiveSessions(ctx context.Context, branchID *uuid.UUID) ([]domain.WorkSession, error)
	ExecInTx(ctx context.Context, fn func(txRepo WorkRepository) error) error
}

type gormWorkRepository struct {
	db *gorm.DB
}

func NewWorkRepository(db *gorm.DB) WorkRepository {
	return &gormWorkRepository{db: db}
}

func (r *gormWorkRepository) ExecInTx(ctx context.Context, fn func(txRepo WorkRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&gormWorkRepository{db: tx})
	})
}

func (r *gormWorkRepository) Create(ctx context.Context, session *domain.WorkSession) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *gormWorkRepository) Update(ctx context.Context, session *domain.WorkSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *gormWorkRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.WorkSession, error) {
	var session domain.WorkSession
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gormWorkRepository) FindActiveSessionByEmployeeID(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error) {
	var session domain.WorkSession
	err := r.db.WithContext(ctx).Where("employee_id = ? AND status = ?", empID, "ACTIVE").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gormWorkRepository) FindLastCompletedSession(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error) {
	var session domain.WorkSession
	err := r.db.WithContext(ctx).Where("employee_id = ? AND status = ?", empID, "COMPLETED").Order("end_time DESC").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gormWorkRepository) FindLastSessionByMotorcycle(ctx context.Context, motorcycleNumber string) (*domain.WorkSession, error) {
	var session domain.WorkSession
	err := r.db.WithContext(ctx).Where("motorcycle_number = ? AND status = ?", motorcycleNumber, "COMPLETED").Order("end_time DESC").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gormWorkRepository) CountTodaySessions(ctx context.Context, empID uuid.UUID) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.db.WithContext(ctx).Model(&domain.WorkSession{}).
		Where("employee_id = ? AND DATE(start_time) = ?", empID, today).
		Count(&count).Error
	return count, err
}

func (r *gormWorkRepository) GetActiveSessions(ctx context.Context, branchID *uuid.UUID) ([]domain.WorkSession, error) {
	var sessions []domain.WorkSession
	q := r.db.WithContext(ctx).Where("status = ?", "ACTIVE")
	if branchID != nil && *branchID != uuid.Nil {
		q = q.Joins("JOIN employees ON employees.id = work_sessions.employee_id").
			Where("employees.branch_id = ?", *branchID)
	}
	err := q.Find(&sessions).Error
	return sessions, err
}
