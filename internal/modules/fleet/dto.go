package fleet

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CreateVehicleRequest struct {
	PlateNumber       string     `json:"plate_number" binding:"required"`
	VehicleType       string     `json:"vehicle_type"` // "motorcycle" or "car"
	Brand             string     `json:"brand"`
	ModelYear         string     `json:"model_year"`
	KeyNumber         string     `json:"key_number"`
	CurrentKM         float64    `json:"current_km"`
	LastOilChangeKM   float64    `json:"last_oil_change_km"`
	IsOdometerBroken  bool       `json:"is_odometer_broken"`
	RegistrationImage string     `json:"registration_image"`
	BranchID          *uuid.UUID `json:"branch_id"`
	Notes             string     `json:"notes"`
}

type UpdateVehicleRequest struct {
	PlateNumber       *string    `json:"plate_number"`
	VehicleType       *string    `json:"vehicle_type"`
	Brand             *string    `json:"brand"`
	ModelYear         *string    `json:"model_year"`
	KeyNumber         *string    `json:"key_number"`
	CurrentKM         *float64   `json:"current_km"`
	LastOilChangeKM   *float64   `json:"last_oil_change_km"`
	LastOilChangeDate *time.Time `json:"last_oil_change_date"`
	IsOdometerBroken  *bool      `json:"is_odometer_broken"`
	RegistrationImage *string    `json:"registration_image"`
	Status            *string    `json:"status"`
	BranchID          *uuid.UUID `json:"branch_id"`
	Notes             *string    `json:"notes"`
}

type VehicleFilter struct {
	BranchID         *uuid.UUID `form:"-"`
	BranchIDStr      string     `form:"branch_id"`
	VehicleType      string     `form:"vehicle_type"`
	Status           string     `form:"status"`
	IsOdometerBroken *bool      `form:"is_odometer_broken"`
	Search           string     `form:"search"`
	Page             int        `form:"page,default=1"`
	Limit            int        `form:"limit,default=50"`
	PageSize         int        `form:"page_size"`
}

type VehicleResponse struct {
	Vehicle        Vehicle `json:"vehicle"`
	NeedsOilChange bool    `json:"needs_oil_change"`
	RemainingOilKM float64 `json:"remaining_oil_km"`
}

// VehicleSummaryDTO carries isolated vehicle summary data for other modules (Work, Maintenance, Custody)
type VehicleSummaryDTO struct {
	ID               uuid.UUID  `json:"id"`
	PlateNumber      string     `json:"plate_number"`
	VehicleType      string     `json:"vehicle_type"`
	CurrentKM        float64    `json:"current_km"`
	LastOilChangeKM  float64    `json:"last_oil_change_km"`
	IsOdometerBroken bool       `json:"is_odometer_broken"`
	Status           string     `json:"status"`
	BranchID         *uuid.UUID `json:"branch_id"`
}

// IVehicleFleetContract defines what other modules can do with Fleet without importing repository
type IVehicleFleetContract interface {
	GetVehicleByPlate(ctx context.Context, plateNumber string) (*VehicleSummaryDTO, error)
	GetVehicleByID(ctx context.Context, id uuid.UUID) (*VehicleSummaryDTO, error)
	UpdateOdometer(ctx context.Context, id uuid.UUID, newKM float64) error
	RecordOilChangeKM(ctx context.Context, id uuid.UUID, km float64) error
}

// IStorageContract for base64 image saving
type IStorageContract interface {
	SaveBase64Image(base64Str string, folder string) (string, error)
}

// IAuditContract for audit logging
type IAuditContract interface {
	LogAction(ctx context.Context, userName, action, details, ip string, branchID *uuid.UUID) error
}

// IInventoryContract for stock deduction on oil change
type IInventoryContract interface {
	DispenseOilForVehicle(ctx context.Context, plateNumber string, currentKM float64, branchID *uuid.UUID, adminName string) error
}
