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

	"delivery-backend/internal/modules/custody/contracts"
	"delivery-backend/internal/modules/custody/controller"
	"delivery-backend/internal/modules/custody/domain"
	"delivery-backend/internal/modules/custody/dto"
)

type mockCustodyService struct {
	listFunc   func(ctx context.Context, branchID *uuid.UUID) ([]dto.CustodyDayResponse, error)
	createFunc func(ctx context.Context, req dto.CreateCustodyDayRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error)
}

func (m *mockCustodyService) List(ctx context.Context, branchID *uuid.UUID) ([]dto.CustodyDayResponse, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, branchID)
	}
	return nil, nil
}

func (m *mockCustodyService) Create(ctx context.Context, req dto.CreateCustodyDayRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, req, adminID, adminName, adminUsername)
	}
	return nil, nil
}

func (m *mockCustodyService) AddAmount(ctx context.Context, req dto.AddCustodyAmountRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error) {
	return nil, nil
}

func (m *mockCustodyService) AddExpense(ctx context.Context, dayID uuid.UUID, branchID *uuid.UUID, req dto.CreateCustodyExpenseRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error) {
	return nil, nil
}

func (m *mockCustodyService) DeleteExpense(ctx context.Context, expenseID uuid.UUID, branchID *uuid.UUID, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error) {
	return nil, nil
}

func (m *mockCustodyService) GetLogs(ctx context.Context, filter dto.CustodyLogFilter) ([]domain.CustodyLog, int64, error) {
	return nil, 0, nil
}

func (m *mockCustodyService) DeleteLog(ctx context.Context, id uuid.UUID, adminID *uuid.UUID, adminName, adminUsername string) error {
	return nil
}

func (m *mockCustodyService) GetDaySummary(ctx context.Context, branchID *uuid.UUID, date string) (*contracts.CustodySummaryDTO, error) {
	return nil, nil
}

func TestCustodyHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dayID := uuid.New()
	mockSvc := &mockCustodyService{
		createFunc: func(ctx context.Context, req dto.CreateCustodyDayRequest, adminID *uuid.UUID, adminName, adminUsername string) (*dto.CustodyDayResponse, error) {
			return &dto.CustodyDayResponse{
				ID:             dayID,
				Date:           req.Date,
				OpeningBalance: 0,
				AddedAmount:    req.AddedAmount,
				CustodyValue:   req.AddedAmount,
				ClosingBalance: req.AddedAmount,
				CreatedAt:      time.Now(),
			}, nil
		},
	}

	h := controller.NewCustodyHandler(mockSvc)
	r := gin.New()
	r.POST("/api/v1/custody", h.Create)

	body, _ := json.Marshal(dto.CreateCustodyDayRequest{
		Date:        "2026-10-03",
		AddedAmount: 1000.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/custody", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCustodyHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockCustodyService{
		listFunc: func(ctx context.Context, branchID *uuid.UUID) ([]dto.CustodyDayResponse, error) {
			return []dto.CustodyDayResponse{
				{
					ID:             uuid.New(),
					Date:           "2026-10-03",
					OpeningBalance: 100,
					AddedAmount:    500,
					ClosingBalance: 600,
				},
			}, nil
		},
	}

	h := controller.NewCustodyHandler(mockSvc)
	r := gin.New()
	r.GET("/api/v1/custody", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/custody", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}
