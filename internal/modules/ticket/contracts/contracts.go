package contracts

import (
	"context"

	"github.com/google/uuid"
)

// EmployeeSummaryDTO carries isolated basic information about an employee
type EmployeeSummaryDTO struct {
	ID       uuid.UUID  `json:"id"`
	Name     string     `json:"name"`
	Phone    string     `json:"phone"`
	BranchID *uuid.UUID `json:"branch_id"`
}

// IEmployeeContract defines what Ticket module can query about employees without importing employee repo or GORM
type IEmployeeContract interface {
	GetEmployeeSummary(ctx context.Context, employeeID uuid.UUID) (*EmployeeSummaryDTO, error)
}

// PushNotificationDTO carries isolated payload for push notifications
type PushNotificationDTO struct {
	TargetUserID uuid.UUID
	Title        string
	Body         string
	Data         map[string]string
}

// INotificationContract defines async notification capabilities
type INotificationContract interface {
	SendAsyncNotification(ctx context.Context, notif PushNotificationDTO)
}
