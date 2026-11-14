package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SettingHandler struct {
	svc SettingService
}

func NewSettingHandler(svc SettingService) *SettingHandler {
	return &SettingHandler{svc: svc}
}

func (h *SettingHandler) GetPublicSettings(c *gin.Context) {
	settings, err := h.svc.GetAllSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب الإعدادات"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *SettingHandler) GetSettings(c *gin.Context) {
	settings, err := h.svc.GetAllSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب الإعدادات"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *SettingHandler) UpdateSettings(c *gin.Context) {
	var req UpdateAppSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	if err := h.svc.UpdateAppSettings(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تحديث الإعدادات"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث الإعدادات بنجاح"})
}
