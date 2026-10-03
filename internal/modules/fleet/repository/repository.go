package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/fleet/domain"
	"delivery-backend/internal/modules/fleet/dto"
)

type VehicleRepository interface {
	Create(ctx context.Context, vehicle *domain.Vehicle) error
	Update(ctx context.Context, vehicle *domain.Vehicle) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Vehicle, error)
	FindByPlateNumber(ctx context.Context, plateNumber string) (*domain.Vehicle, error)
	FindByPlateNumberUnscoped(ctx context.Context, plateNumber string) (*domain.Vehicle, error)
	RestoreVehicle(ctx context.Context, vehicle *domain.Vehicle) error
	FindAll(ctx context.Context, filter dto.VehicleFilter) ([]domain.Vehicle, int64, error)
	CountAll(ctx context.Context, branchID *uuid.UUID) (int64, error)
	UpdateOdometer(ctx context.Context, plateNumber string, newKM float64, distance float64) error
	RecordOilChange(ctx context.Context, id uuid.UUID, currentKM float64) error
	FindLatestVehicleKM(ctx context.Context, plateNumber string) (float64, error)
}

type gormVehicleRepository struct {
	db *gorm.DB
}

func NewVehicleRepository(db *gorm.DB) VehicleRepository {
	return &gormVehicleRepository{db: db}
}

func (r *gormVehicleRepository) Create(ctx context.Context, vehicle *domain.Vehicle) error {
	return r.db.WithContext(ctx).Create(vehicle).Error
}

func (r *gormVehicleRepository) Update(ctx context.Context, vehicle *domain.Vehicle) error {
	return r.db.WithContext(ctx).Save(vehicle).Error
}

func (r *gormVehicleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Vehicle{}, "id = ?", id).Error
}

func (r *gormVehicleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Vehicle, error) {
	var vehicle domain.Vehicle
	err := r.db.WithContext(ctx).First(&vehicle, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func (r *gormVehicleRepository) FindByPlateNumber(ctx context.Context, plateNumber string) (*domain.Vehicle, error) {
	clean := strings.TrimSpace(plateNumber)
	if clean == "" {
		return nil, errors.New("رقم اللوحة فارغ")
	}
	var vehicle domain.Vehicle
	// 1. Exact match
	if err := r.db.WithContext(ctx).Where("TRIM(plate_number) = ?", clean).First(&vehicle).Error; err == nil {
		return &vehicle, nil
	}
	// 2. Prefix match
	if err := r.db.WithContext(ctx).Where("plate_number ILIKE ?", clean+"%").First(&vehicle).Error; err == nil {
		return &vehicle, nil
	}
	// 3. Substring match
	if err := r.db.WithContext(ctx).Where("plate_number ILIKE ?", "%"+clean+"%").First(&vehicle).Error; err == nil {
		return &vehicle, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *gormVehicleRepository) FindByPlateNumberUnscoped(ctx context.Context, plateNumber string) (*domain.Vehicle, error) {
	clean := strings.TrimSpace(plateNumber)
	if clean == "" {
		return nil, errors.New("رقم اللوحة فارغ")
	}
	var vehicle domain.Vehicle
	if err := r.db.WithContext(ctx).Unscoped().Where("TRIM(plate_number) = ? OR plate_number ILIKE ? OR plate_number ILIKE ?", clean, clean+"%", "%"+clean+"%").First(&vehicle).Error; err == nil {
		return &vehicle, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *gormVehicleRepository) RestoreVehicle(ctx context.Context, vehicle *domain.Vehicle) error {
	return r.db.WithContext(ctx).Unscoped().Model(vehicle).Updates(map[string]interface{}{
		"deleted_at":         nil,
		"brand":              vehicle.Brand,
		"model_year":         vehicle.ModelYear,
		"key_number":         vehicle.KeyNumber,
		"current_km":         vehicle.CurrentKM,
		"last_oil_change_km": vehicle.LastOilChangeKM,
		"status":             vehicle.Status,
		"is_odometer_broken": vehicle.IsOdometerBroken,
		"branch_id":          vehicle.BranchID,
		"notes":              vehicle.Notes,
		"vehicle_type":       vehicle.VehicleType,
	}).Error
}

func (r *gormVehicleRepository) FindAll(ctx context.Context, filter dto.VehicleFilter) ([]domain.Vehicle, int64, error) {
	var vehicles []domain.Vehicle
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Vehicle{})

	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.VehicleType != "" {
		query = query.Where("vehicle_type = ?", filter.VehicleType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("plate_number ILIKE ? OR brand ILIKE ? OR key_number ILIKE ? OR notes ILIKE ?", search, search, search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 && filter.PageSize > 0 {
		limit = filter.PageSize
	}
	if limit <= 0 {
		limit = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	err := query.Order("plate_number ASC").Offset(offset).Limit(limit).Find(&vehicles).Error
	if err != nil {
		return nil, 0, err
	}

	return vehicles, total, nil
}

func (r *gormVehicleRepository) CountAll(ctx context.Context, branchID *uuid.UUID) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Vehicle{})
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *gormVehicleRepository) UpdateOdometer(ctx context.Context, plateNumber string, newKM float64, distance float64) error {
	var vehicle domain.Vehicle
	err := r.db.WithContext(ctx).First(&vehicle, "plate_number = ?", plateNumber).Error
	if err != nil {
		newVehicle := domain.Vehicle{
			PlateNumber:     plateNumber,
			VehicleType:     "motorcycle",
			CurrentKM:       newKM,
			LastOilChangeKM: 0,
			TotalDistance:   distance,
			Status:          domain.VehicleStatusAvailable,
		}
		return r.db.WithContext(ctx).Create(&newVehicle).Error
	}

	updates := map[string]interface{}{
		"current_km":     newKM,
		"total_distance": gorm.Expr("total_distance + ?", distance),
		"status":         domain.VehicleStatusAvailable,
		"updated_at":     time.Now(),
	}
	return r.db.WithContext(ctx).Model(&vehicle).Updates(updates).Error
}

func (r *gormVehicleRepository) RecordOilChange(ctx context.Context, id uuid.UUID, currentKM float64) error {
	return r.db.WithContext(ctx).Model(&domain.Vehicle{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_oil_change_km": currentKM,
		"updated_at":         time.Now(),
	}).Error
}

func (r *gormVehicleRepository) FindLatestVehicleKM(ctx context.Context, plateNumber string) (float64, error) {
	cleanPlate := strings.TrimSpace(plateNumber)
	if cleanPlate == "" {
		return 0, errors.New("رقم اللوحة فارغ")
	}

	var vehicle domain.Vehicle
	if err := r.db.WithContext(ctx).Where("plate_number = ?", cleanPlate).First(&vehicle).Error; err == nil {
		return vehicle.CurrentKM, nil
	}
	return 0, nil
}
