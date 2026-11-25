package employee

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// verifiedDocCache stores server-verified OCR results keyed by image_url (expires after 15 min)
// so POST /employees/me/profile immediately uses the trusted result from /employees/scan-document
// without a redundant second Gemini Vision call.
type cachedDocScan struct {
	Result    ScanDocumentResult
	ExpiresAt time.Time
}

var verifiedDocCache sync.Map

func cacheVerifiedDocScan(imageURL string, res ScanDocumentResult) {
	u := strings.TrimSpace(imageURL)
	if u == "" {
		return
	}
	verifiedDocCache.Store(u, cachedDocScan{
		Result:    res,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	})
}

func getCachedDocScan(imageURL string) (ScanDocumentResult, bool) {
	u := strings.TrimSpace(imageURL)
	if u == "" {
		return ScanDocumentResult{}, false
	}
	if val, ok := verifiedDocCache.Load(u); ok {
		if entry, okType := val.(cachedDocScan); okType && time.Now().Before(entry.ExpiresAt) {
			return entry.Result, true
		}
		verifiedDocCache.Delete(u)
	}
	return ScanDocumentResult{}, false
}

type ScanDocumentRequest struct {
	ImageURL    string `json:"image_url"`
	ImageBase64 string `json:"image_base64"`
	DocType     string `json:"doc_type"` // "iqama" | "license" | "auto"
}

type ScanDocumentResult struct {
	Success               bool   `json:"success"`
	DocType               string `json:"doc_type"`
	NameAr                string `json:"name_ar"`
	NameEn                string `json:"name_en"`
	NationalID            string `json:"national_id"`
	DateOfBirth           string `json:"date_of_birth"`
	IqamaExpirationDate   string `json:"iqama_expiration_date"`
	LicenseExpirationDate string `json:"license_expiration_date"`
	Nationality           string `json:"nationality"`
	Engine                string `json:"engine"`
	ElapsedMs             int64  `json:"elapsed_ms"`
}

func (h *EmployeeHandler) ScanDocument(c *gin.Context) {
	t0 := time.Now()
	var req ScanDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	docType := strings.ToLower(strings.TrimSpace(req.DocType))
	if docType == "" {
		docType = "auto"
	}

	imgB64, err := resolveDocumentImageBase64(req.ImageBase64, req.ImageURL)
	if err != nil || imgB64 == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "تعذر قراءة صورة المستند"})
		return
	}

	// Tier 1: Local Zero-Cost Python OCR (/ocr-document on 127.0.0.1:8091)
	res := callLocalDocumentOCR(imgB64, docType)
	res.Engine = "local_ocr"

	// Tier 2: Optional Gemini Vision fallback whenever critical fields (e.g. Arabic-Indic Iqama dates) are missing
	needsFallback := res.DateOfBirth == "" || res.NationalID == "" || res.Nationality == "" ||
		(docType == "iqama" && res.IqamaExpirationDate == "") ||
		(docType == "license" && res.LicenseExpirationDate == "")

	if needsFallback {
		if gemRes, errGem := callGeminiDocumentOCR(imgB64, docType); errGem == nil && gemRes != nil {
			if res.DocType == "" {
				res.DocType = gemRes.DocType
			}
			if res.NameAr == "" {
				res.NameAr = gemRes.NameAr
			}
			if res.NameEn == "" {
				res.NameEn = gemRes.NameEn
			}
			if res.NationalID == "" {
				res.NationalID = gemRes.NationalID
			}
			if res.DateOfBirth == "" {
				res.DateOfBirth = gemRes.DateOfBirth
			}
			if res.IqamaExpirationDate == "" && (docType == "iqama" || (docType == "auto" && gemRes.DocType == "iqama")) {
				res.IqamaExpirationDate = gemRes.IqamaExpirationDate
			}
			if res.LicenseExpirationDate == "" && (docType == "license" || (docType == "auto" && gemRes.DocType == "license")) {
				res.LicenseExpirationDate = gemRes.LicenseExpirationDate
			}
			if res.Nationality == "" {
				res.Nationality = gemRes.Nationality
			}
			res.Engine = "hybrid_local_gemini"
		}
	}

	res.Nationality = normalizeDocNationality(res.Nationality)
	res.Success = res.NationalID != "" || res.DateOfBirth != "" || res.IqamaExpirationDate != "" ||
		res.LicenseExpirationDate != "" || res.Nationality != "" || res.NameEn != "" || res.NameAr != ""
	res.ElapsedMs = time.Since(t0).Milliseconds()

	// Strict ownership check for self-service drivers: reject another person's ID/License card
	if empIDVal, exists := c.Get("employee_id"); exists && res.NationalID != "" {
		roleVal, _ := c.Get("role")
		roleStr := strings.ToUpper(fmt.Sprint(roleVal))
		if roleStr == "" || roleStr == "DRIVER" || roleStr == "EMPLOYEE" {
			if empID, ok := empIDVal.(uuid.UUID); ok {
				if emp, errEmp := h.svc.FindByID(c.Request.Context(), empID); errEmp == nil && emp != nil {
					ownID := extractASCIIOnlyDigits(emp.NationalID)
					cardID := extractASCIIOnlyDigits(res.NationalID)
					if len(ownID) >= 8 && len(cardID) >= 8 && ownID != cardID {
						c.JSON(http.StatusForbidden, gin.H{
							"error": "هذه الوثيقة لا تخصك — رقم الهوية في الصورة لا يطابق رقم هويتك المسجل",
						})
						return
					}
				}
			}
		}
	}

	if req.ImageURL != "" && res.Success {
		cacheVerifiedDocScan(req.ImageURL, res)
	}

	c.JSON(http.StatusOK, res)
}

