package work

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	
)

type gormEmployeeAdapter struct {
	db *gorm.DB
}

func (a *gormEmployeeAdapter) GetEmployee(ctx context.Context, id uuid.UUID) (*EmployeeData, error) {
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

	return &EmployeeData{
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
	if cleanBike := strings.TrimSpace(motorcycleNumber); cleanBike != "" {
		updates["motorcycle_number"] = cleanBike
	}
	return a.db.WithContext(ctx).Table("employees").Where("id = ?", id).Updates(updates).Error
}

func (a *gormEmployeeAdapter) UpdateEmployeeOnEndWork(ctx context.Context, id uuid.UUID, addedDistance float64, totalOrders int) error {
	var activeCount int64
	_ = a.db.WithContext(ctx).Table("work_sessions").
		Where("employee_id = ? AND status = 'ACTIVE' AND deleted_at IS NULL", id).
		Count(&activeCount).Error

	updates := map[string]interface{}{
		"total_distance": gorm.Expr("total_distance + ?", addedDistance),
		"total_orders":   gorm.Expr("total_orders + ?", totalOrders),
		"is_working":     false,
	}

	return a.db.WithContext(ctx).Table("employees").Where("id = ?", id).Updates(updates).Error
}
