package handler

import (
	"io"
	"net/http"
	"strings"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TargetHandler struct {
	targetService      service.TargetService
	excelImportService service.ExcelImportService
}

func NewTargetHandler(targetService service.TargetService, excelImportService service.ExcelImportService) *TargetHandler {
	return &TargetHandler{
		targetService:      targetService,
		excelImportService: excelImportService,
	}
}

// PreviewExcelImport parses uploaded excel and returns stats & duplicate analysis
func (h *TargetHandler) PreviewExcelImport(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "يرجى اختيار ملف إكسل بصيغة .xlsx (Excel file is required)"})
		return
	}

	customDate := strings.TrimSpace(c.PostForm("date"))

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فشل في فتح الملف المرفوع"})
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في قراءة محتوى الملف"})
		return
	}

	preview, err := h.excelImportService.ParseAndPreviewExcel(c.Request.Context(), fileBytes, fileHeader.Filename, customDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preview)
}

// ConfirmExcelImport saves the verified preview rows to PostgreSQL
func (h *TargetHandler) ConfirmExcelImport(c *gin.Context) {
	var req dto.ConfirmImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات الطلب غير صالحة: " + err.Error()})
		return
	}

	adminIDVal, _ := c.Get("admin_id")
	adminID, _ := adminIDVal.(uuid.UUID)
	adminNameVal, _ := c.Get("admin_name")
	adminName, _ := adminNameVal.(string)
	if adminName == "" {
		adminName = "الأدمن"
	}

	resp, err := h.excelImportService.ConfirmImport(c.Request.Context(), req, adminID, adminName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetDashboardSummary returns top metrics, charts, and alert counts
func (h *TargetHandler) GetDashboardSummary(c *gin.Context) {
	month := strings.TrimSpace(c.Query("month"))
	summary, err := h.targetService.GetDashboardSummary(c.Request.Context(), month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// ListIdentifiers returns all identifiers with target performance & status
func (h *TargetHandler) ListIdentifiers(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))
	status := strings.TrimSpace(c.Query("status"))
	month := strings.TrimSpace(c.Query("month"))

	list, err := h.targetService.ListIdentifiers(c.Request.Context(), search, status, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetIdentifierDetails returns deep breakdown: linked drivers, apps, predictions
func (h *TargetHandler) GetIdentifierDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	month := strings.TrimSpace(c.Query("month"))
	details, err := h.targetService.GetIdentifierDetails(c.Request.Context(), id, month)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, details)
}

// CreateIdentifier creates a new identifier (Admin only)
func (h *TargetHandler) CreateIdentifier(c *gin.Context) {
	var body struct {
		Name          string `json:"name" binding:"required"`
		AppName       string `json:"app_name"`
		Code          string `json:"code"`
		MonthlyTarget int    `json:"monthly_target"`
		DailyTarget   int    `json:"daily_target"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "اسم المعرف مطلوب"})
		return
	}

	ident, err := h.targetService.CreateIdentifier(c.Request.Context(), body.Name, body.AppName, body.Code, body.MonthlyTarget, body.DailyTarget)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ident)
}

// UpdateIdentifier updates identifier info (Admin only)
func (h *TargetHandler) UpdateIdentifier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	var body struct {
		Name          string `json:"name"`
		AppName       string `json:"app_name"`
		Code          string `json:"code"`
		MonthlyTarget int    `json:"monthly_target"`
		DailyTarget   int    `json:"daily_target"`
		IsActive      *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	isActive := true
	if body.IsActive != nil {
		isActive = *body.IsActive
	}

	if err := h.targetService.UpdateIdentifier(c.Request.Context(), id, body.Name, body.AppName, body.Code, body.MonthlyTarget, body.DailyTarget, isActive); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث بيانات المعرف بنجاح"})
}

// DeleteIdentifier deletes an identifier (Admin only)
func (h *TargetHandler) DeleteIdentifier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف غير صالح"})
		return
	}

	if err := h.targetService.DeleteIdentifier(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف المعرف بنجاح"})
}

// ListDrivers returns all drivers with order summaries
func (h *TargetHandler) ListDrivers(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))
	month := strings.TrimSpace(c.Query("month"))

	list, err := h.targetService.ListDrivers(c.Request.Context(), search, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// ListAlerts returns underperformance alerts
func (h *TargetHandler) ListAlerts(c *gin.Context) {
	date := strings.TrimSpace(c.Query("date"))
	unresolved := c.Query("unresolved_only") == "true"

	list, err := h.targetService.ListAlerts(c.Request.Context(), date, unresolved)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// ResolveAlert marks an alert as resolved
func (h *TargetHandler) ResolveAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف التنبيه غير صالح"})
		return
	}

	if err := h.targetService.ResolveAlert(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم تسوية التنبيه بنجاح"})
}

// GetTargetSettings returns default targets
func (h *TargetHandler) GetTargetSettings(c *gin.Context) {
	settings, err := h.targetService.GetTargetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// UpdateTargetSettings updates default target settings (Admin only)
func (h *TargetHandler) UpdateTargetSettings(c *gin.Context) {
	var body struct {
		DefaultMonthlyTarget int `json:"default_monthly_target"`
		DefaultDailyTarget   int `json:"default_daily_target"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	if err := h.targetService.UpdateTargetSettings(c.Request.Context(), body.DefaultMonthlyTarget, body.DefaultDailyTarget); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث إعدادات التارچت بنجاح"})
}

// ListImportBatches returns recent imported batches (daily sheets)
func (h *TargetHandler) ListImportBatches(c *gin.Context) {
	batches, err := h.targetService.ListImportBatches(c.Request.Context(), 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, batches)
}

// DeleteImportBatch deletes a specific batch and its associated daily orders
func (h *TargetHandler) DeleteImportBatch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الرفعة غير صالح"})
		return
	}

	if err := h.targetService.DeleteImportBatch(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في حذف الشيت: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف تقرير الشيت وجميع طلباته المرتبطة بنجاح"})
}

// DeleteSheetByDate deletes all imported orders and batches for a specific order date
func (h *TargetHandler) DeleteSheetByDate(c *gin.Context) {
	orderDate := strings.TrimSpace(c.Param("orderDate"))
	if orderDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "تاريخ الشيت مطلوب"})
		return
	}

	if err := h.targetService.DeleteSheetByDate(c.Request.Context(), orderDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في حذف شيت التاريخ: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف شيت وجميع طلبات تاريخ " + orderDate + " بنجاح"})
}
