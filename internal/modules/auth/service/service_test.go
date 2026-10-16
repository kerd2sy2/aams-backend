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
	"delivery-backend/pkg/jwt"
)

var (
	_ repository.AdminRepository  = (*mockAdminRepo)(nil)
	_ repository.RoleRepository   = (*mockRoleRepo)(nil)
	_ repository.BranchRepository = (*mockBranchRepo)(nil)
)

type mockAdminRepo struct {
	admins    map[uuid.UUID]*domain.Admin
	employees map[string]map[string]interface{}
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
func (m *mockAdminRepo) FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*domain.Admin, error) {
	for _, a := range m.admins {
		if a.EmployeeID != nil && *a.EmployeeID == employeeID {
			return a, nil
		}
	}
	return nil, nil
}
func (m *mockAdminRepo) FindEmployeeByLogin(ctx context.Context, login string) (map[string]interface{}, error) {
	if m.employees != nil {
		if emp, ok := m.employees[login]; ok {
			return emp, nil
		}
	}
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

func TestGooglePlayReviewerAccount(t *testing.T) {
	adminRepo := &mockAdminRepo{admins: make(map[uuid.UUID]*domain.Admin)}
	branchRepo := &mockBranchRepo{branches: make(map[uuid.UUID]*domain.Branch)}
	cfg := &config.Config{JWTSecret: "test-secret-key-1234567890"}
	authSvc := service.NewAuthService(adminRepo, branchRepo, cfg)
	ctx := context.Background()

	resp, err := authSvc.Login(ctx, dto.LoginRequest{
		Username: "1234596",
		Password: "1234",
	})
	if err != nil {
		t.Fatalf("expected login to succeed for demo reviewer: %v", err)
	}

	if resp.Type != "employee" {
		t.Errorf("expected type to be employee, got %s", resp.Type)
	}
	if !resp.IsEmployee {
		t.Errorf("expected is_employee to be true")
	}
	if resp.Admin != nil {
		t.Errorf("expected Admin info to be nil for security, but was non-nil")
	}
	if resp.User != nil {
		t.Errorf("expected User info to be nil for security, but was non-nil")
	}
	empMap, ok := resp.Employee.(map[string]interface{})
	if !ok {
		t.Fatalf("expected employee payload to be map")
	}
	if empMap["role"] != "DRIVER" {
		t.Errorf("expected role DRIVER, got %v", empMap["role"])
	}
	claims, err := jwt.ValidateToken(resp.Token, cfg.JWTSecret)
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}
	if claims.Role != "EMPLOYEE" {
		t.Errorf("expected token role to be EMPLOYEE, got %s", claims.Role)
	}
}

func TestPromoteEmployee(t *testing.T) {
	empID := uuid.New()
	adminRepo := &mockAdminRepo{
		admins: make(map[uuid.UUID]*domain.Admin),
		employees: map[string]map[string]interface{}{
			empID.String(): {
				"id":              empID.String(),
				"name":            "أحمد المندوب",
				"phone":           "0551234567",
				"employee_number": "EMP-100",
			},
		},
	}
	adminSvc := service.NewAdminService(adminRepo)
	ctx := context.Background()

	// 1. First promotion: creates new Admin linked to EmployeeID with must_change_password=true
	admin, err := adminSvc.PromoteEmployee(ctx, dto.PromoteEmployeeRequest{
		EmployeeID:  empID,
		Role:        "SUPERVISOR",
		Permissions: []string{"work.view", "work.start"},
	})
	if err != nil {
		t.Fatalf("promotion failed: %v", err)
	}
	if admin.Role != "SUPERVISOR" {
		t.Errorf("expected role SUPERVISOR, got %s", admin.Role)
	}
	if admin.EmployeeID == nil || *admin.EmployeeID != empID {
		t.Errorf("expected linked EmployeeID %s", empID)
	}
	if !admin.MustChangePassword {
		t.Errorf("expected MustChangePassword to be true for new promoted admin")
	}
	if !admin.PermissionsConfigured {
		t.Errorf("expected PermissionsConfigured to be true")
	}

	// 2. Second promotion: updates existing Admin without duplicating
	updatedAdmin, err := adminSvc.PromoteEmployee(ctx, dto.PromoteEmployeeRequest{
		EmployeeID:  empID,
		Role:        "ADMIN",
		Permissions: []string{"*"},
	})
	if err != nil {
		t.Fatalf("second promotion failed: %v", err)
	}
	if updatedAdmin.ID != admin.ID {
		t.Errorf("expected same admin ID %s, got %s (duplicate created)", admin.ID, updatedAdmin.ID)
	}
	if updatedAdmin.Role != "ADMIN" {
		t.Errorf("expected updated role ADMIN, got %s", updatedAdmin.Role)
	}
	if len(adminRepo.admins) != 1 {
		t.Errorf("expected exactly 1 admin in repo, got %d", len(adminRepo.admins))
	}
}

func TestHasAppAccess(t *testing.T) {
	// 1. Super Admin has access to all apps
	if !service.HasAppAccess("SUPER_ADMIN", []string{}, "ohda") {
		t.Errorf("SUPER_ADMIN should have access to ohda")
	}
	if !service.HasAppAccess("SUPER_ADMIN", []string{}, "eshraf") {
		t.Errorf("SUPER_ADMIN should have access to eshraf")
	}

	// 2. Granular permission apps.ohda
	if !service.HasAppAccess("ACCOUNTANT", []string{"apps.ohda"}, "ohda") {
		t.Errorf("user with apps.ohda should access ohda")
	}
	if service.HasAppAccess("ACCOUNTANT", []string{"apps.ohda"}, "edara") {
		t.Errorf("user with only apps.ohda should not access edara")
	}

	// 3. Driver default access
	if !service.HasAppAccess("DRIVER", []string{}, "manadeeb") {
		t.Errorf("DRIVER should have access to manadeeb")
	}
	if service.HasAppAccess("DRIVER", []string{}, "dashboard") {
		t.Errorf("DRIVER should not have access to dashboard without permissions")
	}
}

func TestLogin_AppPermissionEnforcement(t *testing.T) {
	adminRepo := &mockAdminRepo{admins: make(map[uuid.UUID]*domain.Admin)}
	branchRepo := &mockBranchRepo{branches: make(map[uuid.UUID]*domain.Branch)}
	cfg := &config.Config{JWTSecret: "test-secret-key-1234567890123456"}
	svc := service.NewAuthService(adminRepo, branchRepo, cfg)

	// Admin with only custody permissions
	hash, _ := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)
	adminID := uuid.New()
	adminRepo.admins[adminID] = &domain.Admin{
		ID:          adminID,
		Username:    "custody_user",
		Phone:       "0555551234",
		Password:    string(hash),
		Role:        "ACCOUNTANT",
		Permissions: `["apps.ohda","custody.view"]`,
	}

	ctx := context.Background()

	// Allowed app login
	resp, err := svc.Login(ctx, dto.LoginRequest{
		Phone:    "0555551234",
		Password: "1234",
		App:      "ohda",
	})
	if err != nil {
		t.Fatalf("expected successful login to ohda, got error: %v", err)
	}
	if resp.AppPermissions == nil || !resp.AppPermissions["ohda"] {
		t.Errorf("expected ohda permission in app_permissions map")
	}

	// Denied app login
	_, err = svc.Login(ctx, dto.LoginRequest{
		Phone:    "0555551234",
		Password: "1234",
		App:      "edara",
	})
	if err == nil {
		t.Fatalf("expected error when logging in to edara without permission, got nil")
	}
}
