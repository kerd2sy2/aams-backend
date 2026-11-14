package auth

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"delivery-backend/pkg/config"
	"delivery-backend/pkg/jwt"
	"delivery-backend/pkg/middleware"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	adminRepo  AdminRepository
	branchRepo BranchRepository
	cfg        *config.Config
}

func NewAuthService(adminRepo AdminRepository, branchRepo BranchRepository, cfg *config.Config) AuthService {
	return &authService{
		adminRepo:  adminRepo,
		branchRepo: branchRepo,
		cfg:        cfg,
	}
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
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
		return &LoginResponse{
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

	phoneVariations := NormalizeSaudiPhoneVariations(userKey)
	admin, _ := s.adminRepo.FindByLogin(ctx, userKey, phoneVariations)

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
			empName, _ := empMap["name"].(string)
			empBranchID := parseBranchIDFromMap(empMap)
			middleware.InvalidateAuthCache(empID)

			token, refreshToken, err := jwt.GenerateTokens(empID, userKey, empName, "EMPLOYEE", empBranchID, s.cfg.JWTSecret, s.cfg.JWTRefreshSecret)
			if err != nil {
				return nil, err
			}
			exp := time.Now().Add(7 * 24 * time.Hour)
			resp, empPerms := buildEmployeeLoginResponse(token, refreshToken, exp, empMap)
			empRole, _ := empMap["job_role"].(string)
			if empRole == "" {
				empRole = "DRIVER"
			}
			if req.App != "" && !HasAppAccess(empRole, empPerms, req.App) {
				return nil, errors.New("ليس لديك صلاحية للدخول إلى التطبيق، يرجى مراجعة الإدارة")
			}
			return resp, nil
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

	var linkedEmpMap map[string]interface{}
	if admin.EmployeeID != nil && (!adminPasswordMatched || admin.BranchID == nil) {
		linkedEmpMap, _ = s.adminRepo.FindEmployeeByLogin(ctx, admin.EmployeeID.String())
	}

	if !adminPasswordMatched && linkedEmpMap != nil {
		if phoneVal, ok := linkedEmpMap["phone"].(string); ok && phoneVal != "" {
			empPhoneDigits := extractDigits(phoneVal)
			if len(empPhoneDigits) >= 4 && cleanPassword == empPhoneDigits[len(empPhoneDigits)-4:] {
				adminPasswordMatched = true
			}
		}
		if !adminPasswordMatched {
			if natID, ok := linkedEmpMap["national_id"].(string); ok && natID != "" {
				empNatDigits := extractDigits(natID)
				if len(empNatDigits) >= 4 && cleanPassword == empNatDigits[len(empNatDigits)-4:] {
					adminPasswordMatched = true
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

	branchName := ""
	if admin.Branch != nil {
		branchName = admin.Branch.Name
	}
	if admin.BranchID == nil && linkedEmpMap != nil {
		if bid := parseBranchIDFromMap(linkedEmpMap); bid != nil {
			admin.BranchID = bid
			if bn, ok := linkedEmpMap["branch_name"].(string); ok && branchName == "" {
				branchName = bn
			}
		}
	}

	middleware.InvalidateAuthCache(admin.ID)
	token, refreshToken, err := jwt.GenerateTokens(admin.ID, admin.Email, admin.Name, admin.Role, admin.BranchID, s.cfg.JWTSecret, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, err
	}
	exp := time.Now().Add(7 * 24 * time.Hour)

	var perms []string
	if admin.Permissions != "" {
		_ = json.Unmarshal([]byte(admin.Permissions), &perms)
	}

	if req.App != "" && !HasAppAccess(admin.Role, perms, req.App) {
		return nil, errors.New("ليس لديك صلاحية للدخول إلى التطبيق، يرجى مراجعة الإدارة")
	}

	return buildAdminLoginResponse(admin, branchName, token, refreshToken, exp, perms), nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	claims, err := jwt.ValidateTokenForRefresh(refreshToken, s.cfg.JWTSecret, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, errors.New("رمز التحديث غير صالح أو منتهي الصلاحية")
	}

	admin, err := s.adminRepo.FindByID(ctx, claims.AdminID)
	if err != nil || admin == nil {
		// Fallback to Employee (Delegate) refresh if user is in employees table
		empMap, empErr := s.adminRepo.FindEmployeeByLogin(ctx, claims.AdminID.String())
		if (empErr != nil || empMap == nil) && claims.Email != "" {
			empMap, empErr = s.adminRepo.FindEmployeeByLogin(ctx, claims.Email)
		}
		if empErr == nil && empMap != nil {
			delete(empMap, "password_hash")

			empID, _ := uuid.Parse(empMap["id"].(string))
			empName, _ := empMap["name"].(string)
			empBranchID := parseBranchIDFromMap(empMap)
			userKey, _ := empMap["phone"].(string)
			if userKey == "" {
				userKey = claims.Email
			}

			token, newRefreshToken, genErr := jwt.GenerateTokens(empID, userKey, empName, "EMPLOYEE", empBranchID, s.cfg.JWTSecret, s.cfg.JWTRefreshSecret)
			if genErr != nil {
				return nil, genErr
			}
			exp := time.Now().Add(7 * 24 * time.Hour)
			resp, _ := buildEmployeeLoginResponse(token, newRefreshToken, exp, empMap)
			return resp, nil
		}

		return nil, errors.New("المستخدم غير موجود")
	}

	branchName := ""
	if admin.Branch != nil {
		branchName = admin.Branch.Name
	}
	if admin.BranchID == nil && admin.EmployeeID != nil {
		if empMap, empErr := s.adminRepo.FindEmployeeByLogin(ctx, admin.EmployeeID.String()); empErr == nil && empMap != nil {
			if bid := parseBranchIDFromMap(empMap); bid != nil {
				admin.BranchID = bid
				if bn, ok := empMap["branch_name"].(string); ok && branchName == "" {
					branchName = bn
				}
			}
		}
	}

	token, newRefreshToken, err := jwt.GenerateTokens(admin.ID, admin.Email, admin.Name, admin.Role, admin.BranchID, s.cfg.JWTSecret, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, err
	}
	exp := time.Now().Add(7 * 24 * time.Hour)

	var perms []string
	if admin.Permissions != "" {
		_ = json.Unmarshal([]byte(admin.Permissions), &perms)
	}

	return buildAdminLoginResponse(admin, branchName, token, newRefreshToken, exp, perms), nil
}

func (s *authService) GetAdminByID(ctx context.Context, id uuid.UUID) (*AdminDTO, error) {
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
	// Fallback to role permissions ONLY if permissions have not been explicitly configured
	if !a.PermissionsConfigured && len(perms) == 0 && a.RoleObj != nil && a.RoleObj.Permissions != "" {
		_ = json.Unmarshal([]byte(a.RoleObj.Permissions), &perms)
	}
	roleUpper := strings.ToUpper(strings.TrimSpace(a.Role))
	if roleUpper == "SUPER_ADMIN" || roleUpper == "SUPERADMIN" || roleUpper == "OWNER" || roleUpper == "GENERAL_MANAGER" {
		hasWildcard := false
		for _, p := range perms {
			if p == "*" {
				hasWildcard = true
				break
			}
		}
		if !hasWildcard {
			perms = append([]string{"*"}, perms...)
		}
	}
	if perms == nil {
		perms = []string{}
	}

	effectiveAvatar := a.Avatar
	if effectiveAvatar == "" {
		effectiveAvatar = a.GoogleAvatar
	}
	return &AdminDTO{
		ID:                    a.ID,
		Email:                 a.Email,
		Username:              a.Username,
		Name:                  a.Name,
		Role:                  a.Role,
		BranchID:              a.BranchID,
		BranchName:            branchName,
		Permissions:           perms,
		GoogleEmail:           a.GoogleEmail,
		GoogleAvatar:          a.GoogleAvatar,
		Avatar:                effectiveAvatar,
		PermissionsConfigured: a.PermissionsConfigured,
	}, nil
}

func (s *authService) GetAdminByEmail(ctx context.Context, email string) (*AdminDTO, error) {
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
