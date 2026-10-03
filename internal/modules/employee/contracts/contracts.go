package contracts

import (
	"context"

	"github.com/google/uuid"
)

type EmployeeDTO struct {
	ID                    uuid.UUID  `json:"id"`
	Name                  string     `json:"name"`
	JobRole               string     `json:"job_role"`
	EmployeeNumber        string     `json:"employee_number"`
	Phone                 string     `json:"phone"`
	NationalID            string     `json:"national_id"`
	BranchID              *uuid.UUID `json:"branch_id"`
	BranchName            string     `json:"branch_name"`
	VehicleType           string     `json:"vehicle_type"`
	MotorcycleNumber      string     `json:"motorcycle_number"`
	ApplicationID         string     `json:"application_id"`
	ApplicationType       string     `json:"application_type"`
	TotalDistance         float64    `json:"total_distance"`
	LastOilChangeDistance float64    `json:"last_oil_change_distance"`
	Latitude              *float64   `json:"latitude"`
	Longitude             *float64   `json:"longitude"`
}

type IEmployeeContract interface {
	GetEmployee(ctx context.Context, id uuid.UUID) (*EmployeeDTO, error)
	FindAll(ctx context.Context, branchID *uuid.UUID) ([]EmployeeDTO, error)
	FindByID(ctx context.Context, id uuid.UUID) (*EmployeeDTO, error)
	FindByNationalID(ctx context.Context, nationalID string) (*EmployeeDTO, error)
	UpdateOnStartWork(ctx context.Context, id uuid.UUID, appID, appType, motorcycleNumber string) error
	UpdateOnEndWork(ctx context.Context, id uuid.UUID, addedDistance float64, totalOrders int) error
}
