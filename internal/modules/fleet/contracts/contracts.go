package contracts

import (
	"context"

	"github.com/google/uuid"
)

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
