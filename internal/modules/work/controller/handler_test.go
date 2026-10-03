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

	"delivery-backend/internal/modules/work/contracts"
	"delivery-backend/internal/modules/work/controller"
	"delivery-backend/internal/modules/work/domain"
	"delivery-backend/internal/modules/work/dto"
	"delivery-backend/internal/modules/work/service"
)

type mockWorkService struct {
	service.WorkService
}

func (m *mockWorkService) StartWork(ctx context.Context, req dto.StartWorkRequest) (*domain.WorkSession, error) {
	empID, _ := uuid.Parse(req.EmployeeID)
	return &domain.WorkSession{
		ID:         uuid.New(),
		EmployeeID: &empID,
		StartTime:  time.Now(),
		StartKM:    req.StartKM,
		Status:     "ACTIVE",
	}, nil
}

func (m *mockWorkService) EndWork(ctx context.Context, req dto.EndWorkRequest, reviewerID *uuid.UUID, reviewerName string, isSupervisor bool) (*domain.WorkSession, error) {
	empID, _ := uuid.Parse(req.EmployeeID)
	now := time.Now()
	return &domain.WorkSession{
		ID:          uuid.New(),
		EmployeeID:  &empID,
		EndTime:     &now,
		EndKM:       req.EndKM,
		OrdersCount: req.OrdersCount,
		Status:      "COMPLETED",
	}, nil
}

func (m *mockWorkService) GetLastSessionOrVehicleKM(ctx context.Context, empID uuid.UUID, motorcycleNumber string) (*dto.LastKMResponse, error) {
	return &dto.LastKMResponse{
		LastKM: 1200,
	}, nil
}

func (m *mockWorkService) CountTodaySessions(ctx context.Context, empID uuid.UUID) (int64, error) {
	return 2, nil
}

func (m *mockWorkService) CheckOilChange(ctx context.Context, empID uuid.UUID, motorcycleNumber string) (*dto.OilChangeCheckResponse, error) {
	return &dto.OilChangeCheckResponse{
		NeedsOilChange: false,
		Interval:       950,
	}, nil
}

func (m *mockWorkService) GetActiveSession(ctx context.Context, empID uuid.UUID) (*contracts.WorkSessionDTO, error) {
	return &contracts.WorkSessionDTO{
		ID:      uuid.New(),
		Status:  "ACTIVE",
		StartKM: 1000,
	}, nil
}

func TestWorkHandler_Endpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := controller.NewWorkHandler(&mockWorkService{})

	r := gin.New()
	r.POST("/work/start", h.StartWork)
	r.POST("/work/end", h.EndWork)
	r.GET("/work/last-km", h.GetLastKM)
	r.GET("/work/today-count", h.TodayCount)
	r.GET("/work/check-oil", h.CheckOilChange)

	empID := uuid.New()

	// 1. Start Work
	startReq, _ := json.Marshal(dto.StartWorkRequest{
		EmployeeID: empID.String(),
		StartKM:    500,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/work/start", bytes.NewBuffer(startReq))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("expected 200/201, got %d", w.Code)
	}

	// 2. End Work
	endReq, _ := json.Marshal(dto.EndWorkRequest{
		EmployeeID:  empID.String(),
		EndKM:       550,
		OrdersCount: 10,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/work/end", bytes.NewBuffer(endReq))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// 3. Last KM
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/work/last-km?employee_id="+empID.String(), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
