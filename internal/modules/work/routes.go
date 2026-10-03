package work

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/work/controller"
	"delivery-backend/internal/modules/work/repository"
	"delivery-backend/internal/modules/work/service"
	legacyService "delivery-backend/internal/service"
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

func (a *gormEmployeeAdapter) SendShiftApprovalNotification(ctx context.Context, empID uuid.UUID, sessionID uuid.UUID, ordersCount int, fuelCost float64) error {
	var emp struct {
		Name      string     `gorm:"column:name"`
		Language  string     `gorm:"column:language"`
		PushToken string     `gorm:"column:push_token"`
		BranchID  *uuid.UUID `gorm:"column:branch_id"`
	}
	if err := a.db.WithContext(ctx).Table("employees").Where("id = ? AND deleted_at IS NULL", empID).First(&emp).Error; err != nil {
		return err
	}

	lang := strings.ToLower(strings.TrimSpace(emp.Language))
	var title, body string
	switch lang {
	case "en":
		title = "Shift Approved ✅"
		body = fmt.Sprintf("Supervisor approved your shift: %d Orders | Fuel: %.2f SAR", ordersCount, fuelCost)
	case "bn":
		title = "শিফট অনুমোদিত হয়েছে ✅"
		body = fmt.Sprintf("সুপারভাইজার আপনার শিফট অনুমোদন করেছেন: %d টি অর্ডার | জ্বালানী: %.2f SAR", ordersCount, fuelCost)
	case "ur":
		title = "شفت کی تصدیق ہو گئی ✅"
		body = fmt.Sprintf("نگران نے آپ کے آرڈرز کی تصدیق کر دی: %d آرڈرز | ایندھن: %.2f ريال", ordersCount, fuelCost)
	default: // "ar"
		title = "تمت المصادقة على شفت العمل ✅"
		body = fmt.Sprintf("وافق المشرف على طلباتك: %d طلب | بنزين: %.2f ريال", ordersCount, fuelCost)
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
			"type":        "SESSION_APPROVED",
			"sessionId":   sid,
			"session_id":  sid,
			"ordersCount": fmt.Sprintf("%d", ordersCount),
			"fuelCost":    fmt.Sprintf("%.2f", fuelCost),
		})
	}
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

func (v *gormVehicleAdapter) UpdateVehicleKM(ctx context.Context, plateNumber string, km float64) error {
	clean := strings.TrimSpace(plateNumber)
	if clean == "" {
		return nil
	}
	var vehicle struct {
		ID        uuid.UUID `gorm:"column:id"`
		CurrentKM float64   `gorm:"column:current_km"`
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
		updates := map[string]interface{}{}
		if km > vehicle.CurrentKM {
			updates["current_km"] = km
		}
		if delta > 0 {
			updates["total_distance"] = gorm.Expr("COALESCE(total_distance, 0) + ?", delta)
		}
		if len(updates) > 0 {
			return v.db.WithContext(ctx).Table("vehicles").
				Where("id = ?", vehicle.ID).
				Updates(updates).Error
		}
	}
	return nil
}

func (v *gormVehicleAdapter) HasActiveMaintenance(ctx context.Context, motorcycleNumber string) (bool, error) {
	var count int64
	clean := strings.TrimSpace(motorcycleNumber)
	err := v.db.WithContext(ctx).Table("maintenance_requests").
		Where("(TRIM(motorcycle_number) = ? OR motorcycle_number ILIKE ?) AND status IN ?", clean, clean+"%", []string{"pending", "in_progress", "OPEN"}).
		Count(&count).Error
	return count > 0, err
}

type Module struct {
	Service service.WorkService
	Handler *controller.WorkHandler
}

func NewModule(db *gorm.DB, storage service.StorageProvider) *Module {
	repo := repository.NewWorkRepository(db)
	empAdapter := &gormEmployeeAdapter{db: db}
	vehicleAdapter := &gormVehicleAdapter{db: db}
	svc := service.NewWorkService(repo, empAdapter, vehicleAdapter, storage)
	handler := controller.NewWorkHandler(svc)
	return &Module{
		Service: svc,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/work/start", m.Handler.StartWork)
	rg.POST("/work/end", m.Handler.EndWork)
	rg.PUT("/work/:id", m.Handler.UpdateWorkSession)
	rg.PUT("/work/:id/review", m.Handler.ReviewWorkSession)
	rg.PUT("/work/sessions/:id/review", m.Handler.ReviewWorkSession)
	rg.GET("/work/sessions/:id", m.Handler.GetSessionByID)
	rg.GET("/work/active", m.Handler.GetActiveSession)
	rg.GET("/work/last-km", m.Handler.GetLastKM)
	rg.POST("/work/scan-plate", m.Handler.ScanPlate)
	rg.GET("/work/today-count", m.Handler.TodayCount)
	rg.GET("/work/check-oil", m.Handler.CheckOilChange)
}
