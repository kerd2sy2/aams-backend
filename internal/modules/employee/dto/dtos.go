package dto

import (
	"github.com/google/uuid"
)

type CreateEmployeeRequest struct {
	Name                     string     `json:"name" binding:"required"`
	JobRole                  string     `json:"job_role"`
	EmployeeNumber           string     `json:"employee_number"`
	Phone                    string     `json:"phone"`
	PersonalImage            string     `json:"personal_image"`
	NationalID               string     `json:"national_id" binding:"required"`
	IqamaExpirationDate      *string    `json:"iqama_expiration_date"`
	NationalIDImage          string     `json:"national_id_image"`
	DrivingLicenseImage      string     `json:"driving_license_image"`
	PassportImage            string     `json:"passport_image"`
	VehicleRegistrationImage string     `json:"vehicle_registration_image"`
	KeyNumber                string     `json:"key_number"`
	MotorcycleNumber         string     `json:"motorcycle_number"`
	ApplicationID            string     `json:"application_id"`
	ApplicationType          string     `json:"application_type"`
	VehicleType              string     `json:"vehicle_type"`
	Shift                    string     `json:"shift"`
	BranchID                 *uuid.UUID `json:"branch_id"`
}

type UpdateEmployeeRequest struct {
	Name                     string     `json:"name"`
	JobRole                  string     `json:"job_role"`
	EmployeeNumber           string     `json:"employee_number"`
	Phone                    string     `json:"phone"`
	PersonalImage            string     `json:"personal_image"`
	NationalID               string     `json:"national_id"`
	IqamaExpirationDate      *string    `json:"iqama_expiration_date"`
	NationalIDImage          string     `json:"national_id_image"`
	DrivingLicenseImage      string     `json:"driving_license_image"`
	PassportImage            string     `json:"passport_image"`
	VehicleRegistrationImage string     `json:"vehicle_registration_image"`
	KeyNumber                string     `json:"key_number"`
	MotorcycleNumber         string     `json:"motorcycle_number"`
	ApplicationID            string     `json:"application_id"`
	ApplicationType          string     `json:"application_type"`
	VehicleType              string     `json:"vehicle_type"`
	Shift                    string     `json:"shift"`
	BranchID                 *uuid.UUID `json:"branch_id"`
}

type SetPhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type UpdateLocationRequest struct {
	Latitude       float64  `json:"latitude" binding:"required"`
	Longitude      float64  `json:"longitude" binding:"required"`
	Speed          *float64 `json:"speed"`
	Heading        *float64 `json:"heading"`
	IsVPN          *bool    `json:"is_vpn"`
	IsMockLocation *bool    `json:"is_mock_location"`
}

type EmployeeLocationDTO struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	JobRole            string   `json:"job_role"`
	EmployeeNumber     string   `json:"employee_number"`
	Phone              string   `json:"phone"`
	PersonalImage      string   `json:"personal_image"`
	NationalID         string   `json:"national_id"`
	KeyNumber          string   `json:"key_number"`
	MotorcycleNumber   string   `json:"motorcycle_number"`
	ApplicationType    string   `json:"application_type"`
	Shift              string   `json:"shift"`
	BranchID           *string  `json:"branch_id"`
	BranchName         string   `json:"branch_name"`
	Latitude           *float64 `json:"latitude"`
	Longitude          *float64 `json:"longitude"`
	LastLocationAt     *string  `json:"last_location_at"`
	IsShiftActive      bool     `json:"is_shift_active"`
	ActiveSessionID    *string  `json:"active_session_id"`
	IsVPN              bool     `json:"is_vpn"`
	IsMockLocation     bool     `json:"is_mock_location"`
	OutOfZone          bool     `json:"out_of_zone"`
	DistanceFromTaifKm *float64 `json:"distance_from_taif_km"`
}

type EmployeeFilter struct {
	Search          string     `form:"search"`
	ApplicationID   string     `form:"application_id"`
	ApplicationType string     `form:"application_type"`
	BranchID        *uuid.UUID `form:"branch_id"`
	Page            int        `form:"page,default=1"`
	Limit           int        `form:"limit,default=10"`
	SortBy          string     `form:"sort_by,default=created_at"`
	Order           string     `form:"order,default=desc"`
}

type PaginatedEmployeeResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

type OilSetupEntry struct {
	EmployeeID            string  `json:"employee_id" binding:"required,uuid"`
	LastOilChangeDistance float64 `json:"last_oil_change_distance"`
}

type BatchOilSetupRequest struct {
	Entries []OilSetupEntry `json:"entries" binding:"required,dive"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
