package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
	LedgerImageURL string           `gorm:"type:varchar(500)" json:"ledger_image_url"`
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
	ExportNumber      string     `gorm:"type:varchar(50);index" json:"export_number"`
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
	ExportNumber  string     `gorm:"type:varchar(50)" json:"export_number"`
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

