package ticket

import (
	"time"

	"github.com/google/uuid"
)

type CreateSupportTicketRequest struct {
	EmployeeID  *uuid.UUID `json:"employee_id"`
	ViolationID *uuid.UUID `json:"violation_id"`
	Subject     string     `json:"subject" binding:"required"`
	Category    string     `json:"category" binding:"required"`
	Priority    string     `json:"priority"`
	Description string     `json:"description" binding:"required"`
	BranchID    *uuid.UUID `json:"branch_id"`
}

type UpdateSupportTicketRequest struct {
	Subject     *string `json:"subject"`
	Category    *string `json:"category"`
	Priority    *string `json:"priority"`
	Status      *string `json:"status"`
	Description *string `json:"description"`
	Resolution  *string `json:"resolution"`
}

type SupportTicketFilter struct {
	EmployeeID    *uuid.UUID `form:"-"`
	BranchID      *uuid.UUID `form:"-"`
	EmployeeIDStr string     `form:"employee_id"`
	BranchIDStr   string     `form:"branch_id"`
	Status        string     `form:"status"`
	Category      string     `form:"category"`
	Priority      string     `form:"priority"`
	Search        string     `form:"search"`
	Page          int        `form:"page,default=1"`
	Limit         int        `form:"limit,default=50"`
	PageSize      int        `form:"page_size"`
}

type SupportTicketResponse struct {
	ID           uuid.UUID  `json:"id"`
	TicketNumber string     `json:"ticket_number"`
	EmployeeID   *uuid.UUID `json:"employee_id"`
	ViolationID  *uuid.UUID `json:"violation_id,omitempty"`
	EmployeeName string     `json:"employee_name,omitempty"`
	Subject      string     `json:"subject"`
	Category     string     `json:"category"`
	Priority     string     `json:"priority"`
	Status       string     `json:"status"`
	Description   string     `json:"description"`
	Resolution    string     `json:"resolution"`
	RepliedByName string     `json:"replied_by_name,omitempty"`
	RepliedAt     *time.Time `json:"replied_at,omitempty"`
	HistoryLog    string     `json:"history_log,omitempty"`
	BranchID      *uuid.UUID `json:"branch_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
