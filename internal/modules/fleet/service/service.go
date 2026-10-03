package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/fleet/contracts"
	"delivery-backend/internal/modules/fleet/domain"
	"delivery-backend/internal/modules/fleet/dto"
	"delivery-backend/internal/modules/fleet/repository"
)

type VehicleService interface {
	Create(ctx context.Context, req dto.CreateVehicleRequest) (*domain.Vehicle, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateVehicleRequest) (*domain.Vehicle, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*dto.VehicleResponse, error)
	GetByPlateNumber(ctx context.Context, plateNumber string) (*domain.Vehicle, error)
	GetAll(ctx context.Context, filter dto.VehicleFilter) ([]dto.VehicleResponse, int64, error)
	GetLatestKM(ctx context.Context, plateNumber string) (float64, error)
	RecordOilChange(ctx context.Context, id uuid.UUID) error

	// Contract implementations
	GetVehicleByPlate(ctx context.Context, plateNumber string) (*contracts.VehicleSummaryDTO, error)
	GetVehicleByID(ctx context.Context, id uuid.UUID) (*contracts.VehicleSummaryDTO, error)
	UpdateOdometer(ctx context.Context, id uuid.UUID, newKM float64) error
	RecordOilChangeKM(ctx context.Context, id uuid.UUID, km float64) error
}

type vehicleService struct {
	vehicleRepo    repository.VehicleRepository
	storageService contracts.IStorageContract
}

func NewVehicleService(vehicleRepo repository.VehicleRepository, storageService contracts.IStorageContract) VehicleService {
	return &vehicleService{vehicleRepo: vehicleRepo, storageService: storageService}
}

func oilChangeInterval(vehicleType string) float64 {
	if strings.EqualFold(vehicleType, "car") {
		return 10000
	}
	return 950 // motorcycle default
}

func (s *vehicleService) Create(ctx context.Context, req dto.CreateVehicleRequest) (*domain.Vehicle, error) {
	plate := strings.TrimSpace(req.PlateNumber)
	if plate == "" {
		return nil, errors.New("رقم اللوحة مطلوب")
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
			existingUnscoped.Status = domain.VehicleStatusAvailable
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
			existingUnscoped.Status = domain.VehicleStatusAvailable
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

	vehicle := &domain.Vehicle{
		ID:                uuid.New(),
		PlateNumber:       plate,
		VehicleType:       vType,
		Brand:             req.Brand,
		ModelYear:         req.ModelYear,
		KeyNumber:         req.KeyNumber,
		CurrentKM:         req.CurrentKM,
		LastOilChangeKM:   req.LastOilChangeKM,
		IsOdometerBroken:  req.IsOdometerBroken,
		RegistrationImage: regImage,
		TotalDistance:     0,
		Status:            domain.VehicleStatusAvailable,
		BranchID:          req.BranchID,
		Notes:             req.Notes,
	}

	if err := s.vehicleRepo.Create(ctx, vehicle); err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (s *vehicleService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateVehicleRequest) (*domain.Vehicle, error) {
	vehicle, err := s.vehicleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("المركبة غير موجودة")
	}

	if req.PlateNumber != nil {
		vehicle.PlateNumber = strings.TrimSpace(*req.PlateNumber)
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

func (s *vehicleService) GetByID(ctx context.Context, id uuid.UUID) (*dto.VehicleResponse, error) {
	vehicle, err := s.vehicleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("المركبة غير موجودة")
	}

	interval := oilChangeInterval(vehicle.VehicleType)
	drivenSinceOil := vehicle.CurrentKM - vehicle.LastOilChangeKM
	if drivenSinceOil < 0 {
		drivenSinceOil = 0
	}
	remaining := interval - drivenSinceOil
	if remaining < 0 {
		remaining = 0
	}

	return &dto.VehicleResponse{
		Vehicle:        *vehicle,
		NeedsOilChange: drivenSinceOil >= interval,
		RemainingOilKM: remaining,
	}, nil
}

func (s *vehicleService) GetByPlateNumber(ctx context.Context, plateNumber string) (*domain.Vehicle, error) {
	return s.vehicleRepo.FindByPlateNumber(ctx, plateNumber)
}

func (s *vehicleService) GetAll(ctx context.Context, filter dto.VehicleFilter) ([]dto.VehicleResponse, int64, error) {
	vehicles, total, err := s.vehicleRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.VehicleResponse, len(vehicles))
	for i, v := range vehicles {
		interval := oilChangeInterval(v.VehicleType)
		drivenSinceOil := v.CurrentKM - v.LastOilChangeKM
		if drivenSinceOil < 0 {
			drivenSinceOil = 0
		}
		remaining := interval - drivenSinceOil
		if remaining < 0 {
			remaining = 0
		}

		responses[i] = dto.VehicleResponse{
			Vehicle:        v,
			NeedsOilChange: drivenSinceOil >= interval,
			RemainingOilKM: remaining,
		}
	}

	return responses, total, nil
}

func (s *vehicleService) GetLatestKM(ctx context.Context, plateNumber string) (float64, error) {
	return s.vehicleRepo.FindLatestVehicleKM(ctx, plateNumber)
}

func (s *vehicleService) RecordOilChange(ctx context.Context, id uuid.UUID) error {
	vehicle, err := s.vehicleRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("المركبة غير موجودة")
	}
	return s.vehicleRepo.RecordOilChange(ctx, id, vehicle.CurrentKM)
}

// ---------------------------------------------------------
// Contract Implementation for other modules
// ---------------------------------------------------------

func (s *vehicleService) GetVehicleByPlate(ctx context.Context, plateNumber string) (*contracts.VehicleSummaryDTO, error) {
	v, err := s.vehicleRepo.FindByPlateNumber(ctx, plateNumber)
	if err != nil {
		return nil, err
	}
	return &contracts.VehicleSummaryDTO{
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

func (s *vehicleService) GetVehicleByID(ctx context.Context, id uuid.UUID) (*contracts.VehicleSummaryDTO, error) {
	v, err := s.vehicleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &contracts.VehicleSummaryDTO{
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
