package auth

import (

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

)

type RoleHandler struct {
	svc RoleService
}

func NewRoleHandler(svc RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الدور غير صالحة"})
		return
	}

	role, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "تم إنشاء الدور بنجاح", "data": role})
}

func (h *RoleHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الدور غير صالح"})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	role, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث الدور بنجاح", "data": role})
}

func (h *RoleHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الدور غير صالح"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الدور بنجاح"})
}

func (h *RoleHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الدور غير صالح"})
		return
	}

	role, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الدور غير موجود"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": role})
}

func (h *RoleHandler) GetAll(c *gin.Context) {
	roles, err := h.svc.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (h *RoleHandler) GetPermissions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"permissions": []string{
		"*",
		"apps.ohda", "apps.eshraf", "apps.edara", "apps.manadeeb", "apps.dashboard",
		"all_branches", "branches.manage",
		"custody", "custody.access", "custody.view", "custody.add", "custody.expense", "custody.delete", "custody.manage",
		"employees.view", "employees.create", "employees.edit", "employees.delete", "employees.cards",
		"work.view", "work.start", "work.end",
		"vehicles.view", "vehicles.manage", "vehicles.oil",
		"fuel.view", "fuel.manage",
		"maintenance.view", "maintenance.manage",
		"attendance.view", "investigations.view", "investigations.create", "investigations.approve",
		"inventory.view", "inventory.manage", "inventory.dispense",
		"tickets.view", "tickets.manage",
		"bank_accounts.view", "bank_accounts.manage",
		"violations.view", "violations.manage",
		"reports.view", "reports.export",
		"documents.view", "documents.manage",
		"leaves.view", "leaves.manage",
		"partners.view", "partners.manage",
		"users.manage", "roles.manage", "settings.manage",
		"archive.view",
		"audit_logs.view", "error_logs.view",
	}})
}

// AdminHandler