func resolveDocumentImageBase64(rawB64, imageURL string) (string, error) {
	if strings.TrimSpace(rawB64) != "" {
		parts := strings.Split(rawB64, ",")
		return strings.TrimSpace(parts[len(parts)-1]), nil
	}
	u := strings.TrimSpace(imageURL)
	if u == "" {
		return "", fmt.Errorf("missing image")
	}

	// Fast local disk read for /uploads/... files on this server
	if idx := strings.Index(u, "/uploads/"); idx >= 0 {
		rel := strings.TrimPrefix(u[idx:], "/")
		rel = filepath.Clean(rel)
		if strings.HasPrefix(rel, "uploads") && !strings.Contains(rel, "..") {
			for _, base := range []string{".", "/opt/aams-backend"} {
				full := filepath.Join(base, rel)
				if data, err := os.ReadFile(full); err == nil && len(data) > 0 {
					return base64.StdEncoding.EncodeToString(data), nil
				}
			}
		}
	}

	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		client := &http.Client{Timeout: 8 * time.Second}
		resp, err := client.Get(u)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(data), nil
	}
	return "", fmt.Errorf("unsupported image source")
}

func callLocalDocumentOCR(imgB64, docType string) ScanDocumentResult {
	out := ScanDocumentResult{DocType: docType}
	payload, _ := json.Marshal(map[string]string{
		"image_base64": imgB64,
		"doc_type":     docType,
	})
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Post("http://127.0.0.1:8091/ocr-document", "application/json", bytes.NewReader(payload))
	if err != nil {
		return out
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return out
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return out
}

func callGeminiDocumentOCR(imgB64, docType string) (*ScanDocumentResult, error) {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		return nil, fmt.Errorf("no gemini key")
	}

	prompt := fmt.Sprintf(`Extract data from this Saudi %s card image (Iqama / Resident ID or Driving License).
Return ONLY valid JSON with these exact keys (all dates in Gregorian YYYY-MM-DD with ASCII digits 0-9; convert Arabic-Indic digits ٠١٢٣٤٥٦٧٨٩ to 0-9):
{
  "doc_type": "iqama" or "license",
  "name_ar": "Arabic full name",
  "name_en": "English full name in UPPERCASE",
  "national_id": "10-digit ID starting with 1 or 2 (NOT employer ID starting with 7)",
  "date_of_birth": "YYYY-MM-DD",
  "iqama_expiration_date": "YYYY-MM-DD if Iqama, else empty string",
  "license_expiration_date": "YYYY-MM-DD if Driving License, else empty string",
  "nationality": "Arabic country/nationality name (e.g. بنجلاديش, باكستان, مصري, سعودي)"
}`, docType)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
					{"inline_data": map[string]string{"mime_type": "image/jpeg", "data": imgB64}},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":      0,
			"responseMimeType": "application/json",
		},
	}

	b, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-3.1-flash-lite:generateContent?key=%s", apiKey)
	client := &http.Client{Timeout: 15 * time.Second}

	var resp *http.Response
	var err error
	for attempt := 0; attempt < 2; attempt++ {
		resp, err = client.Post(url, "application/json", bytes.NewReader(b))
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			bodySnippet, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
			log.Printf("[DocumentOCR] Gemini attempt %d failed status=%d body=%s", attempt+1, resp.StatusCode, string(bodySnippet))
			resp.Body.Close()
		} else {
			log.Printf("[DocumentOCR] Gemini attempt %d error: %v", attempt+1, err)
		}
		if attempt == 0 {
			time.Sleep(600 * time.Millisecond)
		}
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini status %d", resp.StatusCode)
	}

	var raw struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw.Candidates) == 0 || len(raw.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty gemini response")
	}

	txt := strings.TrimSpace(raw.Candidates[0].Content.Parts[0].Text)
	txt = strings.TrimPrefix(txt, "```json")
	txt = strings.TrimPrefix(txt, "```")
	txt = strings.TrimSuffix(txt, "```")
	txt = strings.TrimSpace(txt)

	var parsed ScanDocumentResult
	if err := json.Unmarshal([]byte(txt), &parsed); err != nil {
		return nil, err
	}
	return &parsed, nil
}

