package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("vehicle not found")
)

const (
	VehicleStatusAvailable   = "AVAILABLE"
	VehicleStatusInUse       = "IN_USE"
	VehicleStatusMaintenance = "MAINTENANCE"
)

// Vehicle model for motorcycle & car assets
type Vehicle struct {
	ID                uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	PlateNumber       string         `gorm:"type:varchar(50);index:idx_vehicles_plate_number,unique;not null" json:"plate_number"` // رقم اللوحة / الدباب
	VehicleType       string         `gorm:"type:varchar(20);default:'motorcycle'" json:"vehicle_type"`                            // "motorcycle" or "car"
	Brand             string         `gorm:"type:varchar(100)" json:"brand"`                                                       // ماركة الدباب
	ModelYear         string         `gorm:"type:varchar(20)" json:"model_year"`                                                   // سنة الصنع
	KeyNumber         string         `gorm:"type:varchar(50)" json:"key_number"`                                                   // رقم المفتاح المرتبط
	CurrentKM         float64        `gorm:"default:0" json:"current_km"`                                                          // العداد الحالي المسجل
	LastOilChangeKM   float64        `gorm:"default:0" json:"last_oil_change_km"`                                                  // قراءة العداد عند آخر تغيير زيت
	TotalDistance     float64        `gorm:"default:0" json:"total_distance"`                                                      // إجمالي الكيلومترات المقطوعة
	Status            string         `gorm:"type:varchar(20);default:'AVAILABLE';index" json:"status"`                             // "AVAILABLE", "IN_USE", "MAINTENANCE"
	IsOdometerBroken  bool           `gorm:"default:false;index" json:"is_odometer_broken"`                                        // عداد المسافات تالف / معطل
	RegistrationImage string         `gorm:"type:text" json:"registration_image"`                                                  // صورة استمارة المركبة / الدباب
	BranchID          *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Notes             string         `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (v *Vehicle) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}
