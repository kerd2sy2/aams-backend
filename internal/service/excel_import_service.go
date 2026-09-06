package service

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/dto"
	"delivery-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type ExcelImportService interface {
	ParseAndPreviewExcel(ctx context.Context, fileBytes []byte, filename string, customDate string) (*dto.ExcelImportPreviewResponse, error)
	ConfirmImport(ctx context.Context, req dto.ConfirmImportRequest, adminID uuid.UUID, adminName string) (*dto.ConfirmImportResponse, error)
}

type excelImportService struct {
	targetRepo repository.TargetRepository
}

func NewExcelImportService(targetRepo repository.TargetRepository) ExcelImportService {
	return &excelImportService{targetRepo: targetRepo}
}

// ParseAndPreviewExcel inspects the uploaded file, parses rows 4+, filters out footers, and checks duplicates
func (s *excelImportService) ParseAndPreviewExcel(ctx context.Context, fileBytes []byte, filename string, customDate string) (*dto.ExcelImportPreviewResponse, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("فشل في فتح ملف الإكسل: %w", err)
	}
	defer f.Close()

	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("الملف لا يحتوي على أي صفحات (Sheets)")
	}
	sheetName := sheetList[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("فشل في قراءة بيانات الصفحة %s: %w", sheetName, err)
	}

	if len(rows) < 4 {
		return nil, fmt.Errorf("ملف الإكسل لا يحتوي على صفوف بيانات كافية (الحد الأدنى 4 صفوف)")
	}

	// 1. Determine Date (from customDate, or filename regex, or fallback to today)
	orderDate := extractDateFromFilename(filename)
	if customDate != "" {
		if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, customDate); matched {
			orderDate = customDate
		}
	}
	if orderDate == "" {
		orderDate = time.Now().Format("2006-01-02")
	}

	// 2. Identify header columns from Row 3 (index 2 in 0-indexed slice)
	headerRowIndex := 2
	headerRow := rows[headerRowIndex]

	colSerial := 0 // A: م
	colIdent := 1  // B: المعرف
	colApp := 2    // C: التطبيق
	colBranch := -1 // الفرع
	colDriver := 3 // D: الاسم / المندوب
	colNinja := 4  // E: نينجا
	colKeeta := 5  // F: كيتا
	colToyo := 6   // G: تويو
	colPlate := 7  // H: رقم اللوحة
	colNotes := 8  // I: ملاحظات

	// Dynamic detection of header row & columns if headers are slightly shifted
	for rIdx := 0; rIdx < len(rows) && rIdx < 5; rIdx++ {
		r := rows[rIdx]
		hasIdent := false
		for cIdx, cell := range r {
			trimCell := strings.TrimSpace(cell)
			if strings.Contains(trimCell, "المعرف") {
				hasIdent = true
				colIdent = cIdx
			}
			if strings.Contains(trimCell, "التطبيق") {
				colApp = cIdx
			}
			if strings.Contains(trimCell, "الفرع") {
				colBranch = cIdx
			}
			if strings.Contains(trimCell, "الاسم") || strings.Contains(trimCell, "المندوب") {
				colDriver = cIdx
			}
			if strings.Contains(trimCell, "نينجا") {
				colNinja = cIdx
			}
			if strings.Contains(trimCell, "كيتا") || strings.Contains(trimCell, "كينتا") {
				colKeeta = cIdx
			}
			if strings.Contains(trimCell, "تويو") {
				colToyo = cIdx
			}
			if strings.Contains(trimCell, "اللوحة") {
				colPlate = cIdx
			}
			if strings.Contains(trimCell, "ملاحظ") {
				colNotes = cIdx
			}
		}
		if hasIdent {
			headerRowIndex = rIdx
			break
		}
	}
	_ = headerRow

	// 3. Process Data Rows (starting from headerRowIndex + 1)
	var parsedRows []dto.ParsedExcelRow
	identMap := make(map[string]bool)
	driverMap := make(map[string]bool)
	totalOrders := 0
	duplicatesCount := 0
	emptyIdentCount := 0

	for i := headerRowIndex + 1; i < len(rows); i++ {
		r := rows[i]
		if isFooterOrEmptyRow(r) {
			continue
		}

		identName := getCell(r, colIdent)
		driverName := getCell(r, colDriver)
		appName := getCell(r, colApp)

		// Must have at least a driver name to process row
		if identName == "" && driverName == "" {
			continue
		}

		// If identifier is empty, leave it empty (track for warning)
		if identName == "" {
			emptyIdentCount++
		}
		// If driver name is empty, use identifier as fallback
		if driverName == "" {
			driverName = identName
		}

		// Normalize app name first (authoritative source is the app column)
		if appName == "كينتا" {
			appName = "كيتا"
		}

		ninjaCount := parseInt(getCell(r, colNinja))
		keetaCount := parseInt(getCell(r, colKeeta))
		toyoCount := parseInt(getCell(r, colToyo))
		// Total orders is sum of ALL count columns regardless of app name
		rowTotal := ninjaCount + keetaCount + toyoCount

		// Fallback: If no app columns but app cell has orders count or general column
		if rowTotal == 0 {
			if n := parseInt(appName); n > 0 {
				rowTotal = n
				appName = "عام"
			}
		}

		plate := getCell(r, colPlate)
		notes := getCell(r, colNotes)
		serial := getCell(r, colSerial)
		if serial == "" {
			serial = strconv.Itoa(len(parsedRows) + 1)
		}

		// If app column is empty, try to infer from count columns
		if appName == "" {
			if ninjaCount > 0 && keetaCount == 0 && toyoCount == 0 {
				appName = "نينجا"
			} else if keetaCount > 0 && ninjaCount == 0 && toyoCount == 0 {
				appName = "كيتا"
			} else if toyoCount > 0 && ninjaCount == 0 && keetaCount == 0 {
				appName = "تويو"
			} else {
				appName = "كيتا" // default fallback
			}
		}

		// Check for duplicate in database for this date, identifier, driver, and app
		isDup := false
		existingCount := 0
		var identID *uuid.UUID
		if identName != "" {
			if identObj, _ := s.targetRepo.FindIdentifierByNameAndApp(ctx, identName, appName); identObj != nil {
				identID = &identObj.ID
			}
		}
		driverObj, _ := s.targetRepo.FindDriverByName(ctx, driverName)
		if driverObj != nil {
			dup, _ := s.targetRepo.CheckDuplicates(ctx, orderDate, identID, driverObj.ID, appName)
			if dup {
				isDup = true
				duplicatesCount++
				existingCount = 1
			}
		}

		branchVal := ""
		if colBranch >= 0 {
			branchVal = getCell(r, colBranch)
		}

		parsedRow := dto.ParsedExcelRow{
			Serial:        serial,
			Identifier:    identName,
			App:           appName,
			Branch:        branchVal,
			DriverName:    driverName,
			NinjaOrders:   ninjaCount,
			KeetaOrders:   keetaCount,
			ToyoOrders:    toyoCount,
			TotalOrders:   rowTotal,
			PlateNumber:   plate,
			Notes:         notes,
			IsDuplicate:   isDup,
			ExistingCount: existingCount,
		}

		parsedRows = append(parsedRows, parsedRow)
		if identName != "" {
			identKey := identName
			if appName != "" {
				identKey = fmt.Sprintf("%s (%s)", identName, appName)
			}
			identMap[identKey] = true
		}
		if driverName != "" {
			driverMap[driverName] = true
		}
		totalOrders += rowTotal
	}

	var identifiersList []string
	for k := range identMap {
		identifiersList = append(identifiersList, k)
	}
	var driversList []string
	for k := range driverMap {
		driversList = append(driversList, k)
	}

	// Build warnings
	var warnings []string
	if emptyIdentCount > 0 {
		warnings = append(warnings, fmt.Sprintf("⚠️ يوجد %d صف بدون معرف (المعرف فارغ)", emptyIdentCount))
	}

	return &dto.ExcelImportPreviewResponse{
		FileName:              filename,
		OrderDate:             orderDate,
		TotalRows:             len(parsedRows),
		TotalOrders:           totalOrders,
		IdentifiersCount:      len(identifiersList),
		Identifiers:           identifiersList,
		DriversCount:          len(driversList),
		Drivers:               driversList,
		DuplicatesCount:       duplicatesCount,
		HasDuplicates:         duplicatesCount > 0,
		EmptyIdentifiersCount: emptyIdentCount,
		Warnings:              warnings,
		Rows:                  parsedRows,
	}, nil
}

