package handler

import (
	"net/http"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/dto"
	"delivery-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CustodyHandler struct {
	svc service.CustodyService
}

func NewCustodyHandler(svc service.CustodyService) *CustodyHandler {
	return &CustodyHandler{svc: svc}
}

// custodyBranchID returns the branch from the admin context (for supervisors),
// or an explicit branch_id query param (for general admins).
func custodyBranchID(c *gin.Context) *uuid.UUID {
	if bid := getBranchID(c); bid != nil {
		return bid
	}
	if s := c.Query("branch_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			return &id
		}
	}
	return nil
}

func getAdminContext(c *gin.Context) *domain.Admin {
	// First try the full admin object (some handlers may set it directly)
	if a, exists := c.Get("admin"); exists {
		if adminObj, ok := a.(*domain.Admin); ok {
			return adminObj
		}
	}

	// Fall back to individual keys set by AuthMiddleware
	adminIDVal, hasID := c.Get("admin_id")
	adminNameVal, hasName := c.Get("admin_name")
	if !hasID && !hasName {
		return nil
	}

	admin := &domain.Admin{}
	if hasID {
		if id, ok := adminIDVal.(uuid.UUID); ok {
			admin.ID = id
		}
	}
	if hasName {
		if name, ok := adminNameVal.(string); ok {
			admin.Name = name
		}
	}
	if emailVal, ok := c.Get("admin_email"); ok {
		if email, ok := emailVal.(string); ok {
			admin.Email = email
		}
	}
	if roleVal, ok := c.Get("admin_role"); ok {
		if role, ok := roleVal.(string); ok {
			admin.Role = role
		}
	}
	// Derive username from email (part before @) if not set separately
	if admin.Username == "" && admin.Email != "" {
		if at := len(admin.Email); at > 0 {
			for i, ch := range admin.Email {
				if ch == '@' {
					admin.Username = admin.Email[:i]
					break
				}
			}
		}
	}
	if admin.Username == "" {
		admin.Username = admin.Name
	}
	return admin
}

func (h *CustodyHandler) List(c *gin.Context) {
	branchID := custodyBranchID(c)
	result, err := h.svc.List(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CustodyHandler) Create(c *gin.Context) {
	var req dto.CreateCustodyDayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى إدخال التاريخ ومبلغ العهدة بشكل صحيح"})
		return
	}

	// Supervisors are ALWAYS bound to their own branch. Ignore any branch_id
	// sent by the client (it may come from a stale localStorage value) and
	// force the branch from the authenticated admin context.
	if adminBranch := getBranchID(c); adminBranch != nil {
		req.BranchID = adminBranch
	}
	if req.BranchID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يجب تحديد الفرع"})
		return
	}

	admin := getAdminContext(c)
	result, err := h.svc.Create(c.Request.Context(), req, admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *CustodyHandler) AddAmount(c *gin.Context) {
	var req dto.AddCustodyAmountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى إدخال مبلغ العهدة الإضافي بشكل صحيح"})
		return
	}

	// Supervisors are ALWAYS bound to their own branch — ignore client input.
	if adminBranch := getBranchID(c); adminBranch != nil {
		req.BranchID = adminBranch
	}

	admin := getAdminContext(c)
	result, err := h.svc.AddAmount(c.Request.Context(), req, admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *CustodyHandler) AddExpense(c *gin.Context) {
	dayID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف العهدة غير صالح"})
		return
	}

	var req dto.CreateCustodyExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى إدخال بيانات المصروف بشكل صحيح"})
		return
	}

	branchID := getBranchID(c)
	admin := getAdminContext(c)
	result, err := h.svc.AddExpense(c.Request.Context(), dayID, branchID, req, admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *CustodyHandler) DeleteExpense(c *gin.Context) {
	expenseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المصروف غير صالح"})
		return
	}

	branchID := getBranchID(c)
	admin := getAdminContext(c)
	result, err := h.svc.DeleteExpense(c.Request.Context(), expenseID, branchID, admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *CustodyHandler) GetLogs(c *gin.Context) {
	var filter dto.CustodyLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معايير تصفية الحركة غير صالحة"})
		return
	}

	if filter.BranchID == "" {
		if bid := custodyBranchID(c); bid != nil {
			filter.BranchID = bid.String()
		}
	}

	logs, total, err := h.svc.GetLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"data":        logs,
		"total":       total,
		"page":        filter.Page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

func (h *CustodyHandler) DeleteLog(c *gin.Context) {
	logID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الحركة غير صالح"})
		return
	}

	branchID := getBranchID(c)
	admin := getAdminContext(c)
	if err := h.svc.DeleteLog(c.Request.Context(), logID, branchID, admin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف/إلغاء الحركة بنجاح وإعادة احتساب المبالغ"})
}
