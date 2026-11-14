package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Session status constants
const (
	StatusActive    = "ACTIVE"
	StatusCompleted = "COMPLETED"
)

// WorkSession model
type WorkSession struct {
	ID                   uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID           *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"`
	Employee             *Employee      `gorm:"foreignKey:EmployeeID;constraint:OnDelete:SET NULL" json:"employee,omitempty"`
	StartTime            time.Time      `gorm:"not null" json:"start_time"`
	EndTime              *time.Time     `json:"end_time"`
	StartKM              float64        `gorm:"not null" json:"start_km"`
	EndKM                float64        `gorm:"default:0" json:"end_km"`
	Distance             float64        `gorm:"default:0" json:"distance"`
	OrdersCount          int            `gorm:"default:0" json:"orders_count"`
	FuelCost             float64        `gorm:"default:0" json:"fuel_cost"`
	ApplicationID        string         `gorm:"type:varchar(50)" json:"application_id"`
	ApplicationType      string         `gorm:"type:varchar(50)" json:"application_type"`
	VehicleType          string         `gorm:"type:varchar(20)" json:"vehicle_type"`               // override for this shift: "car" or "motorcycle"
	MotorcycleNumber     string         `gorm:"type:varchar(50)" json:"motorcycle_number"`          // رقم الدباب لهذا الشفت (قد يختلف عن المسجل)
	StartPlateImage      string         `gorm:"type:text" json:"start_plate_image"`                 // صورة لوحة الدباب الموثقة بالكاميرا عند البدء
	StartKMImage         string         `gorm:"type:text" json:"start_km_image"`                    // صورة عداد البداية
	EndKMImage           string         `gorm:"type:text" json:"end_km_image"`                      // صورة عداد النهاية
	IsReviewed           bool           `gorm:"default:false;index" json:"is_reviewed"`             // حالة مراجعة وتصديق المشرف
	ReviewNotes          string         `gorm:"type:text" json:"review_notes"`                      // ملاحظات المشرف
	ReviewedBy           *uuid.UUID     `gorm:"type:char(36)" json:"reviewed_by"`                   // المشرف المراجع
	ReviewVideoURL       string         `gorm:"type:text" json:"review_video_url"`                   // رابط فيديو التوثيق الصامت للمراجعة والطلبات
	IsEditedBySupervisor bool           `gorm:"default:false;index" json:"is_edited_by_supervisor"` // هل تم تعديل البيانات بواسطة المشرف
	EditedByName         string         `gorm:"type:varchar(100)" json:"edited_by_name"`            // اسم المشرف الذي قام بالتعديل
	OriginalOrdersCount  int            `gorm:"default:0" json:"original_orders_count"`             // عدد الطلبات الأصلي المدخل من المندوب
	OriginalEndKM        float64        `gorm:"default:0" json:"original_end_km"`                   // عداد النهاية الأصلي المدخل من المندوب
	OriginalStartKM      float64        `gorm:"default:0" json:"original_start_km"`                 // عداد البداية الأصلي المدخل من المندوب
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

// AuditLog model
type AuditLog struct {
	ID        uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	AdminName string     `gorm:"type:varchar(100);not null" json:"admin_name"`
	Action    string     `gorm:"type:varchar(100);not null;index" json:"action"`
	Details   string     `gorm:"type:text" json:"details"`
	IPAddress string     `gorm:"type:varchar(50)" json:"ip_address"`
	BranchID  *uuid.UUID `gorm:"type:char(36);index" json:"branch_id"`
	CreatedAt time.Time  `gorm:"index" json:"created_at"`
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}


// MaintenanceLog model for tracking oil changes and maintenance
type MaintenanceLog struct {
	ID         uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID *uuid.UUID `gorm:"type:char(36);index;not null" json:"employee_id"`
	Employee   *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Type       string     `gorm:"type:varchar(50);not null;index" json:"type"` // "oil_change", "spare_part"
	Details    string     `gorm:"type:text" json:"details"`
	DistanceAt float64    `gorm:"not null" json:"distance_at"` // total distance reading at time of maintenance
	Cost       float64    `gorm:"default:0" json:"cost"`
	AdminName  string     `gorm:"type:varchar(100)" json:"admin_name"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (m *MaintenanceLog) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}


// Vehicle status constants
const (
	VehicleStatusAvailable   = "AVAILABLE"
	VehicleStatusInUse       = "IN_USE"
	VehicleStatusMaintenance = "MAINTENANCE"
)

// Vehicle model for motorcycle & car assets (ط§ظ„ط«ظˆط§ط¨طھ / ط§ظ„ط¯ط¨ط§ط¨ط§طھ)
type Vehicle struct {
	ID               uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	PlateNumber      string         `gorm:"type:varchar(50);index:idx_vehicles_plate_number,unique;not null" json:"plate_number"` // رقم اللوحة / الدباب
	VehicleType      string         `gorm:"type:varchar(20);default:'motorcycle'" json:"vehicle_type"` // "motorcycle" or "car"
	Brand            string         `gorm:"type:varchar(100)" json:"brand"`                            // ظ…ط§ط±ظƒط© ط§ظ„ط¯ط¨ط§ط¨ (ظ‡ظˆظ†ط¯ط§طŒ ط³ظˆط²ظˆظƒظٹ...)
	ModelYear        string         `gorm:"type:varchar(20)" json:"model_year"`                        // ط³ظ†ط© ط§ظ„طµظ†ط¹
	KeyNumber        string         `gorm:"type:varchar(50)" json:"key_number"`                        // ط±ظ‚ظ… ط§ظ„ظ…ظپطھط§ط­ ط§ظ„ظ…ط±طھط¨ط·
	CurrentKM        float64        `gorm:"default:0" json:"current_km"`                               // ط§ظ„ط¹ط¯ط§ط¯ ط§ظ„ط­ط§ظ„ظٹ ط§ظ„ظ…ط³ط¬ظ„
	LastOilChangeKM  float64        `gorm:"default:0" json:"last_oil_change_km"`                       // ظ‚ط±ط§ط،ط© ط§ظ„ط¹ط¯ط§ط¯ ط¹ظ†ط¯ ط¢ط®ط± طھط؛ظٹظٹط± ط²ظٹطھ
	LastOilChangeDate *time.Time    `json:"last_oil_change_date"`                                      // تاريخ آخر تغيير زيت
	TotalDistance    float64        `gorm:"default:0" json:"total_distance"`                           // ط¥ط¬ظ…ط§ظ„ظٹ ط§ظ„ظƒظٹظ„ظˆظ…طھط±ط§طھ ط§ظ„ظ…ظ‚ط·ظˆط¹ط©
	Status           string         `gorm:"type:varchar(20);default:'AVAILABLE';index" json:"status"`  // "AVAILABLE", "IN_USE", "MAINTENANCE"
	IsOdometerBroken bool           `gorm:"default:false;index" json:"is_odometer_broken"`             // عداد المسافات تالف / معطل
	RegistrationImage string        `gorm:"type:text" json:"registration_image"`                       // صورة استمارة المركبة / الدباب
	BranchID         *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Branch           *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Notes            string         `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (v *Vehicle) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

// ------------------------------------------------------------------
// 1. FuelLog Model (ط³ط¬ظ„ط§طھ ط§ظ„ظˆظ‚ظˆط¯)
// ------------------------------------------------------------------
type FuelLog struct {
	ID              uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID      *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"`
	Employee        *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	VehiclePlate    string         `gorm:"type:varchar(50);index" json:"vehicle_plate"`
	ShiftID         *uuid.UUID     `gorm:"type:char(36);index" json:"shift_id"`
	Amount          float64        `gorm:"not null;default:0" json:"amount"` // ط§ظ„ظ…ط¨ظ„ط؛ ط¨ط§ظ„ط±ظٹط§ظ„
	Liters          float64        `gorm:"default:0" json:"liters"`          // ط¹ط¯ط¯ ط§ظ„ظ„طھط±ط§طھ
	FuelDate        time.Time      `gorm:"index" json:"fuel_date"`
	StationName     string         `gorm:"type:varchar(150)" json:"station_name"`
	InvoiceImageURL string         `gorm:"type:varchar(500)" json:"invoice_image_url"`
	BranchID        *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Branch          *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Notes           string         `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (f *FuelLog) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// ------------------------------------------------------------------
// 2. TrafficViolation Model (ط§ظ„ظ…ط®ط§ظ„ظپط§طھ ط§ظ„ظ…ط±ظˆط±ظٹط©)
// ------------------------------------------------------------------
type TrafficViolation struct {
	ID              uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	ViolationNumber string         `gorm:"type:varchar(100);index" json:"violation_number"`
	EmployeeID      *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"`
	Employee        *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	VehiclePlate    string         `gorm:"type:varchar(50);index" json:"vehicle_plate"`
	Amount          float64        `gorm:"not null;default:0" json:"amount"`
	PaidAmount      float64        `gorm:"not null;default:0" json:"paid_amount"`
	Reason          string         `gorm:"type:varchar(255);not null" json:"reason"` // سرعة، إشارة، حزام...
	ViolationDate   time.Time      `gorm:"index" json:"violation_date"`
	City            string         `gorm:"type:varchar(100)" json:"city"`
	Status          string         `gorm:"type:varchar(30);default:'RECORDED';index" json:"status"` // RECORDED, PARTIAL, DEDUCTED, DISPUTED, PAID
	BranchID        *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Branch          *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Notes           string         `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (v *TrafficViolation) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

// ------------------------------------------------------------------
// 3. MaintenanceRequest Model (ط·ظ„ط¨ط§طھ طµظٹط§ظ†ط© ط§ظ„ظ…ط±ظƒط¨ط§طھ)
// ------------------------------------------------------------------
type MaintenanceRequest struct {
	ID               uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	VehiclePlate     string         `gorm:"type:varchar(50);index;not null" json:"vehicle_plate"`
	EmployeeID       *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"`
	Employee         *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	IssueDescription string         `gorm:"type:text;not null" json:"issue_description"`
	Priority         string         `gorm:"type:varchar(20);default:'MEDIUM';index" json:"priority"` // LOW, MEDIUM, HIGH, URGENT
	EstimatedCost    float64        `gorm:"default:0" json:"estimated_cost"`
	ActualCost       float64        `gorm:"default:0" json:"actual_cost"`
	WorkshopName     string         `gorm:"type:varchar(150)" json:"workshop_name"`
	Status           string         `gorm:"type:varchar(30);default:'OPEN';index" json:"status"` // OPEN, IN_PROGRESS, RESOLVED, CLOSED
	BranchID         *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Branch           *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
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

