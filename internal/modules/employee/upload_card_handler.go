package employee

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/pkg/middleware"
)

func (h *EmployeeHandler) GetBarcode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	emp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil || !scope.AllowsBranch(emp.BranchID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	if emp.Barcode == "" {
		emp.Barcode, _ = h.svc.GenerateBarcode(emp.NationalID)
	}

	c.JSON(http.StatusOK, gin.H{"barcode": emp.Barcode})
}

func (h *EmployeeHandler) GetQRCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	emp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil || !scope.AllowsBranch(emp.BranchID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	if emp.QRCode == "" {
		emp.QRCode, _ = h.svc.GenerateQRCode(emp.NationalID)
	}

	c.JSON(http.StatusOK, gin.H{"qrcode": emp.QRCode})
}

func (h *EmployeeHandler) GetPrintCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	emp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil || !scope.AllowsBranch(emp.BranchID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": emp})
}

func (h *EmployeeHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		file, err = c.FormFile("file")
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "الصورة مطلوبة"})
		return
	}

	// Strict extension allowlist (defense against XSS / HTML / SVG uploads)
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "نوع الصورة غير مدعوم، الصيغ المسموحة: jpg, jpeg, png, webp"})
		return
	}

	// Validate file size (max 15MB for images)
	if file.Size > 15*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "حجم الصورة يتجاوز الحد الأقصى المسموح (15 ميجابايت)"})
		return
	}

	// Inspect magic bytes for real image content
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "تعذر قراءة ملف الصورة"})
		return
	}
	buf := make([]byte, 512)
	n, readErr := src.Read(buf)
	_ = src.Close()
	if readErr != nil && n == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ملف الصورة فارغ أو تالف"})
		return
	}

	mimeType := http.DetectContentType(buf[:n])
	isValidImage := mimeType == "image/jpeg" || mimeType == "image/png" || mimeType == "image/webp"
	if !isValidImage && ext == ".webp" && n >= 12 {
		if string(buf[0:4]) == "RIFF" && string(buf[8:12]) == "WEBP" {
			isValidImage = true
		}
	}
	if !isValidImage {
		c.JSON(http.StatusBadRequest, gin.H{"error": "محتوى الملف لا يطابق صيغة صورة صالحة (JPEG, PNG, WEBP فقط)"})
		return
	}

	if err := os.MkdirAll("uploads", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في تهيئة مسار الحفظ"})
		return
	}

	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	savePath := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في حفظ الصورة"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + filename})
}

func (h *EmployeeHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "الملف مطلوب"})
		return
	}

	// Strict document & attachment allowlist
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".pdf":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".xlsx": true,
		".xls":  true,
		".csv":  true,
		".docx": true,
		".doc":  true,
		".txt":  true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "نوع الملف غير مدعوم، الصيغ المسموحة: pdf, xlsx, xls, csv, docx, doc, txt, والـ صور"})
		return
	}

	// Validate file size (max 50MB)
	if file.Size > 50*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "حجم الملف يتجاوز الحد الأقصى المسموح (50 ميجابايت)"})
		return
	}

	// Inspect magic bytes
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "تعذر قراءة الملف"})
		return
	}
	buf := make([]byte, 512)
	n, readErr := src.Read(buf)
	_ = src.Close()
	if readErr != nil && n == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "الملف فارغ أو تالف"})
		return
	}

	detectedMime := http.DetectContentType(buf[:n])
	lowerMime := strings.ToLower(detectedMime)
	if strings.Contains(lowerMime, "html") || strings.Contains(lowerMime, "javascript") || strings.Contains(lowerMime, "svg") || strings.Contains(lowerMime, "executable") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "نوع الملف يحتوي على محتوى غير آمن أو غير مصرح به"})
		return
	}

	// For PDF, ensure valid PDF magic header (%PDF)
	if ext == ".pdf" && (n < 4 || string(buf[:4]) != "%PDF") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "محتوى الملف لا يطابق صيغة PDF حقيقية"})
		return
	}

	if err := os.MkdirAll("uploads", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في تهيئة مسار الحفظ"})
		return
	}

	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	savePath := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في حفظ الملف"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + filename})
}
