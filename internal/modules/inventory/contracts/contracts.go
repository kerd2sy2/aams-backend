package contracts

import (
	"context"

	"github.com/google/uuid"
)

// StockItemSummaryDTO carries information about an item in inventory
type StockItemSummaryDTO struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Barcode  string    `json:"barcode"`
	Quantity int       `json:"quantity"`
}

// IInventoryStockContract exposes stock operations for other modules (e.g. Maintenance dispensing oil/spare parts)
type IInventoryStockContract interface {
	CheckStockAvailability(ctx context.Context, itemID uuid.UUID, requiredQty int) (bool, error)
	DeductStock(ctx context.Context, itemID uuid.UUID, qty int, employeeID *uuid.UUID, branchID *uuid.UUID, notes string) error
	GetItemByBarcode(ctx context.Context, barcode string) (*StockItemSummaryDTO, error)
}
