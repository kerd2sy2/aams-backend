package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/employee/domain"
	"delivery-backend/internal/modules/employee/dto"
)

type EmployeeRepository interface {
	Create(ctx context.Context, emp *domain.Employee) error
	Update(ctx context.Context, emp *domain.Employee) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error)
	FindByNationalID(ctx context.Context, nationalID string) (*domain.Employee, error)
	FindAll(ctx context.Context, filter dto.EmployeeFilter) ([]domain.Employee, int64, error)
	Search(ctx context.Context, query string, branchID *uuid.UUID) ([]domain.Employee, error)
	GetWorkingEmployees(ctx context.Context, branchID *uuid.UUID) ([]domain.Employee, error)
	GetLocations(ctx context.Context, branchID *uuid.UUID) ([]dto.EmployeeLocationDTO, error)
	UpdateLocation(ctx context.Context, id uuid.UUID, lat, lng float64, isVPN, isMock, outOfZone bool) error
	UpdatePhone(ctx context.Context, id uuid.UUID, phone string) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	BatchSetOilChange(ctx context.Context, entries []dto.OilSetupEntry) (int, error)
}

type gormEmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &gormEmployeeRepository{db: db}
}

func (r *gormEmployeeRepository) Create(ctx context.Context, emp *domain.Employee) error {
	if emp.ID == uuid.Nil {
		emp.ID = uuid.New()
	}
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
	if err := r.db.WithContext(ctx).First(&emp, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *gormEmployeeRepository) FindByNationalID(ctx context.Context, nationalID string) (*domain.Employee, error) {
	var emp domain.Employee
	if err := r.db.WithContext(ctx).Where("national_id = ?", nationalID).First(&emp).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *gormEmployeeRepository) FindAll(ctx context.Context, filter dto.EmployeeFilter) ([]domain.Employee, int64, error) {
	var emps []domain.Employee
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Employee{})
	if filter.Search != "" {
		s := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(name) LIKE ? OR national_id LIKE ? OR phone LIKE ? OR employee_number LIKE ?", s, s, s, s)
	}
	if filter.ApplicationID != "" {
		q = q.Where("application_id = ?", filter.ApplicationID)
	}
	if filter.ApplicationType != "" {
		q = q.Where("application_type = ?", filter.ApplicationType)
	}
	if filter.BranchID != nil && *filter.BranchID != uuid.Nil {
		q = q.Where("branch_id = ?", *filter.BranchID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	orderClause := "name ASC"
	if filter.SortBy != "" {
		orderClause = filter.SortBy
		if filter.Order != "" {
			orderClause += " " + filter.Order
		}
	}

	err := q.Order(orderClause).Offset(offset).Limit(filter.Limit).Find(&emps).Error
	return emps, total, err
}

func (r *gormEmployeeRepository) Search(ctx context.Context, query string, branchID *uuid.UUID) ([]domain.Employee, error) {
	var emps []domain.Employee
	s := "%" + strings.ToLower(query) + "%"
	q := r.db.WithContext(ctx).Where("LOWER(name) LIKE ? OR national_id LIKE ? OR phone LIKE ? OR motorcycle_number LIKE ?", s, s, s, s)
	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("branch_id = ?", *branchID)
	}
	err := q.Limit(50).Find(&emps).Error
	return emps, err
}

func (r *gormEmployeeRepository) GetWorkingEmployees(ctx context.Context, branchID *uuid.UUID) ([]domain.Employee, error) {
	var emps []domain.Employee
	q := r.db.WithContext(ctx).Where("is_working = ?", true)
	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("branch_id = ?", *branchID)
	}
	err := q.Find(&emps).Error
	return emps, err
}

func (r *gormEmployeeRepository) GetLocations(ctx context.Context, branchID *uuid.UUID) ([]dto.EmployeeLocationDTO, error) {
	var results []dto.EmployeeLocationDTO
	q := r.db.WithContext(ctx).Table("employees").
		Select("employees.id, employees.name, employees.job_role, employees.employee_number, employees.phone, employees.personal_image, employees.national_id, employees.key_number, employees.motorcycle_number, employees.application_type, employees.shift, employees.branch_id, branches.name as branch_name, employees.latitude, employees.longitude, employees.last_location_at, employees.is_working as is_shift_active, employees.is_vpn, employees.is_mock_location, employees.out_of_zone").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id").
		Where("employees.deleted_at IS NULL AND employees.latitude IS NOT NULL AND employees.longitude IS NOT NULL")

	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("employees.branch_id = ?", *branchID)
	}

	err := q.Scan(&results).Error
	return results, err
}

func (r *gormEmployeeRepository) UpdateLocation(ctx context.Context, id uuid.UUID, lat, lng float64, isVPN, isMock, outOfZone bool) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Employee{}).Where("id = ?", id).Updates(map[string]interface{}{
		"latitude":         lat,
		"longitude":        lng,
		"last_location_at": &now,
		"is_vpn":           isVPN,
		"is_mock_location": isMock,
		"out_of_zone":      outOfZone,
	}).Error
}

func (r *gormEmployeeRepository) UpdatePhone(ctx context.Context, id uuid.UUID, phone string) error {
	return r.db.WithContext(ctx).Model(&domain.Employee{}).Where("id = ?", id).Update("phone", phone).Error
}

func (r *gormEmployeeRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.db.WithContext(ctx).Model(&domain.Employee{}).Where("id = ?", id).Update("password_hash", passwordHash).Error
}

func (r *gormEmployeeRepository) BatchSetOilChange(ctx context.Context, entries []dto.OilSetupEntry) (int, error) {
	updated := 0
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, e := range entries {
			id, err := uuid.Parse(e.EmployeeID)
			if err != nil {
				continue
			}
			if err := tx.Model(&domain.Employee{}).Where("id = ?", id).Update("last_oil_change_distance", e.LastOilChangeDistance).Error; err == nil {
				updated++
			}
		}
		return nil
	})
	return updated, err
}
