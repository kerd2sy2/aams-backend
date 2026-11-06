package work

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/work/contracts"
	"delivery-backend/internal/modules/work/controller"
	"delivery-backend/internal/modules/work/domain"
	"delivery-backend/internal/modules/work/repository"
	"delivery-backend/internal/modules/work/service"
	legacyService "delivery-backend/internal/service"
	"delivery-backend/pkg/middleware"
)

type gormEmployeeAdapter struct {
	db *gorm.DB
}

func (a *gormEmployeeAdapter) GetEmployee(ctx context.Context, id uuid.UUID) (*service.EmployeeData, error) {
	var emp struct {
		ID                    uuid.UUID  `gorm:"column:id"`
		Name                  string     `gorm:"column:name"`
		NationalID            string     `gorm:"column:national_id"`
		BranchID              *uuid.UUID `gorm:"column:branch_id"`
		VehicleType           string     `gorm:"column:vehicle_type"`
		MotorcycleNumber      string     `gorm:"column:motorcycle_number"`
		ApplicationID         string     `gorm:"column:application_id"`
		ApplicationType       string     `gorm:"column:application_type"`
		TotalDistance         float64    `gorm:"column:total_distance"`
		LastOilChangeDistance float64    `gorm:"column:last_oil_change_distance"`
		LastOilChangeDate     *time.Time `gorm:"column:last_oil_change_date"`
	}

	err := a.db.WithContext(ctx).Table("employees").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&emp).Error
	if err != nil {
		return nil, err
	}

	return &service.EmployeeData{
		ID:                    emp.ID,
		Name:                  emp.Name,
		NationalID:            emp.NationalID,
		BranchID:              emp.BranchID,
		VehicleType:           emp.VehicleType,
		MotorcycleNumber:      emp.MotorcycleNumber,
		ApplicationID:         emp.ApplicationID,
		ApplicationType:       emp.ApplicationType,
		TotalDistance:         emp.TotalDistance,
		LastOilChangeDistance: emp.LastOilChangeDistance,
		LastOilChangeDate:     emp.LastOilChangeDate,
	}, nil
}

func (a *gormEmployeeAdapter) UpdateEmployeeOnStartWork(ctx context.Context, id uuid.UUID, appID, appType, motorcycleNumber string) error {
	updates := map[string]interface{}{
		"is_working": true,
	}
	if appID != "" {
		updates["application_id"] = appID
	}
	if appType != "" {
		updates["application_type"] = appType
	}
	if motorcycleNumber != "" {
		// Only set motorcycle_number if employee doesn't already have one assigned.
		// If they already have an assigned motorcycle, this shift is on a temporary bike,
		// and they revert back to their assigned motorcycle upon ending the shift.
		var currentAssigned *string
		_ = a.db.WithContext(ctx).Table("employees").Select("motorcycle_number").Where("id = ?", id).Scan(&currentAssigned).Error
		if currentAssigned == nil || strings.TrimSpace(*currentAssigned) == "" {
			updates["motorcycle_number"] = motorcycleNumber
		}
	}
	return a.db.WithContext(ctx).Table("employees").Where("id = ?", id).Updates(updates).Error
}

func (a *gormEmployeeAdapter) UpdateEmployeeOnEndWork(ctx context.Context, id uuid.UUID, addedDistance float64, totalOrders int) error {
	return a.db.WithContext(ctx).Table("employees").Where("id = ?", id).Updates(map[string]interface{}{
		"is_working":      false,
		"total_distance":  gorm.Expr("total_distance + ?", addedDistance),
		"total_orders":    gorm.Expr("total_orders + ?", totalOrders),
	}).Error
}

type supervisorNameTranslations struct {
	ar string
	en string
	bn string
	ur string
}

