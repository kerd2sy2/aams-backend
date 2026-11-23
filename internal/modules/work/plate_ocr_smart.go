package work

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlateFeedbackRequest struct {
	Image          string  `json:"image"`
	ConfirmedPlate string  `json:"confirmed_plate" binding:"required"`
	RawOCRText     string  `json:"raw_ocr_text"`
	EngineUsed     string  `json:"engine_used"`
	Confidence     float64 `json:"confidence"`
	WasCorrected   bool    `json:"was_corrected"`
	EmployeeID     string  `json:"employee_id"`
}

type localOCRResponse struct {
	Success       bool    `json:"success"`
	RawText       string  `json:"raw_text"`
	Digits        string  `json:"digits"`
	EnLetters     string  `json:"en_letters"`
	ArLetters     string  `json:"ar_letters"`
	Confidence    float64 `json:"confidence"`
	Engine        string  `json:"engine"`
	LightingScore float64 `json:"lighting_score"`
	ElapsedMs     int     `json:"elapsed_ms"`
}

var saudiEnToArMap = map[rune]string{
	'A': "أ", 'B': "ب", 'J': "ح", 'D': "د", 'R': "ر", 'S': "س",
	'X': "ص", 'T': "ط", 'E': "ع", 'G': "ق", 'K': "ك", 'L': "ل",
	'Z': "م", 'N': "ن", 'H': "هـ", 'U': "و", 'V': "ى",
}

var saudiArToEnMap = map[string]string{
	"ا": "A", "أ": "A", "إ": "A", "آ": "A", "ب": "B", "ح": "J",
	"د": "D", "ر": "R", "س": "S", "ص": "X", "ط": "T", "ع": "E",
	"ق": "G", "ك": "K", "ل": "L", "م": "Z", "ن": "N", "ه": "H",
	"هـ": "H", "و": "U", "ى": "V", "ي": "V",
}

var digitRegex = regexp.MustCompile(`\d{3,4}`)
var nonAlnumRegex = regexp.MustCompile(`[^A-Z0-9]+`)

