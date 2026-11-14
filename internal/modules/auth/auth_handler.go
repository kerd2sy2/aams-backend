package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/domain"
)

type AuthHandler struct {
	svc AuthService
}

func NewAuthHandler(svc AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات تسجيل الدخول غير صالحة: %v", err)})
		return
	}

	if req.App == "" {
		req.App = c.GetHeader("X-App-ID")
		if req.App == "" {
			req.App = c.GetHeader("X-App-Name")
		}
	}

	resp, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رمز التحديث مطلوب"})
		return
	}

	resp, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الدخول بحساب Google غير صالحة"})
		return
	}

	resp, err := h.svc.GoogleLogin(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Me(c *gin.Context) {
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	if isEmp, _ := c.Get("is_employee"); isEmp == true {
		if empVal, exists := c.Get("employee"); exists && empVal != nil {
			if emp, ok := empVal.(*domain.Employee); ok && emp != nil {
				var perms []string
				permsConfigured := false
				effectiveRole := strings.TrimSpace(emp.JobRole)
				branchID := emp.BranchID

				// 1. Check syntheticAdmin (resolved by auth middleware as *domain.Admin)
				if adminVal, adminExists := c.Get("admin"); adminExists && adminVal != nil {
					if synAdmin, synOk := adminVal.(*domain.Admin); synOk && synAdmin != nil {
						permsConfigured = synAdmin.PermissionsConfigured
						if synAdmin.Permissions != "" {
							_ = json.Unmarshal([]byte(synAdmin.Permissions), &perms)
						}
						if effectiveRole == "" && synAdmin.Role != "" {
							effectiveRole = strings.TrimSpace(synAdmin.Role)
						}
						if branchID == nil && synAdmin.BranchID != nil {
							branchID = synAdmin.BranchID
						}
					}
				}
				// 2. Check employee permissions if not configured via admin
				if !permsConfigured && len(perms) == 0 && emp.Permissions != "" {
					_ = json.Unmarshal([]byte(emp.Permissions), &perms)
					permsConfigured = true
				}
				// 3. Fallback to RoleObj permissions ONLY if NOT configured
				if !permsConfigured && len(perms) == 0 && emp.RoleObj != nil && emp.RoleObj.Permissions != "" {
					_ = json.Unmarshal([]byte(emp.RoleObj.Permissions), &perms)
				}
				if perms == nil {
					perms = []string{}
				}

				if effectiveRole == "" {
					if emp.RoleName != "" {
						effectiveRole = strings.TrimSpace(emp.RoleName)
					} else {
						effectiveRole = "DRIVER"
					}
				}

				roleUpper := strings.ToUpper(effectiveRole)
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

				branchName := ""
				if emp.Branch != nil {
					branchName = emp.Branch.Name
				}

				// Strip oversized Base64/blob fields so /me stays fast (<2 KB) on mobile networks
				cleanEmp := *emp
				cleanEmp.Barcode = ""
				cleanEmp.QRCode = ""
				cleanEmp.NationalIDImage = ""
				cleanEmp.DrivingLicenseImage = ""
				cleanEmp.PassportImage = ""
				cleanEmp.VehicleRegistrationImage = ""
				if strings.HasPrefix(cleanEmp.PersonalImage, "data:") && len(cleanEmp.PersonalImage) > 4096 {
					cleanEmp.PersonalImage = ""
				}
				cleanEmp.BranchID = branchID

				effectiveAvatar := cleanEmp.PersonalImage
				appPerms := ComputeAppPermissions(effectiveRole, perms)

				c.JSON(http.StatusOK, gin.H{
					"user":                   &cleanEmp,
					"employee":               &cleanEmp,
					"id":                     emp.ID,
					"name":                   emp.Name,
					"phone":                  emp.Phone,
					"national_id":            emp.NationalID,
					"employee_number":        emp.EmployeeNumber,
					"job_role":               emp.JobRole,
					"role":                   effectiveRole,
					"role_name":              emp.RoleName,
					"role_id":                emp.RoleID,
					"permissions":            perms,
					"app_permissions":        appPerms,
					"permissions_configured": true,
					"branch_id":              branchID,
					"branch_name":            branchName,
					"avatar":                 effectiveAvatar,
					"personal_image":         effectiveAvatar,
					"is_employee":            true,
					"must_change_password":   emp.MustChangePassword,
					"motorcycle_number":      emp.MotorcycleNumber,
					"key_number":             emp.KeyNumber,
					"shift":                  emp.Shift,
					"vehicle_type":           emp.VehicleType,
				})
				return
			}
			c.JSON(http.StatusOK, empVal)
			return
		}
	}

	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	adminID, _ := adminIDVal.(uuid.UUID)

	adminDTO, err := h.svc.GetAdminByID(c.Request.Context(), adminID)
	if err != nil {
		if empVal, empExists := c.Get("employee"); empExists && empVal != nil {
			c.JSON(http.StatusOK, empVal)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "المستخدم غير موجود"})
		return
	}

	if adminDTO.BranchID == nil {
		if bidVal, ok := c.Get("branch_id"); ok {
			if bid, okPtr := bidVal.(*uuid.UUID); okPtr && bid != nil && *bid != uuid.Nil {
				adminDTO.BranchID = bid
			}
		}
	}

	effectiveAvatar := adminDTO.Avatar
	if effectiveAvatar == "" {
		effectiveAvatar = adminDTO.GoogleAvatar
	}

	adminAppPerms := ComputeAppPermissions(adminDTO.Role, adminDTO.Permissions)

	c.JSON(http.StatusOK, gin.H{
		"user":                   adminDTO,
		"id":                     adminDTO.ID,
		"name":                   adminDTO.Name,
		"email":                  adminDTO.Email,
		"username":               adminDTO.Username,
		"role":                   adminDTO.Role,
		"permissions":            adminDTO.Permissions,
		"app_permissions":        adminAppPerms,
		"permissions_configured": adminDTO.PermissionsConfigured,
		"branch_id":              adminDTO.BranchID,
		"branch_name":            adminDTO.BranchName,
		"is_google_linked":       adminDTO.GoogleEmail != "",
		"google_email":           adminDTO.GoogleEmail,
		"google_avatar":          adminDTO.GoogleAvatar,
		"avatar":                 effectiveAvatar,
		"personal_image":         effectiveAvatar,
	})
}

func (h *AuthHandler) UpdateAvatar(c *gin.Context) {
	var req UpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "الصورة مطلوبة"})
		return
	}

	cleanAvatar := strings.TrimSpace(req.Avatar)
	if cleanAvatar == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رابط الصورة غير صالح"})
		return
	}

	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	adminID, _ := adminIDVal.(uuid.UUID)

	if err := h.svc.UpdateAvatar(c.Request.Context(), adminID, cleanAvatar); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "تم حفظ وتحديث الصورة الشخصية بنجاح",
		"avatar":  cleanAvatar,
	})
}

func (h *AuthHandler) LinkGoogle(c *gin.Context) {
	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	adminID, _ := adminIDVal.(uuid.UUID)

	var req GoogleLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الحساب غير صالحة"})
		return
	}

	if err := h.svc.LinkGoogle(c.Request.Context(), adminID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم ربط حساب Google بنجاح"})
}

func (h *AuthHandler) UnlinkGoogle(c *gin.Context) {
	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	adminID, _ := adminIDVal.(uuid.UUID)

	if err := h.svc.UnlinkGoogle(c.Request.Context(), adminID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم إلغاء ربط حساب Google بنجاح"})
}

// OTPHandler