func normalizeNameToken(token string) string {
	token = strings.TrimSpace(token)
	token = strings.ReplaceAll(token, "أ", "ا")
	token = strings.ReplaceAll(token, "إ", "ا")
	token = strings.ReplaceAll(token, "آ", "ا")
	token = strings.ReplaceAll(token, "ة", "ه")
	token = strings.ReplaceAll(token, "ى", "ي")
	return token
}

func translateSupervisorName(rawName string, lang string) string {
	rawName = strings.TrimSpace(rawName)
	if rawName == "" {
		switch lang {
		case "en":
			return "Supervisor"
		case "bn":
			return "সুপারভাইজার"
		case "ur":
			return "نگران"
		default:
			return "المشرف"
		}
	}

	knownSupervisors := map[string]supervisorNameTranslations{
		"ابراهيم محمود": {
			ar: "إبراهيم محمود",
			en: "Ibrahim Mahmoud",
			bn: "ইব্রাহিম মাহমুদ",
			ur: "ابراہیم محمود",
		},
		"اسلام ابوالعلا": {
			ar: "إسلام أبو العلا",
			en: "Islam Abulela",
			bn: "ইসলাম আবু আল-এلا",
			ur: "اسلام ابو العلا",
		},
		"اسلام ابو العلا": {
			ar: "إسلام أبو العلا",
			en: "Islam Abulela",
			bn: "ইসলাম আবু আল-এلا",
			ur: "اسلام ابو العلا",
		},
		"سيد احمد": {
			ar: "سيد أحمد",
			en: "Sayed Ahmed",
			bn: "সৈয়দ আহমেদ",
			ur: "سید احمد",
		},
		"محمد هشام": {
			ar: "محمد هشام",
			en: "Mohamed Hesham",
			bn: "মোহাম্মদ হিশাম",
			ur: "محمد ہشام",
		},
		"هادي نبيل": {
			ar: "هادي نبيل",
			en: "Hadi Nabil",
			bn: "হাদি নাবিল",
			ur: "ہادی نبیل",
		},
		"محمد صبري": {
			ar: "محمد صبري",
			en: "Mohamed Sabry",
			bn: "মোহাম্মদ সাবরি",
			ur: "محمد صبری",
		},
		"هاني الشمراني": {
			ar: "هاني الشمراني",
			en: "Hani Al-Shamrani",
			bn: "হানি আল-শামরানি",
			ur: "ہانی الشمرانی",
		},
		"ساره خالد": {
			ar: "سارة خالد",
			en: "Sarah Khaled",
			bn: "সারাহ খালেদ",
			ur: "سارہ خالد",
		},
		"صفا محمد": {
			ar: "صفا محمد",
			en: "Safa Mohamed",
			bn: "সাফা মোহাম্মদ",
			ur: "صفا محمد",
		},
	}

	normRaw := normalizeNameToken(rawName)
	for key, trans := range knownSupervisors {
		if strings.HasPrefix(normRaw, normalizeNameToken(key)) {
			switch lang {
			case "en":
				return trans.en
			case "bn":
				return trans.bn
			case "ur":
				return trans.ur
			default:
				return trans.ar
			}
		}
	}

	parts := strings.Fields(rawName)
	if len(parts) == 0 {
		return rawName
	}

	firstNames := []string{parts[0]}
	partIdx := 1
	if (parts[0] == "عبد" || parts[0] == "ابو" || parts[0] == "أبو") && len(parts) > 1 {
		firstNames = append(firstNames, parts[1])
		partIdx = 2
	}
	if partIdx < len(parts) {
		firstNames = append(firstNames, parts[partIdx])
		if (parts[partIdx] == "عبد" || parts[partIdx] == "ابو" || parts[partIdx] == "أبو") && len(parts) > partIdx+1 {
			firstNames = append(firstNames, parts[partIdx+1])
		}
	}

	twoNames := strings.Join(firstNames, " ")
	normTwo := normalizeNameToken(twoNames)
	for key, trans := range knownSupervisors {
		if strings.HasPrefix(normTwo, normalizeNameToken(key)) {
			switch lang {
			case "en":
				return trans.en
			case "bn":
				return trans.bn
			case "ur":
				return trans.ur
			default:
				return trans.ar
			}
		}
	}

	wordDict := map[string]supervisorNameTranslations{
		"محمد":        {ar: "محمد", en: "Mohamed", bn: "মোহাম্মদ", ur: "محمد"},
		"احمد":        {ar: "أحمد", en: "Ahmed", bn: "আহমেদ", ur: "احمد"},
		"محمود":       {ar: "محمود", en: "Mahmoud", bn: "মাহমুদ", ur: "محمود"},
		"ابراهيم":     {ar: "إبراهيم", en: "Ibrahim", bn: "ইব্রাহিম", ur: "ابراہیم"},
		"علي":         {ar: "علي", en: "Ali", bn: "আলী", ur: "علی"},
		"حسن":         {ar: "حسن", en: "Hassan", bn: "হাসান", ur: "حسن"},
		"حسين":        {ar: "حسين", en: "Hussein", bn: "হোসেন", ur: "حسین"},
		"عمر":         {ar: "عمر", en: "Omar", bn: "ওমর", ur: "عمر"},
		"عثمان":       {ar: "عثمان", en: "Othman", bn: "ওসমান", ur: "عثمان"},
		"خالد":        {ar: "خالد", en: "Khaled", bn: "খালেদ", ur: "خالد"},
		"سيد":         {ar: "سيد", en: "Sayed", bn: "সৈয়দ", ur: "سید"},
		"هشام":        {ar: "هشام", en: "Hesham", bn: "হিশাম", ur: "ہشام"},
		"نبيل":        {ar: "نبيل", en: "Nabil", bn: "নাবিল", ur: "نبیل"},
		"هادي":        {ar: "هادي", en: "Hadi", bn: "হাদি", ur: "ہادی"},
		"صبري":        {ar: "صبري", en: "Sabry", bn: "সাবরি", ur: "صبری"},
		"اسلام":       {ar: "إسلام", en: "Islam", bn: "ইসলাম", ur: "اسلام"},
		"طارق":        {ar: "طارق", en: "Tariq", bn: "তারেক", ur: "طارق"},
		"يوسف":        {ar: "يوسف", en: "Youssef", bn: "ইউসুফ", ur: "یوسف"},
		"ياسر":        {ar: "ياسر", en: "Yasser", bn: "ইয়াসির", ur: "یاسر"},
		"مصطفي":       {ar: "مصطفى", en: "Mustafa", bn: "মুস্তফা", ur: "مصطفی"},
		"عبدالله":     {ar: "عبد الله", en: "Abdullah", bn: "আব্দুল্লাহ", ur: "عبد اللہ"},
		"عبد":         {ar: "عبد", en: "Abd", bn: "আব্দ", ur: "عبد"},
		"الله":        {ar: "الله", en: "Allah", bn: "আল্লাহ", ur: "اللہ"},
		"الرحمن":      {ar: "الرحمن", en: "Al-Rahman", bn: "আর-রহমান", ur: "الرحمن"},
		"العزيز":      {ar: "العزيز", en: "Al-Aziz", bn: "আল-আজিজ", ur: "العزیز"},
		"ساره":        {ar: "سارة", en: "Sarah", bn: "সারাহ", ur: "سارہ"},
		"صفا":         {ar: "صفا", en: "Safa", bn: "সাফা", ur: "صفا"},
		"نور":         {ar: "نور", en: "Nour", bn: "নুর", ur: "نور"},
		"علم":         {ar: "علم", en: "Alam", bn: "আলম", ur: "علم"},
		"سالم":        {ar: "سالم", en: "Salem", bn: "সালেম", ur: "سالم"},
		"سلمان":       {ar: "سلمان", en: "Salman", bn: "সালমান", ur: "سلمان"},
		"فهد":         {ar: "فهد", en: "Fahad", bn: "ফাহাদ", ur: "فہد"},
		"سلطان":       {ar: "سلطان", en: "Sultan", bn: "সুলতান", ur: "سلطان"},
		"فيصل":        {ar: "فيصل", en: "Faisal", bn: "ফয়সাল", ur: "فیصل"},
		"بندر":        {ar: "بندر", en: "Bandar", bn: "বান্দার", ur: "بندر"},
		"تركي":        {ar: "تركي", en: "Turki", bn: "তুর্কি", ur: "ترکی"},
		"ماجد":        {ar: "ماجد", en: "Majed", bn: "মাজেদ", ur: "ماجد"},
		"وليد":        {ar: "وليد", en: "Waleed", bn: "ওয়ালিদ", ur: "ولید"},
		"عادل":        {ar: "عادل", en: "Adel", bn: "আদেল", ur: "عادل"},
		"عمرو":        {ar: "عمرو", en: "Amr", bn: "আমর", ur: "عمرو"},
		"كريم":        {ar: "كريم", en: "Kareem", bn: "করিম", ur: "کریم"},
		"سامي":        {ar: "سامي", en: "Sami", bn: "সামি", ur: "سامی"},
		"رامي":        {ar: "رامي", en: "Rami", bn: "রামি", ur: "رامی"},
		"شريف":        {ar: "شريف", en: "Sherif", bn: "শরিফ", ur: "شریف"},
		"ايمن":        {ar: "أيمن", en: "Ayman", bn: "আয়মান", ur: "ایمن"},
		"انس":         {ar: "أنس", en: "Anas", bn: "আনাস", ur: "انس"},
		"زياد":        {ar: "زياد", en: "Ziyad", bn: "জিয়াদ", ur: "زیاد"},
		"حمزه":        {ar: "حمزة", en: "Hamza", bn: "হামজা", ur: "حمزہ"},
		"بلال":        {ar: "بلال", en: "Bilal", bn: "বেলাল", ur: "بلال"},
		"باسم":        {ar: "باسم", en: "Bassem", bn: "বাসেম", ur: "باسم"},
		"حاتم":        {ar: "حاتم", en: "Hatem", bn: "হাতেম", ur: "حاتم"},
		"عصام":        {ar: "عصام", en: "Essam", bn: "এসাম", ur: "عصام"},
		"وائل":        {ar: "وائل", en: "Wael", bn: "ওয়ায়েল", ur: "وائل"},
		"مهند":        {ar: "مهند", en: "Mohannad", bn: "মুহান্নাদ", ur: "مہند"},
		"حسام":        {ar: "حسام", en: "Hossam", bn: "হোসাম", ur: "حسام"},
		"بهاء":        {ar: "بهاء", en: "Bahaa", bn: "বাহা", ur: "بہاء"},
		"علاء":        {ar: "علاء", en: "Alaa", bn: "আলা", ur: "علاء"},
		"رضا":         {ar: "رضا", en: "Reda", bn: "রেজা", ur: "رضا"},
		"فاروق":       {ar: "فاروق", en: "Farouk", bn: "ফারুক", ur: "فاروق"},
		"الشمراني":    {ar: "الشمراني", en: "Al-Shamrani", bn: "আল-শামরানি", ur: "الشمرانی"},
		"المالكي":     {ar: "المالكي", en: "Al-Malki", bn: "আল-মালকি", ur: "المالکی"},
		"الكرداس":     {ar: "الكرداس", en: "Al-Kerdas", bn: "আল-কেরদাস", ur: "الکرداس"},
		"الكرداسي":    {ar: "الكرداسي", en: "Al-Kerdasy", bn: "আল-কেরদাসী", ur: "الکرداسی"},
		"ندا":         {ar: "ندا", en: "Neda", bn: "নেদা", ur: "ندا"},
		"شلبه":        {ar: "شلبه", en: "Shelba", bn: "শেলবা", ur: "شلبہ"},
		"بركه":        {ar: "بركه", en: "Baraka", bn: "বারাকা", ur: "برکہ"},
	}

	translatedParts := make([]string, 0, len(firstNames))
	for _, p := range firstNames {
		normP := normalizeNameToken(p)
		if entry, ok := wordDict[normP]; ok {
			switch lang {
			case "en":
				translatedParts = append(translatedParts, entry.en)
			case "bn":
				translatedParts = append(translatedParts, entry.bn)
			case "ur":
				translatedParts = append(translatedParts, entry.ur)
			default:
				translatedParts = append(translatedParts, entry.ar)
			}
		} else {
			translatedParts = append(translatedParts, p)
		}
	}

	if len(translatedParts) > 0 {
		return strings.Join(translatedParts, " ")
	}

	return twoNames
}

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

	// Calculate remaining KM for oil change
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
	default: // "ar"
		title = "تمت المصادقة"
		body = fmt.Sprintf("وافق المشرف %s على طلباتك\n%d طلب | بنزين: %s\n%s", supervisorLabelAr, ordersCount, fuelFormatted, oilInfo)
	}

	// 1. Save in-app notification in DB
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

	// 2. Send push notification to courier's phone
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

