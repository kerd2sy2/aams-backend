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

type mockAuthRepo struct {
	repository.AuthRepository
	admins      map[uuid.UUID]*domain.Admin
	roles       map[uuid.UUID]*domain.Role
	branches    map[uuid.UUID]*domain.Branch
	otpRequests map[uuid.UUID]*domain.OTPRequest
}

func newMockAuthRepo() *mockAuthRepo {
	return &mockAuthRepo{
		admins:      make(map[uuid.UUID]*domain.Admin),
		roles:       make(map[uuid.UUID]*domain.Role),
		branches:    make(map[uuid.UUID]*domain.Branch),
		otpRequests: make(map[uuid.UUID]*domain.OTPRequest),
	}
}

func (m *mockAuthRepo) CreateAdmin(ctx context.Context, admin *domain.Admin) error {
	m.admins[admin.ID] = admin
	return nil
}

func (m *mockAuthRepo) FindAdminByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error) {
	if a, ok := m.admins[id]; ok {
		return a, nil
	}
	return nil, nil
}

func (m *mockAuthRepo) FindAdminByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	for _, a := range m.admins {
		if a.Username == username {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockAuthRepo) FindAdminByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	for _, a := range m.admins {
		if a.Email == email {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockAuthRepo) UpdateAdmin(ctx context.Context, admin *domain.Admin) error {
	m.admins[admin.ID] = admin
	return nil
}

func (m *mockAuthRepo) DeleteAdmin(ctx context.Context, id uuid.UUID) error {
	delete(m.admins, id)
	return nil
}

func (m *mockAuthRepo) FindAllAdmins(ctx context.Context) ([]domain.Admin, error) {
	var list []domain.Admin
	for _, a := range m.admins {
		list = append(list, *a)
	}
	return list, nil
}

func (m *mockAuthRepo) CreateRole(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}

func (m *mockAuthRepo) FindRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	if r, ok := m.roles[id]; ok {
		return r, nil
	}
	return nil, nil
}

func (m *mockAuthRepo) FindRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, nil
}

func (m *mockAuthRepo) UpdateRole(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}

func (m *mockAuthRepo) DeleteRole(ctx context.Context, id uuid.UUID) error {
	delete(m.roles, id)
	return nil
}

func (m *mockAuthRepo) FindAllRoles(ctx context.Context) ([]domain.Role, error) {
	var list []domain.Role
	for _, r := range m.roles {
		list = append(list, *r)
	}
	return list, nil
}

func (m *mockAuthRepo) CreateBranch(ctx context.Context, branch *domain.Branch) error {
	m.branches[branch.ID] = branch
	return nil
}

func (m *mockAuthRepo) FindBranchByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error) {
	if b, ok := m.branches[id]; ok {
		return b, nil
	}
	return nil, nil
}

func (m *mockAuthRepo) FindBranchByName(ctx context.Context, name string) (*domain.Branch, error) {
	for _, b := range m.branches {
		if b.Name == name {
			return b, nil
		}
	}
	return nil, nil
}

func (m *mockAuthRepo) UpdateBranch(ctx context.Context, branch *domain.Branch) error {
	m.branches[branch.ID] = branch
	return nil
}

func (m *mockAuthRepo) DeleteBranch(ctx context.Context, id uuid.UUID) error {
	delete(m.branches, id)
	return nil
}

func (m *mockAuthRepo) FindAllBranches(ctx context.Context) ([]domain.Branch, error) {
	var list []domain.Branch
	for _, b := range m.branches {
		list = append(list, *b)
	}
	return list, nil
}

func TestAuthService_Login(t *testing.T) {
	repo := newMockAuthRepo()
	cfg := &config.Config{
		JWTSecret: "test-secret-key-1234567890",
	}
	authSvc := service.NewAuthService(repo, cfg)
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	admin := &domain.Admin{
		ID:           uuid.New(),
		Username:     "admin",
		Email:        "admin@aams.com",
		PasswordHash: string(hash),
		Role:         "admin",
	}
	_ = repo.CreateAdmin(ctx, admin)

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
	repo := newMockAuthRepo()
	roleSvc := service.NewRoleService(repo)
	branchSvc := service.NewBranchService(repo)
	ctx := context.Background()

	// 1. Create Role
	role, err := roleSvc.Create(ctx, dto.CreateRoleRequest{
		Name:        "Manager",
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
		City: "Riyadh",
	})
	if err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}
	if branch.Name != "Riyadh Main" {
		t.Errorf("branch name mismatch: %s", branch.Name)
	}
}
