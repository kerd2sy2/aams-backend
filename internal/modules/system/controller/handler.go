package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/system/dto"
	"delivery-backend/internal/modules/system/service"
)

// Helper to extract branch ID from gin context
func getBranchID(c *gin.Context) *uuid.UUID {
	val, exists := c.Get("branch_id")
	if !exists || val == nil {
		return nil
	}
	if bID, ok := val.(uuid.UUID); ok && bID != uuid.Nil {
		return &bID
	}
	return nil
}

// ---------------- Dashboard Handler ----------------
type DashboardHandler struct {
	svc service.DashboardService
}

func NewDashboardHandler(svc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	branchID := getBranchID(c)
	stats, err := h.svc.GetStats(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تحميل إحصائيات لوحة التحكم"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// ---------------- Report Handler ----------------
type ReportHandler struct {
	svc service.ReportService
}

func NewReportHandler(svc service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

func (h *ReportHandler) GetReports(c *gin.Context) {
	var filter dto.ReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	branchID := getBranchID(c)
	if branchID != nil {
		filter.BranchID = branchID
	}

	reports, total, err := h.svc.GetReports(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب التقارير"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  reports,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

func (h *ReportHandler) ExportReports(c *gin.Context) {
	var filter dto.ReportFilter
	_ = c.ShouldBindQuery(&filter)

	branchID := getBranchID(c)
	if branchID != nil {
		filter.BranchID = branchID
	}

	csvData, err := h.svc.ExportReports(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تصدير التقارير"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=work_reports_%s.csv", time.Now().Format("2006-01-02")))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvData)
}

func (h *ReportHandler) GetDailyReport(c *gin.Context) {
	dateStr := c.Query("date")
	branchID := getBranchID(c)

	rep, err := h.svc.GetDailyReport(c.Request.Context(), dateStr, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب التقرير اليومي"})
		return
	}

	c.JSON(http.StatusOK, rep)
}

func (h *ReportHandler) ExportDailyReport(c *gin.Context) {
	dateStr := c.Query("date")
	branchID := getBranchID(c)

	csvData, err := h.svc.ExportDailyReport(c.Request.Context(), dateStr, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تصدير التقرير اليومي"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=daily_report_%s.csv", dateStr))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvData)
}

// ---------------- Audit Handler ----------------
type AuditHandler struct {
	svc service.AuditService
}

func NewAuditHandler(svc service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) GetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	branchID := getBranchID(c)

	logs, total, err := h.svc.GetLogs(c.Request.Context(), branchID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب سجلات التدقيق"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *AuditHandler) ClearLogs(c *gin.Context) {
	branchID := getBranchID(c)
	if err := h.svc.ClearLogs(c.Request.Context(), branchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل مسح السجلات"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم مسح سجلات التدقيق بنجاح"})
}

func (h *AuditHandler) BulkDeleteLogs(c *gin.Context) {
	var req dto.BulkDeleteAuditLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرفات غير صالحة"})
		return
	}

	if err := h.svc.BulkDeleteLogs(c.Request.Context(), req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حذف السجلات"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف السجلات المحددة بنجاح"})
}

func (h *AuditHandler) DeleteLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف سجل غير صالح"})
		return
	}

	if err := h.svc.DeleteLog(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حذف السجل"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف السجل بنجاح"})
}

// ---------------- Setting Handler ----------------
type SettingHandler struct {
	svc service.SettingService
}

func NewSettingHandler(svc service.SettingService) *SettingHandler {
	return &SettingHandler{svc: svc}
}

func (h *SettingHandler) GetPublicSettings(c *gin.Context) {
	settings, err := h.svc.GetAllSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب الإعدادات"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *SettingHandler) GetSettings(c *gin.Context) {
	settings, err := h.svc.GetAllSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب الإعدادات"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *SettingHandler) UpdateSettings(c *gin.Context) {
	var req dto.UpdateAppSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	if err := h.svc.UpdateAppSettings(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تحديث الإعدادات"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث الإعدادات بنجاح"})
}

// ---------------- Notification Handler ----------------
type NotificationHandler struct {
	svc service.NotificationService
}

func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
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
	var req dto.CreateBroadcastRequest
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

func (h *NotificationHandler) GetEmployeeBroadcasts(c *gin.Context) {
	empIDVal, exists := c.Get("employee_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	empID, _ := empIDVal.(uuid.UUID)
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
	empIDVal, exists := c.Get("employee_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	empID, _ := empIDVal.(uuid.UUID)
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

	empIDVal, exists := c.Get("employee_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	empID, _ := empIDVal.(uuid.UUID)

	if err := h.svc.MarkEmployeeBroadcastRead(c.Request.Context(), empID, broadcastID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تحديث حالة القراءة"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تعيين الإشعار كمقروء"})
}

func (h *NotificationHandler) MarkAllEmployeeBroadcastsRead(c *gin.Context) {
	empIDVal, exists := c.Get("employee_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	empID, _ := empIDVal.(uuid.UUID)
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

	empIDVal, exists := c.Get("employee_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	empID, _ := empIDVal.(uuid.UUID)

	var req dto.SubmitPollVoteRequest
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
	empIDVal, exists := c.Get("employee_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	empID, _ := empIDVal.(uuid.UUID)

	var req dto.PushTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رمز الإشعارات مطلوب"})
		return
	}

	if err := h.svc.SaveEmployeePushToken(c.Request.Context(), empID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل حفظ رمز الإشعارات"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حفظ رمز الإشعارات بنجاح"})
}

// ---------------- Archive Handler ----------------
type ArchiveHandler struct {
	svc service.ArchiveService
}

func NewArchiveHandler(svc service.ArchiveService) *ArchiveHandler {
	return &ArchiveHandler{svc: svc}
}

func (h *ArchiveHandler) GetArchived(c *gin.Context) {
	var filter dto.ArchiveFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	branchID := getBranchID(c)
	if branchID != nil {
		filter.BranchID = branchID
	}

	res, err := h.svc.GetArchivedItems(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *ArchiveHandler) Restore(c *gin.Context) {
	var req dto.RestoreArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الاسترجاع غير مكتملة", "details": err.Error()})
		return
	}

	if err := h.svc.Restore(c.Request.Context(), req.Type, req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم استرجاع العنصر بنجاح"})
}

func (h *ArchiveHandler) PermanentDelete(c *gin.Context) {
	var req dto.PermanentDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		itemType := c.Query("type")
		idStr := c.Query("id")
		parsedID, errParse := uuid.Parse(idStr)
		if itemType == "" || errParse != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الحذف النهائي غير صالحة"})
			return
		}
		req.Type = itemType
		req.ID = parsedID
	}

	if err := h.svc.PermanentDelete(c.Request.Context(), req.Type, req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف العنصر نهائياً وبنجاح"})
}

func (h *ArchiveHandler) BulkRestore(c *gin.Context) {
	var req dto.BulkArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الاسترجاع الجماعي غير صالحة", "details": err.Error()})
		return
	}

	if err := h.svc.BulkRestore(c.Request.Context(), req.Type, req.IDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("تم استرجاع %d عناصر بنجاح", len(req.IDs))})
}

func (h *ArchiveHandler) BulkPermanentDelete(c *gin.Context) {
	var req dto.BulkArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الحذف الجماعي غير صالحة", "details": err.Error()})
		return
	}

	if err := h.svc.BulkPermanentDelete(c.Request.Context(), req.Type, req.IDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("تم حذف %d عناصر نهائياً بنجاح", len(req.IDs))})
}
