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

	"delivery-backend/internal/modules/auth/controller"
	"delivery-backend/internal/modules/auth/domain"
	"delivery-backend/internal/modules/auth/dto"
)

type mockBranchService struct {
	createFunc func(ctx context.Context, req dto.CreateBranchRequest) (*domain.Branch, error)
	findAllFunc func(ctx context.Context) ([]dto.BranchResponse, error)
	findByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.Branch, error)
	updateFunc func(ctx context.Context, id uuid.UUID, req dto.UpdateBranchRequest) (*domain.Branch, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockBranchService) Create(ctx context.Context, req dto.CreateBranchRequest) (*domain.Branch, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, req)
	}
	return &domain.Branch{ID: uuid.New(), Name: req.Name}, nil
}
func (m *mockBranchService) FindAll(ctx context.Context) ([]dto.BranchResponse, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return []dto.BranchResponse{}, nil
}
func (m *mockBranchService) FindByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return &domain.Branch{ID: id, Name: "Test Branch"}, nil
}
func (m *mockBranchService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateBranchRequest) (*domain.Branch, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, req)
	}
	return &domain.Branch{ID: id, Name: req.Name}, nil
}
func (m *mockBranchService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

type mockRoleService struct {
	createFunc func(ctx context.Context, req dto.CreateRoleRequest) (*domain.Role, error)
	findAllFunc func(ctx context.Context) ([]dto.RoleResponse, error)
	findByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	updateFunc func(ctx context.Context, id uuid.UUID, req dto.UpdateRoleRequest) (*domain.Role, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockRoleService) Create(ctx context.Context, req dto.CreateRoleRequest) (*domain.Role, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, req)
	}
	return &domain.Role{ID: uuid.New(), Name: req.Name, Description: req.Description}, nil
}
func (m *mockRoleService) FindAll(ctx context.Context) ([]dto.RoleResponse, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx)
	}
	return []dto.RoleResponse{}, nil
}
func (m *mockRoleService) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return &domain.Role{ID: id, Name: "Test Role"}, nil
}
func (m *mockRoleService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateRoleRequest) (*domain.Role, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, req)
	}
	return &domain.Role{ID: id, DisplayName: req.DisplayName}, nil
}
func (m *mockRoleService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestBranchHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockBranchService{}
	h := controller.NewBranchHandler(mockSvc)

	r := gin.New()
	r.POST("/api/v1/branches", h.Create)

	reqBody, _ := json.Marshal(dto.CreateBranchRequest{
		Name: "Jeddah Branch",
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/branches", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRoleHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockRoleService{}
	h := controller.NewRoleHandler(mockSvc)

	r := gin.New()
	r.POST("/api/v1/roles", h.Create)

	reqBody, _ := json.Marshal(dto.CreateRoleRequest{
		Name:        "Supervisor",
		DisplayName: "مشرف",
		Description: "Shift supervisor",
		Permissions: []string{"read", "edit"},
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}
}
