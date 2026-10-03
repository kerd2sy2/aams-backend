package contracts

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WorkSessionDTO struct {
	ID               uuid.UUID  `json:"id"`
	EmployeeID       *uuid.UUID `json:"employee_id"`
	StartTime        time.Time  `json:"start_time"`
	EndTime          *time.Time `json:"end_time"`
	StartKM          float64    `json:"start_km"`
	EndKM            float64    `json:"end_km"`
	Distance         float64    `json:"distance"`
	OrdersCount      int        `json:"orders_count"`
	FuelCost         float64    `json:"fuel_cost"`
	ApplicationID    string     `json:"application_id"`
	ApplicationType  string     `json:"application_type"`
	VehicleType      string     `json:"vehicle_type"`
	MotorcycleNumber string     `json:"motorcycle_number"`
	Status           string     `json:"status"`
}

type IWorkContract interface {
	GetActiveSession(ctx context.Context, empID uuid.UUID) (*WorkSessionDTO, error)
	GetLastCompletedSession(ctx context.Context, empID uuid.UUID) (*WorkSessionDTO, error)
	CountTodaySessions(ctx context.Context, empID uuid.UUID) (int64, error)
}
