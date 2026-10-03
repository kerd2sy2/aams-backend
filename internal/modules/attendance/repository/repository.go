package repository

import (
	"context"

	"delivery-backend/internal/modules/attendance/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AttendanceRepository interface {
	Upsert(ctx context.Context, attendance *domain.Attendance) error
	FindByDate(ctx context.Context, date string) ([]domain.Attendance, error)
	FindByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date string) (*domain.Attendance, error)
	DeleteByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date string) error
}

type gormAttendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &gormAttendanceRepository{db: db}
}

func (r *gormAttendanceRepository) Upsert(ctx context.Context, attendance *domain.Attendance) error {
	if attendance.ID == uuid.Nil {
		attendance.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "employee_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "note", "updated_at"}),
	}).Create(attendance).Error
}

func (r *gormAttendanceRepository) FindByDate(ctx context.Context, date string) ([]domain.Attendance, error) {
	var records []domain.Attendance
	err := r.db.WithContext(ctx).Where("date = ?", date).Find(&records).Error
	return records, err
}

func (r *gormAttendanceRepository) FindByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date string) (*domain.Attendance, error) {
	var record domain.Attendance
	err := r.db.WithContext(ctx).Where("employee_id = ? AND date = ?", employeeID, date).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *gormAttendanceRepository) DeleteByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date string) error {
	return r.db.WithContext(ctx).Where("employee_id = ? AND date = ?", employeeID, date).Delete(&domain.Attendance{}).Error
}
