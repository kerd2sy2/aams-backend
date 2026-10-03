package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/ticket/dto"
	"delivery-backend/internal/modules/ticket/service"
)

type SupportTicketHandler struct {
	ticketService service.SupportTicketService
}

func NewSupportTicketHandler(ticketService service.SupportTicketService) *SupportTicketHandler {
	return &SupportTicketHandler{ticketService: ticketService}
}

func (h *SupportTicketHandler) Create(c *gin.Context) {
	var req dto.CreateSupportTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(uuid.UUID); ok {
			branchID = &val
		}
	}

	ticket, err := h.ticketService.Create(c.Request.Context(), req, branchID)
	if err != nil {
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

	var req dto.UpdateSupportTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	ticket, err := h.ticketService.Update(c.Request.Context(), id, req)
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
	var filter dto.SupportTicketFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
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
