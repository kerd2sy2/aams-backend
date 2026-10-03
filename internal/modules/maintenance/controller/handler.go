package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/maintenance/dto"
	"delivery-backend/internal/modules/maintenance/service"
)

type MaintenanceHandler struct {
	maintService service.MaintenanceService
}

func NewMaintenanceHandler(maintService service.MaintenanceService) *MaintenanceHandler {
	return &MaintenanceHandler{maintService: maintService}
}

func (h *MaintenanceHandler) GetEmployeeLogs(c *gin.Context) {
	empIDStr := c.Query("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	logs, err := h.maintService.GetEmployeeLogs(c.Request.Context(), empID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (h *MaintenanceHandler) GetAllLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	logs, total, err := h.maintService.GetAllLogs(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if limit <= 0 {
		limit = 20
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"data":        logs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

func (h *MaintenanceHandler) CreateRequest(c *gin.Context) {
	var req dto.CreateMaintenanceRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(uuid.UUID); ok {
			branchID = &val
		} else if ptrVal, ok := bID.(*uuid.UUID); ok {
			branchID = ptrVal
		}
	}

	m, err := h.maintService.CreateRequest(c.Request.Context(), req, branchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, m)
}

func (h *MaintenanceHandler) UpdateRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateMaintenanceRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	m, err := h.maintService.UpdateRequest(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, m)
}

func (h *MaintenanceHandler) DeleteRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.maintService.DeleteRequest(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف طلب الصيانة بنجاح"})
}

func (h *MaintenanceHandler) GetAllRequests(c *gin.Context) {
	var filter dto.MaintenanceRequestFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(uuid.UUID); ok {
			branchID = &val
		} else if ptrVal, ok := bID.(*uuid.UUID); ok {
			branchID = ptrVal
		}
	}

	list, total, err := h.maintService.GetAllRequests(c.Request.Context(), filter, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
		"page":  filter.GetEffectivePage(),
		"limit": filter.GetEffectiveLimit(),
	})
}
