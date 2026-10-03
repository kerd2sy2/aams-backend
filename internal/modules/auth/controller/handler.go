package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات تسجيل الدخول غير صالحة"})
		return
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
	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	adminID, _ := adminIDVal.(uuid.UUID)

	adminDTO, err := h.svc.GetAdminByID(c.Request.Context(), adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "المستخدم غير موجود"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": adminDTO})
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
	c.JSON(http.StatusOK, gin.H{"permissions": []string{"*"}})
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
