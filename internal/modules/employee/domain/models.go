package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Employee struct {
	ID                       uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Name                     string         `gorm:"type:varchar(150);not null;index" json:"name"`
	JobRole                  string         `gorm:"type:varchar(50);default:'DRIVER';index" json:"job_role"`
	EmployeeNumber           string         `gorm:"type:varchar(50);index" json:"employee_number"`
	Phone                    string         `gorm:"type:varchar(20);index" json:"phone"`
	PersonalImage            string         `gorm:"type:text" json:"personal_image"`
	NationalID               string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"national_id"`
	PasswordHash             string         `gorm:"type:varchar(255)" json:"-"`
	IqamaExpirationDate      *string        `gorm:"type:varchar(20)" json:"iqama_expiration_date"`
	NationalIDImage          string         `gorm:"type:text" json:"national_id_image"`
	DrivingLicenseImage      string         `gorm:"type:text" json:"driving_license_image"`
	PassportImage            string         `gorm:"type:text" json:"passport_image"`
	VehicleRegistrationImage string         `gorm:"type:text" json:"vehicle_registration_image"`
	KeyNumber                string         `gorm:"type:varchar(50)" json:"key_number"`
	MotorcycleNumber         string         `gorm:"type:varchar(50)" json:"motorcycle_number"`
	ApplicationID            string         `gorm:"type:varchar(50);index" json:"application_id"`
	ApplicationType          string         `gorm:"type:varchar(50);index" json:"application_type"`
	VehicleType              string         `gorm:"type:varchar(20);default:'motorcycle'" json:"vehicle_type"`
	Shift                    string         `gorm:"type:varchar(20);default:'morning'" json:"shift"`
	BranchID                 *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
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
	IsWorking                bool           `gorm:"default:false" json:"is_working"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"-"`
}

func (e *Employee) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
