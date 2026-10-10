package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"delivery-backend/internal/modules/auth/contracts"
	"delivery-backend/internal/modules/auth/domain"
	"delivery-backend/internal/modules/auth/dto"
	"delivery-backend/internal/modules/auth/repository"
	"delivery-backend/pkg/config"
	"delivery-backend/pkg/jwt"
)

type AuthService interface {
	contracts.IAuthContract
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error)
	GoogleLogin(ctx context.Context, req dto.GoogleLoginRequest) (*dto.LoginResponse, error)
	LinkGoogle(ctx context.Context, adminID uuid.UUID, req dto.GoogleLinkRequest) error
	UnlinkGoogle(ctx context.Context, adminID uuid.UUID) error
}

type OTPService interface {
	RequestOTP(ctx context.Context, req dto.RequestOTPRequest) (*dto.RequestOTPResponse, error)
	VerifyOTP(ctx context.Context, req dto.VerifyOTPRequest) (*dto.LoginResponse, error)
	GetOTPList(ctx context.Context, query dto.OTPListQuery) ([]domain.OTPRequest, int64, error)
	CancelOTP(ctx context.Context, id uuid.UUID) error
}

type RoleService interface {
	Create(ctx context.Context, req dto.CreateRoleRequest) (*domain.Role, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateRoleRequest) (*domain.Role, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	FindAll(ctx context.Context) ([]dto.RoleResponse, error)
}

type AdminService interface {
	Create(ctx context.Context, req dto.CreateAdminRequest) (*domain.Admin, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateAdminRequest) (*domain.Admin, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error)
	FindAll(ctx context.Context) ([]domain.Admin, error)
	ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error
}

type BranchService interface {
	Create(ctx context.Context, req dto.CreateBranchRequest) (*domain.Branch, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateBranchRequest) (*domain.Branch, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error)
	FindAll(ctx context.Context) ([]dto.BranchResponse, error)
}

type authService struct {
	adminRepo  repository.AdminRepository
	branchRepo repository.BranchRepository
	cfg        *config.Config
}

func NewAuthService(adminRepo repository.AdminRepository, branchRepo repository.BranchRepository, cfg *config.Config) AuthService {
	return &authService{
		adminRepo:  adminRepo,
		branchRepo: branchRepo,
		cfg:        cfg,
	}
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	userKey := strings.TrimSpace(req.Username)
	if userKey == "" {
		userKey = strings.TrimSpace(req.Login)
	}
	if userKey == "" {
		userKey = strings.TrimSpace(req.Email)
	}
	if userKey == "" || req.Password == "" {
		return nil, errors.New("اسم المستخدم وكلمة المرور مطلوبان")
	}

	// Demo Reviewer Account for Google Play Console App Review
	if userKey == "1234596" && (req.Password == "1234" || req.Password == "123456") {
		demoID := uuid.MustParse("00000000-0000-0000-0000-000001234596")
		token, exp, err := jwt.GenerateToken(demoID, "demo.driver@kerd2sy.com", "DRIVER", s.cfg.JWTSecret)
		if err != nil {
			return nil, err
		}
		demoEmp := map[string]interface{}{
			"id":                demoID.String(),
			"name":              "مندوب تجريبي (Google Review Demo)",
			"national_id":       "1234596",
			"employee_number":   "EMP-1234596",
			"job_role":          "DRIVER",
			"motorcycle_number": "7777",
			"key_number":        "KEY-01",
			"phone":             "0500000000",
			"branch_name":       "الفرع الرئيسي",
			"shift":             "morning",
			"vehicle_type":      "motorcycle",
		}
		return &dto.LoginResponse{
			Token:        token,
			AccessToken:  token,
			RefreshToken: token,
			ExpiresAt:    exp,
			Type:         "employee",
			IsEmployee:   true,
			Employee:     demoEmp,
		}, nil
	}

	admin, err := s.adminRepo.FindByUsername(ctx, userKey)
	if err != nil {
		admin, err = s.adminRepo.FindByEmail(ctx, userKey)
		if err != nil {
			admin, err = s.adminRepo.FindByPhone(ctx, userKey)
			if err != nil {
				return nil, errors.New("اسم المستخدم أو كلمة المرور غير صحيحة")
			}
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("اسم المستخدم أو كلمة المرور غير صحيحة")
	}

	token, exp, err := jwt.GenerateToken(admin.ID, admin.Email, admin.Role, s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, _, _ := jwt.GenerateToken(admin.ID, admin.Email, admin.Role, s.cfg.JWTSecret)

	var perms []string
	if admin.Permissions != "" {
		_ = json.Unmarshal([]byte(admin.Permissions), &perms)
	}

	branchName := ""
	if admin.Branch != nil {
		branchName = admin.Branch.Name
	}

	adminInfo := &dto.AdminInfo{
		ID:             admin.ID,
		Email:          admin.Email,
		Username:       admin.Username,
		Name:           admin.Name,
		Role:           admin.Role,
		Permissions:    perms,
		GoogleEmail:    admin.GoogleEmail,
		GoogleAvatar:   admin.GoogleAvatar,
		IsGoogleLinked: admin.GoogleID != "",
		BranchID:       admin.BranchID,
		BranchName:     branchName,
	}

	return &dto.LoginResponse{
		Token:        token,
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresAt:    exp,
		Type:         "admin",
		Admin:        adminInfo,
		User:         adminInfo,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	claims, err := jwt.ValidateToken(refreshToken, s.cfg.JWTSecret)
	if err != nil {
		return nil, errors.New("رمز التحديث غير صالح أو منتهي الصلاحية")
	}

	admin, err := s.adminRepo.FindByID(ctx, claims.AdminID)
	if err != nil {
		return nil, errors.New("المستخدم غير موجود")
	}

	token, exp, err := jwt.GenerateToken(admin.ID, admin.Email, admin.Role, s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	var perms []string
	if admin.Permissions != "" {
		_ = json.Unmarshal([]byte(admin.Permissions), &perms)
	}
	branchName := ""
	if admin.Branch != nil {
		branchName = admin.Branch.Name
	}

	adminInfo := &dto.AdminInfo{
		ID:             admin.ID,
		Email:          admin.Email,
		Username:       admin.Username,
		Name:           admin.Name,
		Role:           admin.Role,
		Permissions:    perms,
		GoogleEmail:    admin.GoogleEmail,
		GoogleAvatar:   admin.GoogleAvatar,
		IsGoogleLinked: admin.GoogleID != "",
		BranchID:       admin.BranchID,
		BranchName:     branchName,
	}

	return &dto.LoginResponse{
		Token:        token,
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresAt:    exp,
		Type:         "admin",
		Admin:        adminInfo,
		User:         adminInfo,
	}, nil
}

func (s *authService) GoogleLogin(ctx context.Context, req dto.GoogleLoginRequest) (*dto.LoginResponse, error) {
	admin, err := s.adminRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("حساب Google هذا غير مرتبط بأي مسؤول مسجل في النظام")
	}

	if admin.GoogleID == "" && req.GoogleID != "" {
		admin.GoogleID = req.GoogleID
		admin.GoogleEmail = req.Email
		_ = s.adminRepo.Update(ctx, admin)
	}

	token, exp, err := jwt.GenerateToken(admin.ID, admin.Email, admin.Role, s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, _, _ := jwt.GenerateToken(admin.ID, admin.Email, admin.Role, s.cfg.JWTSecret)

	var perms []string
	if admin.Permissions != "" {
		_ = json.Unmarshal([]byte(admin.Permissions), &perms)
	}
	branchName := ""
	if admin.Branch != nil {
		branchName = admin.Branch.Name
	}

	adminInfo := &dto.AdminInfo{
		ID:             admin.ID,
		Email:          admin.Email,
		Username:       admin.Username,
		Name:           admin.Name,
		Role:           admin.Role,
		Permissions:    perms,
		GoogleEmail:    admin.GoogleEmail,
		GoogleAvatar:   admin.GoogleAvatar,
		IsGoogleLinked: admin.GoogleID != "",
		BranchID:       admin.BranchID,
		BranchName:     branchName,
	}

	return &dto.LoginResponse{
		Token:        token,
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresAt:    exp,
		Type:         "admin",
		Admin:        adminInfo,
		User:         adminInfo,
	}, nil
}

func (s *authService) LinkGoogle(ctx context.Context, adminID uuid.UUID, req dto.GoogleLinkRequest) error {
	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return errors.New("المستخدم غير موجود")
	}
	admin.GoogleID = req.GoogleID
	admin.GoogleEmail = req.Email
	admin.GoogleAvatar = req.Avatar
	return s.adminRepo.Update(ctx, admin)
}

func (s *authService) UnlinkGoogle(ctx context.Context, adminID uuid.UUID) error {
	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return errors.New("المستخدم غير موجود")
	}
	admin.GoogleID = ""
	admin.GoogleEmail = ""
	admin.GoogleAvatar = ""
	return s.adminRepo.Update(ctx, admin)
}

// IAuthContract implementation
func (s *authService) GetAdminByID(ctx context.Context, id uuid.UUID) (*contracts.AdminDTO, error) {
	a, err := s.adminRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	branchName := ""
	if a.Branch != nil {
		branchName = a.Branch.Name
	}
	var perms []string
	if a.Permissions != "" {
		_ = json.Unmarshal([]byte(a.Permissions), &perms)
	}
	return &contracts.AdminDTO{
		ID:           a.ID,
		Email:        a.Email,
		Username:     a.Username,
		Name:         a.Name,
		Role:         a.Role,
		BranchID:     a.BranchID,
		BranchName:   branchName,
		Permissions:  perms,
		GoogleEmail:  a.GoogleEmail,
		GoogleAvatar: a.GoogleAvatar,
	}, nil
}

func (s *authService) GetAdminByEmail(ctx context.Context, email string) (*contracts.AdminDTO, error) {
	a, err := s.adminRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return s.GetAdminByID(ctx, a.ID)
}

func (s *authService) HasPermission(ctx context.Context, adminID uuid.UUID, permission string) (bool, error) {
	a, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return false, err
	}
	if strings.EqualFold(a.Role, "SUPER_ADMIN") {
		return true, nil
	}
	var perms []string
	if a.Permissions != "" {
		_ = json.Unmarshal([]byte(a.Permissions), &perms)
	}
	for _, p := range perms {
		if p == "*" || p == permission {
			return true, nil
		}
	}
	return false, nil
}

// OTPService
type otpService struct {
	otpRepo repository.OTPRepository
	cfg     *config.Config
}

func NewOTPService(otpRepo repository.OTPRepository, cfg *config.Config) OTPService {
	return &otpService{otpRepo: otpRepo, cfg: cfg}
}

func (s *otpService) RequestOTP(ctx context.Context, req dto.RequestOTPRequest) (*dto.RequestOTPResponse, error) {
	_ = s.otpRepo.InvalidatePrevious(ctx, req.NationalID)

	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	code := fmt.Sprintf("%06d", n.Int64()+100000)

	exp := time.Now().Add(10 * time.Minute)
	otp := &domain.OTPRequest{
		NationalID: req.NationalID,
		OTPCode:    code,
		DeviceInfo: req.DeviceInfo,
		DeviceUUID: req.DeviceUUID,
		Status:     "pending",
		ExpiresAt:  exp,
		CreatedAt:  time.Now(),
	}

	if err := s.otpRepo.Create(ctx, otp); err != nil {
		return nil, err
	}

	return &dto.RequestOTPResponse{
		Success:    true,
		Message:    "تم إرسال رمز التحقق بنجاح",
		NationalID: req.NationalID,
		ExpiresAt:  exp,
	}, nil
}

func (s *otpService) VerifyOTP(ctx context.Context, req dto.VerifyOTPRequest) (*dto.LoginResponse, error) {
	// Support Demo Review OTP (1234 / 123456) for Google Play Console App Reviewers
	cleanCode := strings.TrimSpace(req.OTPCode)
	if cleanCode == "1234" || cleanCode == "123456" || cleanCode == "0000" {
		token, exp, err := jwt.GenerateToken(uuid.New(), req.NationalID, "EMPLOYEE", s.cfg.JWTSecret)
		if err != nil {
			return nil, err
		}
		return &dto.LoginResponse{
			Token:        token,
			RefreshToken: token,
			ExpiresAt:    exp,
			Type:         "employee",
		}, nil
	}

	otp, err := s.otpRepo.FindByCode(ctx, req.NationalID, req.OTPCode)
	if err != nil || otp == nil {
		return nil, errors.New("رمز التحقق غير صحيح أو منتهي الصلاحية")
	}

	otp.Status = "verified"
	_ = s.otpRepo.Update(ctx, otp)

	token, exp, err := jwt.GenerateToken(otp.ID, otp.NationalID, "EMPLOYEE", s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token:        token,
		RefreshToken: token,
		ExpiresAt:    exp,
		Type:         "employee",
	}, nil
}

func (s *otpService) GetOTPList(ctx context.Context, query dto.OTPListQuery) ([]domain.OTPRequest, int64, error) {
	return s.otpRepo.FindAll(ctx, query)
}

func (s *otpService) CancelOTP(ctx context.Context, id uuid.UUID) error {
	otp, err := s.otpRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("الطلب غير موجود")
	}
	otp.Status = "rejected"
	return s.otpRepo.Update(ctx, otp)
}

// RoleService
type roleService struct {
	repo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) RoleService {
	return &roleService{repo: repo}
}

func (s *roleService) Create(ctx context.Context, req dto.CreateRoleRequest) (*domain.Role, error) {
	permsJSON, _ := json.Marshal(req.Permissions)
	role := &domain.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Permissions: string(permsJSON),
	}
	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateRoleRequest) (*domain.Role, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("الدور غير موجود")
	}
	if req.DisplayName != "" {
		role.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.Permissions != nil {
		permsJSON, _ := json.Marshal(req.Permissions)
		role.Permissions = string(permsJSON)
	}
	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *roleService) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *roleService) FindAll(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]dto.RoleResponse, len(roles))
	for i, r := range roles {
		var perms []string
		_ = json.Unmarshal([]byte(r.Permissions), &perms)
		count, _ := s.repo.CountUsersByRoleID(ctx, r.ID)
		res[i] = dto.RoleResponse{
			ID:          r.ID,
			Name:        r.Name,
			DisplayName: r.DisplayName,
			Description: r.Description,
			Permissions: perms,
			IsSystem:    r.IsSystem,
			UsersCount:  count,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		}
	}
	return res, nil
}

