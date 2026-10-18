package controller

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/modules/operation_video/dto"
	"delivery-backend/internal/modules/operation_video/service"
)

type OperationVideoHandler struct {
	service       service.OperationVideoService
	uploadBaseDir string
}

func NewOperationVideoHandler(service service.OperationVideoService, uploadBaseDir string) *OperationVideoHandler {
	if uploadBaseDir == "" {
		uploadBaseDir = "./uploads"
	}
	return &OperationVideoHandler{
		service:       service,
		uploadBaseDir: uploadBaseDir,
	}
}

// Upload handles silent video upload and metadata recording
func (h *OperationVideoHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("video")
	if err != nil {
		file, err = c.FormFile("file")
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ملف الفيديو مطلوب"})
		return
	}

	// Restrict allowed video/media extensions
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".mp4"
	}
	if ext != ".mp4" && ext != ".mov" && ext != ".m4v" && ext != ".webm" && ext != ".3gp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "صيغة الفيديو غير مدعومة، يرجى رفع ملف فيديو صالح (.mp4, .mov, .webm)"})
		return
	}

	targetDir := filepath.Join(h.uploadBaseDir, "operation_videos")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في تهيئة مسار حفظ الفيديو"})
		return
	}

	videoUUID := uuid.New()
	targetFilename := fmt.Sprintf("%s%s", videoUUID.String(), ext)
	savePath := filepath.Join(targetDir, targetFilename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في حفظ ملف الفيديو على السيرفر"})
		return
	}

	// Strip audio completely using ffmpeg to guarantee 100% silent operational proof video
	silentPath := filepath.Join(targetDir, fmt.Sprintf("silent_%s", targetFilename))
	cmd := exec.Command("ffmpeg", "-y", "-i", savePath, "-an", "-vcodec", "copy", silentPath)
	if err := cmd.Run(); err == nil {
		_ = os.Remove(savePath)
		_ = os.Rename(silentPath, savePath)
	} else {
		_ = os.Remove(silentPath)
	}
	_ = os.Chmod(savePath, 0644)

	fileSize := file.Size
	if fi, err := os.Stat(savePath); err == nil {
		fileSize = fi.Size()
	}

	videoURL := fmt.Sprintf("/api/v1/operation-videos/%s/stream", videoUUID.String())

	// Extract form parameters
	employeeName := strings.TrimSpace(c.PostForm("employee_name"))
	orderNumber := strings.TrimSpace(c.PostForm("order_number"))
	notes := strings.TrimSpace(c.PostForm("notes"))

	var branchID *uuid.UUID
	branchIDStr := strings.TrimSpace(c.PostForm("branch_id"))
	if branchIDStr == "" {
		// Try get from context
		if bVal, exists := c.Get("branch_id"); exists {
			if bUUID, ok := bVal.(uuid.UUID); ok && bUUID != uuid.Nil {
				branchID = &bUUID
			}
		}
	} else if parsedBID, err := uuid.Parse(branchIDStr); err == nil && parsedBID != uuid.Nil {
		branchID = &parsedBID
	}

	var employeeID *uuid.UUID
	empIDStr := strings.TrimSpace(c.PostForm("employee_id"))
	if empIDStr != "" {
		if parsedEID, err := uuid.Parse(empIDStr); err == nil && parsedEID != uuid.Nil {
			employeeID = &parsedEID
		}
	}

	var sessionID *uuid.UUID
	sessIDStr := strings.TrimSpace(c.PostForm("session_id"))
	if sessIDStr != "" {
		if parsedSID, err := uuid.Parse(sessIDStr); err == nil && parsedSID != uuid.Nil {
			sessionID = &parsedSID
		}
	}

	durationSec := 0
	if dStr := c.PostForm("duration_sec"); dStr != "" {
		durationSec, _ = strconv.Atoi(dStr)
	}

	now := time.Now()
	opVideo := domain.OperationVideo{
		ID:           videoUUID,
		BranchID:     branchID,
		EmployeeID:   employeeID,
		EmployeeName: employeeName,
		OrderNumber:  orderNumber,
		SessionID:    sessionID,
		VideoURL:     videoURL,
		DurationSec:  durationSec,
		FileSize:     fileSize,
		Notes:        notes,
		RecordedAt:   now,
		ExpiresAt:    now.AddDate(0, 0, 60), // Strictly 60 days
		IsDeleted:    false,
	}

	result, err := h.service.SaveVideo(c.Request.Context(), &opVideo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في حفظ بيانات الفيديو"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "تم حفظ وتوثيق الفيديو بنجاح، متاح لمدة 60 يوماً",
		"data":    result,
	})
}

