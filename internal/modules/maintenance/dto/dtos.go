package dto

import "github.com/google/uuid"

type CreateMaintenanceRequestRequest struct {
	VehiclePlate     string     `json:"vehicle_plate" binding:"required"`
	EmployeeID       *uuid.UUID `json:"employee_id"`
	IssueDescription string     `json:"issue_description" binding:"required"`
	Priority         string     `json:"priority"` // LOW, MEDIUM, HIGH, URGENT
	EstimatedCost    float64    `json:"estimated_cost"`
	ActualCost       float64    `json:"actual_cost"`
	WorkshopName     string     `json:"workshop_name"`
	Status           string     `json:"status"` // OPEN, IN_PROGRESS, RESOLVED, CLOSED
	BranchID         *uuid.UUID `json:"branch_id"`
	Notes            string     `json:"notes"`
}

type UpdateMaintenanceRequestRequest struct {
	VehiclePlate     *string    `json:"vehicle_plate"`
	EmployeeID       *uuid.UUID `json:"employee_id"`
	IssueDescription *string    `json:"issue_description"`
	Priority         *string    `json:"priority"`
	EstimatedCost    *float64   `json:"estimated_cost"`
	ActualCost       *float64   `json:"actual_cost"`
	WorkshopName     *string    `json:"workshop_name"`
	Status           *string    `json:"status"`
	BranchID         *uuid.UUID `json:"branch_id"`
	Notes            *string    `json:"notes"`
}

type MaintenanceRequestFilter struct {
	VehiclePlate string     `form:"vehicle_plate"`
	EmployeeID   *uuid.UUID `form:"employee_id"`
	BranchID     *uuid.UUID `form:"branch_id"`
	Status       string     `form:"status"`
	Priority     string     `form:"priority"`
	Page         int        `form:"page,default=1"`
	Limit        int        `form:"limit,default=20"`
	PageSize     int        `form:"page_size"`
}

func (f *MaintenanceRequestFilter) GetEffectiveLimit() int {
	if f.PageSize > 0 {
		return f.PageSize
	}
	if f.Limit > 0 {
		return f.Limit
	}
	return 20
}

func (f *MaintenanceRequestFilter) GetEffectivePage() int {
	if f.Page <= 0 {
		return 1
	}
	return f.Page
}
