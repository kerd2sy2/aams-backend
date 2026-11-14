package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/domain"
)

type OTPHandler struct {
	svc OTPService
}

func NewOTPHandler(svc OTPService) *OTPHandler {
	return &OTPHandler{svc: svc}
}

func (h *OTPHandler) RequestOTP(c *gin.Context) {
	var req RequestOTPRequest
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
	var req VerifyOTPRequest
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
	var query OTPListQuery
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