func (a *gormEmployeeAdapter) SendShiftEndedNotificationToSupervisor(ctx context.Context, empID uuid.UUID, session *domain.WorkSession) error {
	sessionCopy := *session
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		var emp struct {
			Name             string     `gorm:"column:name"`
			KeyNumber        string     `gorm:"column:key_number"`
			EmployeeNumber   string     `gorm:"column:employee_number"`
			MotorcycleNumber string     `gorm:"column:motorcycle_number"`
			BranchID         *uuid.UUID `gorm:"column:branch_id"`
		}
		if err := a.db.WithContext(bgCtx).Table("employees").Where("id = ? AND deleted_at IS NULL", empID).First(&emp).Error; err != nil {
			return
		}

		empName := strings.TrimSpace(emp.Name)
		if empName == "" {
			empName = "المندوب"
		}
		bikeNum := strings.TrimSpace(sessionCopy.MotorcycleNumber)
		if bikeNum == "" {
			bikeNum = strings.TrimSpace(emp.MotorcycleNumber)
		}

		// Look up key number: first from employee, then from assigned vehicle
		keyNum := strings.TrimSpace(emp.KeyNumber)
		if keyNum == "" && bikeNum != "" {
			var v struct {
				KeyNumber string `gorm:"column:key_number"`
			}
			cleanBike := strings.ReplaceAll(bikeNum, " ", "")
			if err := a.db.WithContext(bgCtx).Table("vehicles").
				Where("(plate_number = ? OR REPLACE(plate_number, ' ', '') = ?) AND deleted_at IS NULL", bikeNum, cleanBike).
				First(&v).Error; err == nil {
				keyNum = strings.TrimSpace(v.KeyNumber)
			}
		}

		title := "إنهاء وردية مندوب"
		var body string
		if keyNum != "" {
			body = fmt.Sprintf("أنهى المندوب %s - مفتاح رقم %s", empName, keyNum)
		} else if bikeNum != "" {
			body = fmt.Sprintf("أنهى المندوب %s - دباب رقم %s", empName, bikeNum)
		} else {
			body = fmt.Sprintf("أنهى المندوب %s الشفت وبانتظار الاعتماد", empName)
		}

		var tokens []string
		tokenSet := make(map[string]bool)

		// A) Query supervisors from employees table for this employee's branch ONLY
		type supervisorRec struct {
			ID        uuid.UUID  `gorm:"column:id"`
			BranchID  *uuid.UUID `gorm:"column:branch_id"`
			PushToken string     `gorm:"column:push_token"`
		}
		var empSupervisors []supervisorRec
		empQuery := a.db.WithContext(bgCtx).Table("employees").
			Where("(LOWER(job_role) = 'supervisor' OR LOWER(role_name) = 'supervisor') AND deleted_at IS NULL AND push_token IS NOT NULL AND push_token != ''")
		if emp.BranchID != nil {
			empQuery = empQuery.Where("branch_id = ?", *emp.BranchID)
		} else {
			empQuery = empQuery.Where("branch_id IS NULL")
		}
		_ = empQuery.Find(&empSupervisors).Error

		for _, s := range empSupervisors {
			tok := strings.TrimSpace(s.PushToken)
			if tok != "" && !tokenSet[tok] {
				tokenSet[tok] = true
				tokens = append(tokens, tok)
			}
			_ = a.db.WithContext(bgCtx).Table("notifications").Create(map[string]interface{}{
				"id":          uuid.New(),
				"employee_id": s.ID,
				"branch_id":   s.BranchID,
				"title":       title,
				"body":        body,
				"type":        "SHIFT_COMPLETED_PENDING_REVIEW",
				"status":      "unread",
				"created_at":  time.Now(),
				"updated_at":  time.Now(),
			}).Error
		}

		// B) Query supervisors from admins table for this employee's branch ONLY
		type adminRec struct {
			ID        uuid.UUID  `gorm:"column:id"`
			BranchID  *uuid.UUID `gorm:"column:branch_id"`
			PushToken string     `gorm:"column:push_token"`
			Role      string     `gorm:"column:role"`
		}
		var adminSupervisors []adminRec
		adminQuery := a.db.WithContext(bgCtx).Table("admins").
			Where("(LOWER(role) = 'supervisor') AND deleted_at IS NULL AND push_token IS NOT NULL AND push_token != ''")
		if emp.BranchID != nil {
			adminQuery = adminQuery.Where("branch_id = ?", *emp.BranchID)
		} else {
			adminQuery = adminQuery.Where("branch_id IS NULL")
		}
		_ = adminQuery.Find(&adminSupervisors).Error

		for _, adm := range adminSupervisors {
			tok := strings.TrimSpace(adm.PushToken)
			if tok != "" && !tokenSet[tok] {
				tokenSet[tok] = true
				tokens = append(tokens, tok)
			}
			_ = a.db.WithContext(bgCtx).Table("notifications").Create(map[string]interface{}{
				"id":         uuid.New(),
				"admin_id":   adm.ID,
				"branch_id":  adm.BranchID,
				"title":      title,
				"body":       body,
				"type":       "SHIFT_COMPLETED_PENDING_REVIEW",
				"status":     "unread",
				"created_at": time.Now(),
				"updated_at": time.Now(),
			}).Error
		}

		// 2. Send push notification to all collected supervisor devices
		if len(tokens) > 0 {
			sid := sessionCopy.ID.String()
			legacyService.SendFCMBroadcast(tokens, title, body, map[string]string{
				"type":         "SHIFT_COMPLETED_PENDING_REVIEW",
				"employee_id":  empID.String(),
				"employeeName": empName,
				"keyNumber":    keyNum,
				"key_number":   keyNum,
				"bikeNumber":   bikeNum,
				"sessionId":    sid,
				"session_id":   sid,
				"ordersCount":  fmt.Sprintf("%d", sessionCopy.OrdersCount),
				"distance":     fmt.Sprintf("%.1f", sessionCopy.Distance),
			})
		}
	}()

	return nil
}

