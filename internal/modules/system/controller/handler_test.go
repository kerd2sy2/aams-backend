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

	"delivery-backend/internal/modules/system/controller"
	"delivery-backend/internal/modules/system/dto"
)

type mockDashboardService struct {
	getStatsFunc func(ctx context.Context, branchID *uuid.UUID) (*dto.DashboardStatsResponse, error)
}

func (m *mockDashboardService) GetStats(ctx context.Context, branchID *uuid.UUID) (*dto.DashboardStatsResponse, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx, branchID)
	}
	return &dto.DashboardStatsResponse{TotalEmployees: 15, ActiveEmployees: 10}, nil
}

type mockSettingService struct {
	getAllSettingsFunc func(ctx context.Context) (dto.AppSettingsResponse, error)
	updateAppSettingsFunc func(ctx context.Context, req dto.UpdateAppSettingsRequest) error
}

func (m *mockSettingService) GetAllSettings(ctx context.Context) (dto.AppSettingsResponse, error) {
	if m.getAllSettingsFunc != nil {
		return m.getAllSettingsFunc(ctx)
	}
	return dto.AppSettingsResponse{SiteName: "AAMS Portal"}, nil
}

func (m *mockSettingService) GetSettingByKey(ctx context.Context, key string) (*domainAppSettingMock, error) {
	return nil, nil
}

type domainAppSettingMock struct{}

func (m *mockSettingService) UpdateAppSettings(ctx context.Context, req dto.UpdateAppSettingsRequest) error {
	if m.updateAppSettingsFunc != nil {
		return m.updateAppSettingsFunc(ctx, req)
	}
	return nil
}

func (m *mockSettingService) UpdateSetting(ctx context.Context, key, value string) error {
	return nil
}

func TestDashboardHandler_GetStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dashSvc := &mockDashboardService{}
	h := controller.NewDashboardHandler(dashSvc)

	r := gin.New()
	r.GET("/api/v1/dashboard", h.GetStats)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp dto.DashboardStatsResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.TotalEmployees != 15 {
		t.Errorf("expected 15 employees, got %d", resp.TotalEmployees)
	}
}

func TestSettingHandler_GetPublicSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	settingSvc := &mockSettingService{}
	h := controller.NewSettingHandler(settingSvc)

	r := gin.New()
	r.GET("/api/v1/settings/public", h.GetPublicSettings)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/settings/public", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp dto.AppSettingsResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.SiteName != "AAMS Portal" {
		t.Errorf("expected site_name 'AAMS Portal', got %s", resp.SiteName)
	}
}

func TestSettingHandler_UpdateSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	settingSvc := &mockSettingService{}
	h := controller.NewSettingHandler(settingSvc)

	r := gin.New()
	r.PUT("/api/v1/settings", h.UpdateSettings)

	body, _ := json.Marshal(dto.UpdateAppSettingsRequest{
		SiteName: "AAMS Fleet Hub",
	})
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
}
