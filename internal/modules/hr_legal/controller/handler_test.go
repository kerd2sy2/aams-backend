package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/hr_legal/contracts"
	"delivery-backend/internal/modules/hr_legal/controller"
	"delivery-backend/internal/modules/hr_legal/domain"
	"delivery-backend/internal/modules/hr_legal/dto"
)

type mockHRLegalService struct {
	createInvFunc   func(ctx context.Context, req dto.CreateInvestigationRequest, supervisorID uuid.UUID) (*dto.InvestigationResponse, error)
	createDocFunc   func(ctx context.Context, req dto.CreateEmployeeDocumentRequest) (*domain.EmployeeDocument, error)
	createLeaveFunc func(ctx context.Context, req dto.CreateLeaveRequestRequest) (*domain.LeaveRequest, error)
	createVioFunc   func(ctx context.Context, req dto.CreateTrafficViolationRequest) (*domain.TrafficViolation, error)
}

func (m *mockHRLegalService) CreateInvestigation(ctx context.Context, req dto.CreateInvestigationRequest, supervisorID uuid.UUID) (*dto.InvestigationResponse, error) {
	if m.createInvFunc != nil {
		return m.createInvFunc(ctx, req, supervisorID)
	}
	return nil, nil
}

func (m *mockHRLegalService) UpdateInvestigation(ctx context.Context, id uuid.UUID, req dto.UpdateInvestigationRequest) (*dto.InvestigationResponse, error) {
	return nil, nil
}

func (m *mockHRLegalService) GetAllInvestigations(ctx context.Context, branchID *uuid.UUID) ([]dto.InvestigationResponse, error) {
	return nil, nil
}

func (m *mockHRLegalService) GetInvestigationByID(ctx context.Context, id uuid.UUID) (*dto.InvestigationResponse, error) {
	return nil, nil
}

func (m *mockHRLegalService) ApproveInvestigation(ctx context.Context, id uuid.UUID, adminID uuid.UUID, adminName, adminUsername string) (*dto.InvestigationResponse, error) {
	return nil, nil
}

func (m *mockHRLegalService) RejectInvestigation(ctx context.Context, id uuid.UUID, adminID uuid.UUID, adminName, adminUsername string) (*dto.InvestigationResponse, error) {
	return nil, nil
}

func (m *mockHRLegalService) GetPendingInvestigationCount(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockHRLegalService) CreateDocument(ctx context.Context, req dto.CreateEmployeeDocumentRequest) (*domain.EmployeeDocument, error) {
	if m.createDocFunc != nil {
		return m.createDocFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockHRLegalService) UpdateDocument(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeDocumentRequest) (*domain.EmployeeDocument, error) {
	return nil, nil
}

func (m *mockHRLegalService) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockHRLegalService) GetDocumentByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeDocument, error) {
	return nil, nil
}

func (m *mockHRLegalService) GetAllDocuments(ctx context.Context, filter dto.EmployeeDocumentFilter) ([]domain.EmployeeDocument, int64, error) {
	return nil, 0, nil
}

func (m *mockHRLegalService) GetExpiringDocuments(ctx context.Context, days int) ([]domain.EmployeeDocument, error) {
	return nil, nil
}

func (m *mockHRLegalService) CreateBankAccount(ctx context.Context, req dto.CreateEmployeeBankAccountRequest) (*domain.EmployeeBankAccount, error) {
	return nil, nil
}

func (m *mockHRLegalService) UpdateBankAccount(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeBankAccountRequest) (*domain.EmployeeBankAccount, error) {
	return nil, nil
}

func (m *mockHRLegalService) DeleteBankAccount(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockHRLegalService) GetBankAccountByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeBankAccount, error) {
	return nil, nil
}

func (m *mockHRLegalService) GetAllBankAccounts(ctx context.Context, filter dto.EmployeeBankAccountFilter) ([]domain.EmployeeBankAccount, int64, error) {
	return nil, 0, nil
}