type gormVehicleAdapter struct {
	db *gorm.DB
}

func (v *gormVehicleAdapter) GetVehicleLastKM(ctx context.Context, plateNumber string) (float64, error) {
	info, err := v.GetVehicleInfo(ctx, plateNumber)
	if err != nil || info == nil {
		return 0, err
	}
	return info.CurrentKM, nil
}

func (v *gormVehicleAdapter) GetVehicleInfo(ctx context.Context, plateNumber string) (*service.VehicleData, error) {
	clean := strings.TrimSpace(plateNumber)
	if clean == "" {
		return nil, errors.New("رقم اللوحة فارغ")
	}

	var row struct {
		ID                uuid.UUID `gorm:"column:id"`
		PlateNumber       string    `gorm:"column:plate_number"`
		VehicleType       string    `gorm:"column:vehicle_type"`
		CurrentKM         float64   `gorm:"column:current_km"`
		LastOilChangeKM   float64   `gorm:"column:last_oil_change_km"`
		TotalDistance     float64   `gorm:"column:total_distance"`
		IsOdometerBroken  bool      `gorm:"column:is_odometer_broken"`
		RegistrationImage string    `gorm:"column:registration_image"`
		Status            string    `gorm:"column:status"`
	}

	err := v.db.WithContext(ctx).Table("vehicles").
		Where("deleted_at IS NULL AND (TRIM(plate_number) = ? OR plate_number ILIKE ? OR plate_number ILIKE ?)", clean, clean+"%", "%"+clean+"%").
		Order("CASE WHEN TRIM(plate_number) = '" + clean + "' THEN 0 ELSE 1 END").
		First(&row).Error
	if err != nil {
		return nil, err
	}

	return &service.VehicleData{
		ID:                row.ID,
		PlateNumber:       row.PlateNumber,
		VehicleType:       row.VehicleType,
		CurrentKM:         row.CurrentKM,
		LastOilChangeKM:   row.LastOilChangeKM,
		TotalDistance:     row.TotalDistance,
		IsOdometerBroken:  row.IsOdometerBroken,
		RegistrationImage: row.RegistrationImage,
		Status:            row.Status,
	}, nil
}