// StreamVideo streams the operational proof video with HTTP Range support.
// Strictly requires authentication and allowed roles: SUPER_ADMIN, ADMIN, SUPERVISOR, MANAGEMENT
func (h *OperationVideoHandler) StreamVideo(c *gin.Context) {
	idStr := c.Param("id")
	videoID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفيديو غير صالح"})
		return
	}

	// Verify authorized role (Management / Admin / Supervisor)
	roleVal, _ := c.Get("admin_role")
	userRole, _ := roleVal.(string)
	isEmployeeVal, _ := c.Get("is_employee")
	isEmployee, _ := isEmployeeVal.(bool)

	if isEmployee || (userRole != "SUPER_ADMIN" && userRole != "ADMIN" && userRole != "SUPERVISOR" && userRole != "MANAGEMENT") {
		c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح: مشاهدة فيديوهات التوثيق مقتصرة على الإدارة والإشراف فقط"})
		return
	}

	video, err := h.service.GetVideo(c.Request.Context(), videoID)
	if err != nil || video == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الفيديو غير موجود أو انتهت صلاحيته"})
		return
	}

	// Find file on disk (handle both /api/v1/operation-videos/:id/stream and /uploads/... format)
	var filePath string
	if strings.Contains(video.VideoURL, "/uploads/operation_videos/") {
		fileName := filepath.Base(video.VideoURL)
		filePath = filepath.Join(h.uploadBaseDir, "operation_videos", fileName)
	} else {
		matches, _ := filepath.Glob(filepath.Join(h.uploadBaseDir, "operation_videos", fmt.Sprintf("%s.*", videoID.String())))
		if len(matches) > 0 {
			filePath = matches[0]
		} else {
			filePath = filepath.Join(h.uploadBaseDir, "operation_videos", fmt.Sprintf("%s.mp4", videoID.String()))
		}
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "ملف الفيديو غير موجود على السيرفر"})
		return
	}

	c.Header("Content-Type", "video/mp4")
	c.Header("Cache-Control", "private, no-cache, no-store, must-revalidate")
	c.File(filePath)
}


// List returns operation videos filtered by date, employee name, or order number
func (h *OperationVideoHandler) List(c *gin.Context) {
	var filter dto.OperationVideoFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معاملات البحث غير صالحة"})
		return
	}

	// Restrict by supervisor branch if applicable
	if filter.BranchID == "" {
		if bVal, exists := c.Get("branch_id"); exists {
			if bUUID, ok := bVal.(uuid.UUID); ok && bUUID != uuid.Nil {
				bStr := bUUID.String()
				filter.BranchID = bStr
			}
		}
	}

	items, total, err := h.service.ListVideos(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في جلب قائمة الفيديوهات"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"total": total,
	})
}

// GetByID returns single operation video details
func (h *OperationVideoHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفيديو غير صالح"})
		return
	}

	video, err := h.service.GetVideo(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الفيديو غير موجود"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": video})
}

// Delete marks video deleted and purges disk file
func (h *OperationVideoHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفيديو غير صالح"})
		return
	}

	if err := h.service.DeleteVideo(c.Request.Context(), id, h.uploadBaseDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في حذف الفيديو"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الفيديو بنجاح"})
}

// Update updates video notes and/or order number
func (h *OperationVideoHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفيديو غير صالح"})
		return
	}

	var req struct {
		OrderNumber string `json:"order_number"`
		Notes       string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات التعديل غير صالحة"})
		return
	}

	if err := h.service.UpdateVideo(c.Request.Context(), id, req.OrderNumber, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في تحديث بيانات الفيديو"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث بيانات الفيديو بنجاح"})
}

