package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Branch model
type Branch struct {
	ID        uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *Branch) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// Role model
type Role struct {
	ID          uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	DisplayName string         `gorm:"type:varchar(150);not null" json:"display_name"`
	Description string         `gorm:"type:text" json:"description"`
	Permissions string         `gorm:"type:text" json:"permissions"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// Admin model
type Admin struct {
	ID           uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Email        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Username     string         `gorm:"type:varchar(50);uniqueIndex" json:"username"`
	Phone        string         `gorm:"type:varchar(20);uniqueIndex" json:"phone"`
	Password     string         `gorm:"type:varchar(255);not null" json:"-"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	Role         string         `gorm:"type:varchar(50);default:ADMIN" json:"role"`
	RoleID       *uuid.UUID     `gorm:"type:char(36);index" json:"role_id"`
	RoleObj      *Role          `gorm:"foreignKey:RoleID" json:"role_obj,omitempty"`
	Permissions  string         `gorm:"type:text" json:"permissions"`
	GoogleID     string         `gorm:"type:varchar(100);index" json:"google_id,omitempty"`
	GoogleEmail  string         `gorm:"type:varchar(150);index" json:"google_email,omitempty"`
	GoogleAvatar string         `gorm:"type:varchar(255)" json:"google_avatar,omitempty"`
	BranchID     *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Branch       *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *Admin) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// OTPRequest model
type OTPRequest struct {
	ID           uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	NationalID   string     `gorm:"type:varchar(50);not null;index" json:"national_id"`
	EmployeeID   *uuid.UUID `gorm:"type:char(36);index" json:"employee_id,omitempty"`
	EmployeeName string     `gorm:"type:varchar(150)" json:"employee_name"`
	BranchID     *uuid.UUID `gorm:"type:char(36);index" json:"branch_id,omitempty"`
	BranchName   string     `gorm:"type:varchar(100)" json:"branch_name"`
	OTPCode      string     `gorm:"type:varchar(10);not null" json:"otp_code"`
	DeviceInfo   string     `gorm:"type:varchar(255)" json:"device_info"`
	DeviceUUID   string     `gorm:"type:varchar(100);index" json:"device_uuid"`
	Status       string     `gorm:"type:varchar(20);default:'pending';index" json:"status"` // pending, approved, verified, expired, rejected
	ExpiresAt    time.Time  `gorm:"not null" json:"expires_at"`
	CreatedAt    time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (o *OTPRequest) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
