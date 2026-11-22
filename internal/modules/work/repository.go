package work

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	
)

type WorkRepository interface {
	Create(ctx context.Context, session *WorkSession) error
	Update(ctx context.Context, session *WorkSession) error
	FindByID(ctx context.Context, id uuid.UUID) (*WorkSession, error)
	FindActiveSessionByEmployeeID(ctx context.Context, empID uuid.UUID) (*WorkSession, error)
	FindActiveSessionByEmployeeIDForUpdate(ctx context.Context, empID uuid.UUID) (*WorkSession, error)
	FindLastCompletedSession(ctx context.Context, empID uuid.UUID) (*WorkSession, error)
	FindCompletedSessionByEmployeeAndDate(ctx context.Context, empID uuid.UUID, dateStr string) (*WorkSession, error)
	FindCompletedSessionsByEmployeeAndDate(ctx context.Context, empID uuid.UUID, dateStr string) ([]WorkSession, error)
	SoftDeleteByIDs(ctx context.Context, ids []uuid.UUID) error
	FindLastSessionByMotorcycle(ctx context.Context, motorcycleNumber string) (*WorkSession, error)
	CountTodaySessions(ctx context.Context, empID uuid.UUID) (int64, error)
	GetActiveSessions(ctx context.Context, branchID *uuid.UUID) ([]WorkSession, error)
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

func (r *gormWorkRepository) Create(ctx context.Context, session *WorkSession) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *gormWorkRepository) Update(ctx context.Context, session *WorkSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *gormWorkRepository) FindByID(ctx context.Context, id uuid.UUID) (*WorkSession, error) {
	var session WorkSession
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gormWorkRepository) FindActiveSessionByEmployeeID(ctx context.Context, empID uuid.UUID) (*WorkSession, error) {
	var session WorkSession
	err := r.db.WithContext(ctx).Where("employee_id = ? AND status = ?", empID, "ACTIVE").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gormWorkRepository) FindActiveSessionByEmployeeIDForUpdate(ctx context.Context, empID uuid.UUID) (*WorkSession, error) {
	var session WorkSession
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("employee_id = ? AND status = ?", empID, "ACTIVE").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gormWorkRepository) FindLastCompletedSession(ctx context.Context, empID uuid.UUID) (*WorkSession, error) {
	var session WorkSession
	err := r.db.WithContext(ctx).Where("employee_id = ? AND status = ? AND is_reviewed = ?", empID, "COMPLETED", true).Order("end_time DESC").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func riyadhDayBounds(dateStr string) (time.Time, time.Time, error) {
	riyadh := time.FixedZone("Asia/Riyadh", 3*60*60)
	t, err := time.ParseInLocation("2006-01-02", dateStr, riyadh)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, riyadh)
	return startOfDay, startOfDay.Add(24 * time.Hour), nil
}

func (r *gormWorkRepository) FindCompletedSessionByEmployeeAndDate(ctx context.Context, empID uuid.UUID, dateStr string) (*WorkSession, error) {
	sessions, err := r.FindCompletedSessionsByEmployeeAndDate(ctx, empID, dateStr)
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &sessions[0], nil
}

func (r *gormWorkRepository) FindCompletedSessionsByEmployeeAndDate(ctx context.Context, empID uuid.UUID, dateStr string) ([]WorkSession, error) {
	var sessions []WorkSession
	q := r.db.WithContext(ctx).Where("employee_id = ? AND status = ?", empID, "COMPLETED")
	if startOfDay, endOfDay, err := riyadhDayBounds(dateStr); err == nil {
		q = q.Where("start_time >= ? AND start_time < ?", startOfDay, endOfDay)
	} else {
		q = q.Where("DATE(start_time AT TIME ZONE 'Asia/Riyadh') = ?", dateStr)
	}
	err := q.Order("start_time DESC").Find(&sessions).Error
	return sessions, err
}

func (r *gormWorkRepository) SoftDeleteByIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&WorkSession{}).Error
}

func (r *gormWorkRepository) FindLastSessionByMotorcycle(ctx context.Context, motorcycleNumber string) (*WorkSession, error) {
	var session WorkSession
	err := r.db.WithContext(ctx).Where("motorcycle_number = ? AND status = ? AND is_reviewed = ?", motorcycleNumber, "COMPLETED", true).Order("end_time DESC").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *gormWorkRepository) CountTodaySessions(ctx context.Context, empID uuid.UUID) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.db.WithContext(ctx).Model(&WorkSession{}).
		Where("employee_id = ? AND DATE(start_time) = ?", empID, today).
		Count(&count).Error
	return count, err
}

func (r *gormWorkRepository) GetActiveSessions(ctx context.Context, branchID *uuid.UUID) ([]WorkSession, error) {
	var sessions []WorkSession
	q := r.db.WithContext(ctx).Where("status = ?", "ACTIVE")
	if branchID != nil && *branchID != uuid.Nil {
		q = q.Joins("JOIN employees ON employees.id = work_sessions.employee_id").
			Where("employees.branch_id = ?", *branchID)
	}
	err := q.Find(&sessions).Error
	return sessions, err
}
