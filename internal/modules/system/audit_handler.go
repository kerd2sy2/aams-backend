package system

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuditHandler struct {
	svc AuditService
}

func NewAuditHandler(svc AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) GetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	branchID := getBranchID(c)

	logs, total, err := h.svc.GetLogs(c.Request.Context(), branchID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب سجلات التدقيق"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *AuditHandler) ClearLogs(c *gin.Context) {
	branchID := getBranchID(c)
	if err := h.svc.ClearLogs(c.Request.Context(), branchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل مسح السجلات"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم مسح سجلات التدقيق بنجاح"})
}

func (h *AuditHandler) BulkDeleteLogs(c *gin.Context) {
	var req BulkDeleteAuditLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرفات غير صالحة"})
		return
	}

	if err := h.svc.BulkDeleteLogs(c.Request.Context(), req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حذف السجلات"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف السجلات المحددة بنجاح"})
}

func (h *AuditHandler) DeleteLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف سجل غير صالح"})
		return
	}

	if err := h.svc.DeleteLog(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حذف السجل"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف السجل بنجاح"})
}
