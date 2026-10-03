package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/hr_legal/dto"
	"delivery-backend/internal/modules/hr_legal/service"
)

type HRLegalHandler struct {
	svc service.HRLegalService
}

func NewHRLegalHandler(svc service.HRLegalService) *HRLegalHandler {
	return &HRLegalHandler{svc: svc}
}

func getBranchID(c *gin.Context) *uuid.UUID {
	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if val, ok := bid.(*uuid.UUID); ok && val != nil {
			return val
		}
		if val, ok := bid.(uuid.UUID); ok && val != uuid.Nil {
			return &val
		}
	}
	return nil
}

// ------------------------------------------------------------------
// 1. Investigation Endpoints
// ------------------------------------------------------------------
func (h *HRLegalHandler) CreateInvestigation(c *gin.Context) {
	var req dto.CreateInvestigationRequest
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

	resp, err := h.svc.CreateInvestigation(c.Request.Context(), req, supervisorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
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

func (h *HRLegalHandler) GetPendingInvestigationCount(c *gin.Context) {
	count, err := h.svc.GetPendingInvestigationCount(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pending_count": count})
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

func (h *HRLegalHandler) UpdateInvestigation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateInvestigationRequest
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

func (h *HRLegalHandler) ApproveInvestigation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	adminIDVal, _ := c.Get("admin_id")
	adminID, _ := adminIDVal.(uuid.UUID)
	adminName := c.GetString("admin_name")
	adminUsername := c.GetString("admin_email")

	resp, err := h.svc.ApproveInvestigation(c.Request.Context(), id, adminID, adminName, adminUsername)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ------------------------------------------------------------------
// 2. Document Endpoints
// ------------------------------------------------------------------
func (h *HRLegalHandler) CreateDocument(c *gin.Context) {
	var req dto.CreateEmployeeDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	doc, err := h.svc.CreateDocument(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, doc)
}

func (h *HRLegalHandler) UpdateDocument(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateEmployeeDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	doc, err := h.svc.UpdateDocument(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (h *HRLegalHandler) DeleteDocument(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.svc.DeleteDocument(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف المستند بنجاح"})
}

func (h *HRLegalHandler) GetDocumentByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	doc, err := h.svc.GetDocumentByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "المستند غير موجود"})
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (h *HRLegalHandler) GetAllDocuments(c *gin.Context) {
	var filter dto.EmployeeDocumentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	list, total, err := h.svc.GetAllDocuments(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
		"page":  filter.Page,
		"limit": filter.GetEffectiveLimit(),
	})
}

func (h *HRLegalHandler) GetExpiringDocuments(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	list, err := h.svc.GetExpiringDocuments(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// ------------------------------------------------------------------
// 3. Bank Account Endpoints
// ------------------------------------------------------------------
func (h *HRLegalHandler) CreateBankAccount(c *gin.Context) {
	var req dto.CreateEmployeeBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	acc, err := h.svc.CreateBankAccount(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, acc)
}

func (h *HRLegalHandler) UpdateBankAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateEmployeeBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	acc, err := h.svc.UpdateBankAccount(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, acc)
}

func (h *HRLegalHandler) DeleteBankAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.svc.DeleteBankAccount(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الحساب البنكي بنجاح"})
}

func (h *HRLegalHandler) GetAllBankAccounts(c *gin.Context) {
	var filter dto.EmployeeBankAccountFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	list, total, err := h.svc.GetAllBankAccounts(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
		"page":  filter.Page,
		"limit": filter.GetEffectiveLimit(),
	})
}

// ------------------------------------------------------------------
// 4. Leave Request Endpoints
// ------------------------------------------------------------------
func (h *HRLegalHandler) CreateLeaveRequest(c *gin.Context) {
	var req dto.CreateLeaveRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	leave, err := h.svc.CreateLeaveRequest(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, leave)
}

func (h *HRLegalHandler) UpdateLeaveRequestStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateLeaveRequestStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	if req.ApprovedByName == "" {
		req.ApprovedByName = c.GetString("admin_name")
	}

	leave, err := h.svc.UpdateLeaveRequestStatus(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, leave)
}

func (h *HRLegalHandler) ApproveLeaveRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	adminName := c.GetString("admin_name")
	leave, err := h.svc.UpdateLeaveRequestStatus(c.Request.Context(), id, dto.UpdateLeaveRequestStatusRequest{
		Status:         "APPROVED",
		ApprovedByName: adminName,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, leave)
}

func (h *HRLegalHandler) RejectLeaveRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	adminName := c.GetString("admin_name")
	leave, err := h.svc.UpdateLeaveRequestStatus(c.Request.Context(), id, dto.UpdateLeaveRequestStatusRequest{
		Status:         "REJECTED",
		ApprovedByName: adminName,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, leave)
}

func (h *HRLegalHandler) DeleteLeaveRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.svc.DeleteLeaveRequest(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف طلب الإجازة بنجاح"})
}

func (h *HRLegalHandler) GetAllLeaveRequests(c *gin.Context) {
	var filter dto.LeaveRequestFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	list, total, err := h.svc.GetAllLeaveRequests(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
		"page":  filter.Page,
		"limit": filter.GetEffectiveLimit(),
	})
}

func (h *HRLegalHandler) GetPendingLeaveCount(c *gin.Context) {
	count, err := h.svc.GetPendingLeaveCount(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pending_count": count})
}

// ------------------------------------------------------------------
// 5. Traffic Violation Endpoints
// ------------------------------------------------------------------
func (h *HRLegalHandler) CreateViolation(c *gin.Context) {
	var req dto.CreateTrafficViolationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	if bid := getBranchID(c); bid != nil {
		req.BranchID = bid
	}

	v, err := h.svc.CreateViolation(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *HRLegalHandler) UpdateViolation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateTrafficViolationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	v, err := h.svc.UpdateViolation(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *HRLegalHandler) DeleteViolation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.svc.DeleteViolation(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف المخالفة بنجاح"})
}

func (h *HRLegalHandler) GetAllViolations(c *gin.Context) {
	var filter dto.TrafficViolationFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	// If the authenticated user is an employee (delegate), FORCE filter to strictly their own employee_id
	if isEmp, _ := c.Get("is_employee"); isEmp == true {
		if empIDVal, exists := c.Get("employee_id"); exists {
			if empID, ok := empIDVal.(uuid.UUID); ok {
				filter.EmployeeID = &empID
			}
		}
	}

	list, total, totalAmount, deductedAmount, err := h.svc.GetAllViolations(c.Request.Context(), filter, getBranchID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":            list,
		"total":           total,
		"total_amount":    totalAmount,
		"deducted_amount": deductedAmount,
		"total_count":     total,
		"page":            filter.Page,
		"limit":           filter.GetEffectiveLimit(),
	})
}

// ------------------------------------------------------------------
// 6. Fuel Log Endpoints
// ------------------------------------------------------------------
func (h *HRLegalHandler) CreateFuelLog(c *gin.Context) {
	var req dto.CreateFuelLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	log, err := h.svc.CreateFuelLog(c.Request.Context(), req, getBranchID(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, log)
}

func (h *HRLegalHandler) UpdateFuelLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req dto.UpdateFuelLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	log, err := h.svc.UpdateFuelLog(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, log)
}

func (h *HRLegalHandler) DeleteFuelLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.svc.DeleteFuelLog(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف سجل الوقود بنجاح"})
}

func (h *HRLegalHandler) GetAllFuelLogs(c *gin.Context) {
	var filter dto.FuelLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	list, total, err := h.svc.GetAllFuelLogs(c.Request.Context(), filter, getBranchID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
		"page":  filter.Page,
		"limit": filter.GetEffectiveLimit(),
	})
}
