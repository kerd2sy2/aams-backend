package employee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadVehicleRegistrationRequest struct {
	ImageURL         string `json:"image_url" binding:"required"`
	MotorcycleNumber string `json:"motorcycle_number"`
}

type istimaraOcrResult struct {
	IsIstimara     bool   `json:"is_istimara"`
	PlateDigits    string `json:"plate_digits"`
	PlateLettersEn string `json:"plate_letters_en"`
	PlateFull      string `json:"plate_full"`
}

var arabicIndicReplacer = strings.NewReplacer(
	"٠", "0", "١", "1", "٢", "2", "٣", "3", "٤", "4",
	"٥", "5", "٦", "6", "٧", "7", "٨", "8", "٩", "9",
	"۰", "0", "۱", "1", "۲", "2", "۳", "3", "۴", "4",
	"۵", "5", "۶", "6", "۷", "7", "۸", "8", "۹", "9",
)

func normalizePlateDigits(s string) string {
	ascii := arabicIndicReplacer.Replace(strings.TrimSpace(s))
	return extractASCIIOnlyDigits(ascii)
}

func (h *EmployeeHandler) UploadMyVehicleRegistration(c *gin.Context) {
	empIDVal, exists := c.Get("employee_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	empID, _ := empIDVal.(uuid.UUID)

	var req UploadVehicleRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.ImageURL) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "صورة الاستمارة مطلوبة"})
		return
	}

	emp, err := h.svc.FindByID(c.Request.Context(), empID)
	if err != nil || emp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	expectedPlate := strings.TrimSpace(req.MotorcycleNumber)
	if expectedPlate == "" {
		expectedPlate = strings.TrimSpace(emp.MotorcycleNumber)
	}
	expectedDigits := normalizePlateDigits(expectedPlate)
	if expectedDigits == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "لا يوجد دباب مربوط بحسابك حالياً"})
		return
	}

	imgB64, err := resolveDocumentImageBase64("", req.ImageURL)
	if err != nil || imgB64 == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "تعذر قراءة صورة الاستمارة"})
		return
	}

	ocr, errOcr := verifyIstimaraPlate(imgB64)
	if errOcr != nil || ocr == nil || !ocr.IsIstimara || normalizePlateDigits(ocr.PlateDigits) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "تعذر قراءة رقم اللوحة من الاستمارة — يرجى تصوير رخصة السير (الاستمارة) بشكل واضح ومستقيم",
		})
		return
	}

	cardDigits := normalizePlateDigits(ocr.PlateDigits)
	if cardDigits != expectedDigits {
		foundLabel := strings.TrimSpace(ocr.PlateFull)
		if foundLabel == "" {
			foundLabel = cardDigits
		}
		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("هذه الاستمارة تخص دباب آخر (%s) ولا تطابق دبابك المربوط (%s)", foundLabel, expectedPlate),
		})
		return
	}

	cleanURL := strings.TrimSpace(req.ImageURL)
	if err := h.svc.SaveVehicleRegistrationImage(c.Request.Context(), empID, expectedPlate, expectedDigits, cleanURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "تعذر حفظ الاستمارة وربطها بالدباب"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "تم التحقق من الاستمارة وحفظها وربطها بالدباب بنجاح",
		"registration_image": cleanURL,
		"plate_number":       ocr.PlateFull,
	})
}

func verifyIstimaraPlate(imgB64 string) (*istimaraOcrResult, error) {
	// Primary: Gemini 3.1 Flash Lite (reads even worn/faded physical Saudi Istimara cards accurately)
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey != "" {
		prompt := `Analyze this image. Is it a Saudi Vehicle Registration card (رخصة سير / VEHICLES REGISTRATION / استمارة)?
If YES, extract the plate number (رقم اللوحة) which appears in Arabic and English (e.g. "5452 A J", "7037 A J", "6543 A D").
Return ONLY valid JSON:
{
  "is_istimara": true or false,
  "plate_digits": "ASCII digits only e.g. 5452",
  "plate_letters_en": "uppercase English letters e.g. AJ",
  "plate_full": "e.g. 5452 AJ"
}`
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
		client := &http.Client{Timeout: 12 * time.Second}
		resp, err := client.Post(url, "application/json", bytes.NewReader(b))
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var raw struct {
					Candidates []struct {
						Content struct {
							Parts []struct {
								Text string `json:"text"`
							} `json:"parts"`
						} `json:"content"`
					} `json:"candidates"`
				}
				if errDec := json.NewDecoder(resp.Body).Decode(&raw); errDec == nil &&
					len(raw.Candidates) > 0 && len(raw.Candidates[0].Content.Parts) > 0 {
					txt := strings.TrimSpace(raw.Candidates[0].Content.Parts[0].Text)
					txt = strings.TrimPrefix(txt, "```json")
					txt = strings.TrimPrefix(txt, "```")
					txt = strings.TrimSuffix(txt, "```")
					var parsed istimaraOcrResult
					if json.Unmarshal([]byte(strings.TrimSpace(txt)), &parsed) == nil && parsed.PlateDigits != "" {
						return &parsed, nil
					}
				}
			}
		}
	}

	// Fallback: local plate-ocr service
	payload, _ := json.Marshal(map[string]string{"image_base64": imgB64})
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Post("http://127.0.0.1:8091/ocr", "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var local struct {
		Success   bool   `json:"success"`
		Digits    string `json:"digits"`
		EnLetters string `json:"en_letters"`
		RawText   string `json:"raw_text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&local); err != nil {
		return nil, err
	}
	digits := normalizePlateDigits(local.Digits)
	if digits == "" {
		re := regexp.MustCompile(`\b(\d{3,4})\s*[A-Z]\s*[A-Z]\b`)
		if m := re.FindStringSubmatch(strings.ToUpper(local.RawText)); len(m) > 1 {
			digits = m[1]
		}
	}
	if digits == "" {
		return &istimaraOcrResult{IsIstimara: false}, nil
	}
	return &istimaraOcrResult{
		IsIstimara:     true,
		PlateDigits:    digits,
		PlateLettersEn: local.EnLetters,
		PlateFull:      strings.TrimSpace(digits + " " + local.EnLetters),
	}, nil
}
