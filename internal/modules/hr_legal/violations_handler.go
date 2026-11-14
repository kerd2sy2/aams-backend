package hr_legal

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *HRLegalHandler) CreateViolation(c *gin.Context) {
	var req CreateTrafficViolationRequest
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

	var req UpdateTrafficViolationRequest
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
	var filter TrafficViolationFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	// 1. Explicitly parse employee_id if provided in query
	if empStr := c.Query("employee_id"); empStr != "" {
		if parsed, err := uuid.Parse(empStr); err == nil {
			filter.EmployeeID = &parsed
		}
	}

	// 2. Explicitly parse branch_id if provided in query
	if branchStr := c.Query("branch_id"); branchStr != "" {
		if parsed, err := uuid.Parse(branchStr); err == nil {
			filter.BranchID = &parsed
		}
	}

	// 3. If the authenticated user is an employee (delegate), FORCE filter to strictly their own employee_id
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
