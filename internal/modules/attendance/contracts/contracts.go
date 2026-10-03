package contracts

import (
	"context"

	"github.com/google/uuid"
)

type AttendanceStatusDTO struct {
	EmployeeID uuid.UUID `json:"employee_id"`
	Date       string    `json:"date"`
	Status     string    `json:"status"` // present / absent
	Note       string    `json:"note"`
}

type IAttendanceContract interface {
	GetEmployeeAttendanceStatus(ctx context.Context, empID uuid.UUID, date string) (*AttendanceStatusDTO, error)
	MarkPresent(ctx context.Context, empID uuid.UUID, date string, note string) error
}
