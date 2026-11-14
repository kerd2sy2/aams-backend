package system

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ArchiveHandler struct {
	svc ArchiveService
}

func NewArchiveHandler(svc ArchiveService) *ArchiveHandler {
	return &ArchiveHandler{svc: svc}
}

func (h *ArchiveHandler) GetArchived(c *gin.Context) {
	var filter ArchiveFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	branchID := getBranchID(c)
	if branchID != nil {
		filter.BranchID = branchID
	}

	res, err := h.svc.GetArchivedItems(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *ArchiveHandler) Restore(c *gin.Context) {
	var req RestoreArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الاسترجاع غير مكتملة", "details": err.Error()})
		return
	}

	if err := h.svc.Restore(c.Request.Context(), req.Type, req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم استرجاع العنصر بنجاح"})
}

func (h *ArchiveHandler) PermanentDelete(c *gin.Context) {
	var req PermanentDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		itemType := c.Query("type")
		idStr := c.Query("id")
		parsedID, errParse := uuid.Parse(idStr)
		if itemType == "" || errParse != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الحذف النهائي غير صالحة"})
			return
		}
		req.Type = itemType
		req.ID = parsedID
	}

	if err := h.svc.PermanentDelete(c.Request.Context(), req.Type, req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف العنصر نهائياً وبنجاح"})
}

func (h *ArchiveHandler) BulkRestore(c *gin.Context) {
	var req BulkArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الاسترجاع الجماعي غير صالحة", "details": err.Error()})
		return
	}

	if err := h.svc.BulkRestore(c.Request.Context(), req.Type, req.IDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("تم استرجاع %d عناصر بنجاح", len(req.IDs))})
}

func (h *ArchiveHandler) BulkPermanentDelete(c *gin.Context) {
	var req BulkArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الحذف الجماعي غير صالحة", "details": err.Error()})
		return
	}

	if err := h.svc.BulkPermanentDelete(c.Request.Context(), req.Type, req.IDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("تم حذف %d عناصر نهائياً بنجاح", len(req.IDs))})
}
