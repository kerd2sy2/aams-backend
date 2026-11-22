package ticket

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func isFieldEmployeeCaller(c *gin.Context) (uuid.UUID, bool) {
	isEmp, _ := c.Get("is_employee")
	role := strings.ToUpper(strings.TrimSpace(c.GetString("admin_role")))
	isSupervisorOrAdmin := role == "SUPERVISOR" || role == "ADMIN" || role == "SUPER_ADMIN" ||
		role == "OWNER" || role == "GENERAL_MANAGER" || role == "OPERATIONS_MANAGER" || role == "ACCOUNTANT"
	if !isSupervisorOrAdmin && (isEmp == true || role == "DRIVER" || role == "EMPLOYEE" || role == "DELEGATE") {
		if authEmpID, exists := c.Get("employee_id"); exists {
			if aUUID, ok := authEmpID.(uuid.UUID); ok && aUUID != uuid.Nil {
				return aUUID, true
			}
		}
	}
	return uuid.Nil, false
}

type SupportTicketHandler struct {
	ticketService SupportTicketService
}

func NewSupportTicketHandler(ticketService SupportTicketService) *SupportTicketHandler {
	return &SupportTicketHandler{ticketService: ticketService}
}

func (h *SupportTicketHandler) Create(c *gin.Context) {
	var req CreateSupportTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	enforceCooldown := false
	var branchID *uuid.UUID
	if empUUID, isFieldEmp := isFieldEmployeeCaller(c); isFieldEmp {
		req.EmployeeID = &empUUID
		enforceCooldown = true
	} else if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(uuid.UUID); ok {
			branchID = &val
		}
	}

	ticket, err := h.ticketService.Create(c.Request.Context(), req, branchID, enforceCooldown)
	if err != nil {
		if errors.Is(err, ErrComplaintCooldown) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error(), "code": "COMPLAINT_COOLDOWN"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

func (h *SupportTicketHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var req UpdateSupportTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	adminName := strings.TrimSpace(c.GetString("admin_name"))
	if adminName == "" {
		adminName = strings.TrimSpace(c.GetString("user_name"))
	}

	ticket, err := h.ticketService.Update(c.Request.Context(), id, req, adminName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (h *SupportTicketHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.ticketService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف التذكرة بنجاح"})
}

func (h *SupportTicketHandler) GetAll(c *gin.Context) {
	var filter SupportTicketFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}
	if s := strings.TrimSpace(filter.EmployeeIDStr); s != "" {
		if parsed, err := uuid.Parse(s); err == nil {
			filter.EmployeeID = &parsed
		}
	}
	if s := strings.TrimSpace(filter.BranchIDStr); s != "" {
		if parsed, err := uuid.Parse(s); err == nil {
			filter.BranchID = &parsed
		}
	}

	var branchID *uuid.UUID
	if empUUID, isFieldEmp := isFieldEmployeeCaller(c); isFieldEmp {
		filter.EmployeeID = &empUUID
		branchID = nil
	} else if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(uuid.UUID); ok {
			branchID = &val
		}
	}

	list, total, err := h.ticketService.GetAll(c.Request.Context(), filter, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	limit := filter.Limit
	if limit == 0 && filter.PageSize > 0 {
		limit = filter.PageSize
	}
	if limit == 0 {
		limit = 50
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
		"page":  filter.Page,
		"limit": limit,
	})
}
