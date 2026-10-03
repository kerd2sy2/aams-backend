package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/maintenance/contracts"
	"delivery-backend/internal/modules/maintenance/domain"
	"delivery-backend/internal/modules/maintenance/dto"
	"delivery-backend/internal/modules/maintenance/repository"
)

type MaintenanceService interface {
	// Maintenance Logs
	GetEmployeeLogs(ctx context.Context, empID uuid.UUID, limit int) ([]domain.MaintenanceLog, error)
	GetAllLogs(ctx context.Context, page, limit int) ([]domain.MaintenanceLog, int64, error)

	// Maintenance Requests
	CreateRequest(ctx context.Context, req dto.CreateMaintenanceRequestRequest, adminBranchID *uuid.UUID) (*domain.MaintenanceRequest, error)
	UpdateRequest(ctx context.Context, id uuid.UUID, req dto.UpdateMaintenanceRequestRequest) (*domain.MaintenanceRequest, error)
	DeleteRequest(ctx context.Context, id uuid.UUID) error
	GetRequestByID(ctx context.Context, id uuid.UUID) (*domain.MaintenanceRequest, error)
	GetAllRequests(ctx context.Context, filter dto.MaintenanceRequestFilter, adminBranchID *uuid.UUID) ([]domain.MaintenanceRequest, int64, error)

	// Contract implementations
	RecordMaintenanceLog(ctx context.Context, empID *uuid.UUID, maintType string, details string, distanceAt, cost float64, adminName string) error
	GetLastOilChange(ctx context.Context, empID uuid.UUID) (*contracts.MaintenanceLogContractDTO, error)
}

type maintenanceService struct {
	repo repository.MaintenanceRepository
}

func NewMaintenanceService(repo repository.MaintenanceRepository) MaintenanceService {
	return &maintenanceService{repo: repo}
}

func (s *maintenanceService) GetEmployeeLogs(ctx context.Context, empID uuid.UUID, limit int) ([]domain.MaintenanceLog, error) {
	return s.repo.FindByEmployeeID(ctx, empID, limit)
}

func (s *maintenanceService) GetAllLogs(ctx context.Context, page, limit int) ([]domain.MaintenanceLog, int64, error) {
	return s.repo.FindAllLogs(ctx, page, limit)
}

func (s *maintenanceService) CreateRequest(ctx context.Context, req dto.CreateMaintenanceRequestRequest, adminBranchID *uuid.UUID) (*domain.MaintenanceRequest, error) {
	priority := "MEDIUM"
	if req.Priority != "" {
		priority = req.Priority
	}
	status := "OPEN"
	if req.Status != "" {
		status = req.Status
	}

	branchID := req.BranchID
	if branchID == nil && adminBranchID != nil {
		branchID = adminBranchID
	}

	m := &domain.MaintenanceRequest{
		ID:               uuid.New(),
		VehiclePlate:     req.VehiclePlate,
		EmployeeID:       req.EmployeeID,
		IssueDescription: req.IssueDescription,
		Priority:         priority,
		EstimatedCost:    req.EstimatedCost,
		ActualCost:       req.ActualCost,
		WorkshopName:     req.WorkshopName,
		Status:           status,
		BranchID:         branchID,
		Notes:            req.Notes,
	}

	if err := s.repo.CreateRequest(ctx, m); err != nil {
		return nil, err
	}
	return s.repo.FindRequestByID(ctx, m.ID)
}

func (s *maintenanceService) UpdateRequest(ctx context.Context, id uuid.UUID, req dto.UpdateMaintenanceRequestRequest) (*domain.MaintenanceRequest, error) {
	m, err := s.repo.FindRequestByID(ctx, id)
	if err != nil {
		return nil, errors.New("طلب الصيانة غير موجود")
	}

	if req.VehiclePlate != nil {
		m.VehiclePlate = *req.VehiclePlate
	}
	if req.EmployeeID != nil {
		m.EmployeeID = req.EmployeeID
	}
	if req.IssueDescription != nil {
		m.IssueDescription = *req.IssueDescription
	}
	if req.Priority != nil {
		m.Priority = *req.Priority
	}
	if req.EstimatedCost != nil {
		m.EstimatedCost = *req.EstimatedCost
	}
	if req.ActualCost != nil {
		m.ActualCost = *req.ActualCost
	}
	if req.WorkshopName != nil {
		m.WorkshopName = *req.WorkshopName
	}
	if req.Status != nil {
		m.Status = *req.Status
	}
	if req.Notes != nil {
		m.Notes = *req.Notes
	}
	if req.BranchID != nil {
		m.BranchID = req.BranchID
	}

	if err := s.repo.UpdateRequest(ctx, m); err != nil {
		return nil, err
	}
	return s.repo.FindRequestByID(ctx, id)
}

func (s *maintenanceService) DeleteRequest(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteRequest(ctx, id)
}

func (s *maintenanceService) GetRequestByID(ctx context.Context, id uuid.UUID) (*domain.MaintenanceRequest, error) {
	return s.repo.FindRequestByID(ctx, id)
}

func (s *maintenanceService) GetAllRequests(ctx context.Context, filter dto.MaintenanceRequestFilter, adminBranchID *uuid.UUID) ([]domain.MaintenanceRequest, int64, error) {
	if adminBranchID != nil {
		filter.BranchID = adminBranchID
	}
	return s.repo.FindAllRequests(ctx, filter)
}

// Contract implementations
func (s *maintenanceService) RecordMaintenanceLog(ctx context.Context, empID *uuid.UUID, maintType string, details string, distanceAt, cost float64, adminName string) error {
	log := &domain.MaintenanceLog{
		ID:         uuid.New(),
		EmployeeID: empID,
		Type:       maintType,
		Details:    details,
		DistanceAt: distanceAt,
		Cost:       cost,
		AdminName:  adminName,
	}
	return s.repo.CreateLog(ctx, log)
}

func (s *maintenanceService) GetLastOilChange(ctx context.Context, empID uuid.UUID) (*contracts.MaintenanceLogContractDTO, error) {
	log, err := s.repo.FindLastOilChange(ctx, empID)
	if err != nil {
		return nil, err
	}
	return &contracts.MaintenanceLogContractDTO{
		ID:         log.ID,
		EmployeeID: log.EmployeeID,
		Type:       log.Type,
		Details:    log.Details,
		DistanceAt: log.DistanceAt,
		Cost:       log.Cost,
		AdminName:  log.AdminName,
		CreatedAt:  log.CreatedAt,
	}, nil
}
