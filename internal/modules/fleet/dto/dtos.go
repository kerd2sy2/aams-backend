package dto

import (
	"github.com/google/uuid"

	"delivery-backend/internal/modules/fleet/domain"
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
	IsOdometerBroken  *bool      `json:"is_odometer_broken"`
	RegistrationImage *string    `json:"registration_image"`
	Status            *string    `json:"status"`
	BranchID          *uuid.UUID `json:"branch_id"`
	Notes             *string    `json:"notes"`
}

type VehicleFilter struct {
	BranchID         *uuid.UUID `form:"branch_id"`
	VehicleType      string     `form:"vehicle_type"`
	Status           string     `form:"status"`
	IsOdometerBroken *bool      `form:"is_odometer_broken"`
	Search           string     `form:"search"`
	Page             int        `form:"page,default=1"`
	Limit            int        `form:"limit,default=50"`
	PageSize         int        `form:"page_size"`
}

type VehicleResponse struct {
	Vehicle        domain.Vehicle `json:"vehicle"`
	NeedsOilChange bool           `json:"needs_oil_change"`
	RemainingOilKM float64        `json:"remaining_oil_km"`
}
