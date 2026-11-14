package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Branch model for multi-branch support
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

// Role model for RBAC
type Role struct {
	ID          uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	DisplayName string         `gorm:"type:varchar(150);not null" json:"display_name"`
	Description string         `gorm:"type:text" json:"description"`
	Permissions string         `gorm:"type:text" json:"permissions"` // JSON array string e.g. ["employees.view", "*"]
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

// Admin model for JWT Authentication
type Admin struct {
	ID           uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Email        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Username     string         `gorm:"type:varchar(50);uniqueIndex" json:"username"`
	Phone        string         `gorm:"type:varchar(20);uniqueIndex" json:"phone"`
	Password     string         `gorm:"type:varchar(255);not null" json:"-"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	Role         string         `gorm:"type:varchar(50);default:ADMIN" json:"role"`
	RoleID                *uuid.UUID     `gorm:"type:char(36);index" json:"role_id"`
	RoleObj               *Role          `gorm:"foreignKey:RoleID" json:"role_obj,omitempty"`
	EmployeeID            *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id,omitempty"`
	Employee              *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Permissions           string         `gorm:"type:text" json:"permissions"` // Optional user-specific permissions
	PermissionsConfigured bool           `gorm:"default:false" json:"permissions_configured"`
	MustChangePassword    bool           `gorm:"default:false" json:"must_change_password"`
	PushToken    string         `gorm:"type:text" json:"push_token"`
	GoogleID     string         `gorm:"type:varchar(100);index" json:"google_id,omitempty"`
	GoogleEmail  string         `gorm:"type:varchar(150);index" json:"google_email,omitempty"`
	GoogleAvatar string         `gorm:"type:varchar(255)" json:"google_avatar,omitempty"`
	Avatar       string         `gorm:"type:text" json:"avatar,omitempty"`
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

// Employee model
type Employee struct {
	ID                       uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Name                     string         `gorm:"type:varchar(150);not null;index" json:"name"`
	JobRole                  string         `gorm:"type:varchar(50);default:'DRIVER';index" json:"job_role"`
	EmployeeNumber           string         `gorm:"type:varchar(50);index" json:"employee_number"`
	Phone                    string         `gorm:"type:varchar(20);index" json:"phone"`
	PersonalImage            string         `gorm:"type:text" json:"personal_image"`
	NationalID               string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"national_id"`
	PasswordHash             string         `gorm:"type:varchar(255)" json:"-"`
	IqamaExpirationDate      *string        `gorm:"type:varchar(20)" json:"iqama_expiration_date"` // YYYY-MM-DD
	NationalIDImage          string         `gorm:"type:text" json:"national_id_image"`
	DrivingLicenseImage      string         `gorm:"type:text" json:"driving_license_image"`
	PassportImage            string         `gorm:"type:text" json:"passport_image"`
	VehicleRegistrationImage string         `gorm:"type:text" json:"vehicle_registration_image"`
	KeyNumber                string         `gorm:"type:varchar(50)" json:"key_number"`
	MotorcycleNumber         string         `gorm:"type:varchar(50)" json:"motorcycle_number"`
	ApplicationID            string         `gorm:"type:varchar(50);index" json:"application_id"`
	ApplicationType          string         `gorm:"type:varchar(50);index" json:"application_type"`
	VehicleType              string         `gorm:"type:varchar(20);default:'motorcycle'" json:"vehicle_type"` // "car" or "motorcycle"
	Shift                    string         `gorm:"type:varchar(20);default:'morning'" json:"shift"`           // "morning" or "evening"
	BranchID                 *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Branch                   *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Barcode                  string         `gorm:"type:text" json:"barcode"`
	QRCode                   string         `gorm:"type:text" json:"qr_code"`
	TotalDistance            float64        `gorm:"default:0" json:"total_distance"`
	LastOilChangeDistance    float64        `gorm:"default:0" json:"last_oil_change_distance"`
	Latitude                 *float64       `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude                *float64       `gorm:"type:decimal(11,8)" json:"longitude"`
	LastLocationAt           *time.Time     `json:"last_location_at"`
	IsVPN                    bool           `gorm:"default:false" json:"is_vpn"`
	IsMockLocation           bool           `gorm:"default:false" json:"is_mock_location"`
	OutOfZone                bool           `gorm:"default:false" json:"out_of_zone"`
	PushToken                string         `gorm:"type:text" json:"push_token"`
	DeviceUUID               string         `gorm:"type:varchar(100)" json:"device_uuid"`
	Language                 string         `gorm:"type:varchar(10);default:'ar'" json:"language"`
	MustChangePassword       bool           `gorm:"default:false" json:"must_change_password"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"-"`
	RoleID                   *uuid.UUID     `gorm:"type:char(36);index" json:"role_id"`
	RoleObj                  *Role          `gorm:"foreignKey:RoleID" json:"role_obj,omitempty"`
	RoleName                 string         `gorm:"type:varchar(50)" json:"role_name"`
	Permissions              string         `gorm:"type:text" json:"permissions"`
	WorkSessions []WorkSession `gorm:"foreignKey:EmployeeID;constraint:OnDelete:SET NULL" json:"work_sessions,omitempty"`
}

func (e *Employee) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

