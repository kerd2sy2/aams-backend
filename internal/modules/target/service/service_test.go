package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/target/domain"
	"delivery-backend/internal/modules/target/repository"
	"delivery-backend/internal/modules/target/service"
)

type mockTargetRepo struct {
	repository.TargetRepository
	idents   map[uuid.UUID]*domain.Identifier
	settings map[string]string
}

func newMockTargetRepo() *mockTargetRepo {
	return &mockTargetRepo{
		idents:   make(map[uuid.UUID]*domain.Identifier),
		settings: make(map[string]string),
	}
}

func (m *mockTargetRepo) CreateIdentifier(ctx context.Context, ident *domain.Identifier) error {
	m.idents[ident.ID] = ident
	return nil
}

func (m *mockTargetRepo) FindIdentifierByID(ctx context.Context, id uuid.UUID) (*domain.Identifier, error) {
	if i, ok := m.idents[id]; ok {
		return i, nil
	}
	return nil, nil
}

func (m *mockTargetRepo) FindIdentifierByNameAndApp(ctx context.Context, name, appName string) (*domain.Identifier, error) {
	for _, i := range m.idents {
		if i.Name == name && i.AppName == appName {
			return i, nil
		}
	}
	return nil, nil
}

func (m *mockTargetRepo) ListIdentifiers(ctx context.Context, search string, isActive *bool) ([]domain.Identifier, error) {
	var list []domain.Identifier
	for _, i := range m.idents {
		list = append(list, *i)
	}
	return list, nil
}

func (m *mockTargetRepo) GetOrdersForMonth(ctx context.Context, monthPrefix string) ([]domain.DailyOrder, error) {
	return []domain.DailyOrder{}, nil
}

func (m *mockTargetRepo) ListTargetAlerts(ctx context.Context, alertDate string, unresolvedOnly bool) ([]domain.TargetAlert, error) {
	return []domain.TargetAlert{}, nil
}

func (m *mockTargetRepo) GetTargetSetting(ctx context.Context, key string) (string, error) {
	return m.settings[key], nil
}

func (m *mockTargetRepo) SetTargetSetting(ctx context.Context, key, val string) error {
	m.settings[key] = val
	return nil
}

func (m *mockTargetRepo) UpdateAllIdentifiersTargets(ctx context.Context, monthlyTarget, dailyTarget int) error {
	for _, i := range m.idents {
		if monthlyTarget > 0 {
			i.MonthlyTarget = monthlyTarget
		}
		if dailyTarget > 0 {
			i.DailyTarget = dailyTarget
		}
	}
	return nil
}

func TestTargetService_CreateAndSettings(t *testing.T) {
	repo := newMockTargetRepo()
	svc := service.NewTargetService(repo)
	ctx := context.Background()

	// 1. Create Identifier
	ident, err := svc.CreateIdentifier(ctx, "فهد", "نينجا", "FHD", 500, 20)
	if err != nil {
		t.Fatalf("unexpected error creating identifier: %v", err)
	}
	if ident.Name != "فهد" || ident.MonthlyTarget != 500 {
		t.Errorf("unexpected identifier values: name=%s target=%d", ident.Name, ident.MonthlyTarget)
	}

	// 2. Cannot duplicate
	_, err = svc.CreateIdentifier(ctx, "فهد", "نينجا", "FHD", 500, 20)
	if err == nil {
		t.Fatalf("expected error creating duplicate, got nil")
	}

	// 3. Settings
	_ = svc.UpdateTargetSettings(ctx, 480, 19)
	settings, err := svc.GetTargetSettings(ctx)
	if err != nil {
		t.Fatalf("unexpected error getting settings: %v", err)
	}
	if settings.DefaultMonthlyTarget != 480 || settings.DefaultDailyTarget != 19 {
		t.Errorf("unexpected settings values: %+v", settings)
	}
}
