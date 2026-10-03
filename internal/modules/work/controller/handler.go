package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/work/dto"
	"delivery-backend/internal/modules/work/service"
)

type WorkHandler struct {
	svc service.WorkService
}

func NewWorkHandler(svc service.WorkService) *WorkHandler {
	return &WorkHandler{svc: svc}
}

func getAdminInfo(c *gin.Context) (*uuid.UUID, string) {
	// If the authenticated user is an employee / courier, they are NOT an admin or supervisor
	if isEmp, exists := c.Get("is_employee"); exists {
		if b, ok := isEmp.(bool); ok && b {
			return nil, ""
		}
	}
	role := strings.ToUpper(c.GetString("admin_role"))
	if role == "DRIVER" || role == "EMPLOYEE" {
		return nil, ""
	}

	var adminID *uuid.UUID
	if idVal, exists := c.Get("admin_id"); exists && idVal != nil {
		if id, ok := idVal.(uuid.UUID); ok && id != uuid.Nil {
			adminID = &id
		}
	}
	adminName := c.GetString("admin_name")
	return adminID, adminName
}

func (h *WorkHandler) StartWork(c *gin.Context) {
	var req dto.StartWorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	session, err := h.svc.StartWork(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *WorkHandler) EndWork(c *gin.Context) {
	var req dto.EndWorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	adminID, adminName := getAdminInfo(c)
	isSupervisor := adminID != nil

	session, err := h.svc.EndWork(c.Request.Context(), req, adminID, adminName, isSupervisor)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *WorkHandler) UpdateWorkSession(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الشفت غير صالح"})
		return
	}

	var req dto.UpdateWorkSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	_, adminName := getAdminInfo(c)
	session, err := h.svc.UpdateWorkSession(c.Request.Context(), sessionID, req, adminName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *WorkHandler) ReviewWorkSession(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الشفت غير صالح"})
		return
	}

	var req dto.ReviewWorkSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	adminID, adminName := getAdminInfo(c)
	if adminID == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح: هذه العملية مخصصة للمشرفين فقط"})
		return
	}
	session, err := h.svc.ReviewSession(c.Request.Context(), sessionID, req, adminID, adminName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *WorkHandler) GetSessionByID(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الشفت غير صالح"})
		return
	}

	session, err := h.svc.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الشفت غير موجود"})
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *WorkHandler) GetActiveSession(c *gin.Context) {
	empIDStr := c.Query("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	session, err := h.svc.GetActiveSession(c.Request.Context(), empID)
	if err != nil || session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "لا يوجد شفت نشط حالياً"})
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *WorkHandler) GetLastKM(c *gin.Context) {
	empIDStr := c.Query("employee_id")
	empID, _ := uuid.Parse(empIDStr)

	motorcycleNumber := c.Query("motorcycle_number")
	if motorcycleNumber == "" {
		motorcycleNumber = c.Query("plate")
	}

	resp, err := h.svc.GetLastSessionOrVehicleKM(c.Request.Context(), empID, motorcycleNumber)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"last_end_km":        0,
			"last_start_km":      0,
			"is_odometer_broken": false,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *WorkHandler) ScanPlate(c *gin.Context) {
	var req struct {
		Image string `json:"image" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يجب إرفاق الصورة"})
		return
	}

	res, err := h.svc.ScanPlateImage(c.Request.Context(), req.Image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *WorkHandler) TodayCount(c *gin.Context) {
	empIDStr := c.Query("employee_id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	count, err := h.svc.CountTodaySessions(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count, "today_count": count})
}

func (h *WorkHandler) CheckOilChange(c *gin.Context) {
	empIDStr := c.Query("employee_id")
	empID, _ := uuid.Parse(empIDStr)

	motorcycleNumber := c.Query("motorcycle_number")
	res, err := h.svc.CheckOilChange(c.Request.Context(), empID, motorcycleNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

