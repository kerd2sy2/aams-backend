package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	svc DashboardService
}

func NewDashboardHandler(svc DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	branchID := getBranchID(c)
	stats, err := h.svc.GetStats(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تحميل إحصائيات لوحة التحكم"})
		return
	}
	c.JSON(http.StatusOK, stats)
}
