package handler

import (
	"net/http"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/repository"
	"delivery-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InvestigationHandler struct {
	svc     service.InvestigationService
	empRepo repository.EmployeeRepository
}

func NewInvestigationHandler(svc service.InvestigationService, empRepo repository.EmployeeRepository) *InvestigationHandler {
	return &InvestigationHandler{svc: svc, empRepo: empRepo}
}

func (h *InvestigationHandler) Create(c *gin.Context) {
	var req dto.CreateInvestigationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى إدخال بيانات صحيحة"})
		return
	}

	if req.Type == "investigation" {
		if len(req.Questions) == 0 || len(req.Answers) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "يجب إدخال الأسئلة والإجابات"})
			return
		}

		if len(req.Questions) != len(req.Answers) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "عدد الإجابات يجب أن يساوي عدد الأسئلة"})
			return
		}
	}

	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	adminID := adminIDVal.(uuid.UUID)

	// Verify branch access for the employee being investigated
	empID, parseErr := uuid.Parse(req.EmployeeID)
	if parseErr == nil {
		emp, empErr := h.empRepo.FindByID(c.Request.Context(), empID)
		if empErr == nil && !checkEmployeeBranchAccess(c, emp) {
			return
		}
	}

	result, err := h.svc.Create(c.Request.Context(), req, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *InvestigationHandler) GetAll(c *gin.Context) {
	branchID := getBranchID(c)
	result, err := h.svc.GetAll(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *InvestigationHandler) PendingCount(c *gin.Context) {
	count, err := h.svc.GetPendingApprovalCount(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في جلب عدد الطلبات"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *InvestigationHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	// Verify branch access
	inv, svcErr := h.svc.GetByID(c.Request.Context(), id)
	if svcErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "التحقيق غير موجود"})
		return
	}
	empID := inv.EmployeeID
	emp, empErr := h.empRepo.FindByID(c.Request.Context(), empID)
	if empErr == nil && !checkEmployeeBranchAccess(c, emp) {
		return
	}

	var req dto.UpdateInvestigationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى إدخال بيانات صحيحة"})
		return
	}

	result, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *InvestigationHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	inv, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "التحقيق غير موجود"})
		return
	}

	// Verify branch access
	empID := inv.EmployeeID
	emp, empErr := h.empRepo.FindByID(c.Request.Context(), empID)
	if empErr == nil && !checkEmployeeBranchAccess(c, emp) {
		return
	}

	c.JSON(http.StatusOK, inv)
}

func (h *InvestigationHandler) GetPublicByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	inv, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الوثيقة غير موجودة"})
		return
	}

	c.JSON(http.StatusOK, inv)
}

func (h *InvestigationHandler) Approve(c *gin.Context) {
	h.handleApproval(c, true)
}

func (h *InvestigationHandler) Reject(c *gin.Context) {
	h.handleApproval(c, false)
}

func (h *InvestigationHandler) handleApproval(c *gin.Context, approve bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if c.GetString("admin_role") != "ADMIN" {
		c.JSON(http.StatusForbidden, gin.H{"error": "هذه العملية متاحة للمدير (الأدمن) فقط"})
		return
	}

	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	adminID := adminIDVal.(uuid.UUID)

	// Verify branch access
	inv, svcErr := h.svc.GetByID(c.Request.Context(), id)
	if svcErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "التحقيق غير موجود"})
		return
	}
	emp, empErr := h.empRepo.FindByID(c.Request.Context(), inv.EmployeeID)
	if empErr == nil && !checkEmployeeBranchAccess(c, emp) {
		return
	}

	var result *dto.InvestigationResponse
	if approve {
		result, err = h.svc.Approve(c.Request.Context(), id, adminID)
	} else {
		result, err = h.svc.Reject(c.Request.Context(), id, adminID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