func toArabicPlateDigits(digits string) string {
	ar := []string{"٠", "١", "٢", "٣", "٤", "٥", "٦", "٧", "٨", "٩"}
	var b strings.Builder
	for _, r := range digits {
		if r >= '0' && r <= '9' {
			b.WriteString(ar[r-'0'])
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func formatEnToArLetters(en string) string {
	en = strings.ToUpper(strings.TrimSpace(en))
	var parts []string
	for _, r := range en {
		if v, ok := saudiEnToArMap[r]; ok {
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, " ")
}

func normalizePlateParts(raw string) (digits, enLetters, arLetters string) {
	cleaned := strings.TrimSpace(raw)
	// Convert Arabic-Indic digits
	arReplacer := strings.NewReplacer(
		"٠", "0", "١", "1", "٢", "2", "٣", "3", "٤", "4",
		"٥", "5", "٦", "6", "٧", "7", "٨", "8", "٩", "9",
		"۰", "0", "۱", "1", "۲", "2", "۳", "3", "۴", "4",
		"۵", "5", "۶", "6", "۷", "7", "۸", "8", "۹", "9",
	)
	cleaned = arReplacer.Replace(cleaned)
	if m := digitRegex.FindString(cleaned); m != "" {
		digits = m
	}

	// Check English letters first
	var enBuf strings.Builder
	for _, r := range strings.ToUpper(cleaned) {
		if _, ok := saudiEnToArMap[r]; ok {
			enBuf.WriteRune(r)
		}
	}
	if enBuf.Len() >= 2 {
		enLetters = enBuf.String()
		if len(enLetters) > 3 {
			enLetters = enLetters[:2]
		}
	} else {
		// Check Arabic letters
		var arEn strings.Builder
		cleaned = strings.ReplaceAll(cleaned, "هـ", "ه")
		for _, r := range cleaned {
			if mapped, ok := saudiArToEnMap[string(r)]; ok {
				arEn.WriteString(mapped)
			}
		}
		if arEn.Len() >= 2 {
			enLetters = arEn.String()
			if len(enLetters) > 3 {
				enLetters = enLetters[:2]
			}
		}
	}

	if enLetters != "" {
		arLetters = formatEnToArLetters(enLetters)
	}
	return digits, enLetters, arLetters
}

func makeCorrectionRawKey(raw string) string {
	upper := strings.ToUpper(strings.TrimSpace(raw))
	key := nonAlnumRegex.ReplaceAllString(upper, "")
	if len(key) > 64 {
		key = key[:64]
	}
	return key
}

func (s *workService) getWorkDB() *gorm.DB {
	if r, ok := s.repo.(*gormWorkRepository); ok && r != nil {
		return r.db
	}
	return nil
}

func (s *workService) callLocalPlateOCR(ctx context.Context, imageBase64 string) (*localOCRResponse, error) {
	url := strings.TrimSpace(os.Getenv("PLATE_OCR_LOCAL_URL"))
	if url == "" {
		url = "http://127.0.0.1:8091/ocr"
	}
	payload, _ := json.Marshal(map[string]string{"image": imageBase64})
	reqCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("local ocr status %d", resp.StatusCode)
	}
	var out localOCRResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *workService) callGeminiPlateVision(ctx context.Context, imageBase64 string) (digits, enLetters, arLetters string, err error) {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		return "", "", "", fmt.Errorf("gemini key not configured")
	}

	cleanB64 := imageBase64
	if idx := strings.Index(cleanB64, ","); idx != -1 {
		cleanB64 = cleanB64[idx+1:]
	}
	cleanB64 = strings.TrimSpace(cleanB64)

	prompt := `Read the Saudi motorcycle or car license plate in this image.
Return ONLY valid JSON with no markdown: {"digits":"7569","en_letters":"BE"}
Rules:
- digits: 3 to 4 English numbers (0-9).
- en_letters: 2 English uppercase letters for motorcycles (or 3 for cars) from Saudi plate alphabet (A,B,J,D,R,S,X,T,E,G,K,L,Z,N,H,U,V). Ignore "KSA".
- If no plate is visible, return {"digits":"","en_letters":""}.`

	bodyMap := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
					{"inline_data": map[string]string{
						"mime_type": "image/jpeg",
						"data":      cleanB64,
					}},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0,
			"maxOutputTokens": 60,
		},
	}

	payload, _ := json.Marshal(bodyMap)
	reqCtx, cancel := context.WithTimeout(ctx, 7*time.Second)
	defer cancel()

	endpoint := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("gemini status %d", resp.StatusCode)
	}

	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBytes, &parsed); err != nil || len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", "", "", fmt.Errorf("gemini empty response")
	}

	raw := strings.TrimSpace(parsed.Candidates[0].Content.Parts[0].Text)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var plateJSON struct {
		Digits    string `json:"digits"`
		EnLetters string `json:"en_letters"`
	}
	if err := json.Unmarshal([]byte(raw), &plateJSON); err == nil {
		d, e, a := normalizePlateParts(plateJSON.Digits + " " + plateJSON.EnLetters)
		if d != "" && e != "" {
			return d, e, a, nil
		}
	}

	d, e, a := normalizePlateParts(raw)
	if d != "" && e != "" {
		return d, e, a, nil
	}
	return "", "", "", fmt.Errorf("plate not found by gemini")
}

func (s *workService) lookupLearnedCorrection(ctx context.Context, rawText string) (digits, enLetters, arLetters string, found bool) {
	db := s.getWorkDB()
	if db == nil {
		return "", "", "", false
	}
	key := makeCorrectionRawKey(rawText)
	if len(key) < 4 {
		return "", "", "", false
	}
	var row struct {
		ConfirmedDigits    string `gorm:"column:confirmed_digits"`
		ConfirmedEnLetters string `gorm:"column:confirmed_en_letters"`
		ConfirmedArLetters string `gorm:"column:confirmed_ar_letters"`
	}
	err := db.WithContext(ctx).Table("plate_ocr_corrections").
		Select("confirmed_digits, confirmed_en_letters, confirmed_ar_letters").
		Where("raw_key = ?", key).
		First(&row).Error
	if err == nil && row.ConfirmedDigits != "" && row.ConfirmedEnLetters != "" {
		return row.ConfirmedDigits, row.ConfirmedEnLetters, row.ConfirmedArLetters, true
	}
	return "", "", "", false
}