// ConfirmImport persists the previewed rows into PostgreSQL
func (s *excelImportService) ConfirmImport(ctx context.Context, req dto.ConfirmImportRequest, adminID uuid.UUID, adminName string) (*dto.ConfirmImportResponse, error) {
	if len(req.Rows) == 0 {
		return nil, fmt.Errorf("لا توجد بيانات للاستيراد")
	}

	if req.DeduplicationAction == "CANCEL" {
		return nil, fmt.Errorf("تم إلغاء عملية الاستيراد من قِبل المستخدم")
	}

	// 1. Create Import Batch Record
	batch := domain.ImportBatch{
		FileName:       req.FileName,
		OrderDate:      req.OrderDate,
		UploadedBy:     adminID,
		UploadedByName: adminName,
		TotalRows:      len(req.Rows),
		Status:         "COMPLETED",
	}
	if err := s.targetRepo.CreateImportBatch(ctx, &batch); err != nil {
		return nil, fmt.Errorf("فشل في إنشاء سجل رفعة الإكسل: %w", err)
	}

	importedCount := 0
	skippedCount := 0
	replacedCount := 0
	totalOrders := 0

	// Cache identifier and driver UUIDs to minimize DB queries
	identCache := make(map[string]*domain.Identifier)
	driverCache := make(map[string]*domain.Driver)

	var ordersToInsert []domain.DailyOrder
	activeIdentifiersForAlerts := make(map[uuid.UUID]*domain.Identifier)

	for _, row := range req.Rows {
		// Handle duplicate according to user choice
		if row.IsDuplicate {
			if req.DeduplicationAction == "IGNORE_DUPLICATES" {
				skippedCount++
				continue
			}
		}

		var identID *uuid.UUID
		var ident *domain.Identifier

		// Find or Create Identifier ONLY IF row.Identifier is NOT empty!
		identName := strings.TrimSpace(row.Identifier)
		if identName != "" {
			identCacheKey := fmt.Sprintf("%s___%s", identName, row.App)
			var ok bool
			ident, ok = identCache[identCacheKey]
			if !ok {
				var err error
				ident, err = s.targetRepo.FindOrCreateIdentifier(ctx, identName, row.App)
				if err != nil {
					return nil, fmt.Errorf("فشل في تسجيل المعرف %s (%s): %w", identName, row.App, err)
				}
				identCache[identCacheKey] = ident
			}
			identID = &ident.ID
			activeIdentifiersForAlerts[ident.ID] = ident
			if row.Branch != "" && ident.Branch != row.Branch {
				ident.Branch = row.Branch
				_ = s.targetRepo.UpdateIdentifier(ctx, ident)
			}
		}

		// Find or Create Driver
		driver, ok := driverCache[row.DriverName]
		if !ok {
			var err error
			driver, err = s.targetRepo.FindOrCreateDriver(ctx, row.DriverName)
			if err != nil {
				return nil, fmt.Errorf("فشل في تسجيل المندوب %s: %w", row.DriverName, err)
			}
			driverCache[row.DriverName] = driver
		}

		// Link Driver to Identifier ONLY if identifier exists
		if identID != nil {
			_ = s.targetRepo.LinkDriverToIdentifier(ctx, *identID, driver.ID, req.OrderDate)
		}

		// If Replace Duplicates: delete previous matching record
		if row.IsDuplicate && req.DeduplicationAction == "REPLACE_DUPLICATES" {
			_ = s.targetRepo.DeleteOrdersByDateAndApp(ctx, req.OrderDate, row.App, identID, driver.ID)
			replacedCount++
		}

		// Record daily orders
		dailyOrder := domain.DailyOrder{
			ImportBatchID: batch.ID,
			OrderDate:     req.OrderDate,
			IdentifierID:  identID,
			DriverID:      driver.ID,
			AppName:       row.App,
			Branch:        row.Branch,
			OrdersCount:   row.TotalOrders,
			PlateNumber:   row.PlateNumber,
			Notes:         row.Notes,
		}
		ordersToInsert = append(ordersToInsert, dailyOrder)
		importedCount++
		totalOrders += row.TotalOrders
	}

	// Bulk insert orders
	if err := s.targetRepo.CreateDailyOrdersBatch(ctx, ordersToInsert); err != nil {
		return nil, fmt.Errorf("فشل في حفظ سجلات الطلبات في قاعدة البيانات: %w", err)
	}

	// Update batch totals
	batch.TotalOrders = totalOrders
	batch.IdentifiersCount = len(identCache)
	batch.DriversCount = len(driverCache)
	_ = s.targetRepo.UpdateImportBatch(ctx, &batch)

	// 2. Generate Daily Alerts for underperforming identifiers (< 17 orders)
	for identID, ident := range activeIdentifiersForAlerts {
		// Calculate total orders for this identifier on this order date
		dayOrders, err := s.targetRepo.GetDailyOrders(ctx, req.OrderDate, &identID)
		if err == nil {
			dayTotal := 0
			for _, ord := range dayOrders {
				dayTotal += ord.OrdersCount
			}
			targetReq := ident.DailyTarget
			if targetReq <= 0 {
				targetReq = 15
			}
			if dayTotal < targetReq {
				alert := domain.TargetAlert{
					IdentifierID: identID,
					AlertDate:    req.OrderDate,
					TargetOrders: targetReq,
					ActualOrders: dayTotal,
					Deficit:      targetReq - dayTotal,
					IsResolved:   false,
				}
				_ = s.targetRepo.CreateTargetAlert(ctx, &alert)
			}
		}
	}

	return &dto.ConfirmImportResponse{
		BatchID:             batch.ID,
		ImportedOrdersCount: importedCount,
		SkippedCount:        skippedCount,
		ReplacedCount:       replacedCount,
		TotalOrders:         totalOrders,
		Message:             fmt.Sprintf("تم استيراد %d سجل بنجاح بإجمالي %d طلب بتاريخ %s", importedCount, totalOrders, req.OrderDate),
	}, nil
}

