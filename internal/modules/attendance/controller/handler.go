package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/attendance/dto"
	"delivery-backend/internal/modules/attendance/service"
)

type AttendanceHandler struct {
	svc service.AttendanceService
}

func NewAttendanceHandler(svc service.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{svc: svc}
}

func attendanceBranchID(c *gin.Context) *uuid.UUID {
	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if val, ok := bid.(*uuid.UUID); ok && val != nil {
			return val
		}
		if val, ok := bid.(uuid.UUID); ok && val != uuid.Nil {
			return &val
		}
	}
	if s := c.Query("branch_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			return &id
		}
	}
	return nil
}

func (h *AttendanceHandler) GetAttendance(c *gin.Context) {
	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	branchID := attendanceBranchID(c)

	attendance, err := h.svc.GetAttendance(c.Request.Context(), date, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("فشل في جلب بيانات الحضور: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date": date,
		"data": attendance,
	})
}

func (h *AttendanceHandler) ToggleAttendance(c *gin.Context) {
	employeeIDStr := c.Param("employee_id")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	var req dto.ToggleAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	if req.Status != "present" && req.Status != "absent" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "الحالة يجب أن تكون present أو absent"})
		return
	}

	result, err := h.svc.ToggleAttendance(c.Request.Context(), uuid.Nil, employeeID, req.Date, req.Status, req.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("فشل في تحديث الحضور: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}
