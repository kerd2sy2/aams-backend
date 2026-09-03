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
	ID          uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Email       string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Username    string         `gorm:"type:varchar(50);uniqueIndex" json:"username"`
	Phone       string         `gorm:"type:varchar(20);uniqueIndex" json:"phone"`
	Password    string         `gorm:"type:varchar(255);not null" json:"-"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Role        string         `gorm:"type:varchar(50);default:ADMIN" json:"role"`
	RoleID      *uuid.UUID     `gorm:"type:char(36);index" json:"role_id"`
	RoleObj     *Role          `gorm:"foreignKey:RoleID" json:"role_obj,omitempty"`
	Permissions string         `gorm:"type:text" json:"permissions"` // Optional user-specific permissions
	GoogleID    string         `gorm:"type:varchar(100);index" json:"google_id,omitempty"`
	GoogleEmail string         `gorm:"type:varchar(150);index" json:"google_email,omitempty"`
	GoogleAvatar string        `gorm:"type:varchar(255)" json:"google_avatar,omitempty"`
	BranchID    *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Branch      *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *Admin) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}


// Employee model
type Employee struct {
	ID                    uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Name                  string         `gorm:"type:varchar(150);not null;index" json:"name"`
	JobRole               string         `gorm:"type:varchar(50);default:'DRIVER';index" json:"job_role"`
	EmployeeNumber        string         `gorm:"type:varchar(50);index" json:"employee_number"`
	PersonalImage         string         `gorm:"type:text" json:"personal_image"`
	NationalID            string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"national_id"`
	PasswordHash          string         `gorm:"type:varchar(255)" json:"-"`
	IqamaExpirationDate   *string        `gorm:"type:varchar(20)" json:"iqama_expiration_date"` // YYYY-MM-DD
	NationalIDImage           string         `gorm:"type:text" json:"national_id_image"`
	DrivingLicenseImage       string         `gorm:"type:text" json:"driving_license_image"`
	PassportImage             string         `gorm:"type:text" json:"passport_image"`
	VehicleRegistrationImage  string         `gorm:"type:text" json:"vehicle_registration_image"`
	KeyNumber             string         `gorm:"type:varchar(50)" json:"key_number"`
	MotorcycleNumber      string         `gorm:"type:varchar(50)" json:"motorcycle_number"`
	ApplicationID         string         `gorm:"type:varchar(50);index" json:"application_id"`
	ApplicationType       string         `gorm:"type:varchar(50);index" json:"application_type"`
	VehicleType           string         `gorm:"type:varchar(20);default:'motorcycle'" json:"vehicle_type"` // "car" or "motorcycle"
	Shift                 string         `gorm:"type:varchar(20);default:'morning'" json:"shift"`           // "morning" or "evening"
	BranchID              *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Branch                *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Barcode               string         `gorm:"type:text" json:"barcode"`
	QRCode                string         `gorm:"type:text" json:"qr_code"`
	TotalDistance         float64        `gorm:"default:0" json:"total_distance"`
	LastOilChangeDistance float64        `gorm:"default:0" json:"last_oil_change_distance"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`

	WorkSessions []WorkSession `gorm:"foreignKey:EmployeeID;constraint:OnDelete:SET NULL" json:"work_sessions,omitempty"`
}

func (e *Employee) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// Session status constants
const (
	StatusActive    = "ACTIVE"
	StatusCompleted = "COMPLETED"
)

