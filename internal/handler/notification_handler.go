package handler

import (
	"fmt"
	"net/http"

	"delivery-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	notifService service.NotificationService
}

func NewNotificationHandler(s service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: s}
}

func getAdminUUID(c *gin.Context) (uuid.UUID, error) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("Unauthorized")
	}
	switch v := adminID.(type) {
	case string:
		return uuid.Parse(v)
	case uuid.UUID:
		return v, nil
	default:
		return uuid.Nil, fmt.Errorf("Invalid type")
	}
}

func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	id, err := getAdminUUID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	status := c.DefaultQuery("status", "")
	notifs, err := h.notifService.GetMyNotifications(c.Request.Context(), id, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifs)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	adminUUID, err := getAdminUUID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	notifIDStr := c.Param("id")
	notifUUID, err := uuid.Parse(notifIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := h.notifService.MarkAsRead(c.Request.Context(), notifUUID, adminUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	adminUUID, err := getAdminUUID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.notifService.MarkAllAsRead(c.Request.Context(), adminUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All marked as read"})
}
