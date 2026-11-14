package hr_legal

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *HRLegalHandler) CreateLeaveRequest(c *gin.Context) {
	var req CreateLeaveRequestRequest
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

	var req UpdateLeaveRequestStatusRequest
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
	leave, err := h.svc.UpdateLeaveRequestStatus(c.Request.Context(), id, UpdateLeaveRequestStatusRequest{
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
	leave, err := h.svc.UpdateLeaveRequestStatus(c.Request.Context(), id, UpdateLeaveRequestStatusRequest{
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
	var filter LeaveRequestFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	// Strict Ownership: If caller is an employee/courier, force filter to their own leaves
	if isEmp, _ := c.Get("is_employee"); isEmp == true {
		if empIDVal, exists := c.Get("employee_id"); exists {
			if empID, ok := empIDVal.(uuid.UUID); ok && empID != uuid.Nil {
				filter.EmployeeID = &empID
			}
		}
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
