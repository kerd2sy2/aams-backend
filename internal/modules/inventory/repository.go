package inventory

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InventoryRepository interface {
	CreateItem(ctx context.Context, item *InventoryItem) error
	UpdateItem(ctx context.Context, item *InventoryItem) error
	DeleteItem(ctx context.Context, id uuid.UUID) error
	FindItemByID(ctx context.Context, id uuid.UUID) (*InventoryItem, error)
	FindByBarcode(ctx context.Context, barcode string) (*InventoryItem, error)
	FindAllItems(ctx context.Context, itemType string) ([]InventoryItem, error)
	GetStockByBranch(ctx context.Context, branchID *uuid.UUID) (map[uuid.UUID]int, error)
	GetItemStock(ctx context.Context, itemID uuid.UUID, branchID *uuid.UUID) (int, error)
	FindOilItemsWithStock(ctx context.Context, branchID *uuid.UUID) ([]InventoryItem, error)

	CreateTransaction(ctx context.Context, tx *InventoryTransaction) error
	FindTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]InventoryTransaction, int64, error)
	DeleteAllTransactions(ctx context.Context) error

	CreatePurchaseInvoice(ctx context.Context, invoice *PurchaseInvoice, stockTxs []*InventoryTransaction) error
	FindPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]PurchaseInvoice, int64, error)
	FindPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*PurchaseInvoice, error)
	DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error
	UpdateVehicleOilKM(ctx context.Context, plateNumber string, currentKM float64) error
	FindVehicleByPlateOrID(ctx context.Context, id *uuid.UUID, plateNumber string) (*VehicleInfo, error)
}

type VehicleInfo struct {
	ID          uuid.UUID  `gorm:"column:id"`
	PlateNumber string     `gorm:"column:plate_number"`
	CurrentKM   float64    `gorm:"column:current_km"`
	BranchID    *uuid.UUID `gorm:"column:branch_id"`
}

type gormInventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &gormInventoryRepository{db: db}
}

func (r *gormInventoryRepository) likeOp() string {
	if r.db.Dialector.Name() == "postgres" {
		return "ILIKE"
	}
	return "LIKE"
}

func (r *gormInventoryRepository) CreateItem(ctx context.Context, item *InventoryItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *gormInventoryRepository) UpdateItem(ctx context.Context, item *InventoryItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *gormInventoryRepository) DeleteItem(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&InventoryItem{}, "id = ?", id).Error
}

func (r *gormInventoryRepository) FindItemByID(ctx context.Context, id uuid.UUID) (*InventoryItem, error) {
	var item InventoryItem
	if err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *gormInventoryRepository) FindByBarcode(ctx context.Context, barcode string) (*InventoryItem, error) {
	var item InventoryItem
	if err := r.db.WithContext(ctx).Where("barcode = ?", barcode).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *gormInventoryRepository) FindAllItems(ctx context.Context, itemType string) ([]InventoryItem, error) {
	var items []InventoryItem
	query := r.db.WithContext(ctx).Order("created_at DESC")
	if itemType != "" {
		query = query.Where("type = ?", itemType)
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *gormInventoryRepository) GetStockByBranch(ctx context.Context, branchID *uuid.UUID) (map[uuid.UUID]int, error) {
	type row struct {
		ItemID   uuid.UUID
		Quantity int
	}
	var rows []row

	query := r.db.WithContext(ctx).
		Model(&InventoryTransaction{}).
		Select("item_id, SUM(CASE WHEN type = 'in' THEN quantity ELSE -quantity END) as quantity")

	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	query = query.Group("item_id")

	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]int, len(rows))
	for _, r := range rows {
		result[r.ItemID] = r.Quantity
	}
	return result, nil
}

func (r *gormInventoryRepository) GetItemStock(ctx context.Context, itemID uuid.UUID, branchID *uuid.UUID) (int, error) {
	var total int64

	query := r.db.WithContext(ctx).
		Model(&InventoryTransaction{}).
		Select("COALESCE(SUM(CASE WHEN type = 'in' THEN quantity ELSE -quantity END), 0)").
		Where("item_id = ?", itemID)

	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}

	if err := query.Scan(&total).Error; err != nil {
		return 0, err
	}
	return int(total), nil
}

func (r *gormInventoryRepository) FindOilItemsWithStock(ctx context.Context, branchID *uuid.UUID) ([]InventoryItem, error) {
	var oilItems []InventoryItem
	cond := "type = ? OR LOWER(type) = 'oil' OR name " + r.likeOp() + " ?"
	if err := r.db.WithContext(ctx).Where(cond, "oil", "%زيت%").Find(&oilItems).Error; err != nil {
		return nil, err
	}

	stock, err := r.GetStockByBranch(ctx, branchID)
	if err != nil {
		return nil, err
	}

	var result []InventoryItem
	for _, item := range oilItems {
		if qty, ok := stock[item.ID]; ok && qty > 0 {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *gormInventoryRepository) CreateTransaction(ctx context.Context, tx *InventoryTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *gormInventoryRepository) FindTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]InventoryTransaction, int64, error) {
	var txs []InventoryTransaction
	var total int64

	query := r.db.WithContext(ctx).Model(&InventoryTransaction{}).Preload("Item")
	if itemID != nil {
		query = query.Where("item_id = ?", *itemID)
	}
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&txs).Error
	return txs, total, err
}

