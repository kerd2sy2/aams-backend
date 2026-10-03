package work

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/work/controller"
	"delivery-backend/internal/modules/work/repository"
	"delivery-backend/internal/modules/work/service"
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
		updates["motorcycle_number"] = motorcycleNumber
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

type gormVehicleAdapter struct {
	db *gorm.DB
}

func (v *gormVehicleAdapter) GetVehicleLastKM(ctx context.Context, plateNumber string) (float64, error) {
	var km float64
	err := v.db.WithContext(ctx).Table("vehicles").
		Select("COALESCE(current_km, 0)").
		Where("plate_number = ? AND deleted_at IS NULL", plateNumber).
		Scan(&km).Error
	return km, err
}

func (v *gormVehicleAdapter) UpdateVehicleKM(ctx context.Context, plateNumber string, km float64) error {
	return v.db.WithContext(ctx).Table("vehicles").
		Where("plate_number = ? AND deleted_at IS NULL AND current_km < ?", plateNumber, km).
		Update("current_km", km).Error
}

func (v *gormVehicleAdapter) HasActiveMaintenance(ctx context.Context, motorcycleNumber string) (bool, error) {
	var count int64
	err := v.db.WithContext(ctx).Table("maintenance_requests").
		Where("motorcycle_number = ? AND status IN ?", motorcycleNumber, []string{"pending", "in_progress"}).
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
