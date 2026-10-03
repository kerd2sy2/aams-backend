package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkSession struct {
	ID                   uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID           *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"`
	StartTime            time.Time      `gorm:"not null" json:"start_time"`
	EndTime              *time.Time     `json:"end_time"`
	StartKM              float64        `gorm:"not null" json:"start_km"`
	EndKM                float64        `gorm:"default:0" json:"end_km"`
	Distance             float64        `gorm:"default:0" json:"distance"`
	OrdersCount          int            `gorm:"default:0" json:"orders_count"`
	FuelCost             float64        `gorm:"default:0" json:"fuel_cost"`
	ApplicationID        string         `gorm:"type:varchar(50)" json:"application_id"`
	ApplicationType      string         `gorm:"type:varchar(50)" json:"application_type"`
	VehicleType          string         `gorm:"type:varchar(20)" json:"vehicle_type"`
	MotorcycleNumber     string         `gorm:"type:varchar(50)" json:"motorcycle_number"`
	StartPlateImage      string         `gorm:"type:text" json:"start_plate_image"`
	StartKMImage         string         `gorm:"type:text" json:"start_km_image"`
	EndKMImage           string         `gorm:"type:text" json:"end_km_image"`
	IsReviewed           bool           `gorm:"default:false;index" json:"is_reviewed"`
	ReviewNotes          string         `gorm:"type:text" json:"review_notes"`
	ReviewedBy           *uuid.UUID     `gorm:"type:char(36)" json:"reviewed_by"`
	IsEditedBySupervisor bool           `gorm:"default:false;index" json:"is_edited_by_supervisor"`
	EditedByName         string         `gorm:"type:varchar(100)" json:"edited_by_name"`
	OriginalOrdersCount  int            `gorm:"default:0" json:"original_orders_count"`
	OriginalEndKM        float64        `gorm:"default:0" json:"original_end_km"`
	OriginalStartKM      float64        `gorm:"default:0" json:"original_start_km"`
	Notes                string         `gorm:"type:text" json:"notes"`
	Status               string         `gorm:"type:varchar(20);default:ACTIVE;index" json:"status"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (w *WorkSession) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}
