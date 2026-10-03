package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrLogNotFound     = errors.New("maintenance log not found")
	ErrRequestNotFound = errors.New("maintenance request not found")
)

// MaintenanceLog model for tracking oil changes and maintenance
type MaintenanceLog struct {
	ID         uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID *uuid.UUID     `gorm:"type:char(36);index;not null" json:"employee_id"`
	Type       string         `gorm:"type:varchar(50);not null;index" json:"type"` // "oil_change", "spare_part"
	Details    string         `gorm:"type:text" json:"details"`
	DistanceAt float64        `gorm:"not null" json:"distance_at"` // total distance reading at time of maintenance
	Cost       float64        `gorm:"default:0" json:"cost"`
	AdminName  string         `gorm:"type:varchar(100)" json:"admin_name"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (m *MaintenanceLog) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// MaintenanceRequest Model (طلبات صيانة المركبات)
type MaintenanceRequest struct {
	ID               uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	VehiclePlate     string         `gorm:"type:varchar(50);index;not null" json:"vehicle_plate"`
	EmployeeID       *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"`
	IssueDescription string         `gorm:"type:text;not null" json:"issue_description"`
	Priority         string         `gorm:"type:varchar(20);default:'MEDIUM';index" json:"priority"` // LOW, MEDIUM, HIGH, URGENT
	EstimatedCost    float64        `gorm:"default:0" json:"estimated_cost"`
	ActualCost       float64        `gorm:"default:0" json:"actual_cost"`
	WorkshopName     string         `gorm:"type:varchar(150)" json:"workshop_name"`
	Status           string         `gorm:"type:varchar(30);default:'OPEN';index" json:"status"` // OPEN, IN_PROGRESS, RESOLVED, CLOSED
	BranchID         *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Notes            string         `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (m *MaintenanceRequest) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