func (r *gormInventoryRepository) DeleteAllTransactions(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("1 = 1").Delete(&InventoryTransaction{}).Error
}

func (r *gormInventoryRepository) CreatePurchaseInvoice(ctx context.Context, invoice *PurchaseInvoice, stockTxs []*InventoryTransaction) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(invoice).Error; err != nil {
			return err
		}
		for _, stx := range stockTxs {
			if err := tx.Create(stx).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *gormInventoryRepository) FindPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]PurchaseInvoice, int64, error) {
	var invoices []PurchaseInvoice
	var total int64

	query := r.db.WithContext(ctx).Model(&PurchaseInvoice{}).Preload("Items.Item")

	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if search != "" {
		s := "%" + search + "%"
		query = query.Where("invoice_number LIKE ? OR supplier_name LIKE ? OR notes LIKE ?", s, s, s)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("invoice_date DESC, created_at DESC").Offset(offset).Limit(limit).Find(&invoices).Error
	return invoices, total, err
}

func (r *gormInventoryRepository) FindPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*PurchaseInvoice, error) {
	var invoice PurchaseInvoice
	if err := r.db.WithContext(ctx).Preload("Items.Item").First(&invoice, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *gormInventoryRepository) DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&PurchaseInvoice{}, "id = ?", id).Error
}

func extractDigitsFromPlate(plate string) string {
	var b strings.Builder
	for _, r := range plate {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else if r >= '٠' && r <= '٩' {
			b.WriteRune('0' + (r - '٠'))
		}
	}
	return b.String()
}

func (r *gormInventoryRepository) UpdateVehicleOilKM(ctx context.Context, plateNumber string, currentKM float64) error {
	clean := strings.TrimSpace(plateNumber)
	if clean == "" {
		return nil
	}
	now := time.Now()
	updates := map[string]interface{}{
		"last_oil_change_km":   currentKM,
		"last_oil_change_date": &now,
		"updated_at":           now,
	}
	if currentKM > 0 {
		updates["current_km"] = currentKM
	}
	like := r.likeOp()
	cond1 := "deleted_at IS NULL AND (TRIM(plate_number) = ? OR plate_number " + like + " ? OR plate_number " + like + " ?)"
	tx := r.db.WithContext(ctx).Table("vehicles").
		Where(cond1, clean, clean+"%", "%"+clean+"%").
		Updates(updates)
	if tx.Error == nil && tx.RowsAffected == 0 {
		digits := extractDigitsFromPlate(clean)
		if len(digits) >= 3 {
			cond2 := "deleted_at IS NULL AND (TRIM(plate_number) = ? OR plate_number " + like + " ? OR plate_number " + like + " ?)"
			return r.db.WithContext(ctx).Table("vehicles").
				Where(cond2, digits, digits+" %", digits+"%").
				Updates(updates).Error
		}
	}
	return tx.Error
}

func (r *gormInventoryRepository) FindVehicleByPlateOrID(ctx context.Context, id *uuid.UUID, plateNumber string) (*VehicleInfo, error) {
	like := r.likeOp()
	query := r.db.WithContext(ctx).Table("vehicles").Where("deleted_at IS NULL")
	clean := strings.TrimSpace(plateNumber)
	if id != nil && *id != uuid.Nil {
		query = query.Where("id = ?", *id)
	} else if clean != "" {
		cond := "TRIM(plate_number) = ? OR plate_number " + like + " ? OR plate_number " + like + " ?"
		query = query.Where(cond, clean, clean+"%", "%"+clean+"%")
	} else {
		return nil, errors.New("لا توجد بيانات للبحث عن الدباب")
	}

	var v VehicleInfo
	if err := query.First(&v).Error; err != nil {
		if id == nil || *id == uuid.Nil {
			digits := extractDigitsFromPlate(clean)
			if len(digits) >= 3 {
				cond2 := "deleted_at IS NULL AND (TRIM(plate_number) = ? OR plate_number " + like + " ? OR plate_number " + like + " ?)"
				if err2 := r.db.WithContext(ctx).Table("vehicles").
					Where(cond2, digits, digits+" %", digits+"%").
					First(&v).Error; err2 == nil {
					return &v, nil
				}
			}
		}
		return nil, err
	}
	return &v, nil
}
