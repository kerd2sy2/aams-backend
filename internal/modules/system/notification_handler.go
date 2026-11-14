package system

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	svc NotificationService
}

func NewNotificationHandler(svc NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	adminID, _ := adminIDVal.(uuid.UUID)
	branchID := getBranchID(c)
	status := c.Query("status")

	notifs, err := h.svc.GetMyNotifications(c.Request.Context(), adminID, branchID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب الإشعارات"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": notifs})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	adminID, _ := adminIDVal.(uuid.UUID)

	if err := h.svc.MarkAllAsRead(c.Request.Context(), adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تحديث حالة الإشعارات"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديد جميع الإشعارات كمقروءة"})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف إشعار غير صالح"})
		return
	}

	adminIDVal, _ := c.Get("admin_id")
	adminID, _ := adminIDVal.(uuid.UUID)

	if err := h.svc.MarkAsRead(c.Request.Context(), id, adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تحديث الإشعار"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث الإشعار كمقروء"})
}

func (h *NotificationHandler) SendBroadcast(c *gin.Context) {
	var req CreateBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الإشعار غير مكتملة"})
		return
	}

	createdBy := c.GetString("admin_name")
	if createdBy == "" {
		createdBy = "الإدارة"
	}

	broadcast, err := h.svc.SendBroadcast(c.Request.Context(), req, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "تم إرسال الإشعار الجماعي بنجاح", "data": broadcast})
}

func (h *NotificationHandler) GetBroadcasts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	branchID := getBranchID(c)

	broadcasts, total, err := h.svc.GetBroadcasts(c.Request.Context(), branchID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب الإعلانات الجماعية"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  broadcasts,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *NotificationHandler) DeleteBroadcast(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.svc.DeleteBroadcast(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حذف الإشعار الجماعي"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الإشعار الجماعي بنجاح"})
}

func (h *NotificationHandler) GetBroadcastVotes(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	votes, err := h.svc.GetBroadcastVotes(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب أصوات الاستبيان"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": votes})
}

func resolveActorEmployeeID(c *gin.Context) (uuid.UUID, bool) {
	if empIDVal, exists := c.Get("employee_id"); exists {
		if id, ok := empIDVal.(uuid.UUID); ok && id != uuid.Nil {
			return id, true
		}
	}
	if qID := c.Query("employee_id"); qID != "" {
		if id, err := uuid.Parse(qID); err == nil && id != uuid.Nil {
			return id, true
		}
	}
	if aIDVal, exists := c.Get("admin_id"); exists {
		if id, ok := aIDVal.(uuid.UUID); ok && id != uuid.Nil {
			return id, true
		}
	}
	return uuid.Nil, false
}

func (h *NotificationHandler) GetEmployeeBroadcasts(c *gin.Context) {
	empID, exists := resolveActorEmployeeID(c)
	if !exists {
		if _, isAdmin := c.Get("admin_id"); isAdmin {
			c.JSON(http.StatusOK, gin.H{"data": []BroadcastItemDTO{}})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	branchID := getBranchID(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	broadcasts, err := h.svc.GetEmployeeBroadcasts(c.Request.Context(), empID, branchID, time.Time{}, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب الإشعارات"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": broadcasts})
}

func (h *NotificationHandler) GetEmployeeUnreadBroadcasts(c *gin.Context) {
	empID, exists := resolveActorEmployeeID(c)
	if !exists {
		if _, isAdmin := c.Get("admin_id"); isAdmin {
			c.JSON(http.StatusOK, gin.H{"data": []BroadcastItemDTO{}, "count": 0})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	branchID := getBranchID(c)

	unread, err := h.svc.GetEmployeeUnreadBroadcasts(c.Request.Context(), empID, branchID, time.Time{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب الإشعارات غير المقروءة"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": unread, "count": len(unread)})
}

func (h *NotificationHandler) MarkEmployeeBroadcastRead(c *gin.Context) {
	idStr := c.Param("id")
	broadcastID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف إشعار غير صالح"})
		return
	}

	empID, exists := resolveActorEmployeeID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}

	if err := h.svc.MarkEmployeeBroadcastRead(c.Request.Context(), empID, broadcastID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تحديث حالة القراءة"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تعيين الإشعار كمقروء"})
}

func (h *NotificationHandler) MarkAllEmployeeBroadcastsRead(c *gin.Context) {
	empID, exists := resolveActorEmployeeID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	branchID := getBranchID(c)

	if err := h.svc.MarkAllEmployeeBroadcastsRead(c.Request.Context(), empID, branchID, time.Time{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تحديث الإشعارات"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تعيين جميع الإشعارات كمقروءة"})
}

func (h *NotificationHandler) SubmitVote(c *gin.Context) {
	idStr := c.Param("id")
	broadcastID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	empID, exists := resolveActorEmployeeID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}

	var req SubmitPollVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "خيار التصويت غير صالح"})
		return
	}

	if err := h.svc.SubmitVote(c.Request.Context(), empID, broadcastID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تسجيل صوتك بنجاح"})
}

func (h *NotificationHandler) SaveEmployeePushToken(c *gin.Context) {
	empID, exists := resolveActorEmployeeID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}

	var req PushTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}
	log.Printf("[PushToken] employee=%s language=%q has_token=%t", empID, req.Language, req.PushToken != "")

	if err := h.svc.SaveEmployeePushToken(c.Request.Context(), empID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حفظ رمز الإشعارات واللغة"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حفظ رمز الإشعارات واللغة بنجاح"})
}
