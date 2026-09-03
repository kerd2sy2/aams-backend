package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OTPRequest represents a 4-digit one-time password request for employee login/device verification
type OTPRequest struct {
	ID           uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID   uuid.UUID      `gorm:"type:char(36);index;not null" json:"employee_id"`
	Employee     *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	NationalID   string         `gorm:"type:varchar(50);index;not null" json:"national_id"`
	EmployeeName string         `gorm:"type:varchar(150);not null" json:"employee_name"`
	OTPCode      string         `gorm:"type:varchar(10);not null" json:"otp_code"` // 4-digit code e.g. "4829"
	DeviceInfo   string         `gorm:"type:varchar(255)" json:"device_info"`
	DeviceUUID   string         `gorm:"type:varchar(100);index" json:"device_uuid"`
	Status       string         `gorm:"type:varchar(20);default:'PENDING';index" json:"status"` // PENDING, VERIFIED, EXPIRED, CANCELLED
	ExpiresAt    time.Time      `gorm:"index;not null" json:"expires_at"`
	VerifiedAt   *time.Time     `json:"verified_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (o *OTPRequest) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
