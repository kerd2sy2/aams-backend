package service

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"delivery-backend/pkg/config"
)

const (
	maxFileSize   = 10 << 20 // 10 MB maximum file size
	maxBase64Size = 15 << 20 // 15 MB maximum base64 file size
)

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

var allowedCategories = map[string]bool{
	"personal":      true,
	"national_id":   true,
	"license":       true,
	"logo":          true,
	"investigation": true,
	"odometer":      true,
}

type StorageService interface {
	SaveImage(file *multipart.FileHeader, category string) (string, error)
	SaveBase64Image(base64Str string, category string) (string, error)
}

type storageService struct {
	cfg *config.Config
}

func NewStorageService(cfg *config.Config) StorageService {
	// Ensure local uploads folder exists
	_ = os.MkdirAll("uploads/personal", 0755)
	_ = os.MkdirAll("uploads/national_id", 0755)
	_ = os.MkdirAll("uploads/license", 0755)
	_ = os.MkdirAll("uploads/logo", 0755)
	_ = os.MkdirAll("uploads/investigation", 0755)
	_ = os.MkdirAll("uploads/odometer", 0755)
	return &storageService{cfg: cfg}
}

func (s *storageService) SaveImage(fileHeader *multipart.FileHeader, category string) (string, error) {
	// Validate file size
	if fileHeader.Size > maxFileSize {
		return "", fmt.Errorf("حجم الملف كبير جداً. الحد الأقصى 10 ميجابايت")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Validate content type via magic bytes
	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file for validation: %w", err)
	}
	mimeType := http.DetectContentType(buf)
	if !allowedMimeTypes[mimeType] {
		return "", fmt.Errorf("نوع الملف غير مسموح به. الأنواع المسموحة: JPEG, PNG, WebP")
	}

	// Reset file reader for saving
	file.Seek(0, io.SeekStart)

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	if !allowedExtensions[ext] {
		return "", fmt.Errorf("امتداد الملف غير مسموح به. الامتدادات المسموحة: .jpg, .jpeg, .png, .webp")
	}

	if !allowedCategories[category] {
		return "", fmt.Errorf("تصنيف ملف غير صالح")
	}

	filename := fmt.Sprintf("%s_%d%s", category, time.Now().UnixNano(), ext)
	dstPath := filepath.Join("uploads", category, filename)

	out, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Returns normalized local URL (e.g. /uploads/personal/...) or S3 URL
	return fmt.Sprintf("/uploads/%s/%s", category, filename), nil
}

func (s *storageService) SaveBase64Image(base64Data string, category string) (string, error) {
	if base64Data == "" {
		return "", nil
	}

	// Check if already a URL or path
	if strings.HasPrefix(base64Data, "http://") || strings.HasPrefix(base64Data, "https://") || strings.HasPrefix(base64Data, "/uploads/") {
		return base64Data, nil
	}

	// Validate size before processing
	if len(base64Data) > maxBase64Size {
		return "", fmt.Errorf("حجم الصورة كبير جداً. الحد الأقصى 15 ميجابايت")
	}

	// Remove data URI prefix if present (e.g., data:image/png;base64,)
	idx := strings.Index(base64Data, ",")
	rawBase64 := base64Data
	ext := ".png"

	if idx != -1 {
		prefix := base64Data[:idx]
		if strings.Contains(prefix, "jpeg") || strings.Contains(prefix, "jpg") {
			ext = ".jpg"
		} else if strings.Contains(prefix, "webp") {
			ext = ".webp"
		}
		rawBase64 = base64Data[idx+1:]
	}

	decoded, err := base64.StdEncoding.DecodeString(rawBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Validate decoded size
	if len(decoded) > maxFileSize {
		return "", fmt.Errorf("حجم الصورة كبير جداً بعد فك التشفير. الحد الأقصى 10 ميجابايت")
	}

	if !allowedCategories[category] {
		return "", fmt.Errorf("تصنيف ملف غير صالح")
	}

	filename := fmt.Sprintf("%s_%d%s", category, time.Now().UnixNano(), ext)
	dstPath := filepath.Join("uploads", category, filename)

	if err := os.WriteFile(dstPath, decoded, 0644); err != nil {
		return "", fmt.Errorf("failed to save base64 image: %w", err)
	}

	return fmt.Sprintf("/uploads/%s/%s", category, filename), nil
}
