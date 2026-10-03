package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/system/domain"
	"delivery-backend/internal/modules/system/dto"
	"delivery-backend/internal/modules/system/repository"
	"delivery-backend/internal/modules/system/service"
)

type mockSystemRepo struct {
	repository.SystemRepository
	logs     []domain.AuditLog
	settings map[string]string
}

func newMockSystemRepo() *mockSystemRepo {
	return &mockSystemRepo{
		settings: make(map[string]string),
	}
}

func (m *mockSystemRepo) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	m.logs = append(m.logs, *log)
	return nil
}

func (m *mockSystemRepo) FindAuditLogs(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error) {
	return m.logs, int64(len(m.logs)), nil
}

func (m *mockSystemRepo) GetSettings(ctx context.Context) ([]domain.AppSetting, error) {
	var list []domain.AppSetting
	for k, v := range m.settings {
		list = append(list, domain.AppSetting{Key: k, Value: v})
	}
	return list, nil
}

func (m *mockSystemRepo) UpsertSetting(ctx context.Context, setting *domain.AppSetting) error {
	m.settings[setting.Key] = setting.Value
	return nil
}

func (m *mockSystemRepo) GetDashboardStats(ctx context.Context, branchID *uuid.UUID) (*dto.DashboardStatsResponse, error) {
	return &dto.DashboardStatsResponse{
		TotalEmployees:   10,
		ActiveEmployees:  8,
		TotalMotorcycles: 5,
		TodayOrders:      120,
	}, nil
}

func TestAuditService_LogAction(t *testing.T) {
	repo := newMockSystemRepo()
	auditSvc := service.NewAuditService(repo)
	ctx := context.Background()

	err := auditSvc.LogAction(ctx, "AdminUser", "LOGIN", "Logged in successfully", "127.0.0.1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logs, total, err := auditSvc.GetLogs(ctx, nil, 1, 10)
	if err != nil || total != 1 {
		t.Fatalf("expected 1 log, got total: %d, err: %v", total, err)
	}
	if logs[0].Action != "LOGIN" {
		t.Errorf("expected action LOGIN, got: %s", logs[0].Action)
	}
}

func TestSettingService_UpdateAndGet(t *testing.T) {
	repo := newMockSystemRepo()
	settingSvc := service.NewSettingService(repo)
	ctx := context.Background()

	err := settingSvc.UpdateAppSettings(ctx, dto.UpdateAppSettingsRequest{
		SiteName: "AAMS Logistics Test",
		LogoURL:  "https://example.com/logo.png",
	})
	if err != nil {
		t.Fatalf("failed to update settings: %v", err)
	}

	settings, err := settingSvc.GetAllSettings(ctx)
	if err != nil {
		t.Fatalf("failed to get settings: %v", err)
	}
	if settings.SiteName != "AAMS Logistics Test" {
		t.Errorf("site name mismatch: %s", settings.SiteName)
	}
}

func TestDashboardService_GetStats(t *testing.T) {
	repo := newMockSystemRepo()
	dashSvc := service.NewDashboardService(repo)
	ctx := context.Background()

	stats, err := dashSvc.GetStats(ctx, nil)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.TotalEmployees != 10 || stats.TodayOrders != 120 {
		t.Errorf("unexpected stats: %+v", stats)
	}
}
