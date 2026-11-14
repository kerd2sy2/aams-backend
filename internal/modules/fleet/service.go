package fleet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type VehicleService interface {
	Create(ctx context.Context, req CreateVehicleRequest) (*Vehicle, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateVehicleRequest) (*Vehicle, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*VehicleResponse, error)
	GetByPlateNumber(ctx context.Context, plateNumber string) (*Vehicle, error)
	GetAll(ctx context.Context, filter VehicleFilter) ([]VehicleResponse, int64, error)
	GetLatestKM(ctx context.Context, plateNumber string) (float64, error)
	RecordOilChange(ctx context.Context, id uuid.UUID, adminName string) error

	// Contract implementations
	GetVehicleByPlate(ctx context.Context, plateNumber string) (*VehicleSummaryDTO, error)
	GetVehicleByID(ctx context.Context, id uuid.UUID) (*VehicleSummaryDTO, error)
	UpdateOdometer(ctx context.Context, id uuid.UUID, newKM float64) error
	RecordOilChangeKM(ctx context.Context, id uuid.UUID, km float64) error
}

type vehicleService struct {
	vehicleRepo       VehicleRepository
	storageService    IStorageContract
	inventoryContract IInventoryContract
}

func NewVehicleService(vehicleRepo VehicleRepository, storageService IStorageContract, inventoryContract IInventoryContract) VehicleService {
	return &vehicleService{
		vehicleRepo:       vehicleRepo,
		storageService:    storageService,
		inventoryContract: inventoryContract,
	}
}

func oilChangeInterval(vehicleType string) float64 {
	if strings.EqualFold(vehicleType, "car") {
		return 10000
	}
	return 950 // motorcycle default
}

