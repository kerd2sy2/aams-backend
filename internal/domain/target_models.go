package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Identifier model represents a company account/identifier (e.g. "فهد", "فردين", "هريدي")
type Identifier struct {
	ID            uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Name          string         `gorm:"type:varchar(100);index;not null" json:"name"`
	AppName       string         `gorm:"type:varchar(50);index;default:''" json:"app_name"`
	Branch        string         `gorm:"type:varchar(50);index;default:''" json:"branch,omitempty"`
	Code          string         `gorm:"type:varchar(50);index" json:"code,omitempty"`
	MonthlyTarget int            `gorm:"default:460" json:"monthly_target"`
	DailyTarget   int            `gorm:"default:18" json:"daily_target"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	Drivers       []Driver       `gorm:"many2many:identifier_drivers;" json:"drivers,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (i *Identifier) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	if i.MonthlyTarget <= 0 {
		i.MonthlyTarget = 460
	}
	if i.DailyTarget <= 0 {
		i.DailyTarget = 15
	}
	return nil
}

// Driver model represents a driver who works under identifiers (e.g. "لابون شندر", "مومن جمان")
type Driver struct {
	ID         uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Name       string         `gorm:"type:varchar(150);uniqueIndex;not null" json:"name"`
	EmployeeID *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id,omitempty"`
	Phone      string         `gorm:"type:varchar(20)" json:"phone,omitempty"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (d *Driver) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// IdentifierDriver maps the 1-to-many relationship between Identifier and Drivers
type IdentifierDriver struct {
	ID            uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	IdentifierID  uuid.UUID `gorm:"type:char(36);index;not null" json:"identifier_id"`
	DriverID      uuid.UUID `gorm:"type:char(36);index;not null" json:"driver_id"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	FirstSeenDate string    `gorm:"type:varchar(10);index" json:"first_seen_date,omitempty"` // YYYY-MM-DD
	LastSeenDate  string    `gorm:"type:varchar(10);index" json:"last_seen_date,omitempty"`  // YYYY-MM-DD
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (id *IdentifierDriver) BeforeCreate(tx *gorm.DB) error {
	if id.ID == uuid.Nil {
		id.ID = uuid.New()
	}
	return nil
}

// ImportBatch tracks each Excel upload action
type ImportBatch struct {
	ID               uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	FileName         string    `gorm:"type:varchar(255);not null" json:"file_name"`
	OrderDate        string    `gorm:"type:varchar(10);index;not null" json:"order_date"` // YYYY-MM-DD
	UploadedBy       uuid.UUID `gorm:"type:char(36);index" json:"uploaded_by"`
	UploadedByName   string    `gorm:"type:varchar(100)" json:"uploaded_by_name"`
	TotalRows        int       `gorm:"default:0" json:"total_rows"`
	TotalOrders      int       `gorm:"default:0" json:"total_orders"`
	IdentifiersCount int       `gorm:"default:0" json:"identifiers_count"`
	DriversCount     int       `gorm:"default:0" json:"drivers_count"`
	Status           string    `gorm:"type:varchar(20);default:'COMPLETED'" json:"status"` // PREVIEW, COMPLETED, CANCELLED
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (b *ImportBatch) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// DailyOrder stores individual order records parsed from daily Excel files
type DailyOrder struct {
	ID            uuid.UUID   `gorm:"type:char(36);primary_key" json:"id"`
	ImportBatchID uuid.UUID   `gorm:"type:char(36);index" json:"import_batch_id"`
	OrderDate     string      `gorm:"type:varchar(10);index;not null" json:"order_date"` // YYYY-MM-DD
	IdentifierID  *uuid.UUID  `gorm:"type:char(36);index" json:"identifier_id,omitempty"`
	DriverID      uuid.UUID   `gorm:"type:char(36);index;not null" json:"driver_id"`
	AppName       string      `gorm:"type:varchar(50);index" json:"app_name"` // نينجا، كيتا، تويو، إلخ
	Branch        string      `gorm:"type:varchar(50);index;default:''" json:"branch,omitempty"` // الفرع 1، الفرع 2، إلخ
	OrdersCount   int         `gorm:"not null" json:"orders_count"`
	PlateNumber   string      `gorm:"type:varchar(50)" json:"plate_number,omitempty"`
	Notes         string      `gorm:"type:text" json:"notes,omitempty"`
	Identifier    *Identifier `gorm:"foreignKey:IdentifierID" json:"identifier,omitempty"`
	Driver        *Driver     `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

func (o *DailyOrder) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

// TargetAlert stores daily deficit/underperformance notifications
type TargetAlert struct {
	ID           uuid.UUID   `gorm:"type:char(36);primary_key" json:"id"`
	IdentifierID uuid.UUID   `gorm:"type:char(36);index;not null" json:"identifier_id"`
	AlertDate    string      `gorm:"type:varchar(10);index;not null" json:"alert_date"` // YYYY-MM-DD
	TargetOrders int         `gorm:"default:17" json:"target_orders"`
	ActualOrders int         `gorm:"default:0" json:"actual_orders"`
	Deficit      int         `gorm:"default:0" json:"deficit"`
	IsResolved   bool        `gorm:"default:false" json:"is_resolved"`
	Identifier   *Identifier `gorm:"foreignKey:IdentifierID" json:"identifier,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

func (a *TargetAlert) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// TargetSetting represents configurable target thresholds
type TargetSetting struct {
	ID           uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	SettingKey   string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"setting_key"`
	SettingValue string    `gorm:"type:varchar(100);not null" json:"setting_value"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *TargetSetting) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
