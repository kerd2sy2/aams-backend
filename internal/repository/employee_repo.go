package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/domain"
)

type gormEmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &gormEmployeeRepository{db: db}
}

func (r *gormEmployeeRepository) Create(ctx context.Context, emp *domain.Employee) error {
	return r.db.WithContext(ctx).Create(emp).Error
}

func (r *gormEmployeeRepository) Update(ctx context.Context, emp *domain.Employee) error {
	return r.db.WithContext(ctx).Save(emp).Error
}

func (r *gormEmployeeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Employee{}, "id = ?", id).Error
}

func (r *gormEmployeeRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	var emp domain.Employee
	if err := r.db.WithContext(ctx).Preload("Branch").First(&emp, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *gormEmployeeRepository) FindByNationalID(ctx context.Context, nationalID string) (*domain.Employee, error) {
	var emp domain.Employee
	if err := r.db.WithContext(ctx).First(&emp, "national_id = ?", nationalID).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *gormEmployeeRepository) FindBySearchTerm(ctx context.Context, term string, branchID *uuid.UUID) ([]domain.Employee, error) {
	var employees []domain.Employee
	term = strings.TrimSpace(term)
	if term == "" {
		return employees, nil
	}

	query := r.db.WithContext(ctx).Preload("Branch")
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}

	// Only return couriers/drivers for work sessions (exclude supervisors and management)
	query = query.Where("job_role = ? OR job_role = '' OR job_role IS NULL", "DRIVER")

	// Try UUID parse first
	parsedUUID, err := uuid.Parse(term)
	if err == nil {
		query.Where("id = ?", parsedUUID).Find(&employees)
		if len(employees) > 0 {
			return employees, nil
		}
	}

	likeTerm := "%" + term + "%"
	err = query.Where(
		"name LIKE ? OR employee_number LIKE ? OR national_id LIKE ? OR motorcycle_number LIKE ? OR key_number LIKE ? OR application_id LIKE ?",
		likeTerm, likeTerm, likeTerm, likeTerm, likeTerm, likeTerm,
	).Limit(15).Find(&employees).Error
	return employees, err
}

func (r *gormEmployeeRepository) CountAll(ctx context.Context, branchID *uuid.UUID) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Employee{})
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *gormEmployeeRepository) UpdateLocation(ctx context.Context, id uuid.UUID, lat, lng float64, isVPN, isMock, outOfZone bool) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Employee{}).Where("id = ?", id).Updates(map[string]interface{}{
		"latitude":         lat,
		"longitude":        lng,
		"last_location_at": now,
		"is_vpn":           isVPN,
		"is_mock_location": isMock,
		"out_of_zone":      outOfZone,
	}).Error
}

func (r *gormEmployeeRepository) GetLocations(ctx context.Context, branchID *uuid.UUID) ([]domain.Employee, error) {
	var employees []domain.Employee
	query := r.db.WithContext(ctx).Model(&domain.Employee{}).Preload("Branch")
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	err := query.Order("name ASC").Find(&employees).Error
	return employees, err
}

// GORM Work Repository