func (s *vehicleService) Create(ctx context.Context, req CreateVehicleRequest) (*Vehicle, error) {
	plate := strings.TrimSpace(req.PlateNumber)
	if plate == "" {
		return nil, ErrPlateRequired
	}

	vType := req.VehicleType
	if vType == "" {
		vType = "motorcycle"
	}

	regImage := strings.TrimSpace(req.RegistrationImage)
	if s.storageService != nil && strings.HasPrefix(regImage, "data:image") {
		if savedUrl, err := s.storageService.SaveBase64Image(regImage, "registration"); err == nil && savedUrl != "" {
			regImage = savedUrl
		}
	}

	existingUnscoped, _ := s.vehicleRepo.FindByPlateNumberUnscoped(ctx, plate)
	if existingUnscoped != nil {
		if existingUnscoped.DeletedAt.Valid {
			existingUnscoped.Brand = req.Brand
			existingUnscoped.ModelYear = req.ModelYear
			existingUnscoped.KeyNumber = req.KeyNumber
			existingUnscoped.VehicleType = vType
			existingUnscoped.Status = VehicleStatusAvailable
			existingUnscoped.BranchID = req.BranchID
			existingUnscoped.Notes = req.Notes
			if regImage != "" {
				existingUnscoped.RegistrationImage = regImage
			}
			if req.CurrentKM > 0 {
				existingUnscoped.CurrentKM = req.CurrentKM
			}
			if req.LastOilChangeKM > 0 {
				existingUnscoped.LastOilChangeKM = req.LastOilChangeKM
			}
			if err := s.vehicleRepo.RestoreVehicle(ctx, existingUnscoped); err != nil {
				return nil, fmt.Errorf("فشل استعادة المركبة: %w", err)
			}
			return existingUnscoped, nil
		}

		if existingUnscoped.BranchID == nil && req.BranchID != nil {
			existingUnscoped.Brand = req.Brand
			existingUnscoped.ModelYear = req.ModelYear
			existingUnscoped.KeyNumber = req.KeyNumber
			existingUnscoped.VehicleType = vType
			existingUnscoped.Status = VehicleStatusAvailable
			existingUnscoped.BranchID = req.BranchID
			existingUnscoped.Notes = req.Notes
			if regImage != "" {
				existingUnscoped.RegistrationImage = regImage
			}
			if req.CurrentKM > 0 {
				existingUnscoped.CurrentKM = req.CurrentKM
			}
			if req.LastOilChangeKM > 0 {
				existingUnscoped.LastOilChangeKM = req.LastOilChangeKM
			}
			if err := s.vehicleRepo.Update(ctx, existingUnscoped); err != nil {
				return nil, fmt.Errorf("فشل تحديث المركبة: %w", err)
			}
			return existingUnscoped, nil
		}

		return nil, fmt.Errorf("المركبة برقم اللوحة %s مسجلة بالفعل", plate)
	}

	lastOilKM := req.LastOilChangeKM
	if lastOilKM <= 0 && req.CurrentKM > 0 {
		lastOilKM = req.CurrentKM
	}

	vehicle := &Vehicle{
		ID:                uuid.New(),
		PlateNumber:       plate,
		VehicleType:       vType,
		Brand:             req.Brand,
		ModelYear:         req.ModelYear,
		KeyNumber:         req.KeyNumber,
		CurrentKM:         req.CurrentKM,
		LastOilChangeKM:   lastOilKM,
		IsOdometerBroken:  req.IsOdometerBroken,
		RegistrationImage: regImage,
		TotalDistance:     0,
		Status:            VehicleStatusAvailable,
		BranchID:          req.BranchID,
		Notes:             req.Notes,
	}

	if err := s.vehicleRepo.Create(ctx, vehicle); err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (s *vehicleService) Update(ctx context.Context, id uuid.UUID, req UpdateVehicleRequest) (*Vehicle, error) {
	vehicle, err := s.vehicleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrVehicleNotFound
	}

	if req.PlateNumber != nil {
		newPlate := strings.TrimSpace(*req.PlateNumber)
		if newPlate != "" && newPlate != vehicle.PlateNumber {
			existing, _ := s.vehicleRepo.FindByPlateNumber(ctx, newPlate)
			if existing != nil && existing.ID != id {
				return nil, ErrPlateAlreadyUsed
			}
			vehicle.PlateNumber = newPlate
		}
	}
	if req.VehicleType != nil {
		vehicle.VehicleType = *req.VehicleType
	}
	if req.Brand != nil {
		vehicle.Brand = *req.Brand
	}
	if req.ModelYear != nil {
		vehicle.ModelYear = *req.ModelYear
	}
	if req.KeyNumber != nil {
		vehicle.KeyNumber = *req.KeyNumber
	}
	if req.CurrentKM != nil {
		vehicle.CurrentKM = *req.CurrentKM
	}
	if req.LastOilChangeKM != nil {
		vehicle.LastOilChangeKM = *req.LastOilChangeKM
	}
	if req.LastOilChangeDate != nil {
		vehicle.LastOilChangeDate = req.LastOilChangeDate
	}
	if req.IsOdometerBroken != nil {
		vehicle.IsOdometerBroken = *req.IsOdometerBroken
	}
	if req.RegistrationImage != nil {
		regImage := strings.TrimSpace(*req.RegistrationImage)
		if s.storageService != nil && strings.HasPrefix(regImage, "data:image") {
			if savedUrl, err := s.storageService.SaveBase64Image(regImage, "registration"); err == nil && savedUrl != "" {
				regImage = savedUrl
			}
		}
		vehicle.RegistrationImage = regImage
	}
	if req.Status != nil {
		vehicle.Status = *req.Status
	}
	if req.BranchID != nil {
		vehicle.BranchID = req.BranchID
	}
	if req.Notes != nil {
		vehicle.Notes = *req.Notes
	}

	if err := s.vehicleRepo.Update(ctx, vehicle); err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (s *vehicleService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.vehicleRepo.Delete(ctx, id)
}

func isVehicleBrokenOdometer(plate string, broken bool) bool {
	if broken {
		return true
	}
	clean := strings.TrimSpace(plate)
	return strings.Contains(clean, "7569") || strings.Contains(clean, "8043")
}

func calculateVehicleOilStatus(v *Vehicle) (needsOil bool, remainingKM float64) {
	isBroken := isVehicleBrokenOdometer(v.PlateNumber, v.IsOdometerBroken)
	if isBroken {
		if v.LastOilChangeDate != nil {
			days := time.Since(*v.LastOilChangeDate).Hours() / 24.0
			if days >= 5.0 {
				return true, 0
			}
			remDays := 5.0 - days
			return false, remDays * 200.0
		}
		return true, 0
	}

	interval := oilChangeInterval(v.VehicleType)
	var drivenSinceOil float64
	if v.LastOilChangeKM > 0 || v.LastOilChangeDate != nil {
		drivenSinceOil = v.CurrentKM - v.LastOilChangeKM
		if drivenSinceOil < 0 {
			drivenSinceOil = 0
		}
	} else if v.TotalDistance > 0 {
		drivenSinceOil = v.TotalDistance
	} else {
		drivenSinceOil = 0
	}
	remaining := interval - drivenSinceOil
	if remaining < 0 {
		remaining = 0
	}
	return drivenSinceOil >= interval, remaining
}

func (s *vehicleService) GetByID(ctx context.Context, id uuid.UUID) (*VehicleResponse, error) {
	vehicle, err := s.vehicleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrVehicleNotFound
	}

	needsOil, remaining := calculateVehicleOilStatus(vehicle)

	return &VehicleResponse{
		Vehicle:        *vehicle,
		NeedsOilChange: needsOil,
		RemainingOilKM: remaining,
	}, nil
}

