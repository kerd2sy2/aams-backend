package contracts

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MaintenanceLogContractDTO struct {
	ID         uuid.UUID  `json:"id"`
	EmployeeID *uuid.UUID `json:"employee_id"`
	Type       string     `json:"type"`
	Details    string     `json:"details"`
	DistanceAt float64    `json:"distance_at"`
	Cost       float64    `json:"cost"`
	AdminName  string     `json:"admin_name"`
	CreatedAt  time.Time  `json:"created_at"`
}

type IMaintenanceContract interface {
	RecordMaintenanceLog(ctx context.Context, empID *uuid.UUID, maintType string, details string, distanceAt, cost float64, adminName string) error
	GetLastOilChange(ctx context.Context, empID uuid.UUID) (*MaintenanceLogContractDTO, error)
}