// WorkSession model
type WorkSession struct {
	ID               uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID       *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"`
	Employee         *Employee      `gorm:"foreignKey:EmployeeID;constraint:OnDelete:SET NULL" json:"employee,omitempty"`
	StartTime        time.Time      `gorm:"not null" json:"start_time"`
	EndTime          *time.Time     `json:"end_time"`
	StartKM          float64        `gorm:"not null" json:"start_km"`
	EndKM            float64        `gorm:"default:0" json:"end_km"`
	Distance         float64        `gorm:"default:0" json:"distance"`
	OrdersCount      int            `gorm:"default:0" json:"orders_count"`
	FuelCost         float64        `gorm:"default:0" json:"fuel_cost"`
	ApplicationID    string         `gorm:"type:varchar(50)" json:"application_id"`
	ApplicationType  string         `gorm:"type:varchar(50)" json:"application_type"`
	VehicleType      string         `gorm:"type:varchar(20)" json:"vehicle_type"`      // override for this shift: "car" or "motorcycle"
	MotorcycleNumber string         `gorm:"type:varchar(50)" json:"motorcycle_number"` // رقم الدباب لهذا الشفت (قد يختلف عن المسجل)
	StartKMImage     string         `gorm:"type:text" json:"start_km_image"`           // صورة عداد البداية
	EndKMImage       string         `gorm:"type:text" json:"end_km_image"`             // صورة عداد النهاية
	IsReviewed           bool           `gorm:"default:false;index" json:"is_reviewed"`                      // حالة مراجعة وتصديق المشرف
	ReviewNotes          string         `gorm:"type:text" json:"review_notes"`                               // ملاحظات المشرف
	ReviewedBy           *uuid.UUID     `gorm:"type:char(36)" json:"reviewed_by"`                            // المشرف المراجع
	IsEditedBySupervisor bool           `gorm:"default:false;index" json:"is_edited_by_supervisor"`           // هل تم تعديل البيانات بواسطة المشرف
	EditedByName         string         `gorm:"type:varchar(100)" json:"edited_by_name"`                     // اسم المشرف الذي قام بالتعديل
	OriginalOrdersCount  int            `gorm:"default:0" json:"original_orders_count"`                      // عدد الطلبات الأصلي المدخل من المندوب
	OriginalEndKM        float64        `gorm:"default:0" json:"original_end_km"`                            // عداد النهاية الأصلي المدخل من المندوب
	OriginalStartKM      float64        `gorm:"default:0" json:"original_start_km"`                          // عداد البداية الأصلي المدخل من المندوب
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

// InventoryItem model for warehouse/inventory management
// Items are global/shared across all branches.
// Quantity per branch is calculated from InventoryTransaction records.
type InventoryItem struct {
	ID        uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Name      string         `gorm:"type:varchar(150);not null;index" json:"name"`
	Type      string         `gorm:"type:varchar(50);not null;index" json:"type"` // "oil" or "spare_part"
	Unit      string         `gorm:"type:varchar(30)" json:"unit"`                // "جركن", "قطعة", etc.
	Barcode   string         `gorm:"type:varchar(100);index" json:"barcode"`      // باركود الصنف
	Quantity  int            `gorm:"not null;default:0" json:"quantity"`          // الكمية الإجمالية (للتخزين فقط - لا تستخدم للعرض)
	MinAlert  int            `gorm:"default:5" json:"min_alert"`                  // minimum quantity alert
	Notes     string         `gorm:"type:text" json:"notes"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (i *InventoryItem) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// InventoryTransaction model for stock movements
type InventoryTransaction struct {
	ID         uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	ItemID     uuid.UUID      `gorm:"type:char(36);index;not null" json:"item_id"`
	Item       *InventoryItem `gorm:"foreignKey:ItemID" json:"item,omitempty"`
	Type       string         `gorm:"type:varchar(20);not null;index" json:"type"` // "in" or "out"
	Quantity   int            `gorm:"not null" json:"quantity"`
	EmployeeID *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"` // who received the item
	Employee   *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	BranchID   *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"` // ط§ظ„ظپط±ط¹ ط§ظ„طھط§ط¨ط¹ ظ„ظ‡
	Notes      string         `gorm:"type:text" json:"notes"`
	CreatedAt  time.Time      `json:"created_at"`
}

func (t *InventoryTransaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// PurchaseInvoice model for tracking supplier purchase bills
type PurchaseInvoice struct {
	ID            uuid.UUID             `gorm:"type:char(36);primary_key" json:"id"`
	InvoiceNumber string                `gorm:"type:varchar(100);index;not null" json:"invoice_number"`
	SupplierName  string                `gorm:"type:varchar(200);index;not null" json:"supplier_name"`
	InvoiceDate   time.Time             `gorm:"not null" json:"invoice_date"`
	Subtotal      float64               `gorm:"type:decimal(12,2);not null;default:0" json:"subtotal"`
	Discount      float64               `gorm:"type:decimal(12,2);not null;default:0" json:"discount"`
	TaxRate       float64               `gorm:"type:decimal(5,2);not null;default:0" json:"tax_rate"`
	TaxAmount     float64               `gorm:"type:decimal(12,2);not null;default:0" json:"tax_amount"`
	TotalAmount   float64               `gorm:"type:decimal(12,2);not null;default:0" json:"total_amount"`
	BranchID      *uuid.UUID            `gorm:"type:char(36);index" json:"branch_id"`
	Branch        *Branch               `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	CreatedByName string                `gorm:"type:varchar(100)" json:"created_by_name"`
	Notes         string                `gorm:"type:text" json:"notes"`
	Items         []PurchaseInvoiceItem `gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
	DeletedAt     gorm.DeletedAt        `gorm:"index" json:"-"`
}

func (p *PurchaseInvoice) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// PurchaseInvoiceItem model for individual line items within a purchase invoice
type PurchaseInvoiceItem struct {
	ID         uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	InvoiceID  uuid.UUID      `gorm:"type:char(36);index;not null" json:"invoice_id"`
	ItemID     uuid.UUID      `gorm:"type:char(36);index;not null" json:"item_id"`
	Item       *InventoryItem `gorm:"foreignKey:ItemID" json:"item,omitempty"`
	Quantity   int            `gorm:"not null" json:"quantity"`
	UnitPrice  float64        `gorm:"type:decimal(10,2);not null;default:0" json:"unit_price"`
	TotalPrice float64        `gorm:"type:decimal(12,2);not null;default:0" json:"total_price"`
	Notes      string         `gorm:"type:text" json:"notes"`
	CreatedAt  time.Time      `json:"created_at"`
}

func (pi *PurchaseInvoiceItem) BeforeCreate(tx *gorm.DB) error {
	if pi.ID == uuid.Nil {
		pi.ID = uuid.New()
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

// Investigation model - supports multiple template types
type Investigation struct {
	ID                 uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID         uuid.UUID  `gorm:"type:char(36);index;not null" json:"employee_id"`
	Employee           *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	SupervisorID       uuid.UUID  `gorm:"type:char(36);index;not null" json:"supervisor_id"`
	Supervisor         *Admin     `gorm:"foreignKey:SupervisorID" json:"supervisor,omitempty"`
	NationalID         string     `gorm:"type:varchar(50)" json:"national_id"`
	Type               string     `gorm:"type:varchar(30);default:investigation" json:"type"` // investigation, supervisor_report, advance, internet_advance, absence, custody
	Questions          string     `gorm:"type:text" json:"questions"`                         // JSON array of question strings (for investigation)
	Answers            string     `gorm:"type:text" json:"answers"`                           // JSON array of answer strings (for investigation)
	ReportText         string     `gorm:"type:text" json:"report_text"`                       // Text content (for supervisor_report, absence)
	Images             string     `gorm:"type:text" json:"images"`                            // JSON array of image URLs (for supervisor_report)
	Amount             *float64   `gorm:"type:decimal(10,2)" json:"amount"`                   // Amount (for advance, internet_advance)
	StartDate          *time.Time `json:"start_date"`                                         // Absence start date
	EndDate            *time.Time `json:"end_date"`                                           // Absence end date
	Items              string     `gorm:"type:text" json:"items"`                             // JSON array of items (for custody)
	IsGuilty           *bool      `gorm:"default:false" json:"is_guilty"`
	Notes              string     `gorm:"type:text" json:"notes"`
	DeductionMonth     string     `gorm:"type:varchar(20)" json:"deduction_month"`              // ط´ظ‡ط± ط§ظ„ط®طµظ… ظ…ظ† ط§ظ„ط±ط§طھط¨ (ظ„ط³ظ„ظپط© ط§ظ„ظ†طھ)
	Status             string     `gorm:"type:varchar(20);default:pending;index" json:"status"` // pending, approved, rejected (ظ„ظ„ط³ظ„ظپط© ظˆط³ظ„ظپط© ط§ظ„ظ†طھ)
	ApprovedByName     string     `gorm:"type:varchar(100)" json:"approved_by_name"`
	ApprovedByUsername string     `gorm:"type:varchar(50)" json:"approved_by_username"`
	RejectedByName     string     `gorm:"type:varchar(100)" json:"rejected_by_name"`
	RejectedByUsername string     `gorm:"type:varchar(50)" json:"rejected_by_username"`
	ApprovedAt         *time.Time `json:"approved_at"`
	RejectedAt         *time.Time `json:"rejected_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

func (i *Investigation) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// --- Attendance Model ---
type Attendance struct {
	ID         uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:idx_emp_date" json:"employee_id"`
	Employee   *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Date       string    `gorm:"type:varchar(10);not null;uniqueIndex:idx_emp_date" json:"date"`
	Status     string    `gorm:"type:varchar(20);not null;default:present" json:"status"` // present / absent
	Note       string    `gorm:"type:text" json:"note"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (a *Attendance) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.Status == "" {
		a.Status = "present"
	}
	return nil
}

type AttendanceInfo struct {
	EmployeeID     uuid.UUID  `json:"employee_id"`
	EmployeeName   string     `json:"employee_name"`
	NationalID     string     `json:"national_id"`
	BranchName     string     `json:"branch_name"`
	VehicleType    string     `json:"vehicle_type"`
	Status         string     `json:"status"`
	Note           string     `json:"note"`
	HasWorkSession bool       `json:"has_work_session"`
	SessionID      *uuid.UUID `json:"session_id"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
}

// CustodyDay model - daily custody per branch.
// OpeningBalance is carried from the previous day's ClosingBalance,
// AddedAmount is the new money added today, and ClosingBalance is the
// remaining balance carried to the next day.
type CustodyDay struct {
	ID             uuid.UUID        `gorm:"type:char(36);primary_key" json:"id"`
	BranchID       *uuid.UUID       `gorm:"type:char(36);uniqueIndex:idx_custody_day_branch_date" json:"branch_id"`
	Branch         *Branch          `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Date           string           `gorm:"type:varchar(10);not null;uniqueIndex:idx_custody_day_branch_date" json:"date"`
	OpeningBalance float64          `gorm:"default:0" json:"opening_balance"`
	AddedAmount    float64          `gorm:"default:0" json:"added_amount"`
	ClosingBalance float64          `gorm:"default:0" json:"closing_balance"`
	Expenses       []CustodyExpense `gorm:"foreignKey:CustodyDayID;constraint:OnDelete:CASCADE" json:"expenses,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

func (c *CustodyDay) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type AppSetting struct {
	ID        uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Key       string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`
	Value     string         `gorm:"type:text" json:"value"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *AppSetting) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// CustodyExpense model - a single expense entry under one of the custody categories.
type CustodyExpense struct {
	ID                uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	CustodyDayID      uuid.UUID  `gorm:"type:char(36);index;not null" json:"custody_day_id"`
	Category          string     `gorm:"type:varchar(20);not null;index" json:"category"` // fuel, license, spare_parts, other
	Amount            float64    `gorm:"default:0" json:"amount"`
	RecipientName     string     `gorm:"type:varchar(150)" json:"recipient_name"`
	CreatedByID       *uuid.UUID `gorm:"type:char(36);index" json:"created_by_id"`
	CreatedByName     string     `gorm:"type:varchar(100)" json:"created_by_name"`
	CreatedByUsername string     `gorm:"type:varchar(50)" json:"created_by_username"`
	CreatedAt         time.Time  `json:"created_at"`
}

func (e *CustodyExpense) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// CustodyLog model - audit log for every custody transaction/movement
type CustodyLog struct {
	ID            uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	BranchID      *uuid.UUID `gorm:"type:char(36);index" json:"branch_id"`
	Branch        *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	CustodyDayID  uuid.UUID  `gorm:"type:char(36);index" json:"custody_day_id"`
	Date          string     `gorm:"type:varchar(10);not null;index" json:"date"`
	ActionType    string     `gorm:"type:varchar(30);not null;index" json:"action_type"` // ADD_CUSTODY, ADD_EXPENSE, DELETE_EXPENSE
	Category      string     `gorm:"type:varchar(30)" json:"category"`                   // fuel, license, spare_parts, other, custody
	Amount        float64    `gorm:"default:0" json:"amount"`
	Description   string     `gorm:"type:varchar(255)" json:"description"`
	RecipientName string     `gorm:"type:varchar(150)" json:"recipient_name"`
	AdminID       *uuid.UUID `gorm:"type:char(36);index" json:"admin_id"`
	AdminName     string     `gorm:"type:varchar(100)" json:"admin_name"`
	AdminUsername string     `gorm:"type:varchar(50)" json:"admin_username"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (l *CustodyLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
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
	ID                    uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	PlateNumber           string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"plate_number"` // ط±ظ‚ظ… ط§ظ„ظ„ظˆط­ط© / ط§ظ„ط¯ط¨ط§ط¨
	VehicleType           string         `gorm:"type:varchar(20);default:'motorcycle'" json:"vehicle_type"` // "motorcycle" or "car"
	Brand                 string         `gorm:"type:varchar(100)" json:"brand"`                           // ظ…ط§ط±ظƒط© ط§ظ„ط¯ط¨ط§ط¨ (ظ‡ظˆظ†ط¯ط§طŒ ط³ظˆط²ظˆظƒظٹ...)
	ModelYear             string         `gorm:"type:varchar(20)" json:"model_year"`                       // ط³ظ†ط© ط§ظ„طµظ†ط¹
	KeyNumber             string         `gorm:"type:varchar(50)" json:"key_number"`                       // ط±ظ‚ظ… ط§ظ„ظ…ظپطھط§ط­ ط§ظ„ظ…ط±طھط¨ط·
	CurrentKM             float64        `gorm:"default:0" json:"current_km"`                              // ط§ظ„ط¹ط¯ط§ط¯ ط§ظ„ط­ط§ظ„ظٹ ط§ظ„ظ…ط³ط¬ظ„
	LastOilChangeKM       float64        `gorm:"default:0" json:"last_oil_change_km"`                      // ظ‚ط±ط§ط،ط© ط§ظ„ط¹ط¯ط§ط¯ ط¹ظ†ط¯ ط¢ط®ط± طھط؛ظٹظٹط± ط²ظٹطھ
	TotalDistance         float64        `gorm:"default:0" json:"total_distance"`                          // ط¥ط¬ظ…ط§ظ„ظٹ ط§ظ„ظƒظٹظ„ظˆظ…طھط±ط§طھ ط§ظ„ظ…ظ‚ط·ظˆط¹ط©
	Status                string         `gorm:"type:varchar(20);default:'AVAILABLE';index" json:"status"` // "AVAILABLE", "IN_USE", "MAINTENANCE"
	BranchID              *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Branch                *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Notes                 string         `gorm:"type:text" json:"notes"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
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
	Reason          string         `gorm:"type:varchar(255);not null" json:"reason"` // ط³ط±ط¹ط©طŒ ط¥ط´ط§ط±ط©طŒ ط­ط²ط§ظ…...
	ViolationDate   time.Time      `gorm:"index" json:"violation_date"`
	City            string         `gorm:"type:varchar(100)" json:"city"`
	Status          string         `gorm:"type:varchar(30);default:'RECORDED';index" json:"status"` // RECORDED, DEDUCTED, DISPUTED, PAID
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

// ------------------------------------------------------------------
// 4. EmployeeDocument Model (ط§ظ„ظ…ط³طھظ†ط¯ط§طھ ظˆط§ظ„ط±ط®طµ ظˆط§ظ„ط´ظ‡ط§ط¯ط§طھ)
// ------------------------------------------------------------------
type EmployeeDocument struct {
	ID         uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID uuid.UUID      `gorm:"type:char(36);index;not null" json:"employee_id"`
	Employee   *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	DocType    string         `gorm:"type:varchar(50);not null;index" json:"doc_type"` // PROMISSORY_NOTE, CONTRACT, DRIVING_LICENSE, VEHICLE_REGISTRATION, CRIMINAL_RECORD, MEDICAL_INSURANCE, OTHER
	Title      string         `gorm:"type:varchar(200);not null" json:"title"`
	DocNumber  string         `gorm:"type:varchar(100)" json:"doc_number"`
	FileURL    string         `gorm:"type:varchar(500)" json:"file_url"`
	IssueDate  *time.Time     `json:"issue_date"`
	ExpiryDate *time.Time     `gorm:"index" json:"expiry_date"`
	Status     string         `gorm:"type:varchar(30);default:'VALID';index" json:"status"` // VALID, EXPIRING_SOON, EXPIRED, PENDING_REVIEW
	Notes      string         `gorm:"type:text" json:"notes"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (d *EmployeeDocument) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// ------------------------------------------------------------------
// 5. EmployeeBankAccount Model (ط§ظ„ط­ط³ط§ط¨ط§طھ ط§ظ„ط¨ظ†ظƒظٹط© ظ„ظ„ظ…ظ†ط§ط¯ظٹط¨)
// ------------------------------------------------------------------
type EmployeeBankAccount struct {
	ID               uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID       uuid.UUID      `gorm:"type:char(36);index;not null" json:"employee_id"`
	Employee         *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	BankName         string         `gorm:"type:varchar(100);not null" json:"bank_name"`
	IBAN             string         `gorm:"type:varchar(50);not null;index" json:"iban"`
	AccountOwnerName string         `gorm:"type:varchar(150);not null" json:"account_owner_name"`
	IsDefault        bool           `gorm:"default:false" json:"is_default"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *EmployeeBankAccount) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// ------------------------------------------------------------------
// 6. LeaveRequest Model (ط·ظ„ط¨ط§طھ ط§ظ„ط¥ط¬ط§ط²ط§طھ)
// ------------------------------------------------------------------
type LeaveRequest struct {
	ID             uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID     uuid.UUID      `gorm:"type:char(36);index;not null" json:"employee_id"`
	Employee       *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	LeaveType      string         `gorm:"type:varchar(50);default:'ANNUAL';index" json:"leave_type"` // ANNUAL, SICK, EMERGENCY, UNPAID
	StartDate      string         `gorm:"type:varchar(10);not null;index" json:"start_date"`         // YYYY-MM-DD
	EndDate        string         `gorm:"type:varchar(10);not null;index" json:"end_date"`           // YYYY-MM-DD
	DaysCount      int            `gorm:"default:1" json:"days_count"`
	Reason         string         `gorm:"type:text" json:"reason"`
	Status         string         `gorm:"type:varchar(30);default:'PENDING';index" json:"status"` // PENDING, APPROVED, REJECTED
	ApprovedByName string         `gorm:"type:varchar(100)" json:"approved_by_name"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (l *LeaveRequest) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

// ------------------------------------------------------------------
// 7. SupportTicket Model (طھط°ط§ظƒط± ط§ظ„ط¯ط¹ظ… ظˆط§ظ„ط´ظƒط§ظˆظ‰)
// ------------------------------------------------------------------
type SupportTicket struct {
	ID           uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	TicketNumber string         `gorm:"type:varchar(50);index;not null" json:"ticket_number"`
	EmployeeID   *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"`
	Employee     *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Subject      string         `gorm:"type:varchar(255);not null" json:"subject"`
	Category     string         `gorm:"type:varchar(50);default:'OPERATIONAL';index" json:"category"` // OPERATIONAL, FINANCIAL, VEHICLE, APPLICATION, OTHER
	Priority     string         `gorm:"type:varchar(20);default:'MEDIUM';index" json:"priority"`      // LOW, MEDIUM, HIGH, URGENT
	Status       string         `gorm:"type:varchar(30);default:'OPEN';index" json:"status"`          // OPEN, IN_PROGRESS, RESOLVED, CLOSED
	Description  string         `gorm:"type:text;not null" json:"description"`
	Resolution   string         `gorm:"type:text" json:"resolution"`
	BranchID     *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Branch       *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
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



type Notification struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID   *uuid.UUID `gorm:"type:uuid;index" json:"branch_id"`
	AdminID    *uuid.UUID `gorm:"type:uuid;index" json:"admin_id"`
	EmployeeID *uuid.UUID `gorm:"type:uuid;index" json:"employee_id"`
	Title      string     `gorm:"type:varchar(255);not null" json:"title"`
	Body       string     `gorm:"type:text;not null" json:"body"`
	Type       string     `gorm:"type:varchar(50);not null;index" json:"type"`
	Status     string     `gorm:"type:varchar(50);default:'unread';index" json:"status"`
	CreatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	Branch   *Branch   `gorm:"foreignKey:BranchID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"branch,omitempty"`
	Admin    *Admin    `gorm:"foreignKey:AdminID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"admin,omitempty"`
	Employee *Employee `gorm:"foreignKey:EmployeeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"employee,omitempty"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