func (v *gormVehicleAdapter) EnsureVehicleExists(ctx context.Context, plateNumber string, vehicleType string, branchID *uuid.UUID, startKM float64) (*service.VehicleData, error) {
	clean := strings.TrimSpace(plateNumber)
	if clean == "" {
		return nil, errors.New("رقم اللوحة فارغ")
	}

	existing, err := v.GetVehicleInfo(ctx, clean)
	if err == nil && existing != nil {
		if startKM > 0 && existing.CurrentKM == 0 {
			_ = v.UpdateVehicleKM(ctx, clean, startKM)
		}
		_ = v.db.WithContext(ctx).Table("vehicles").Where("id = ?", existing.ID).Update("status", "IN_USE").Error
		existing.Status = "IN_USE"
		return existing, nil
	}

	vType := "motorcycle"
	if strings.EqualFold(vehicleType, "car") || strings.Contains(clean, "2459") || strings.Contains(clean, "STA") || strings.Contains(clean, "س ط ا") {
		vType = "car"
	}

	newID := uuid.New()
	now := time.Now()
	newVehicle := map[string]interface{}{
		"id":                 newID,
		"plate_number":       clean,
		"vehicle_type":       vType,
		"current_km":         startKM,
		"last_oil_change_km": startKM,
		"total_distance":     0.0,
		"status":             "IN_USE",
		"branch_id":          branchID,
		"created_at":         now,
		"updated_at":         now,
	}

	if err := v.db.WithContext(ctx).Table("vehicles").Create(newVehicle).Error; err != nil {
		return v.GetVehicleInfo(ctx, clean)
	}

	return &service.VehicleData{
		ID:              newID,
		PlateNumber:     clean,
		VehicleType:     vType,
		CurrentKM:       startKM,
		LastOilChangeKM: startKM,
		Status:          "IN_USE",
	}, nil
}

