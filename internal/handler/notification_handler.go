package handler

import (
	"fmt"
	"net/http"

	"delivery-backend/internal/dto"
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

// ------------------------------------------------------------------
// Broadcast Notifications (الإشعارات الجماعية لجميع الهواتف)
// ------------------------------------------------------------------

func (h *NotificationHandler) SendBroadcast(c *gin.Context) {
	var req dto.CreateBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى كتابة عنوان الإشعار ونصه"})
		return
	}

	adminName := "الإدارة العامة"
	if name, exists := c.Get("admin_name"); exists && name != nil {
		if s, ok := name.(string); ok && s != "" {
			adminName = s
		}
	} else if username, exists := c.Get("username"); exists && username != nil {
		if s, ok := username.(string); ok && s != "" {
			adminName = s
		}
	}

	broadcast, err := h.notifService.SendBroadcast(c.Request.Context(), req, adminName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل إرسال الإشعار الجماعي: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "تم إرسال الإشعار الجماعي بنجاح لجميع الهواتف",
		"broadcast": broadcast,
	})
}

func (h *NotificationHandler) GetBroadcasts(c *gin.Context) {
	var branchID *uuid.UUID
	branchIDStr := c.Query("branch_id")
	if branchIDStr != "" {
		if bid, err := uuid.Parse(branchIDStr); err == nil {
			branchID = &bid
		}
	}

	limit := 50
	offset := 0

	broadcasts, total, err := h.notifService.GetBroadcasts(c.Request.Context(), branchID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب قائمة الإشعارات: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  broadcasts,
		"total": total,
	})
}

func (h *NotificationHandler) DeleteBroadcast(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الإشعار غير صالح"})
		return
	}

	if err := h.notifService.DeleteBroadcast(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حذف الإشعار: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الإشعار بنجاح"})
}

func (h *NotificationHandler) GetEmployeeBroadcasts(c *gin.Context) {
	empIDStr := c.Query("employee_id")
	if empIDStr == "" {
		empIDStr = c.Param("employee_id")
	}
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	broadcasts, err := h.notifService.GetEmployeeBroadcasts(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب الإشعارات: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": broadcasts})
}

func (h *NotificationHandler) GetEmployeeUnreadBroadcasts(c *gin.Context) {
	empIDStr := c.Query("employee_id")
	if empIDStr == "" {
		empIDStr = c.Param("employee_id")
	}
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	unread, err := h.notifService.GetEmployeeUnreadBroadcasts(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل فحص الإشعارات غير المقروءة: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": len(unread),
		"data":         unread,
	})
}

func (h *NotificationHandler) MarkEmployeeBroadcastRead(c *gin.Context) {
	broadcastIDStr := c.Param("id")
	broadcastID, err := uuid.Parse(broadcastIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الإشعار غير صالح"})
		return
	}

	empIDStr := c.Query("employee_id")
	if empIDStr == "" {
		var body struct {
			EmployeeID string `json:"employee_id"`
		}
		_ = c.ShouldBindJSON(&body)
		empIDStr = body.EmployeeID
	}

	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	if err := h.notifService.MarkEmployeeBroadcastRead(c.Request.Context(), broadcastID, empID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تحديث حالة القراءة: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث حالة القراءة بنجاح"})
}

func (h *NotificationHandler) SaveEmployeePushToken(c *gin.Context) {
	empIDStr := c.Query("employee_id")
	if empIDStr == "" {
		empIDStr = c.Param("employee_id")
	}

	var req dto.PushTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رمز الإشعار PushToken مطلوب"})
		return
	}

	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	if err := h.notifService.SaveEmployeePushToken(c.Request.Context(), empID, req.PushToken, req.DeviceUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حفظ رمز الإشعارات: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تسجيل رمز الإشعارات بنجاح"})
}

func (h *NotificationHandler) SubmitVote(c *gin.Context) {
	broadcastIDStr := c.Param("id")
	broadcastID, err := uuid.Parse(broadcastIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الإشعار غير صالح"})
		return
	}

	var req dto.SubmitPollVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى تحديد اختيارك: موافق أو معترض"})
		return
	}

	empIDStr := c.Query("employee_id")
	if empIDStr == "" {
		empIDStr = c.Param("employee_id")
	}
	var empID uuid.UUID
	if empIDStr != "" {
		empID, err = uuid.Parse(empIDStr)
	} else if adminUUID, aErr := getAdminUUID(c); aErr == nil {
		empID = adminUUID
	}

	if empID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير موجود أو غير صالح"})
		return
	}

	if err := h.notifService.RecordVote(c.Request.Context(), broadcastID, empID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تسجيل التصويت: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "تم تسجيل تصويتك بنجاح",
		"response": req.Response,
	})
}

func (h *NotificationHandler) GetBroadcastVotes(c *gin.Context) {
	broadcastIDStr := c.Param("id")
	broadcastID, err := uuid.Parse(broadcastIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الإشعار غير صالح"})
		return
	}

	votes, err := h.notifService.GetBroadcastVotes(c.Request.Context(), broadcastID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب نتائج التصويت: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  votes,
		"total": len(votes),
	})
}
