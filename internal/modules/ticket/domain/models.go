package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("support ticket not found")
)

// SupportTicket Model for internal ticketing and complaints
type SupportTicket struct {
	ID           uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	TicketNumber string         `gorm:"type:varchar(50);index;not null" json:"ticket_number"`
	EmployeeID   *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"`
	Subject      string         `gorm:"type:varchar(255);not null" json:"subject"`
	Category     string         `gorm:"type:varchar(50);default:'OPERATIONAL';index" json:"category"` // OPERATIONAL, FINANCIAL, VEHICLE, APPLICATION, OTHER
	Priority     string         `gorm:"type:varchar(20);default:'MEDIUM';index" json:"priority"`      // LOW, MEDIUM, HIGH, URGENT
	Status       string         `gorm:"type:varchar(30);default:'OPEN';index" json:"status"`          // OPEN, IN_PROGRESS, RESOLVED, CLOSED
	Description  string         `gorm:"type:text;not null" json:"description"`
	Resolution   string         `gorm:"type:text" json:"resolution"`
	BranchID     *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *SupportTicket) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
