package dto

import (
	"time"

	"github.com/google/uuid"
)

// 1. Investigation
type CreateInvestigationRequest struct {
	EmployeeID     string   `json:"employee_id" binding:"required,uuid"`
	Type           string   `json:"type"`
	Questions      []string `json:"questions"`
	Answers        []string `json:"answers"`
	ReportText     string   `json:"report_text"`
	Images         []string `json:"images"`
	Amount         *float64 `json:"amount"`
	StartDate      string   `json:"start_date"`
	EndDate        string   `json:"end_date"`
	Items          []string `json:"items"`
	IsGuilty       bool     `json:"is_guilty"`
	Notes          string   `json:"notes"`
	DeductionMonth string   `json:"deduction_month"`
}

type UpdateInvestigationRequest struct {
	EmployeeID     string   `json:"employee_id" binding:"required,uuid"`
	Type           string   `json:"type"`
	Questions      []string `json:"questions"`
	Answers        []string `json:"answers"`
	ReportText     string   `json:"report_text"`
	Images         []string `json:"images"`
	Amount         *float64 `json:"amount"`
	StartDate      string   `json:"start_date"`
	EndDate        string   `json:"end_date"`
	Items          []string `json:"items"`
	IsGuilty       bool     `json:"is_guilty"`
	Notes          string   `json:"notes"`
	DeductionMonth string   `json:"deduction_month"`
}

type InvestigationResponse struct {
	ID                 uuid.UUID  `json:"id"`
	EmployeeID         uuid.UUID  `json:"employee_id"`
	EmployeeName       string     `json:"employee_name"`
	NationalID         string     `json:"national_id"`
	SupervisorID       uuid.UUID  `json:"supervisor_id"`
	SupervisorName     string     `json:"supervisor_name"`
	Type               string     `json:"type"`
	Questions          []string   `json:"questions"`
	Answers            []string   `json:"answers"`
	ReportText         string     `json:"report_text"`
	Images             []string   `json:"images"`
	Amount             *float64   `json:"amount"`
	StartDate          *time.Time `json:"start_date"`
	EndDate            *time.Time `json:"end_date"`
	Items              []string   `json:"items"`
	IsGuilty           bool       `json:"is_guilty"`
	Notes              string     `json:"notes"`
	DeductionMonth     string     `json:"deduction_month"`
	Status             string     `json:"status"`
	ApprovedByName     string     `json:"approved_by_name"`
	ApprovedByUsername string     `json:"approved_by_username"`
	RejectedByName     string     `json:"rejected_by_name"`
	RejectedByUsername string     `json:"rejected_by_username"`
	ApprovedAt         *time.Time `json:"approved_at"`
	RejectedAt         *time.Time `json:"rejected_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

// 2. EmployeeDocument
type CreateEmployeeDocumentRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" binding:"required"`
	DocType    string    `json:"doc_type" binding:"required"`
	Title      string    `json:"title" binding:"required"`
	DocNumber  string    `json:"doc_number"`
	FileURL    string    `json:"file_url"`
	IssueDate  *string   `json:"issue_date"`
	ExpiryDate *string   `json:"expiry_date"`
	Status     string    `json:"status"`
	Notes      string    `json:"notes"`
}

type UpdateEmployeeDocumentRequest struct {
	DocType    *string `json:"doc_type"`
	Title      *string `json:"title"`
	DocNumber  *string `json:"doc_number"`
	FileURL    *string `json:"file_url"`
	IssueDate  *string `json:"issue_date"`
	ExpiryDate *string `json:"expiry_date"`
	Status     *string `json:"status"`
	Notes      *string `json:"notes"`
}

type EmployeeDocumentFilter struct {
	EmployeeID *uuid.UUID `form:"employee_id"`
	DocType    string     `form:"doc_type"`
	Status     string     `form:"status"`
	Search     string     `form:"search"`
	Page       int        `form:"page,default=1"`
	Limit      int        `form:"limit,default=50"`
	PageSize   int        `form:"page_size"`
}

func (f *EmployeeDocumentFilter) GetEffectiveLimit() int {
	if f.PageSize > 0 {
		return f.PageSize
	}
	if f.Limit > 0 {
		return f.Limit
	}
	return 50
}

// 3. EmployeeBankAccount
type CreateEmployeeBankAccountRequest struct {
	EmployeeID       uuid.UUID `json:"employee_id" binding:"required"`
	BankName         string    `json:"bank_name" binding:"required"`
	IBAN             string    `json:"iban" binding:"required"`
	AccountOwnerName string    `json:"account_owner_name" binding:"required"`
	IsDefault        bool      `json:"is_default"`
}

type UpdateEmployeeBankAccountRequest struct {
	BankName         *string `json:"bank_name"`
	IBAN             *string `json:"iban"`
	AccountOwnerName *string `json:"account_owner_name"`
	IsDefault        *bool   `json:"is_default"`
}

type EmployeeBankAccountFilter struct {
	EmployeeID *uuid.UUID `form:"employee_id"`
	Search     string     `form:"search"`
	Page       int        `form:"page,default=1"`
	Limit      int        `form:"limit,default=50"`
	PageSize   int        `form:"page_size"`
}