func (s *vehicleService) GetByPlateNumber(ctx context.Context, plateNumber string) (*Vehicle, error) {
	return s.vehicleRepo.FindByPlateNumber(ctx, plateNumber)
}

func (s *vehicleService) GetAll(ctx context.Context, filter VehicleFilter) ([]VehicleResponse, int64, error) {
	vehicles, total, err := s.vehicleRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]VehicleResponse, len(vehicles))
	for i, v := range vehicles {
		needsOil, remaining := calculateVehicleOilStatus(&v)

		responses[i] = VehicleResponse{
			Vehicle:        v,
			NeedsOilChange: needsOil,
			RemainingOilKM: remaining,
		}
	}

	return responses, total, nil
}

func (s *vehicleService) GetLatestKM(ctx context.Context, plateNumber string) (float64, error) {
	return s.vehicleRepo.FindLatestVehicleKM(ctx, plateNumber)
}

func (s *vehicleService) RecordOilChange(ctx context.Context, id uuid.UUID, adminName string) error {
	vehicle, err := s.vehicleRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("المركبة غير موجودة")
	}

	if s.inventoryContract != nil {
		if err := s.inventoryContract.DispenseOilForVehicle(ctx, vehicle.PlateNumber, vehicle.CurrentKM, vehicle.BranchID, adminName); err != nil {
			return err
		}
	}

	return s.vehicleRepo.RecordOilChange(ctx, id, vehicle.CurrentKM)
}

// ---------------------------------------------------------
// Contract Implementation for other modules
// ---------------------------------------------------------

func (s *vehicleService) GetVehicleByPlate(ctx context.Context, plateNumber string) (*VehicleSummaryDTO, error) {
	v, err := s.vehicleRepo.FindByPlateNumber(ctx, plateNumber)
	if err != nil {
		return nil, err
	}
	return &VehicleSummaryDTO{
		ID:               v.ID,
		PlateNumber:      v.PlateNumber,
		VehicleType:      v.VehicleType,
		CurrentKM:        v.CurrentKM,
		LastOilChangeKM:  v.LastOilChangeKM,
		IsOdometerBroken: v.IsOdometerBroken,
		Status:           v.Status,
		BranchID:         v.BranchID,
	}, nil
}

func (s *vehicleService) GetVehicleByID(ctx context.Context, id uuid.UUID) (*VehicleSummaryDTO, error) {
	v, err := s.vehicleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &VehicleSummaryDTO{
		ID:               v.ID,
		PlateNumber:      v.PlateNumber,
		VehicleType:      v.VehicleType,
		CurrentKM:        v.CurrentKM,
		LastOilChangeKM:  v.LastOilChangeKM,
		IsOdometerBroken: v.IsOdometerBroken,
		Status:           v.Status,
		BranchID:         v.BranchID,
	}, nil
}

func (s *vehicleService) UpdateOdometer(ctx context.Context, id uuid.UUID, newKM float64) error {
	v, err := s.vehicleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	v.CurrentKM = newKM
	return s.vehicleRepo.Update(ctx, v)
}

func (s *vehicleService) RecordOilChangeKM(ctx context.Context, id uuid.UUID, km float64) error {
	return s.vehicleRepo.RecordOilChange(ctx, id, km)
}
