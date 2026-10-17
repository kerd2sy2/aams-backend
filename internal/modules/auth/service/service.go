package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
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
	GetDevices(ctx context.Context, nationalID string) ([]map[string]interface{}, error)
	RevokeDevice(ctx context.Context, nationalID, deviceUUID string) error
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
	PromoteEmployee(ctx context.Context, req dto.PromoteEmployeeRequest) (*domain.Admin, error)
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

func extractDigits(s string) string {
	eastern := []string{"٠", "١", "٢", "٣", "٤", "٥", "٦", "٧", "٨", "٩"}
	persian := []string{"۰", "۱", "۲", "۳", "۴", "۵", "۶", "۷", "۸", "۹"}
	for i := 0; i <= 9; i++ {
		s = strings.ReplaceAll(s, eastern[i], strconv.Itoa(i))
		s = strings.ReplaceAll(s, persian[i], strconv.Itoa(i))
	}
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func GetAppArabicName(appName string) string {
	switch strings.ToLower(strings.TrimSpace(appName)) {
	case "ohda", "custody":
		return "تطبيق العهدة"
	case "eshraf", "supervision":
		return "تطبيق الإشراف"
	case "edara", "admin", "management":
		return "تطبيق الإدارة"
	case "manadeeb", "delegate", "driver":
		return "تطبيق المناديب"
	case "dashboard":
		return "لوحة التحكم"
	default:
		return appName
	}
}

func HasAppAccess(role string, permissions []string, appName string) bool {
	cleanApp := strings.ToLower(strings.TrimSpace(appName))
	if cleanApp == "" {
		return true
	}

	normRole := strings.ToUpper(strings.TrimSpace(role))
	if normRole == "SUPER_ADMIN" {
		return true
	}

	hasPerm := func(target string) bool {
		for _, p := range permissions {
			pClean := strings.ToLower(strings.TrimSpace(p))
			if pClean == "*" || pClean == target {
				return true
			}
		}
		return false
	}

	switch cleanApp {
	case "ohda", "custody":
		if hasPerm("apps.ohda") || hasPerm("app.ohda") {
			return true
		}
		if len(permissions) > 0 {
			return false
		}
		return normRole == "ACCOUNTANT" || normRole == "CUSTODY" || normRole == "CUSTODY_ADMIN"

	case "eshraf", "supervision":
		if hasPerm("apps.eshraf") || hasPerm("app.eshraf") {
			return true
		}
		if len(permissions) > 0 {
			return false
		}
		return normRole == "SUPERVISOR"

	case "edara", "admin", "management":
		if hasPerm("apps.edara") || hasPerm("app.edara") {
			return true
		}
		if len(permissions) > 0 {
			return false
		}
		return normRole == "ADMIN" || normRole == "OWNER"

	case "manadeeb", "delegate", "driver":
		if hasPerm("apps.manadeeb") || hasPerm("app.manadeeb") {
			return true
		}
		if len(permissions) > 0 {
			return false
		}
		return normRole == "DRIVER" || normRole == "EMPLOYEE" || normRole == ""

	case "dashboard":
		if hasPerm("apps.dashboard") || hasPerm("app.dashboard") {
			return true
		}
		if len(permissions) > 0 {
			return normRole == "ADMIN" || normRole == "SUPERVISOR" || normRole == "ACCOUNTANT" || normRole == "HR" || normRole == "OWNER"
		}
		return normRole == "ADMIN" || normRole == "SUPERVISOR" || normRole == "ACCOUNTANT" || normRole == "HR" || normRole == "OWNER"

	default:
		return hasPerm("apps." + cleanApp) || hasPerm(cleanApp)
	}
}

func ComputeAppPermissions(role string, permissions []string) map[string]bool {
	return map[string]bool{
		"ohda":      HasAppAccess(role, permissions, "ohda"),
		"eshraf":    HasAppAccess(role, permissions, "eshraf"),
		"edara":     HasAppAccess(role, permissions, "edara"),
		"manadeeb":  HasAppAccess(role, permissions, "manadeeb"),
		"dashboard": HasAppAccess(role, permissions, "dashboard"),
	}
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	userKey := strings.TrimSpace(req.Phone)
	if userKey == "" {
		userKey = strings.TrimSpace(req.Username)
	}
	if userKey == "" {
		userKey = strings.TrimSpace(req.Login)
	}
	if userKey == "" {
		userKey = strings.TrimSpace(req.Email)
	}
	if userKey == "" || req.Password == "" {
		return nil, errors.New("رقم الجوال وكلمة المرور مطلوبان")
	}

	cleanUserDigits := extractDigits(userKey)
	cleanPassword := strings.TrimSpace(req.Password)

	// Demo Reviewer Account for Google Play Console App Review
	isDemoUser := (userKey == "1234596" || cleanUserDigits == "1234596" || cleanUserDigits == "0500000000" || cleanUserDigits == "500000000" || cleanUserDigits == "966500000000")
	isDemoPass := (cleanPassword == "1234" || cleanPassword == "123456" || cleanPassword == "0000" || cleanPassword == "596")
	if isDemoUser && isDemoPass {
		demoID := uuid.MustParse("00000000-0000-0000-0000-000001234596")
		token, exp, err := jwt.GenerateToken(demoID, "demo.driver@kerd2sy.com", "EMPLOYEE", s.cfg.JWTSecret)
		if err != nil {
			return nil, err
		}
		demoEmp := map[string]interface{}{
			"id":                demoID.String(),
			"name":              "تست جوجل",
			"national_id":       "1234596",
			"employee_number":   "EMP-1234596",
			"job_role":          "DRIVER",
			"motorcycle_number": "7777",
			"key_number":        "KEY-01",
			"phone":             "0500000000",
			"branch_name":       "الفرع الرئيسي",
			"shift":             "morning",
			"vehicle_type":      "motorcycle",
			"role":              "DRIVER",
			"permissions":       []string{},
			"is_employee":       true,
		}
		return &dto.LoginResponse{
			Token:                 token,
			AccessToken:           token,
			RefreshToken:          token,
			ExpiresAt:             exp,
			Type:                  "employee",
			IsEmployee:            true,
			Employee:              demoEmp,
			Admin:                 nil,
			User:                  nil,
			MustChangePassword:    false,
			PermissionsConfigured: false,
		}, nil
	}

	phoneVariations := repository.NormalizeSaudiPhoneVariations(userKey)

	admin, err := s.adminRepo.FindByUsername(ctx, userKey)
	if err != nil || admin == nil {
		admin, err = s.adminRepo.FindByEmail(ctx, userKey)
		if err != nil || admin == nil {
			for _, p := range phoneVariations {
				if a, pErr := s.adminRepo.FindByPhone(ctx, p); pErr == nil && a != nil {
					admin = a
					break
				}
			}
		}
	}

	if admin == nil {
		// Fallback to Employee (Field Delegate) login
		if empMap, empErr := s.adminRepo.FindEmployeeByLogin(ctx, userKey); empErr == nil && empMap != nil {
			pwdHash, _ := empMap["password_hash"].(string)
			natID, _ := empMap["national_id"].(string)
			phoneVal, _ := empMap["phone"].(string)

			// Default password checks:
			// 1. Last 4 digits of employee phone
			// 2. Last 4 digits of input user phone
			// 3. Last 4 digits of National ID (fallback)
			matched := false

			if pwdHash != "" {
				// User set a custom bcrypt password
				if err := bcrypt.CompareHashAndPassword([]byte(pwdHash), []byte(cleanPassword)); err == nil {
					matched = true
				}
			}

			if !matched && phoneVal != "" {
				phoneDigits := extractDigits(phoneVal)
				if len(phoneDigits) >= 4 && cleanPassword == phoneDigits[len(phoneDigits)-4:] {
					matched = true
				}
			}

			if !matched && len(cleanUserDigits) >= 4 {
				if cleanPassword == cleanUserDigits[len(cleanUserDigits)-4:] {
					matched = true
				}
			}

			if !matched && natID != "" {
				natDigits := extractDigits(natID)
				if len(natDigits) >= 4 && cleanPassword == natDigits[len(natDigits)-4:] {
					matched = true
				}
			}

			if !matched {
				return nil, errors.New("كلمة المرور غير صحيحة، كلمة المرور هي آخر 4 أرقام من رقم الجوال")
			}

			delete(empMap, "password_hash")

			empID, _ := uuid.Parse(empMap["id"].(string))
			token, exp, err := jwt.GenerateToken(empID, userKey, "EMPLOYEE", s.cfg.JWTSecret)
			if err != nil {
				return nil, err
			}

			var empPerms []string
			if rawP, ok := empMap["permissions"].(string); ok && rawP != "" {
				_ = json.Unmarshal([]byte(rawP), &empPerms)
			}
			empRole, _ := empMap["job_role"].(string)
			if empRole == "" {
				empRole = "DRIVER"
			}

			if req.App != "" && !HasAppAccess(empRole, empPerms, req.App) {
				return nil, fmt.Errorf("عذراً، ليس لديك صلاحية للدخول إلى %s (يرجى مراجعة إدارة النظام)", GetAppArabicName(req.App))
			}

			appPerms := ComputeAppPermissions(empRole, empPerms)
			empMap["app_permissions"] = appPerms

			mustChange, _ := empMap["must_change_password"].(bool)
			return &dto.LoginResponse{
				Token:                 token,
				AccessToken:           token,
				RefreshToken:          token,
				ExpiresAt:             exp,
				Type:                  "employee",
				IsEmployee:            true,
				Employee:              empMap,
				MustChangePassword:    mustChange,
				PermissionsConfigured: false,
				AppPermissions:        appPerms,
			}, nil
		}
		return nil, errors.New("رقم الجوال أو كلمة المرور غير صحيحة")
	}

	// Verify Admin password:
	adminPasswordMatched := false
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(cleanPassword)); err == nil {
		adminPasswordMatched = true
	}

	// Default password for admin: last 4 digits of phone
	if !adminPasswordMatched && admin.Phone != "" {
		adminPhoneDigits := extractDigits(admin.Phone)
		if len(adminPhoneDigits) >= 4 && cleanPassword == adminPhoneDigits[len(adminPhoneDigits)-4:] {
			adminPasswordMatched = true
		}
	}

	if !adminPasswordMatched && admin.EmployeeID != nil {
		if empMap, empErr := s.adminRepo.FindEmployeeByLogin(ctx, admin.EmployeeID.String()); empErr == nil && empMap != nil {
			if phoneVal, ok := empMap["phone"].(string); ok && phoneVal != "" {
				empPhoneDigits := extractDigits(phoneVal)
				if len(empPhoneDigits) >= 4 && cleanPassword == empPhoneDigits[len(empPhoneDigits)-4:] {
					adminPasswordMatched = true
				}
			}
			if !adminPasswordMatched {
				if natID, ok := empMap["national_id"].(string); ok && natID != "" {
					empNatDigits := extractDigits(natID)
					if len(empNatDigits) >= 4 && cleanPassword == empNatDigits[len(empNatDigits)-4:] {
						adminPasswordMatched = true
					}
				}
			}
		}
	}

	if !adminPasswordMatched && len(cleanUserDigits) >= 4 {
		if cleanPassword == cleanUserDigits[len(cleanUserDigits)-4:] {
			adminPasswordMatched = true
		}
	}

	if !adminPasswordMatched {
		return nil, errors.New("رقم الجوال أو كلمة المرور غير صحيحة")
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

	if req.App != "" && !HasAppAccess(admin.Role, perms, req.App) {
		return nil, fmt.Errorf("عذراً، ليس لديك صلاحية للدخول إلى %s (يرجى مراجعة إدارة النظام)", GetAppArabicName(req.App))
	}

	appPerms := ComputeAppPermissions(admin.Role, perms)

	branchName := ""
	if admin.Branch != nil {
		branchName = admin.Branch.Name
	}

	adminInfo := &dto.AdminInfo{
		ID:                    admin.ID,
		Email:                 admin.Email,
		Username:              admin.Username,
		Phone:                 admin.Phone,
		EmployeeID:            admin.EmployeeID,
		Name:                  admin.Name,
		Role:                  admin.Role,
		Permissions:           perms,
		PermissionsConfigured: admin.PermissionsConfigured,
		MustChangePassword:    admin.MustChangePassword,
		AppPermissions:        appPerms,
		GoogleEmail:           admin.GoogleEmail,
		GoogleAvatar:          admin.GoogleAvatar,
		IsGoogleLinked:        admin.GoogleID != "",
		BranchID:              admin.BranchID,
		BranchName:            branchName,
	}

	return &dto.LoginResponse{
		Token:                 token,
		AccessToken:           token,
		RefreshToken:          refreshToken,
		ExpiresAt:             exp,
		Type:                  "admin",
		Admin:                 adminInfo,
		User:                  adminInfo,
		MustChangePassword:    admin.MustChangePassword,
		PermissionsConfigured: admin.PermissionsConfigured,
		AppPermissions:        appPerms,
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

	appPerms := ComputeAppPermissions(admin.Role, perms)

	adminInfo := &dto.AdminInfo{
		ID:                    admin.ID,
		Email:                 admin.Email,
		Username:              admin.Username,
		Phone:                 admin.Phone,
		EmployeeID:            admin.EmployeeID,
		Name:                  admin.Name,
		Role:                  admin.Role,
		Permissions:           perms,
		PermissionsConfigured: admin.PermissionsConfigured,
		MustChangePassword:    admin.MustChangePassword,
		AppPermissions:        appPerms,
		GoogleEmail:           admin.GoogleEmail,
		GoogleAvatar:          admin.GoogleAvatar,
		IsGoogleLinked:        admin.GoogleID != "",
		BranchID:              admin.BranchID,
		BranchName:            branchName,
	}

	return &dto.LoginResponse{
		Token:                 token,
		AccessToken:           token,
		RefreshToken:          refreshToken,
		ExpiresAt:             exp,
		Type:                  "admin",
		Admin:                 adminInfo,
		User:                  adminInfo,
		MustChangePassword:    admin.MustChangePassword,
		PermissionsConfigured: admin.PermissionsConfigured,
		AppPermissions:        appPerms,
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

	// Generate 4-digit code (1000 - 9999) to match mobile app input
	n, _ := rand.Int(rand.Reader, big.NewInt(9000))
	code := fmt.Sprintf("%04d", n.Int64()+1000)

	exp := time.Now().Add(10 * time.Minute)
	otp := &domain.OTPRequest{
		NationalID: req.NationalID,
		OTPCode:    code,
		DeviceInfo: req.DeviceInfo,
		DeviceUUID: req.DeviceUUID,
		Status:     "PENDING",
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
		demoID := uuid.MustParse("00000000-0000-0000-0000-000001234596")
		token, exp, err := jwt.GenerateToken(demoID, req.NationalID, "EMPLOYEE", s.cfg.JWTSecret)
		if err != nil {
			return nil, err
		}
		demoEmp := map[string]interface{}{
			"id":                demoID.String(),
			"name":              "تست جوجل",
			"national_id":       req.NationalID,
			"employee_number":   "EMP-1234596",
			"job_role":          "DRIVER",
			"motorcycle_number": "7777",
			"key_number":        "KEY-01",
			"phone":             "0500000000",
			"branch_name":       "الفرع الرئيسي",
			"shift":             "morning",
			"vehicle_type":      "motorcycle",
			"is_employee":       true,
			"role":              "DRIVER",
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

	otp, err := s.otpRepo.FindByCode(ctx, req.NationalID, req.OTPCode)
	if err != nil || otp == nil {
		return nil, errors.New("رمز التحقق غير صحيح أو منتهي الصلاحية")
	}

	otp.Status = "VERIFIED"
	_ = s.otpRepo.Update(ctx, otp)

	// Resolve Employee ID and Full Profile
	var empID uuid.UUID
	var empData map[string]interface{}
	if emp, fErr := s.otpRepo.FindEmployeeByNationalID(ctx, req.NationalID); fErr == nil && emp != nil {
		if idStr, ok := emp["id"].(string); ok {
			if parsed, pErr := uuid.Parse(idStr); pErr == nil {
				empID = parsed
			}
		}
		empData = emp
	}

	if empID == uuid.Nil {
		if otp.EmployeeID != nil && *otp.EmployeeID != uuid.Nil {
			empID = *otp.EmployeeID
		} else {
			empID = otp.ID
		}
		empData = map[string]interface{}{
			"id":          empID.String(),
			"name":        otp.EmployeeName,
			"national_id": otp.NationalID,
			"job_role":    "DRIVER",
			"is_employee": true,
			"role":        "DRIVER",
			"branch_name": otp.BranchName,
		}
	}

	token, exp, err := jwt.GenerateToken(empID, req.NationalID, "EMPLOYEE", s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token:        token,
		AccessToken:  token,
		RefreshToken: token,
		ExpiresAt:    exp,
		Type:         "employee",
		IsEmployee:   true,
		Employee:     empData,
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

func (s *otpService) GetDevices(ctx context.Context, nationalID string) ([]map[string]interface{}, error) {
	return s.otpRepo.GetDevicesByNationalID(ctx, nationalID)
}

func (s *otpService) RevokeDevice(ctx context.Context, nationalID, deviceUUID string) error {
	return s.otpRepo.RevokeDevice(ctx, nationalID, deviceUUID)
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
		Name:                  req.Name,
		Email:                 req.Email,
		Username:              req.Username,
		Phone:                 req.Phone,
		Password:              string(hashedPwd),
		Role:                  req.Role,
		RoleID:                req.RoleID,
		EmployeeID:            req.EmployeeID,
		Permissions:           string(permsJSON),
		PermissionsConfigured: req.Permissions != nil,
		MustChangePassword:    false,
		BranchID:              req.BranchID,
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
	if req.EmployeeID != nil {
		admin.EmployeeID = req.EmployeeID
	}
	if req.Permissions != nil {
		permsJSON, _ := json.Marshal(req.Permissions)
		admin.Permissions = string(permsJSON)
		admin.PermissionsConfigured = true
	}
	if req.PermissionsConfigured != nil {
		admin.PermissionsConfigured = *req.PermissionsConfigured
	}
	if req.ClearBranch {
		admin.BranchID = nil
	} else if req.BranchID != nil {
		admin.BranchID = req.BranchID
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err == nil {
			admin.Password = string(hashed)
		}
	}
	admin.Branch = nil
	admin.RoleObj = nil
	if err := s.repo.Update(ctx, admin); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, admin.ID)
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
		matched := false
		if admin.Phone != "" {
			phoneDigits := extractDigits(admin.Phone)
			if len(phoneDigits) >= 4 && oldPassword == phoneDigits[len(phoneDigits)-4:] {
				matched = true
			}
		}
		if !matched {
			return errors.New("كلمة المرور الحالية غير صحيحة")
		}
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	admin.Password = string(hashed)
	admin.MustChangePassword = false
	return s.repo.Update(ctx, admin)
}

func (s *adminService) PromoteEmployee(ctx context.Context, req dto.PromoteEmployeeRequest) (*domain.Admin, error) {
	if req.EmployeeID == uuid.Nil {
		return nil, errors.New("معرف الموظف مطلوب")
	}
	if req.Role == "" {
		return nil, errors.New("الدور الوظيفي مطلوب")
	}

	empMap, err := s.repo.FindEmployeeByLogin(ctx, req.EmployeeID.String())
	if err != nil || empMap == nil {
		return nil, errors.New("الموظف غير موجود")
	}

	empName, _ := empMap["name"].(string)
	empPhone, _ := empMap["phone"].(string)
	empNumber, _ := empMap["employee_number"].(string)

	targetPhone := strings.TrimSpace(req.Phone)
	if targetPhone == "" {
		targetPhone = empPhone
	}

	admin, err := s.repo.FindByEmployeeID(ctx, req.EmployeeID)
	if err != nil || admin == nil {
		if targetPhone != "" {
			variations := repository.NormalizeSaudiPhoneVariations(targetPhone)
			for _, v := range variations {
				if a, pErr := s.repo.FindByPhone(ctx, v); pErr == nil && a != nil {
					admin = a
					break
				}
			}
		}
	}

	var permsJSON []byte
	hasCustomPerms := false
	if req.Permissions != nil {
		permsJSON, _ = json.Marshal(req.Permissions)
		hasCustomPerms = true
	}

	if admin != nil {
		admin.Role = req.Role
		if req.RoleID != nil {
			admin.RoleID = req.RoleID
		}
		if req.BranchID != nil {
			admin.BranchID = req.BranchID
		}
		if targetPhone != "" {
			admin.Phone = targetPhone
		}
		admin.EmployeeID = &req.EmployeeID
		if hasCustomPerms {
			admin.Permissions = string(permsJSON)
			admin.PermissionsConfigured = true
		}
		if err := s.repo.Update(ctx, admin); err != nil {
			return nil, err
		}
		return admin, nil
	}

	phoneDigits := extractDigits(targetPhone)
	tempPass := "1234"
	if len(phoneDigits) >= 4 {
		tempPass = phoneDigits[len(phoneDigits)-4:]
	}
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(tempPass), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	email := fmt.Sprintf("emp_%s@aams.local", req.EmployeeID.String()[:8])
	username := targetPhone
	if username == "" {
		username = empNumber
	}
	if username == "" {
		username = fmt.Sprintf("emp_%s", req.EmployeeID.String()[:8])
	}

	newAdmin := &domain.Admin{
		Name:                  empName,
		Email:                 email,
		Username:              username,
		Phone:                 targetPhone,
		Password:              string(hashedPwd),
		Role:                  req.Role,
		RoleID:                req.RoleID,
		EmployeeID:            &req.EmployeeID,
		Permissions:           string(permsJSON),
		PermissionsConfigured: hasCustomPerms,
		MustChangePassword:    true,
		BranchID:              req.BranchID,
	}

	if err := s.repo.Create(ctx, newAdmin); err != nil {
		return nil, err
	}
	return newAdmin, nil
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