func (m *mockHRLegalService) CreateLeaveRequest(ctx context.Context, req dto.CreateLeaveRequestRequest) (*domain.LeaveRequest, error) {
	if m.createLeaveFunc != nil {
		return m.createLeaveFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockHRLegalService) UpdateLeaveRequestStatus(ctx context.Context, id uuid.UUID, req dto.UpdateLeaveRequestStatusRequest) (*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockHRLegalService) DeleteLeaveRequest(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockHRLegalService) GetLeaveRequestByID(ctx context.Context, id uuid.UUID) (*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockHRLegalService) GetAllLeaveRequests(ctx context.Context, filter dto.LeaveRequestFilter) ([]domain.LeaveRequest, int64, error) {
	return nil, 0, nil
}

func (m *mockHRLegalService) GetPendingLeaveCount(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockHRLegalService) CreateViolation(ctx context.Context, req dto.CreateTrafficViolationRequest) (*domain.TrafficViolation, error) {
	if m.createVioFunc != nil {
		return m.createVioFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockHRLegalService) UpdateViolation(ctx context.Context, id uuid.UUID, req dto.UpdateTrafficViolationRequest) (*domain.TrafficViolation, error) {
	return nil, nil
}

func (m *mockHRLegalService) DeleteViolation(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockHRLegalService) GetViolationByID(ctx context.Context, id uuid.UUID) (*domain.TrafficViolation, error) {
	return nil, nil
}

func (m *mockHRLegalService) GetAllViolations(ctx context.Context, filter dto.TrafficViolationFilter, adminBranchID *uuid.UUID) ([]domain.TrafficViolation, int64, error) {
	return nil, 0, nil
}

func (m *mockHRLegalService) CreateFuelLog(ctx context.Context, req dto.CreateFuelLogRequest, adminBranchID *uuid.UUID) (*domain.FuelLog, error) {
	return nil, nil
}

func (m *mockHRLegalService) UpdateFuelLog(ctx context.Context, id uuid.UUID, req dto.UpdateFuelLogRequest) (*domain.FuelLog, error) {
	return nil, nil
}

func (m *mockHRLegalService) DeleteFuelLog(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockHRLegalService) GetFuelLogByID(ctx context.Context, id uuid.UUID) (*domain.FuelLog, error) {
	return nil, nil
}

func (m *mockHRLegalService) GetAllFuelLogs(ctx context.Context, filter dto.FuelLogFilter, adminBranchID *uuid.UUID) ([]domain.FuelLog, int64, error) {
	return nil, 0, nil
}

func (m *mockHRLegalService) CheckEmployeeOnLeave(ctx context.Context, empID uuid.UUID, date time.Time) (*contracts.LeaveCheckDTO, error) {
	return nil, nil
}

func (m *mockHRLegalService) GetPendingRequestsCount(ctx context.Context) (int64, int64, error) {
	return 0, 0, nil
}

func TestHRLegalHandler_Endpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	empID := uuid.New()
	supID := uuid.New()

	mockSvc := &mockHRLegalService{
		createInvFunc: func(ctx context.Context, req dto.CreateInvestigationRequest, supervisorID uuid.UUID) (*dto.InvestigationResponse, error) {
			return &dto.InvestigationResponse{
				ID:           uuid.New(),
				EmployeeID:   empID,
				SupervisorID: supID,
				Type:         req.Type,
				ReportText:   req.ReportText,
				Status:       "pending",
			}, nil
		},
		createDocFunc: func(ctx context.Context, req dto.CreateEmployeeDocumentRequest) (*domain.EmployeeDocument, error) {
			return &domain.EmployeeDocument{
				ID:         uuid.New(),
				EmployeeID: req.EmployeeID,
				DocType:    req.DocType,
				Title:      req.Title,
			}, nil
		},
	}

	h := controller.NewHRLegalHandler(mockSvc)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("admin_id", supID)
		c.Next()
	})

	r.POST("/api/v1/investigations", h.CreateInvestigation)
	r.POST("/api/v1/documents", h.CreateDocument)

	// Test Create Investigation
	body, _ := json.Marshal(dto.CreateInvestigationRequest{
		EmployeeID: empID.String(),
		Type:       "absence",
		ReportText: "Absence report",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/investigations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for investigation, got %d: %s", w.Code, w.Body.String())
	}

	// Test Create Document
	docBody, _ := json.Marshal(dto.CreateEmployeeDocumentRequest{
		EmployeeID: empID,
		DocType:    "CONTRACT",
		Title:      "Employment Contract",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/documents", bytes.NewBuffer(docBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for document, got %d: %s", w.Code, w.Body.String())
	}
}
