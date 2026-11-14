package auth

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

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
		if hasPerm("apps.ohda") || hasPerm("app.ohda") || hasPerm("custody") || hasPerm("custody.view") || hasPerm("custody.add") || hasPerm("custody.manage") || hasPerm("custody.create") || hasPerm("custody.delete") {
			return true
		}
		if len(permissions) > 0 {
			return normRole == "ADMIN" || normRole == "SUPER_ADMIN" || normRole == "OWNER" || normRole == "GENERAL_MANAGER"
		}
		return normRole == "ACCOUNTANT" || normRole == "CUSTODY" || normRole == "CUSTODY_ADMIN" || normRole == "ADMIN" || normRole == "SUPER_ADMIN" || normRole == "OWNER" || normRole == "GENERAL_MANAGER"

	case "eshraf", "supervision":
		if hasPerm("apps.eshraf") || hasPerm("app.eshraf") || hasPerm("work.view") || hasPerm("work.start") || hasPerm("work.end") || hasPerm("employees.view") {
			return true
		}
		if len(permissions) > 0 {
			return normRole == "ADMIN" || normRole == "SUPER_ADMIN" || normRole == "OWNER" || normRole == "GENERAL_MANAGER"
		}
		return normRole == "SUPERVISOR" || normRole == "ADMIN" || normRole == "SUPER_ADMIN" || normRole == "OWNER" || normRole == "GENERAL_MANAGER"

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

func parseBranchIDFromMap(m map[string]interface{}) *uuid.UUID {
	if m == nil {
		return nil
	}
	if s, ok := m["branch_id"].(string); ok && strings.TrimSpace(s) != "" {
		if bid, err := uuid.Parse(s); err == nil && bid != uuid.Nil {
			return &bid
		}
	}
	return nil
}

func buildEmployeeLoginResponse(token, refreshToken string, exp time.Time, empMap map[string]interface{}) (*LoginResponse, []string) {
	var empPerms []string
	if rawP, ok := empMap["permissions"].(string); ok && rawP != "" {
		_ = json.Unmarshal([]byte(rawP), &empPerms)
	}
	empRole, _ := empMap["job_role"].(string)
	if empRole == "" {
		empRole = "DRIVER"
	}
	appPerms := ComputeAppPermissions(empRole, empPerms)
	empMap["app_permissions"] = appPerms
	mustChange, _ := empMap["must_change_password"].(bool)

	return &LoginResponse{
		Token:                 token,
		AccessToken:           token,
		RefreshToken:          refreshToken,
		ExpiresAt:             exp,
		Type:                  "employee",
		IsEmployee:            true,
		Employee:              empMap,
		MustChangePassword:    mustChange,
		PermissionsConfigured: false,
		AppPermissions:        appPerms,
	}, empPerms
}

func buildAdminLoginResponse(admin *Admin, branchName, token, refreshToken string, exp time.Time, perms []string) *LoginResponse {
	appPerms := ComputeAppPermissions(admin.Role, perms)
	avatar := admin.Avatar
	if avatar == "" {
		avatar = admin.GoogleAvatar
	}
	adminInfo := &AdminInfo{
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
		Avatar:                avatar,
		IsGoogleLinked:        admin.GoogleID != "",
		BranchID:              admin.BranchID,
		BranchName:            branchName,
	}
	return &LoginResponse{
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
	}
}
