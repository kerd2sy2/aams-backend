package controller_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"delivery-backend/internal/modules/target/controller"
	"delivery-backend/internal/modules/target/dto"
	"delivery-backend/internal/modules/target/service"
)

type mockTargetService struct {
	service.TargetService
}

func (m *mockTargetService) GetDashboardSummary(ctx context.Context, month string, branch string) (*dto.TargetDashboardSummaryDTO, error) {
	return &dto.TargetDashboardSummaryDTO{
		TotalIdentifiers: 10,
		TotalMonthOrders: 3500,
	}, nil
}

func (m *mockTargetService) GetTargetSettings(ctx context.Context) (*dto.TargetSettingsDTO, error) {
	return &dto.TargetSettingsDTO{
		DefaultMonthlyTarget: 460,
		DefaultDailyTarget:   18,
	}, nil
}

func TestTargetHandler_GetDashboardSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := controller.NewTargetHandler(&mockTargetService{}, nil)

	r := gin.New()
	r.GET("/target/dashboard", h.GetDashboardSummary)
	r.GET("/target/settings", h.GetTargetSettings)

	// 1. Dashboard
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/target/dashboard?month=2026-10", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// 2. Settings
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/target/settings", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
