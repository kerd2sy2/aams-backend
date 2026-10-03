package contracts

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type LeaveCheckDTO struct {
	EmployeeID uuid.UUID `json:"employee_id"`
	IsOnLeave  bool      `json:"is_on_leave"`
	LeaveType  string    `json:"leave_type"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
}

type IHRLegalContract interface {
	CheckEmployeeOnLeave(ctx context.Context, empID uuid.UUID, date time.Time) (*LeaveCheckDTO, error)
	GetPendingRequestsCount(ctx context.Context) (investigations int64, leaves int64, err error)
}
