package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"delivery-backend/internal/modules/auth/domain"
	"delivery-backend/internal/modules/auth/dto"
	"delivery-backend/internal/modules/auth/repository"
	"delivery-backend/internal/modules/auth/service"
	"delivery-backend/pkg/config"
)

var (
	_ repository.AdminRepository  = (*mockAdminRepo)(nil)
	_ repository.RoleRepository   = (*mockRoleRepo)(nil)
	_ repository.BranchRepository = (*mockBranchRepo)(nil)
)

type mockAdminRepo struct {
	admins map[uuid.UUID]*domain.Admin
}

func (m *mockAdminRepo) Create(ctx context.Context, admin *domain.Admin) error {
	m.admins[admin.ID] = admin
	return nil
}
func (m *mockAdminRepo) Update(ctx context.Context, admin *domain.Admin) error {
	m.admins[admin.ID] = admin
	return nil
}
func (m *mockAdminRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.admins, id)
	return nil
}
func (m *mockAdminRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error) {
	if a, ok := m.admins[id]; ok {
		return a, nil
	}
	return nil, nil
}
func (m *mockAdminRepo) FindByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	for _, a := range m.admins {
		if a.Email == email {
			return a, nil
		}
	}
	return nil, nil
}
func (m *mockAdminRepo) FindByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	for _, a := range m.admins {
		if a.Username == username {
			return a, nil
		}
	}
	return nil, nil
}
func (m *mockAdminRepo) FindByPhone(ctx context.Context, phone string) (*domain.Admin, error) {
	for _, a := range m.admins {
		if a.Phone == phone {
			return a, nil
		}
	}
	return nil, nil
}
func (m *mockAdminRepo) FindByGoogleID(ctx context.Context, googleID string) (*domain.Admin, error) {
	for _, a := range m.admins {
		if a.GoogleID == googleID {
			return a, nil
		}
	}
	return nil, nil
}
func (m *mockAdminRepo) FindAll(ctx context.Context) ([]domain.Admin, error) {
	var list []domain.Admin
	for _, a := range m.admins {
		list = append(list, *a)
	}
	return list, nil
}
func (m *mockAdminRepo) FindEmployeeByLogin(ctx context.Context, login string) (map[string]interface{}, error) {
	return nil, nil
}

type mockRoleRepo struct {
	roles map[uuid.UUID]*domain.Role
}

func (m *mockRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockRoleRepo) Update(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockRoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.roles, id)
	return nil
}
func (m *mockRoleRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	if r, ok := m.roles[id]; ok {
		return r, nil
	}
	return nil, nil
}
func (m *mockRoleRepo) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, nil
}
func (m *mockRoleRepo) FindAll(ctx context.Context) ([]domain.Role, error) {
	var list []domain.Role
	for _, r := range m.roles {
		list = append(list, *r)
	}
	return list, nil
}
func (m *mockRoleRepo) CountUsersByRoleID(ctx context.Context, roleID uuid.UUID) (int64, error) {
	return 0, nil
}

type mockBranchRepo struct {
	branches map[uuid.UUID]*domain.Branch
}

func (m *mockBranchRepo) Create(ctx context.Context, branch *domain.Branch) error {
	m.branches[branch.ID] = branch
	return nil
}
func (m *mockBranchRepo) Update(ctx context.Context, branch *domain.Branch) error {
	m.branches[branch.ID] = branch
	return nil
}
func (m *mockBranchRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.branches, id)
	return nil
}
func (m *mockBranchRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error) {
	if b, ok := m.branches[id]; ok {
		return b, nil
	}
	return nil, nil
}
func (m *mockBranchRepo) FindByName(ctx context.Context, name string) (*domain.Branch, error) {
	for _, b := range m.branches {
		if b.Name == name {
			return b, nil
		}
	}
	return nil, nil
}
func (m *mockBranchRepo) FindAll(ctx context.Context) ([]domain.Branch, error) {
	var list []domain.Branch
	for _, b := range m.branches {
		list = append(list, *b)
	}
	return list, nil
}
func (m *mockBranchRepo) CountEmployeesByBranchID(ctx context.Context, branchID uuid.UUID) (int64, error) {
	return 0, nil
}

func TestAuthService_Login(t *testing.T) {
	adminRepo := &mockAdminRepo{admins: make(map[uuid.UUID]*domain.Admin)}
	branchRepo := &mockBranchRepo{branches: make(map[uuid.UUID]*domain.Branch)}
	cfg := &config.Config{
		JWTSecret: "test-secret-key-1234567890",
	}
	authSvc := service.NewAuthService(adminRepo, branchRepo, cfg)
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	admin := &domain.Admin{
		ID:       uuid.New(),
		Username: "admin",
		Email:    "admin@aams.com",
		Password: string(hash),
		Role:     "admin",
	}
	_ = adminRepo.Create(ctx, admin)

	// 1. Success login
	resp, err := authSvc.Login(ctx, dto.LoginRequest{
		Username: "admin",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}

	// 2. Invalid password
	_, err = authSvc.Login(ctx, dto.LoginRequest{
		Username: "admin",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Error("expected error for invalid password")
	}
}

func TestRoleAndBranchService(t *testing.T) {
	roleRepo := &mockRoleRepo{roles: make(map[uuid.UUID]*domain.Role)}
	branchRepo := &mockBranchRepo{branches: make(map[uuid.UUID]*domain.Branch)}
	roleSvc := service.NewRoleService(roleRepo)
	branchSvc := service.NewBranchService(branchRepo)
	ctx := context.Background()

	// 1. Create Role
	role, err := roleSvc.Create(ctx, dto.CreateRoleRequest{
		Name:        "Manager",
		DisplayName: "مدير",
		Description: "Manager role",
		Permissions: []string{"read", "write"},
	})
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}
	if role.Name != "Manager" {
		t.Errorf("role name mismatch: %s", role.Name)
	}

	// 2. Create Branch
	branch, err := branchSvc.Create(ctx, dto.CreateBranchRequest{
		Name: "Riyadh Main",
	})
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}
	if branch.Name != "Riyadh Main" {
		t.Errorf("branch name mismatch: %s", branch.Name)
	}
}
