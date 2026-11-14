package maintenance

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MaintenanceRepository interface {
	CreateLog(ctx context.Context, log *MaintenanceLog) error
	FindByEmployeeID(ctx context.Context, empID uuid.UUID, limit int) ([]MaintenanceLog, error)
	FindAllLogs(ctx context.Context, page, limit int) ([]MaintenanceLog, int64, error)
	FindLastOilChange(ctx context.Context, empID uuid.UUID) (*MaintenanceLog, error)

	CreateRequest(ctx context.Context, req *MaintenanceRequest) error
	UpdateRequest(ctx context.Context, req *MaintenanceRequest) error
	DeleteRequest(ctx context.Context, id uuid.UUID) error
	FindRequestByID(ctx context.Context, id uuid.UUID) (*MaintenanceRequest, error)
	FindAllRequests(ctx context.Context, filter MaintenanceRequestFilter) ([]MaintenanceRequest, int64, error)
}

type gormMaintenanceRepository struct {
	db *gorm.DB
}

func NewMaintenanceRepository(db *gorm.DB) MaintenanceRepository {
	return &gormMaintenanceRepository{db: db}
}

func (r *gormMaintenanceRepository) CreateLog(ctx context.Context, log *MaintenanceLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *gormMaintenanceRepository) FindByEmployeeID(ctx context.Context, empID uuid.UUID, limit int) ([]MaintenanceLog, error) {
	var logs []MaintenanceLog
	q := r.db.WithContext(ctx).Where("employee_id = ?", empID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&logs).Error
	return logs, err
}

func (r *gormMaintenanceRepository) FindAllLogs(ctx context.Context, page, limit int) ([]MaintenanceLog, int64, error) {
	var logs []MaintenanceLog
	var total int64

	db := r.db.WithContext(ctx).Model(&MaintenanceLog{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (r *gormMaintenanceRepository) FindLastOilChange(ctx context.Context, empID uuid.UUID) (*MaintenanceLog, error) {
	var log MaintenanceLog
	err := r.db.WithContext(ctx).
		Where("employee_id = ? AND type = ?", empID, "oil_change").
		Order("created_at DESC").
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *gormMaintenanceRepository) CreateRequest(ctx context.Context, req *MaintenanceRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *gormMaintenanceRepository) UpdateRequest(ctx context.Context, req *MaintenanceRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *gormMaintenanceRepository) DeleteRequest(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&MaintenanceRequest{}, "id = ?", id).Error
}

func (r *gormMaintenanceRepository) FindRequestByID(ctx context.Context, id uuid.UUID) (*MaintenanceRequest, error) {
	var m MaintenanceRequest
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *gormMaintenanceRepository) FindAllRequests(ctx context.Context, filter MaintenanceRequestFilter) ([]MaintenanceRequest, int64, error) {
	var requests []MaintenanceRequest
	var total int64

	db := r.db.WithContext(ctx).Model(&MaintenanceRequest{})

	if filter.VehiclePlate != "" {
		db = db.Where("vehicle_plate = ?", filter.VehiclePlate)
	}
	if filter.EmployeeID != nil {
		db = db.Where("employee_id = ?", *filter.EmployeeID)
	}
	if filter.BranchID != nil {
		db = db.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.Priority != "" {
		db = db.Where("priority = ?", filter.Priority)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.GetEffectivePage()
	limit := filter.GetEffectiveLimit()
	offset := (page - 1) * limit

	err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&requests).Error
	return requests, total, err
}
