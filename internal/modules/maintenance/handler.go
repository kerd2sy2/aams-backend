package maintenance

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/pkg/middleware"
)

type MaintenanceHandler struct {
	service MaintenanceService
}

func NewMaintenanceHandler(service MaintenanceService) *MaintenanceHandler {
	return &MaintenanceHandler{service: service}
}

// GetEmployeeLogs handles GET /maintenance/employee-logs
func (h *MaintenanceHandler) GetEmployeeLogs(c *gin.Context) {
	empIDStr := c.Query("employee_id")
	if empIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف مطلوب"})
		return
	}

	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	logs, err := h.service.GetEmployeeLogs(c.Request.Context(), empID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب سجلات الصيانة"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

// GetAllLogs handles GET /maintenance/logs
func (h *MaintenanceHandler) GetAllLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	logs, total, err := h.service.GetAllLogs(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب سجلات الصيانة"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// CreateRequest handles POST /maintenance-requests
func (h *MaintenanceHandler) CreateRequest(c *gin.Context) {
	var req CreateMaintenanceRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الطلب غير صالحة"})
		return
	}

	var adminBranchID *uuid.UUID
	scope, sErr := middleware.ResolveBranchScope(c)
	if sErr == nil && scope != nil && !scope.IsGlobal {
		adminBranchID = scope.BranchID
	}

	res, err := h.service.CreateRequest(c.Request.Context(), req, adminBranchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل إنشاء طلب الصيانة"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": res})
}

// UpdateRequest handles PUT /maintenance-requests/:id
func (h *MaintenanceHandler) UpdateRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الطلب غير صالح"})
		return
	}

	var req UpdateMaintenanceRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات التحديث غير صالحة"})
		return
	}

	res, err := h.service.UpdateRequest(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}

// DeleteRequest handles DELETE /maintenance-requests/:id
func (h *MaintenanceHandler) DeleteRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الطلب غير صالح"})
		return
	}

	if err := h.service.DeleteRequest(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حذف طلب الصيانة"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف طلب الصيانة بنجاح"})
}

// GetAllRequests handles GET /maintenance-requests
func (h *MaintenanceHandler) GetAllRequests(c *gin.Context) {
	var filter MaintenanceRequestFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معاملات البحث غير صالحة"})
		return
	}

	var adminBranchID *uuid.UUID
	scope, sErr := middleware.ResolveBranchScope(c)
	if sErr == nil && scope != nil && !scope.IsGlobal {
		adminBranchID = scope.BranchID
	}

	reqs, total, err := h.service.GetAllRequests(c.Request.Context(), filter, adminBranchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب طلبات الصيانة"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  reqs,
		"total": total,
	})
}