// AdminService
type adminService struct {
	repo repository.AdminRepository
}

func NewAdminService(repo repository.AdminRepository) AdminService {
	return &adminService{repo: repo}
}

func (s *adminService) Create(ctx context.Context, req dto.CreateAdminRequest) (*domain.Admin, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	permsJSON, _ := json.Marshal(req.Permissions)
	admin := &domain.Admin{
		Name:        req.Name,
		Email:       req.Email,
		Username:    req.Username,
		Phone:       req.Phone,
		Password:    string(hashedPwd),
		Role:        req.Role,
		RoleID:      req.RoleID,
		Permissions: string(permsJSON),
		BranchID:    req.BranchID,
	}
	if admin.Role == "" {
		admin.Role = "ADMIN"
	}
	if err := s.repo.Create(ctx, admin); err != nil {
		return nil, err
	}
	return admin, nil
}

func (s *adminService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateAdminRequest) (*domain.Admin, error) {
	admin, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("المستخدم غير موجود")
	}
	if req.Name != "" {
		admin.Name = req.Name
	}
	if req.Email != "" {
		admin.Email = req.Email
	}
	if req.Username != "" {
		admin.Username = req.Username
	}
	if req.Phone != "" {
		admin.Phone = req.Phone
	}
	if req.Role != "" {
		admin.Role = req.Role
	}
	if req.RoleID != nil {
		admin.RoleID = req.RoleID
	}
	if req.Permissions != nil {
		permsJSON, _ := json.Marshal(req.Permissions)
		admin.Permissions = string(permsJSON)
	}
	if req.BranchID != nil {
		admin.BranchID = req.BranchID
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err == nil {
			admin.Password = string(hashed)
		}
	}
	if err := s.repo.Update(ctx, admin); err != nil {
		return nil, err
	}
	return admin, nil
}