func (v *gormVehicleAdapter) UpdateVehicleKM(ctx context.Context, plateNumber string, km float64) error {
	clean := strings.TrimSpace(plateNumber)
	if clean == "" {
		return nil
	}
	var vehicle struct {
		ID              uuid.UUID `gorm:"column:id"`
		CurrentKM       float64   `gorm:"column:current_km"`
		LastOilChangeKM float64   `gorm:"column:last_oil_change_km"`
	}
	err := v.db.WithContext(ctx).Table("vehicles").
		Where("deleted_at IS NULL AND (TRIM(plate_number) = ? OR plate_number ILIKE ? OR plate_number ILIKE ?)", clean, clean+"%", "%"+clean+"%").
		Order("CASE WHEN TRIM(plate_number) = '" + clean + "' THEN 0 ELSE 1 END").
		First(&vehicle).Error
	if err == nil {
		delta := km - vehicle.CurrentKM
		if delta < 0 {
			delta = 0
		}
		updates := map[string]interface{}{
			"status": "AVAILABLE",
		}
		if km > vehicle.CurrentKM {
			updates["current_km"] = km
		}
		if vehicle.LastOilChangeKM <= 0 && vehicle.CurrentKM <= 0 && km > 0 {
			updates["last_oil_change_km"] = km
		}
		if delta > 0 {
			updates["total_distance"] = gorm.Expr("COALESCE(total_distance, 0) + ?", delta)
		}
		return v.db.WithContext(ctx).Table("vehicles").
			Where("id = ?", vehicle.ID).
			Updates(updates).Error
	}
	return nil
}

