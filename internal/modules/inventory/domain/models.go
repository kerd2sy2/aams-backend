package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound          = errors.New("inventory item not found")
	ErrInsufficientStock = errors.New("insufficient inventory stock")
)

// InventoryItem model for warehouse/inventory management
type InventoryItem struct {
	ID        uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Name      string         `gorm:"type:varchar(150);not null;index" json:"name"`
	Type      string         `gorm:"type:varchar(50);not null;index" json:"type"` // "oil" or "spare_part"
	Unit      string         `gorm:"type:varchar(30)" json:"unit"`                // "جركن", "قطعة", etc.
	Barcode   string         `gorm:"type:varchar(100);index" json:"barcode"`      // باركود الصنف
	Quantity  int            `gorm:"not null;default:0" json:"quantity"`          // الكمية الإجمالية
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
	BranchID   *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
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