func (s *workService) matchAgainstLiveFleet(ctx context.Context, digits, enLetters string) (string, string, string, bool) {
	db := s.getWorkDB()
	if db == nil || len(digits) < 3 {
		return digits, enLetters, formatEnToArLetters(enLetters), false
	}
	var plates []string
	if err := db.WithContext(ctx).Table("vehicles").Where("deleted_at IS NULL AND plate_number IS NOT NULL").Pluck("plate_number", &plates).Error; err != nil {
		return digits, enLetters, formatEnToArLetters(enLetters), false
	}

	var exactDigitMatches [][3]string
	for _, p := range plates {
		fd, fe, fa := normalizePlateParts(p)
		if fd == "" || fe == "" {
			continue
		}
		if fd == digits && fe == enLetters {
			return fd, fe, fa, true
		}
		if fd == digits {
			exactDigitMatches = append(exactDigitMatches, [3]string{fd, fe, fa})
		}
	}

	// If only ONE bike in the entire fleet has these exact 4 digits, snap to it automatically!
	if len(exactDigitMatches) == 1 && len(digits) == 4 {
		m := exactDigitMatches[0]
		return m[0], m[1], m[2], true
	}
	return digits, enLetters, formatEnToArLetters(enLetters), false
}

func (s *workService) saveLearnedSampleAsync(rawText, digits, enLetters, arLetters, engine string, conf float64, wasCorrected bool, employeeID string, imageBase64 string) {
	db := s.getWorkDB()
	if db == nil || digits == "" || enLetters == "" {
		return
	}
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		fullPlate := fmt.Sprintf("%s %s", digits, arLetters)

		// 1. Save learned correction mapping so local OCR improves automatically
		if key := makeCorrectionRawKey(rawText); len(key) >= 4 {
			_ = db.WithContext(bgCtx).Exec(`
				INSERT INTO plate_ocr_corrections (raw_key, confirmed_digits, confirmed_en_letters, confirmed_ar_letters, full_plate, hit_count, updated_at)
				VALUES (?, ?, ?, ?, ?, 1, NOW())
				ON CONFLICT (raw_key) DO UPDATE SET
					confirmed_digits = EXCLUDED.confirmed_digits,
					confirmed_en_letters = EXCLUDED.confirmed_en_letters,
					confirmed_ar_letters = EXCLUDED.confirmed_ar_letters,
					full_plate = EXCLUDED.full_plate,
					hit_count = plate_ocr_corrections.hit_count + 1,
					updated_at = NOW()
			`, key, digits, enLetters, arLetters, fullPlate).Error
		}

		// 2. Save labeled image to dataset folder for future YOLO/OCR fine-tuning
		savedPath := ""
		if strings.TrimSpace(imageBase64) != "" {
			cleanB64 := imageBase64
			if idx := strings.Index(cleanB64, ","); idx != -1 {
				cleanB64 = cleanB64[idx+1:]
			}
			if imgBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(cleanB64)); err == nil && len(imgBytes) > 500 {
				folderName := fmt.Sprintf("%s_%s", digits, enLetters)
				dir := filepath.Join("/opt/aams-backend/uploads/plate_dataset", folderName)
				if mkErr := os.MkdirAll(dir, 0755); mkErr == nil {
					fileName := fmt.Sprintf("%d_%s.jpg", time.Now().UnixMilli(), uuid.NewString()[:8])
					fullPath := filepath.Join(dir, fileName)
					if wErr := os.WriteFile(fullPath, imgBytes, 0644); wErr == nil {
						savedPath = fullPath
					}
				}
			}
		}

		var empUUID *uuid.UUID
		if parsedID, err := uuid.Parse(strings.TrimSpace(employeeID)); err == nil {
			empUUID = &parsedID
		}
		_ = db.WithContext(bgCtx).Exec(`
			INSERT INTO plate_training_samples (employee_id, confirmed_plate, raw_ocr_text, engine_used, confidence, was_corrected, image_path, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
		`, empUUID, fullPlate, rawText, engine, conf, wasCorrected, savedPath).Error
	}()
}

