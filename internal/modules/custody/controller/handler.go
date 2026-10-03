package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/custody/dto"
	"delivery-backend/internal/modules/custody/service"
)

type CustodyHandler struct {
	svc service.CustodyService
}

func NewCustodyHandler(svc service.CustodyService) *CustodyHandler {
	return &CustodyHandler{svc: svc}
}

func custodyBranchID(c *gin.Context) *uuid.UUID {
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

func getAdminInfo(c *gin.Context) (*uuid.UUID, string, string) {
	var adminID *uuid.UUID
	if idVal, exists := c.Get("admin_id"); exists && idVal != nil {
		if id, ok := idVal.(uuid.UUID); ok && id != uuid.Nil {
			adminID = &id
		}
	}
	adminName := c.GetString("admin_name")
	adminUsername := adminName
	if email := c.GetString("admin_email"); email != "" {
		parts := strings.Split(email, "@")
		if len(parts) > 0 && parts[0] != "" {
			adminUsername = parts[0]
		}
	}
	return adminID, adminName, adminUsername
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	if bid := custodyBranchID(c); bid != nil {
		req.BranchID = bid
	}

	adminID, adminName, adminUsername := getAdminInfo(c)
	result, err := h.svc.Create(c.Request.Context(), req, adminID, adminName, adminUsername)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *CustodyHandler) AddAmount(c *gin.Context) {
	var req dto.AddCustodyAmountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	if bid := custodyBranchID(c); bid != nil {
		req.BranchID = bid
	}

	adminID, adminName, adminUsername := getAdminInfo(c)
	result, err := h.svc.AddAmount(c.Request.Context(), req, adminID, adminName, adminUsername)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CustodyHandler) AddExpense(c *gin.Context) {
	dayID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف السجل غير صالح"})
		return
	}

	var req dto.CreateCustodyExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	branchID := custodyBranchID(c)
	adminID, adminName, adminUsername := getAdminInfo(c)
	result, err := h.svc.AddExpense(c.Request.Context(), dayID, branchID, req, adminID, adminName, adminUsername)
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

	branchID := custodyBranchID(c)
	adminID, adminName, adminUsername := getAdminInfo(c)
	result, err := h.svc.DeleteExpense(c.Request.Context(), expenseID, branchID, adminID, adminName, adminUsername)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CustodyHandler) GetLogs(c *gin.Context) {
	var filter dto.CustodyLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة: " + err.Error()})
		return
	}

	if bid := custodyBranchID(c); bid != nil {
		filter.BranchID = bid.String()
	}

	logs, total, err := h.svc.GetLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	limit := filter.GetEffectiveLimit()
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"data":        logs,
		"total":       total,
		"page":        filter.GetEffectivePage(),
		"limit":       limit,
		"total_pages": totalPages,
	})
}

func (h *CustodyHandler) DeleteLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف السجل غير صالح"})
		return
	}

	adminID, adminName, adminUsername := getAdminInfo(c)
	if err := h.svc.DeleteLog(c.Request.Context(), id, adminID, adminName, adminUsername); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف سجل التدقيق بنجاح"})
}

func (h *CustodyHandler) CheckAdminOnly(c *gin.Context) bool {
	role, exists := c.Get("admin_role")
	if !exists || (role != "admin" && role != "superadmin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح لك بالوصول"})
		return false
	}
	return true
}

func getIntQuery(c *gin.Context, key string, defaultVal int) int {
	if s := c.Query(key); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	return defaultVal
}
