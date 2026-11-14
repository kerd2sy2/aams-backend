package hr_legal

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *HRLegalHandler) CreateInvestigation(c *gin.Context) {
	var req CreateInvestigationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	supIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	supervisorID, ok := supIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "معرف المشرف غير صالح"})
		return
	}

	if req.SupervisorName == "" {
		req.SupervisorName = c.GetString("admin_name")
	}

	resp, err := h.svc.CreateInvestigation(c.Request.Context(), req, supervisorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *HRLegalHandler) UpdateInvestigation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req UpdateInvestigationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	resp, err := h.svc.UpdateInvestigation(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HRLegalHandler) DeleteInvestigation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.svc.DeleteInvestigation(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف السجل بنجاح"})
}

func (h *HRLegalHandler) GetAllInvestigations(c *gin.Context) {
	branchID := getBranchID(c)
	list, err := h.svc.GetAllInvestigations(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *HRLegalHandler) GetInvestigationByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	resp, err := h.svc.GetInvestigationByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "سجل التحقيق غير موجود"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HRLegalHandler) GetPublicByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	resp, err := h.svc.GetInvestigationByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الوثيقة غير موجودة"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HRLegalHandler) ApproveInvestigation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	adminIDVal, _ := c.Get("admin_id")
	adminID, _ := adminIDVal.(uuid.UUID)
	adminName := c.GetString("admin_name")
	adminUsername := c.GetString("admin_username")
	if adminUsername == "" {
		adminUsername = c.GetString("admin_email")
	}

	resp, err := h.svc.ApproveInvestigation(c.Request.Context(), id, adminID, adminName, adminUsername)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HRLegalHandler) RejectInvestigation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	adminID, _ := adminIDVal.(uuid.UUID)
	adminName := c.GetString("admin_name")
	adminUsername := c.GetString("admin_username")
	if adminUsername == "" {
		adminUsername = c.GetString("admin_email")
	}

	resp, err := h.svc.RejectInvestigation(c.Request.Context(), id, adminID, adminName, adminUsername)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HRLegalHandler) GetPendingInvestigationCount(c *gin.Context) {
	count, err := h.svc.GetPendingInvestigationCount(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pending_count": count})
}