// ScanPlateImageSmart runs: Tier 1 Local OpenCV+RapidOCR -> Tier 2 Self-Learning Memory + Live Fleet Match -> Tier 3 Optional Gemini Vision Fallback (never fails if API has no credits!)
func (s *workService) ScanPlateImageSmart(ctx context.Context, imageBase64 string) (map[string]interface{}, error) {
	var rawText, digits, enLetters, arLetters, engine string
	var conf float64

	// Tier 1: Local Free OpenCV + RapidOCR microservice on VPS (zero API cost)
	if localRes, err := s.callLocalPlateOCR(ctx, imageBase64); err == nil && localRes != nil {
		rawText = localRes.RawText
		digits = localRes.Digits
		enLetters = localRes.EnLetters
		arLetters = localRes.ArLetters
		conf = localRes.Confidence
		engine = localRes.Engine
	}

	// Tier 2A: Check Self-Learning Correction Memory
	if rawText != "" {
		if ld, le, la, ok := s.lookupLearnedCorrection(ctx, rawText); ok {
			digits, enLetters, arLetters = ld, le, la
			conf = 0.96
			engine = "local_learned"
		}
	}

	// Tier 2B: Match against Live Fleet DB (vehicles table)
	if digits != "" {
		fd, fe, fa, matched := s.matchAgainstLiveFleet(ctx, digits, enLetters)
		if matched {
			digits, enLetters, arLetters = fd, fe, fa
			if conf < 0.90 {
				conf = 0.90
			}
		}
	}

	// Tier 3: Optional Gemini Vision fallback ONLY if local engine couldn't get a high-confidence plate.
	// Gracefully ignored if Gemini has no quota/credits or fails!
	if (digits == "" || enLetters == "" || conf < 0.78) && strings.TrimSpace(os.Getenv("GEMINI_API_KEY")) != "" {
		if gd, ge, _, gErr := s.callGeminiPlateVision(ctx, imageBase64); gErr == nil && gd != "" && ge != "" {
			// Snap Gemini output against live fleet too
			fd, fe, fa, _ := s.matchAgainstLiveFleet(ctx, gd, ge)
			digits, enLetters, arLetters = fd, fe, fa
			conf = 0.95
			engine = "gemini_fallback"
			// Auto-learn so next time the free local engine solves it without Gemini!
			s.saveLearnedSampleAsync(rawText, digits, enLetters, arLetters, engine, conf, rawText != "", "", imageBase64)
		}
	}

	if digits == "" || enLetters == "" {
		return map[string]interface{}{
			"success":    false,
			"digits":     digits,
			"letters":    arLetters,
			"raw_text":   rawText,
			"engine":     engine,
			"confidence": conf,
			"message":    "لم يتم التعرف على اللوحة بوضوح",
		}, nil
	}

	if arLetters == "" {
		arLetters = formatEnToArLetters(enLetters)
	}
	fullPlate := strings.TrimSpace(digits + " " + arLetters)

	return map[string]interface{}{
		"success":         true,
		"digits":          digits,
		"letters":         arLetters,
		"arabic_digits":   toArabicPlateDigits(digits),
		"arabic_letters":  arLetters,
		"english_letters": enLetters,
		"full_plate":      fullPlate,
		"raw_text":        rawText,
		"engine":          engine,
		"confidence":      conf,
	}, nil
}

func (s *workService) RecordPlateFeedback(ctx context.Context, req PlateFeedbackRequest) error {
	digits, enLetters, arLetters := normalizePlateParts(req.ConfirmedPlate)
	if digits == "" || enLetters == "" {
		return nil
	}
	s.saveLearnedSampleAsync(
		req.RawOCRText,
		digits,
		enLetters,
		arLetters,
		req.EngineUsed,
		req.Confidence,
		req.WasCorrected,
		req.EmployeeID,
		req.Image,
	)
	return nil
}
