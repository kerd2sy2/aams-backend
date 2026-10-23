package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/work/contracts"
	"delivery-backend/internal/modules/work/dto"
	"delivery-backend/internal/modules/work/service"
	"delivery-backend/pkg/middleware"
)

type WorkHandler struct {
	svc   service.WorkService
	audit contracts.IAuditContract
}

func NewWorkHandler(svc service.WorkService, audit ...contracts.IAuditContract) *WorkHandler {
	var a contracts.IAuditContract
	if len(audit) > 0 {
		a = audit[0]
	}
	return &WorkHandler{svc: svc, audit: a}
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

	// Strict Ownership: If caller is an employee/courier, bind to verified token identity
	if isEmp, _ := c.Get("is_employee"); isEmp == true {
		if authEmpID, exists := c.Get("employee_id"); exists {
			if aUUID, ok := authEmpID.(uuid.UUID); ok && aUUID != uuid.Nil {
				req.EmployeeID = aUUID.String()
			}
		}
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

	// Strict Ownership: If caller is an employee/courier, bind to verified token identity
	if isEmp, _ := c.Get("is_employee"); isEmp == true {
		if authEmpID, exists := c.Get("employee_id"); exists {
			if aUUID, ok := authEmpID.(uuid.UUID); ok && aUUID != uuid.Nil {
				req.EmployeeID = aUUID.String()
			}
		}
	}

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

	// Strict IDOR Check: If caller is an employee/courier, verify ownership
	if isEmp, _ := c.Get("is_employee"); isEmp == true {
		if authEmpID, exists := c.Get("employee_id"); exists {
			if aUUID, ok := authEmpID.(uuid.UUID); ok && aUUID != uuid.Nil {
				if session.EmployeeID == nil || *session.EmployeeID != aUUID {
					c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح: لا يمكنك عرض تفاصيل شفت موظف آخر"})
					return
				}
			}
		}
	}

	c.JSON(http.StatusOK, session)
}

func (h *WorkHandler) GetActiveSession(c *gin.Context) {
	var empID uuid.UUID
	if isEmp, _ := c.Get("is_employee"); isEmp == true {
		if authEmpID, exists := c.Get("employee_id"); exists {
			if aUUID, ok := authEmpID.(uuid.UUID); ok && aUUID != uuid.Nil {
				empID = aUUID
			}
		}
	} else if empIDStr := c.Query("employee_id"); empIDStr != "" {
		parsed, err := uuid.Parse(empIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
			return
		}
		empID = parsed
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف مطلوب"})
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

func (h *WorkHandler) SupervisorStartShift(c *gin.Context) {
	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if isEmpVal, ok := c.Get("is_employee"); ok {
		if b, okBool := isEmpVal.(bool); okBool && b {
			c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح: بدء دوام المشرف مخصص لحسابات الإشراف والإدارة فقط"})
			return
		}
	}
	if strings.EqualFold(scope.Role, "DRIVER") || strings.EqualFold(scope.Role, "EMPLOYEE") {
		c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح للسائقين ببدء دوام المشرف أو تنفيذ تجاوزات"})
		return
	}

	var req dto.SupervisorStartShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	adminIDVal, exists := c.Get("admin_id")
	if !exists || adminIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "جلسة المشرف غير صالحة"})
		return
	}
	supervisorID, _ := adminIDVal.(uuid.UUID)
	supervisorName := c.GetString("admin_name")
	if supervisorName == "" {
		supervisorName = "Supervisor"
	}

	session, err := h.svc.SupervisorStartShift(c.Request.Context(), req, supervisorID, supervisorName, scope.Role, scope.BranchID)
	if err != nil {
		if errors.Is(err, service.ErrActiveSessionExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrCrossBranchForbidden) {
			c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.audit != nil {
		targetDesc := "دوام المشرف الشخصي"
		if req.EmployeeID != "" {
			targetDesc = fmt.Sprintf("تجاوز دوام للموظف %s", req.EmployeeID)
		}
		details := fmt.Sprintf("بدء %s (بداية العداد: %.0f) - السبب: %s", targetDesc, session.StartKM, req.OverrideReason)
		_ = h.audit.LogAction(c.Request.Context(), supervisorName, "بدء شفت مشرف / تجاوز", details, c.ClientIP(), scope.BranchID)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "تم بدء الشفت بنجاح",
		"data":    session,
	})
}

func (h *WorkHandler) SupervisorEndShift(c *gin.Context) {
	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if isEmpVal, ok := c.Get("is_employee"); ok {
		if b, okBool := isEmpVal.(bool); okBool && b {
			c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح: إنهاء دوام المشرف مخصص لحسابات الإشراف والإدارة فقط"})
			return
		}
	}
	if strings.EqualFold(scope.Role, "DRIVER") || strings.EqualFold(scope.Role, "EMPLOYEE") {
		c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح للسائقين بإنهاء دوام المشرف أو تنفيذ تجاوزات"})
		return
	}

	var req dto.SupervisorEndShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	adminIDVal, exists := c.Get("admin_id")
	if !exists || adminIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "جلسة المشرف غير صالحة"})
		return
	}
	supervisorID, _ := adminIDVal.(uuid.UUID)
	supervisorName := c.GetString("admin_name")
	if supervisorName == "" {
		supervisorName = "Supervisor"
	}

	session, err := h.svc.SupervisorEndShift(c.Request.Context(), req, supervisorID, supervisorName, scope.Role, scope.BranchID)
	if err != nil {
		if errors.Is(err, service.ErrNoActiveSession) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrCrossBranchForbidden) {
			c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.audit != nil {
		targetDesc := "دوام المشرف الشخصي"
		if req.EmployeeID != "" {
			targetDesc = fmt.Sprintf("تجاوز دوام للموظف %s", req.EmployeeID)
		}
		details := fmt.Sprintf("إنهاء %s (نهاية العداد: %.0f، المسافة: %.0f كم، الطلبات: %d) - السبب: %s", targetDesc, session.EndKM, session.Distance, session.OrdersCount, req.OverrideReason)
		_ = h.audit.LogAction(c.Request.Context(), supervisorName, "إنهاء شفت مشرف / تجاوز", details, c.ClientIP(), scope.BranchID)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "تم إنهاء الشفت بنجاح",
		"data":    session,
	})
}