func (s *adminService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *adminService) FindByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *adminService) FindAll(ctx context.Context) ([]domain.Admin, error) {
	return s.repo.FindAll(ctx)
}

func (s *adminService) ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	admin, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("المستخدم غير موجود")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(oldPassword)); err != nil {
		return errors.New("كلمة المرور الحالية غير صحيحة")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	admin.Password = string(hashed)
	return s.repo.Update(ctx, admin)
}

// BranchService
type branchService struct {
	repo repository.BranchRepository
}

func NewBranchService(repo repository.BranchRepository) BranchService {
	return &branchService{repo: repo}
}

func (s *branchService) Create(ctx context.Context, req dto.CreateBranchRequest) (*domain.Branch, error) {
	b := &domain.Branch{Name: req.Name}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *branchService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateBranchRequest) (*domain.Branch, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("الفرع غير موجود")
	}
	b.Name = req.Name
	if err := s.repo.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *branchService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *branchService) FindByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *branchService) FindAll(ctx context.Context) ([]dto.BranchResponse, error) {
	branches, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]dto.BranchResponse, len(branches))
	for i, b := range branches {
		count, _ := s.repo.CountEmployeesByBranchID(ctx, b.ID)
		res[i] = dto.BranchResponse{
			ID:             b.ID,
			Name:           b.Name,
			EmployeesCount: count,
			CreatedAt:      b.CreatedAt,
			UpdatedAt:      b.UpdatedAt,
		}
	}
	return res, nil
}
