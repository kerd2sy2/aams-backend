package hr_legal

import (
	"context"

	"time"

	"github.com/google/uuid"

)

type HRLegalService interface {
	// Investigation
	CreateInvestigation(ctx context.Context, req CreateInvestigationRequest, supervisorID uuid.UUID) (*InvestigationResponse, error)
	UpdateInvestigation(ctx context.Context, id uuid.UUID, req UpdateInvestigationRequest) (*InvestigationResponse, error)
	DeleteInvestigation(ctx context.Context, id uuid.UUID) error
	GetAllInvestigations(ctx context.Context, branchID *uuid.UUID) ([]InvestigationResponse, error)
	GetInvestigationByID(ctx context.Context, id uuid.UUID) (*InvestigationResponse, error)
	ApproveInvestigation(ctx context.Context, id uuid.UUID, adminID uuid.UUID, adminName, adminUsername string) (*InvestigationResponse, error)
	RejectInvestigation(ctx context.Context, id uuid.UUID, adminID uuid.UUID, adminName, adminUsername string) (*InvestigationResponse, error)
	GetPendingInvestigationCount(ctx context.Context) (int64, error)

	// Document
	CreateDocument(ctx context.Context, req CreateEmployeeDocumentRequest) (*EmployeeDocument, error)
	UpdateDocument(ctx context.Context, id uuid.UUID, req UpdateEmployeeDocumentRequest) (*EmployeeDocument, error)
	DeleteDocument(ctx context.Context, id uuid.UUID) error
	GetDocumentByID(ctx context.Context, id uuid.UUID) (*EmployeeDocument, error)
	GetAllDocuments(ctx context.Context, filter EmployeeDocumentFilter) ([]EmployeeDocument, int64, error)
	GetExpiringDocuments(ctx context.Context, days int) ([]EmployeeDocument, error)

	// Bank Account
	CreateBankAccount(ctx context.Context, req CreateEmployeeBankAccountRequest) (*EmployeeBankAccount, error)
	UpdateBankAccount(ctx context.Context, id uuid.UUID, req UpdateEmployeeBankAccountRequest) (*EmployeeBankAccount, error)
	DeleteBankAccount(ctx context.Context, id uuid.UUID) error
	GetBankAccountByID(ctx context.Context, id uuid.UUID) (*EmployeeBankAccount, error)
	GetAllBankAccounts(ctx context.Context, filter EmployeeBankAccountFilter) ([]EmployeeBankAccount, int64, error)

	// Leave Request
	CreateLeaveRequest(ctx context.Context, req CreateLeaveRequestRequest) (*LeaveRequest, error)
	UpdateLeaveRequestStatus(ctx context.Context, id uuid.UUID, req UpdateLeaveRequestStatusRequest) (*LeaveRequest, error)
	DeleteLeaveRequest(ctx context.Context, id uuid.UUID) error
	GetLeaveRequestByID(ctx context.Context, id uuid.UUID) (*LeaveRequest, error)
	GetAllLeaveRequests(ctx context.Context, filter LeaveRequestFilter) ([]LeaveRequest, int64, error)
	GetPendingLeaveCount(ctx context.Context) (int64, error)

	// Traffic Violation
	CreateViolation(ctx context.Context, req CreateTrafficViolationRequest) (*TrafficViolation, error)
	UpdateViolation(ctx context.Context, id uuid.UUID, req UpdateTrafficViolationRequest) (*TrafficViolation, error)
	DeleteViolation(ctx context.Context, id uuid.UUID) error
	GetViolationByID(ctx context.Context, id uuid.UUID) (*TrafficViolation, error)
	GetAllViolations(ctx context.Context, filter TrafficViolationFilter, adminBranchID *uuid.UUID) ([]TrafficViolation, int64, float64, float64, error)

	// Fuel Log
	CreateFuelLog(ctx context.Context, req CreateFuelLogRequest, adminBranchID *uuid.UUID) (*FuelLog, error)
	UpdateFuelLog(ctx context.Context, id uuid.UUID, req UpdateFuelLogRequest) (*FuelLog, error)
	DeleteFuelLog(ctx context.Context, id uuid.UUID) error
	GetFuelLogByID(ctx context.Context, id uuid.UUID) (*FuelLog, error)
	GetAllFuelLogs(ctx context.Context, filter FuelLogFilter, adminBranchID *uuid.UUID) ([]FuelLog, int64, error)

	// Contract implementations
	CheckEmployeeOnLeave(ctx context.Context, empID uuid.UUID, date time.Time) (*LeaveCheckDTO, error)
	GetPendingRequestsCount(ctx context.Context) (investigations int64, leaves int64, err error)
}

type hrLegalService struct {
	repo HRLegalRepository
}

func NewHRLegalService(repo HRLegalRepository) HRLegalService {
	return &hrLegalService{repo: repo}
}

// ------------------------------------------------------------------
// 1. Investigation
// ------------------------------------------------------------------