package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateItemRequest struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"` // "oil" or "spare_part"
	Unit     string `json:"unit"`
	Barcode  string `json:"barcode"`
	MinAlert int    `json:"min_alert"`
	Notes    string `json:"notes"`
}

type UpdateItemRequest struct {
	Name     *string `json:"name"`
	Type     *string `json:"type"`
	Unit     *string `json:"unit"`
	Barcode  *string `json:"barcode"`
	MinAlert *int    `json:"min_alert"`
	Notes    *string `json:"notes"`
}

type AddStockRequest struct {
	ItemID   uuid.UUID  `json:"item_id" binding:"required"`
	Quantity int        `json:"quantity" binding:"required,gt=0"`
	Notes    string     `json:"notes"`
	BranchID *uuid.UUID `json:"branch_id"`
}

type RemoveStockRequest struct {
	ItemID     uuid.UUID  `json:"item_id" binding:"required"`
	Quantity   int        `json:"quantity" binding:"required,gt=0"`
	EmployeeID *uuid.UUID `json:"employee_id"`
	Notes      string     `json:"notes"`
	BranchID   *uuid.UUID `json:"branch_id"`
}

type DispenseOilRequest struct {
	EmployeeID uuid.UUID  `json:"employee_id" binding:"required"`
	Quantity   int        `json:"quantity" binding:"required,gt=0"`
	ItemID     *uuid.UUID `json:"item_id"` // اختياري
	Notes      string     `json:"notes"`
	BranchID   *uuid.UUID `json:"branch_id"`
}

type InventoryFilter struct {
	BranchID *uuid.UUID `form:"branch_id"`
	Type     string     `form:"type"`
	Search   string     `form:"search"`
	Page     int        `form:"page,default=1"`
	Limit    int        `form:"limit,default=50"`
	PageSize int        `form:"page_size"`
}

type TransactionFilter struct {
	BranchID   *uuid.UUID `form:"branch_id"`
	ItemID     *uuid.UUID `form:"item_id"`
	EmployeeID *uuid.UUID `form:"employee_id"`
	Type       string     `form:"type"`
	Page       int        `form:"page,default=1"`
	Limit      int        `form:"limit,default=50"`
	PageSize   int        `form:"page_size"`
}

type CreatePurchaseInvoiceItemRequest struct {
	ItemID    uuid.UUID `json:"item_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,gt=0"`
	UnitPrice float64   `json:"unit_price" binding:"required,gte=0"`
	Notes     string    `json:"notes"`
}

type CreatePurchaseInvoiceRequest struct {
	InvoiceNumber string                             `json:"invoice_number" binding:"required"`
	SupplierName  string                             `json:"supplier_name" binding:"required"`
	InvoiceDate   time.Time                          `json:"invoice_date" binding:"required"`
	Discount      float64                            `json:"discount"`
	TaxRate       float64                            `json:"tax_rate"`
	BranchID      *uuid.UUID                         `json:"branch_id"`
	Notes         string                             `json:"notes"`
	Items         []CreatePurchaseInvoiceItemRequest `json:"items" binding:"required,min=1"`
}

type PurchaseInvoiceFilter struct {
	BranchID     *uuid.UUID `form:"branch_id"`
	SupplierName string     `form:"supplier_name"`
	Search       string     `form:"search"`
	Page         int        `form:"page,default=1"`
	Limit        int        `form:"limit,default=50"`
	PageSize     int        `form:"page_size"`
}
