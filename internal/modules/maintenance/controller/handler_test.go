package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/maintenance/contracts"
	"delivery-backend/internal/modules/maintenance/controller"
	"delivery-backend/internal/modules/maintenance/domain"
	"delivery-backend/internal/modules/maintenance/dto"
)

type mockMaintenanceService struct {
	createReqFunc func(ctx context.Context, req dto.CreateMaintenanceRequestRequest, adminBranchID *uuid.UUID) (*domain.MaintenanceRequest, error)
	getAllReqFunc func(ctx context.Context, filter dto.MaintenanceRequestFilter, adminBranchID *uuid.UUID) ([]domain.MaintenanceRequest, int64, error)
	getAllLogs    func(ctx context.Context, page, limit int) ([]domain.MaintenanceLog, int64, error)
}

func (m *mockMaintenanceService) GetEmployeeLogs(ctx context.Context, empID uuid.UUID, limit int) ([]domain.MaintenanceLog, error) {
	return nil, nil
}

func (m *mockMaintenanceService) GetAllLogs(ctx context.Context, page, limit int) ([]domain.MaintenanceLog, int64, error) {
	if m.getAllLogs != nil {
		return m.getAllLogs(ctx, page, limit)
	}
	return nil, 0, nil
}

func (m *mockMaintenanceService) CreateRequest(ctx context.Context, req dto.CreateMaintenanceRequestRequest, adminBranchID *uuid.UUID) (*domain.MaintenanceRequest, error) {
	if m.createReqFunc != nil {
		return m.createReqFunc(ctx, req, adminBranchID)
	}
	return nil, nil
}

func (m *mockMaintenanceService) UpdateRequest(ctx context.Context, id uuid.UUID, req dto.UpdateMaintenanceRequestRequest) (*domain.MaintenanceRequest, error) {
	return nil, nil
}

func (m *mockMaintenanceService) DeleteRequest(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockMaintenanceService) GetRequestByID(ctx context.Context, id uuid.UUID) (*domain.MaintenanceRequest, error) {
	return nil, nil
}

func (m *mockMaintenanceService) GetAllRequests(ctx context.Context, filter dto.MaintenanceRequestFilter, adminBranchID *uuid.UUID) ([]domain.MaintenanceRequest, int64, error) {
	if m.getAllReqFunc != nil {
		return m.getAllReqFunc(ctx, filter, adminBranchID)
	}
	return nil, 0, nil
}

func (m *mockMaintenanceService) RecordMaintenanceLog(ctx context.Context, empID *uuid.UUID, maintType string, details string, distanceAt, cost float64, adminName string) error {
	return nil
}

func (m *mockMaintenanceService) GetLastOilChange(ctx context.Context, empID uuid.UUID) (*contracts.MaintenanceLogContractDTO, error) {
	return nil, nil
}

func TestMaintenanceHandler_CreateRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reqID := uuid.New()
	mockSvc := &mockMaintenanceService{
		createReqFunc: func(ctx context.Context, req dto.CreateMaintenanceRequestRequest, adminBranchID *uuid.UUID) (*domain.MaintenanceRequest, error) {
			return &domain.MaintenanceRequest{
				ID:               reqID,
				VehiclePlate:     req.VehiclePlate,
				IssueDescription: req.IssueDescription,
				Priority:         req.Priority,
				Status:           "OPEN",
			}, nil
		},
	}

	h := controller.NewMaintenanceHandler(mockSvc)
	r := gin.New()
	r.POST("/api/v1/maintenance-requests", h.CreateRequest)

	body, _ := json.Marshal(dto.CreateMaintenanceRequestRequest{
		VehiclePlate:     "1234-XYZ",
		IssueDescription: "Tire replacement",
		Priority:         "URGENT",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/maintenance-requests", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMaintenanceHandler_GetAllLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockMaintenanceService{
		getAllLogs: func(ctx context.Context, page, limit int) ([]domain.MaintenanceLog, int64, error) {
			return []domain.MaintenanceLog{
				{
					ID:         uuid.New(),
					Type:       "oil_change",
					DistanceAt: 4500,
				},
			}, 1, nil
		},
	}

	h := controller.NewMaintenanceHandler(mockSvc)
	r := gin.New()
	r.GET("/api/v1/maintenance/logs", h.GetAllLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/maintenance/logs?page=1&limit=20", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}
