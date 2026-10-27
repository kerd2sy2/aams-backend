package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/modules/auth/dto"
	"delivery-backend/internal/modules/auth/service"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
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
	var req dto.RefreshTokenRequest
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
	var req dto.GoogleLoginRequest
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
				// 1. Check syntheticAdmin (resolved by auth middleware with linkedAdmin)
				if adminVal, adminExists := c.Get("admin"); adminExists && adminVal != nil {
					if synAdmin, synOk := adminVal.(*domain.Admin); synOk && synAdmin != nil {
						permsConfigured = synAdmin.PermissionsConfigured
						if synAdmin.Permissions != "" {
							_ = json.Unmarshal([]byte(synAdmin.Permissions), &perms)
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

				effectiveRole := strings.TrimSpace(emp.JobRole)
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

				effectiveAvatar := emp.PersonalImage
				appPerms := service.ComputeAppPermissions(effectiveRole, perms)

				c.JSON(http.StatusOK, gin.H{
					"user":                   emp,
					"employee":               emp,
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
					"branch_id":              emp.BranchID,
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

	effectiveAvatar := adminDTO.Avatar
	if effectiveAvatar == "" {
		effectiveAvatar = adminDTO.GoogleAvatar
	}

	adminAppPerms := service.ComputeAppPermissions(adminDTO.Role, adminDTO.Permissions)

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
	var req dto.UpdateAvatarRequest
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

	var req dto.GoogleLinkRequest
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
type OTPHandler struct {
	svc service.OTPService
}

func NewOTPHandler(svc service.OTPService) *OTPHandler {
	return &OTPHandler{svc: svc}
}

func (h *OTPHandler) RequestOTP(c *gin.Context) {
	var req dto.RequestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رقم الهوية مطلوب"})
		return
	}

	resp, err := h.svc.RequestOTP(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *OTPHandler) VerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رمز التحقق ورقم الهوية مطلوبان"})
		return
	}

	resp, err := h.svc.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *OTPHandler) GetOTPList(c *gin.Context) {
	var query dto.OTPListQuery
	_ = c.ShouldBindQuery(&query)

	list, total, err := h.svc.GetOTPList(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (h *OTPHandler) CancelOTP(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الطلب غير صالح"})
		return
	}

	if err := h.svc.CancelOTP(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم إلغاء الطلب بنجاح"})
}

func (h *OTPHandler) GetMyDevices(c *gin.Context) {
	nationalID := ""
	if empVal, exists := c.Get("employee"); exists && empVal != nil {
		if emp, ok := empVal.(*domain.Employee); ok && emp.NationalID != "" {
			nationalID = emp.NationalID
		}
	}
	if nationalID == "" {
		if emailVal, exists := c.Get("admin_email"); exists && emailVal != nil {
			nationalID = fmt.Sprintf("%v", emailVal)
		}
	}
	if nationalID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}

	devices, err := h.svc.GetDevices(c.Request.Context(), nationalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": devices})
}

func (h *OTPHandler) RevokeDevice(c *gin.Context) {
	deviceUUID := c.Param("uuid")
	if deviceUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الجهاز مطلوب"})
		return
	}

	nationalID := ""
	if empVal, exists := c.Get("employee"); exists && empVal != nil {
		if emp, ok := empVal.(*domain.Employee); ok && emp.NationalID != "" {
			nationalID = emp.NationalID
		}
	}
	if nationalID == "" {
		if emailVal, exists := c.Get("admin_email"); exists && emailVal != nil {
			nationalID = fmt.Sprintf("%v", emailVal)
		}
	}
	if nationalID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}

	if err := h.svc.RevokeDevice(c.Request.Context(), nationalID, deviceUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم إلغاء توثيق الجهاز بنجاح"})
}

// RoleHandler
type RoleHandler struct {
	svc service.RoleService
}

func NewRoleHandler(svc service.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الدور غير صالحة"})
		return
	}

	role, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "تم إنشاء الدور بنجاح", "data": role})
}

func (h *RoleHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الدور غير صالح"})
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	role, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث الدور بنجاح", "data": role})
}

func (h *RoleHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الدور غير صالح"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الدور بنجاح"})
}

func (h *RoleHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الدور غير صالح"})
		return
	}

	role, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الدور غير موجود"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": role})
}

func (h *RoleHandler) GetAll(c *gin.Context) {
	roles, err := h.svc.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (h *RoleHandler) GetPermissions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"permissions": []string{
		"*",
		"apps.ohda", "apps.eshraf", "apps.edara", "apps.manadeeb", "apps.dashboard",
		"all_branches", "branches.manage",
		"custody", "custody.access", "custody.view", "custody.add", "custody.expense", "custody.delete", "custody.manage",
		"employees.view", "employees.create", "employees.edit", "employees.delete", "employees.cards",
		"work.view", "work.start", "work.end",
		"vehicles.view", "vehicles.manage", "vehicles.oil",
		"fuel.view", "fuel.manage",
		"maintenance.view", "maintenance.manage",
		"attendance.view", "investigations.view", "investigations.create", "investigations.approve",
		"inventory.view", "inventory.manage", "inventory.dispense",
		"tickets.view", "tickets.manage",
		"bank_accounts.view", "bank_accounts.manage",
		"violations.view", "violations.manage",
		"reports.view", "reports.export",
		"documents.view", "documents.manage",
		"leaves.view", "leaves.manage",
		"partners.view", "partners.manage",
		"users.manage", "roles.manage", "settings.manage",
		"archive.view",
		"audit_logs.view", "error_logs.view",
	}})
}

// AdminHandler
type AdminHandler struct {
	svc service.AdminService
}

func NewAdminHandler(svc service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) Create(c *gin.Context) {
	var req dto.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	admin, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "تم إنشاء المستخدم بنجاح", "data": admin})
}

func (h *AdminHandler) Promote(c *gin.Context) {
	var req dto.PromoteEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	admin, err := h.svc.PromoteEmployee(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم ترقية الموظف وربط الحساب بنجاح", "data": admin})
}

func (h *AdminHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المستخدم غير صالح"})
		return
	}

	var req dto.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	admin, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث بيانات المستخدم بنجاح", "data": admin})
}

func (h *AdminHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المستخدم غير صالح"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف المستخدم بنجاح"})
}

func (h *AdminHandler) GetAll(c *gin.Context) {
	admins, err := h.svc.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": admins})
}

func (h *AdminHandler) ChangePassword(c *gin.Context) {
	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	adminID, _ := adminIDVal.(uuid.UUID)

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), adminID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تغيير كلمة المرور بنجاح"})
}

// BranchHandler
type BranchHandler struct {
	svc service.BranchService
}

func NewBranchHandler(svc service.BranchService) *BranchHandler {
	return &BranchHandler{svc: svc}
}

func (h *BranchHandler) Create(c *gin.Context) {
	var req dto.CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "اسم الفرع مطلوب"})
		return
	}

	b, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "تم إنشاء الفرع بنجاح", "data": b})
}

func (h *BranchHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفرع غير صالح"})
		return
	}

	var req dto.UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "اسم الفرع مطلوب"})
		return
	}

	b, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث الفرع بنجاح", "data": b})
}

func (h *BranchHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفرع غير صالح"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الفرع بنجاح"})
}

func (h *BranchHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفرع غير صالح"})
		return
	}

	b, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الفرع غير موجود"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": b})
}

func (h *BranchHandler) GetAll(c *gin.Context) {
	branches, err := h.svc.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": branches})
}
