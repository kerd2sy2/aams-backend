package work

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	legacyService "delivery-backend/internal/service"
)

func (a *gormEmployeeAdapter) SendShiftApprovalNotification(ctx context.Context, empID uuid.UUID, sessionID uuid.UUID, ordersCount int, fuelCost float64, supervisorName string) error {
	var emp struct {
		Name                  string     `gorm:"column:name"`
		Language              string     `gorm:"column:language"`
		PushToken             string     `gorm:"column:push_token"`
		BranchID              *uuid.UUID `gorm:"column:branch_id"`
		MotorcycleNumber      string     `gorm:"column:motorcycle_number"`
		VehicleType           string     `gorm:"column:vehicle_type"`
		TotalDistance         float64    `gorm:"column:total_distance"`
		LastOilChangeDistance float64    `gorm:"column:last_oil_change_distance"`
	}
	if err := a.db.WithContext(ctx).Table("employees").Where("id = ? AND deleted_at IS NULL", empID).First(&emp).Error; err != nil {
		return err
	}

	cleanSupervisor := strings.TrimSpace(supervisorName)
	supervisorLabelAr := translateSupervisorName(cleanSupervisor, "ar")
	supervisorLabelEn := translateSupervisorName(cleanSupervisor, "en")
	supervisorLabelBn := translateSupervisorName(cleanSupervisor, "bn")
	supervisorLabelUr := translateSupervisorName(cleanSupervisor, "ur")

	fuelFormatted := fmt.Sprintf("%.2f ريال", fuelCost)
	if fuelCost == float64(int(fuelCost)) {
		fuelFormatted = fmt.Sprintf("%d ريال", int(fuelCost))
	}

	var session struct {
		MotorcycleNumber string  `gorm:"column:motorcycle_number"`
		VehicleType      string  `gorm:"column:vehicle_type"`
		EndKM            float64 `gorm:"column:end_km"`
		Distance         float64 `gorm:"column:distance"`
	}
	_ = a.db.WithContext(ctx).Table("work_sessions").Where("id = ?", sessionID).First(&session).Error

	bikeNumber := strings.TrimSpace(session.MotorcycleNumber)
	if bikeNumber == "" {
		bikeNumber = strings.TrimSpace(emp.MotorcycleNumber)
	}

	vType := strings.ToLower(strings.TrimSpace(session.VehicleType))
	if vType == "" {
		vType = strings.ToLower(strings.TrimSpace(emp.VehicleType))
	}

	interval := 950.0
	if vType == "car" {
		interval = 10000.0
	}

	var remainingKM float64 = interval
	hasOilCalculation := false

	if bikeNumber != "" {
		var vehicle struct {
			VehicleType     string  `gorm:"column:vehicle_type"`
			CurrentKM       float64 `gorm:"column:current_km"`
			LastOilChangeKM float64 `gorm:"column:last_oil_change_km"`
			TotalDistance   float64 `gorm:"column:total_distance"`
		}
		err := a.db.WithContext(ctx).Table("vehicles").
			Where("deleted_at IS NULL AND (TRIM(plate_number) = ? OR plate_number ILIKE ? OR plate_number ILIKE ?)", bikeNumber, bikeNumber+"%", "%"+bikeNumber+"%").
			Order("CASE WHEN TRIM(plate_number) = '" + bikeNumber + "' THEN 0 ELSE 1 END").
			First(&vehicle).Error
		if err == nil {
			if strings.EqualFold(vehicle.VehicleType, "car") {
				interval = 10000.0
			}
			currentKM := vehicle.CurrentKM
			if currentKM <= 0 && session.EndKM > 0 {
				currentKM = session.EndKM
			}
			if vehicle.LastOilChangeKM > 0 && currentKM >= vehicle.LastOilChangeKM {
				distanceSinceOil := currentKM - vehicle.LastOilChangeKM
				remainingKM = interval - distanceSinceOil
				hasOilCalculation = true
			} else if vehicle.TotalDistance > 0 {
				distanceSinceOil := vehicle.TotalDistance
				remainingKM = interval - distanceSinceOil
				hasOilCalculation = true
			}
		}
	}

	if !hasOilCalculation {
		if emp.LastOilChangeDistance > 0 && emp.TotalDistance >= emp.LastOilChangeDistance {
			distanceSinceOil := emp.TotalDistance - emp.LastOilChangeDistance
			remainingKM = interval - distanceSinceOil
		} else if emp.TotalDistance > 0 {
			distanceSinceOil := emp.TotalDistance
			remainingKM = interval - distanceSinceOil
		}
	}

	remainingInt := int(math.Round(remainingKM))
	if remainingInt < 0 {
		remainingInt = 0
	}

	lang := strings.ToLower(strings.TrimSpace(emp.Language))
	var title, body, oilInfo string

	if remainingInt <= 0 {
		switch {
		case lang == "en":
			oilInfo = "⚠️ Warning: Oil change is due now! Change oil immediately."
		case lang == "bn" || lang == "bengali" || lang == "bd":
			oilInfo = "⚠️ সতর্কতা: এখনই ইঞ্জিন অয়েল পরিবর্তনের সময় হয়েছে! দ্রুত অয়েল পরিবর্তন করুন।"
		case lang == "ur" || lang == "urdu" || lang == "pk":
			oilInfo = "⚠️ توجہ: آئل تبدیل کرنے کا وقت ہو گیا ہے! فوری انجن آئل تبدیل کروائیں۔"
		default:
			oilInfo = "⚠️ تنبيه: حان موعد غيار الزيت الآن! يرجى التوجه لتغيير الزيت فوراً."
		}
	} else {
		switch {
		case lang == "en":
			oilInfo = fmt.Sprintf("Remaining for oil change: %d km", remainingInt)
		case lang == "bn" || lang == "bengali" || lang == "bd":
			oilInfo = fmt.Sprintf("তেল পরিবর্তনের বাকি: %d কিমি", remainingInt)
		case lang == "ur" || lang == "urdu" || lang == "pk":
			oilInfo = fmt.Sprintf("آئل تبدیلی میں باقی: %d کلومیٹر", remainingInt)
		default:
			oilInfo = fmt.Sprintf("باقي على غيار الزيت: %d كم", remainingInt)
		}
	}

	switch {
	case lang == "en":
		title = "Approved"
		fuelEn := fmt.Sprintf("%.2f SAR", fuelCost)
		if fuelCost == float64(int(fuelCost)) {
			fuelEn = fmt.Sprintf("%d SAR", int(fuelCost))
		}
		body = fmt.Sprintf("Supervisor %s approved your orders\n%d Orders | Fuel: %s\n%s", supervisorLabelEn, ordersCount, fuelEn, oilInfo)
	case lang == "bn" || lang == "bengali" || lang == "bd":
		title = "অনুমোদিত"
		fuelBn := fmt.Sprintf("%.2f SAR", fuelCost)
		if fuelCost == float64(int(fuelCost)) {
			fuelBn = fmt.Sprintf("%d SAR", int(fuelCost))
		}
		body = fmt.Sprintf("সুপারভাইজার %s আপনার অর্ডার অনুমোদন করেছেন\n%d টি অর্ডার | জ্বালানী: %s\n%s", supervisorLabelBn, ordersCount, fuelBn, oilInfo)
	case lang == "ur" || lang == "urdu" || lang == "pk":
		title = "تصدیق ہو گئی"
		fuelUr := fmt.Sprintf("%.2f ريال", fuelCost)
		if fuelCost == float64(int(fuelCost)) {
			fuelUr = fmt.Sprintf("%d ريال", int(fuelCost))
		}
		body = fmt.Sprintf("نگران %s نے آپ کے آرڈرز کی تصدیق کر دی\n%d آرڈرز | ایندھن: %s\n%s", supervisorLabelUr, ordersCount, fuelUr, oilInfo)
	default:
		title = "تمت المصادقة"
		body = fmt.Sprintf("وافق المشرف %s على طلباتك\n%d طلب | بنزين: %s\n%s", supervisorLabelAr, ordersCount, fuelFormatted, oilInfo)
	}

	_ = a.db.WithContext(ctx).Table("notifications").Create(map[string]interface{}{
		"id":          uuid.New(),
		"employee_id": empID,
		"branch_id":   emp.BranchID,
		"title":       title,
		"body":        body,
		"type":        "SESSION_APPROVED",
		"status":      "unread",
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}).Error

	if emp.PushToken != "" {
		token := emp.PushToken
		sid := sessionID.String()
		go legacyService.SendFCMBroadcast([]string{token}, title, body, map[string]string{
			"type":             "SESSION_APPROVED",
			"employee_id":      empID.String(),
			"sessionId":        sid,
			"session_id":       sid,
			"ordersCount":      fmt.Sprintf("%d", ordersCount),
			"fuelCost":         fmt.Sprintf("%.2f", fuelCost),
			"remaining_oil_km": fmt.Sprintf("%d", remainingInt),
			"needs_oil_change": fmt.Sprintf("%t", remainingInt <= 0),
		})
	}
	return nil
}