// Helpers

func getCell(row []string, idx int) string {
	if idx >= 0 && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func parseInt(val string) int {
	if val == "" {
		return 0
	}
	// clean any commas, periods or Arabic numbers
	val = strings.ReplaceAll(val, ",", "")
	val = strings.TrimSpace(val)
	n, err := strconv.Atoi(val)
	if err != nil {
		// try float then round
		if f, err2 := strconv.ParseFloat(val, 64); err2 == nil {
			return int(f)
		}
		return 0
	}
	return n
}

// isFooterOrEmptyRow checks if row is summary/signature/empty footer
func isFooterOrEmptyRow(row []string) bool {
	if len(row) == 0 {
		return true
	}
	joined := strings.TrimSpace(strings.Join(row, " "))
	if joined == "" || joined == "`" {
		return true
	}
	// Known footers in company sheets
	lower := strings.ToLower(joined)
	if strings.Contains(lower, "توقيع المشرف") ||
		strings.Contains(lower, "مجموع الطلبات") ||
		strings.Contains(lower, "#ref!") ||
		strings.Contains(lower, "بيان يومي") {
		return true
	}
	return false
}

// extractDateFromFilename extracts YYYY-MM-DD from filenames like "3-9-2026.xlsx" or "2026-09-03.xlsx"
func extractDateFromFilename(filename string) string {
	base := filepath.Base(filename)

	// Format: D-M-YYYY or DD-MM-YYYY (e.g. 3-9-2026.xlsx)
	re1 := regexp.MustCompile(`(\d{1,2})[-_](\d{1,2})[-_](\d{4})`)
	if match := re1.FindStringSubmatch(base); len(match) == 4 {
		d, _ := strconv.Atoi(match[1])
		m, _ := strconv.Atoi(match[2])
		y, _ := strconv.Atoi(match[3])
		return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
	}

	// Format: YYYY-MM-DD or YYYY-M-D (e.g. 2026-09-03.xlsx)
	re2 := regexp.MustCompile(`(\d{4})[-_](\d{1,2})[-_](\d{1,2})`)
	if match := re2.FindStringSubmatch(base); len(match) == 4 {
		y, _ := strconv.Atoi(match[1])
		m, _ := strconv.Atoi(match[2])
		d, _ := strconv.Atoi(match[3])
		return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
	}

	return ""
}