func (f *EmployeeBankAccountFilter) GetEffectiveLimit() int {
	if f.PageSize > 0 {
		return f.PageSize
	}
	if f.Limit > 0 {
		return f.Limit
	}
	return 50
}

// 4. LeaveRequest
type CreateLeaveRequestRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" binding:"required"`
	LeaveType  string    `json:"leave_type"`
	StartDate  string    `json:"start_date" binding:"required"`
	EndDate    string    `json:"end_date" binding:"required"`
	DaysCount  int       `json:"days_count"`
	Reason     string    `json:"reason"`
}

type UpdateLeaveRequestStatusRequest struct {
	Status         string `json:"status" binding:"required"`
	ApprovedByName string `json:"approved_by_name"`
}

type LeaveRequestFilter struct {
	EmployeeID *uuid.UUID `form:"employee_id"`
	Status     string     `form:"status"`
	LeaveType  string     `form:"leave_type"`
	Search     string     `form:"search"`
	Page       int        `form:"page,default=1"`
	Limit      int        `form:"limit,default=50"`
	PageSize   int        `form:"page_size"`
}

func (f *LeaveRequestFilter) GetEffectiveLimit() int {
	if f.PageSize > 0 {
		return f.PageSize
	}
	if f.Limit > 0 {
		return f.Limit
	}
	return 50
}

// 5. TrafficViolation
type CreateTrafficViolationRequest struct {
	ViolationNumber string     `json:"violation_number"`
	EmployeeID      *uuid.UUID `json:"employee_id"`
	VehiclePlate    string     `json:"vehicle_plate"`
	Amount          float64    `json:"amount" binding:"required"`
	PaidAmount      *float64   `json:"paid_amount"`
	Reason          string     `json:"reason" binding:"required"`
	ViolationDate   string     `json:"violation_date"`
	City            string     `json:"city"`
	Status          string     `json:"status"`
	BranchID        *uuid.UUID `json:"branch_id"`
	Notes           string     `json:"notes"`
}

type UpdateTrafficViolationRequest struct {
	ViolationNumber *string    `json:"violation_number"`
	EmployeeID      *uuid.UUID `json:"employee_id"`
	VehiclePlate    *string    `json:"vehicle_plate"`
	Amount          *float64   `json:"amount"`
	PaidAmount      *float64   `json:"paid_amount"`
	AddPayment      *float64   `json:"add_payment"`
	Reason          *string    `json:"reason"`
	ViolationDate   *string    `json:"violation_date"`
	City            *string    `json:"city"`
	Status          *string    `json:"status"`
	Notes           *string    `json:"notes"`
}

type TrafficViolationFilter struct {
	BranchID   *uuid.UUID `form:"branch_id"`
	EmployeeID *uuid.UUID `form:"employee_id"`
	Status     string     `form:"status"`
	Search     string     `form:"search"`
	StartDate  string     `form:"start_date"`
	EndDate    string     `form:"end_date"`
	Page       int        `form:"page"`
	Limit      int        `form:"limit"`
	PageSize   int        `form:"page_size"`
}

func (f *TrafficViolationFilter) GetEffectiveLimit() int {
	if f.PageSize > 0 {
		return f.PageSize
	}
	if f.Limit > 0 {
		return f.Limit
	}
	return 50
}

// 6. FuelLog
type CreateFuelLogRequest struct {
	EmployeeID      *uuid.UUID `json:"employee_id"`
	VehiclePlate    string     `json:"vehicle_plate"`
	ShiftID         *uuid.UUID `json:"shift_id"`
	Amount          float64    `json:"amount" binding:"required"`
	Liters          float64    `json:"liters"`
	FuelDate        string     `json:"fuel_date"`
	StationName     string     `json:"station_name"`
	InvoiceImageURL string     `json:"invoice_image_url"`
	BranchID        *uuid.UUID `json:"branch_id"`
	Notes           string     `json:"notes"`
}

type UpdateFuelLogRequest struct {
	EmployeeID      *uuid.UUID `json:"employee_id"`
	VehiclePlate    *string    `json:"vehicle_plate"`
	Amount          *float64   `json:"amount"`
	Liters          *float64   `json:"liters"`
	FuelDate        *string    `json:"fuel_date"`
	StationName     *string    `json:"station_name"`
	InvoiceImageURL *string    `json:"invoice_image_url"`
	Notes           *string    `json:"notes"`
}

type FuelLogFilter struct {
	BranchID   *uuid.UUID `form:"branch_id"`
	EmployeeID *uuid.UUID `form:"employee_id"`
	Plate      string     `form:"plate"`
	StartDate  string     `form:"start_date"`
	EndDate    string     `form:"end_date"`
	Search     string     `form:"search"`
	Page       int        `form:"page"`
	Limit      int        `form:"limit"`
	PageSize   int        `form:"page_size"`
}

func (f *FuelLogFilter) GetEffectiveLimit() int {
	if f.PageSize > 0 {
		return f.PageSize
	}
	if f.Limit > 0 {
		return f.Limit
	}
	return 50
}