func normalizeDocNationality(raw string) string {
	low := strings.ToLower(strings.TrimSpace(raw))
	if low == "" {
		return ""
	}
	switch {
	case strings.Contains(low, "بنجلاديش") || strings.Contains(low, "بنغلاديش") || strings.Contains(low, "bangladesh"):
		return "بنجلاديش"
	case strings.Contains(low, "باكستان") || strings.Contains(low, "پاکستان") || strings.Contains(low, "pakistan"):
		return "باكستان"
	case strings.Contains(low, "مصر") || strings.Contains(low, "egypt"):
		return "مصري"
	case strings.Contains(low, "سعود") || strings.Contains(low, "saudi"):
		return "سعودي"
	default:
		return strings.TrimSpace(raw)
	}
}

// tryAutoCropDocumentFile calls the local OpenCV /crop-document service to isolate and
// upright-rotate ID/License/Passport/Registration cards from phone screenshots or photos.
func tryAutoCropDocumentFile(savePath string) (string, bool) {
	raw, err := os.ReadFile(savePath)
	if err != nil || len(raw) == 0 {
		return savePath, false
	}
	payload, _ := json.Marshal(map[string]string{
		"image_base64": base64.StdEncoding.EncodeToString(raw),
	})
	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Post("http://127.0.0.1:8091/crop-document", "application/json", bytes.NewReader(payload))
	if err != nil {
		return savePath, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return savePath, false
	}
	var out struct {
		Success     bool   `json:"success"`
		Cropped     bool   `json:"cropped"`
		ImageBase64 string `json:"image_base64"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || !out.Success || !out.Cropped || out.ImageBase64 == "" {
		return savePath, false
	}
	decoded, err := base64.StdEncoding.DecodeString(out.ImageBase64)
	if err != nil || len(decoded) < 2000 {
		return savePath, false
	}
	ext := strings.ToLower(filepath.Ext(savePath))
	targetPath := savePath
	if ext != ".jpg" && ext != ".jpeg" {
		targetPath = strings.TrimSuffix(savePath, filepath.Ext(savePath)) + ".jpg"
	}
	if err := os.WriteFile(targetPath, decoded, 0644); err != nil {
		return savePath, false
	}
	if targetPath != savePath {
		_ = os.Remove(savePath)
	}
	return targetPath, true
}
