package work

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	
)

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

func extractPlateDigits(plate string) string {
	var b strings.Builder
	for _, r := range plate {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else if r >= '٠' && r <= '٩' {
			b.WriteRune('0' + (r - '٠'))
		}
	}
	return b.String()
}

func (v *gormVehicleAdapter) GetVehicleInfo(ctx context.Context, plateNumber string) (*VehicleData, error) {
	clean := strings.TrimSpace(plateNumber)
	if clean == "" {
		return nil, errors.New("رقم اللوحة فارغ")
	}

	var row struct {
		ID                uuid.UUID  `gorm:"column:id"`
		PlateNumber       string     `gorm:"column:plate_number"`
		VehicleType       string     `gorm:"column:vehicle_type"`
		CurrentKM         float64    `gorm:"column:current_km"`
		LastOilChangeKM   float64    `gorm:"column:last_oil_change_km"`
		LastOilChangeDate *time.Time `gorm:"column:last_oil_change_date"`
		TotalDistance     float64    `gorm:"column:total_distance"`
		IsOdometerBroken  bool       `gorm:"column:is_odometer_broken"`
		RegistrationImage string     `gorm:"column:registration_image"`
		Status            string     `gorm:"column:status"`
	}

	err := v.db.WithContext(ctx).Table("vehicles").
		Where("deleted_at IS NULL AND (TRIM(plate_number) = ? OR plate_number ILIKE ? OR plate_number ILIKE ?)", clean, clean+"%", "%"+clean+"%").
		Order("CASE WHEN TRIM(plate_number) = '" + clean + "' THEN 0 ELSE 1 END").
		First(&row).Error
	if err != nil {
		digits := extractPlateDigits(clean)
		if len(digits) >= 3 {
			err = v.db.WithContext(ctx).Table("vehicles").
				Where("deleted_at IS NULL AND (TRIM(plate_number) = ? OR plate_number ILIKE ? OR plate_number ILIKE ?)", digits, digits+" %", digits+"%").
				First(&row).Error
		}
		if err != nil {
			return nil, err
		}
	}

	return &VehicleData{
		ID:                row.ID,
		PlateNumber:       row.PlateNumber,
		VehicleType:       row.VehicleType,
		CurrentKM:         row.CurrentKM,
		LastOilChangeKM:   row.LastOilChangeKM,
		LastOilChangeDate: row.LastOilChangeDate,
		TotalDistance:     row.TotalDistance,
		IsOdometerBroken:  row.IsOdometerBroken,
		RegistrationImage: row.RegistrationImage,
		Status:            row.Status,
	}, nil
}

func (v *gormVehicleAdapter) EnsureVehicleExists(ctx context.Context, plateNumber string, vehicleType string, branchID *uuid.UUID, startKM float64) (*VehicleData, error) {
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

	return &VehicleData{
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