func (v *gormVehicleAdapter) HasActiveMaintenance(ctx context.Context, motorcycleNumber string) (bool, error) {
	var count int64
	clean := strings.TrimSpace(motorcycleNumber)
	err := v.db.WithContext(ctx).Table("maintenance_requests").
		Where("(TRIM(vehicle_plate) = ? OR vehicle_plate ILIKE ?) AND status IN ?", clean, clean+"%", []string{"pending", "in_progress", "OPEN"}).
		Count(&count).Error
	return count > 0, err
}

type Module struct {
	Service service.WorkService
	Handler *controller.WorkHandler
}

func NewModule(db *gorm.DB, storage service.StorageProvider, audit ...contracts.IAuditContract) *Module {
	repo := repository.NewWorkRepository(db)
	empAdapter := &gormEmployeeAdapter{db: db}
	vehicleAdapter := &gormVehicleAdapter{db: db}
	svc := service.NewWorkService(repo, empAdapter, vehicleAdapter, storage)
	var auditContract contracts.IAuditContract
	if len(audit) > 0 {
		auditContract = audit[0]
	}
	handler := controller.NewWorkHandler(svc, auditContract)
	return &Module{
		Service: svc,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/work/start", m.Handler.StartWork)
	rg.POST("/work/end", m.Handler.EndWork)
	rg.POST("/work/manual-entry", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR", "ACCOUNTANT"), m.Handler.SaveManualEntries)
	rg.PUT("/work/:id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.UpdateWorkSession)
	rg.PUT("/work/:id/review", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.ReviewWorkSession)
	rg.PUT("/work/sessions/:id/review", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.ReviewWorkSession)
	rg.GET("/work/sessions/:id", m.Handler.GetSessionByID)
	rg.GET("/work/active", m.Handler.GetActiveSession)
	rg.GET("/work/last-km", m.Handler.GetLastKM)
	rg.POST("/work/scan-plate", m.Handler.ScanPlate)
	rg.GET("/work/today-count", m.Handler.TodayCount)
	rg.GET("/work/check-oil", m.Handler.CheckOilChange)

	// Supervisor shift routes (Strictly restricted to supervisor / management roles)
	supervisorWork := rg.Group("/work/supervisor")
	supervisorWork.Use(middleware.RequireRoles("SUPERVISOR", "ADMIN", "SUPER_ADMIN"))
	{
		supervisorWork.POST("/start-shift", m.Handler.SupervisorStartShift)
		supervisorWork.POST("/end-shift", m.Handler.SupervisorEndShift)
	}
}
