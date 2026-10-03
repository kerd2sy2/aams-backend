package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/inventory/domain"
)

type InventoryRepository interface {
	CreateItem(ctx context.Context, item *domain.InventoryItem) error
	UpdateItem(ctx context.Context, item *domain.InventoryItem) error
	DeleteItem(ctx context.Context, id uuid.UUID) error
	FindItemByID(ctx context.Context, id uuid.UUID) (*domain.InventoryItem, error)
	FindByBarcode(ctx context.Context, barcode string) (*domain.InventoryItem, error)
	FindAllItems(ctx context.Context, itemType string) ([]domain.InventoryItem, error)
	GetStockByBranch(ctx context.Context, branchID *uuid.UUID) (map[uuid.UUID]int, error)
	GetItemStock(ctx context.Context, itemID uuid.UUID, branchID *uuid.UUID) (int, error)
	FindOilItemsWithStock(ctx context.Context, branchID *uuid.UUID) ([]domain.InventoryItem, error)

	CreateTransaction(ctx context.Context, tx *domain.InventoryTransaction) error
	FindTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]domain.InventoryTransaction, int64, error)
	DeleteAllTransactions(ctx context.Context) error

	CreatePurchaseInvoice(ctx context.Context, invoice *domain.PurchaseInvoice, stockTxs []*domain.InventoryTransaction) error
	FindPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]domain.PurchaseInvoice, int64, error)
	FindPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseInvoice, error)
	DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error
}

type gormInventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &gormInventoryRepository{db: db}
}

func (r *gormInventoryRepository) CreateItem(ctx context.Context, item *domain.InventoryItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *gormInventoryRepository) UpdateItem(ctx context.Context, item *domain.InventoryItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *gormInventoryRepository) DeleteItem(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.InventoryItem{}, "id = ?", id).Error
}

func (r *gormInventoryRepository) FindItemByID(ctx context.Context, id uuid.UUID) (*domain.InventoryItem, error) {
	var item domain.InventoryItem
	if err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *gormInventoryRepository) FindByBarcode(ctx context.Context, barcode string) (*domain.InventoryItem, error) {
	var item domain.InventoryItem
	if err := r.db.WithContext(ctx).Where("barcode = ?", barcode).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *gormInventoryRepository) FindAllItems(ctx context.Context, itemType string) ([]domain.InventoryItem, error) {
	var items []domain.InventoryItem
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
		Model(&domain.InventoryTransaction{}).
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
		Model(&domain.InventoryTransaction{}).
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

func (r *gormInventoryRepository) FindOilItemsWithStock(ctx context.Context, branchID *uuid.UUID) ([]domain.InventoryItem, error) {
	var oilItems []domain.InventoryItem
	if err := r.db.WithContext(ctx).Where("type = ?", "oil").Find(&oilItems).Error; err != nil {
		return nil, err
	}

	stock, err := r.GetStockByBranch(ctx, branchID)
	if err != nil {
		return nil, err
	}

	var result []domain.InventoryItem
	for _, item := range oilItems {
		if qty, ok := stock[item.ID]; ok && qty > 0 {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *gormInventoryRepository) CreateTransaction(ctx context.Context, tx *domain.InventoryTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *gormInventoryRepository) FindTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]domain.InventoryTransaction, int64, error) {
	var txs []domain.InventoryTransaction
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.InventoryTransaction{}).Preload("Item")
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
	return r.db.WithContext(ctx).Where("1 = 1").Delete(&domain.InventoryTransaction{}).Error
}

func (r *gormInventoryRepository) CreatePurchaseInvoice(ctx context.Context, invoice *domain.PurchaseInvoice, stockTxs []*domain.InventoryTransaction) error {
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

func (r *gormInventoryRepository) FindPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]domain.PurchaseInvoice, int64, error) {
	var invoices []domain.PurchaseInvoice
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.PurchaseInvoice{}).Preload("Items.Item")

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

func (r *gormInventoryRepository) FindPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseInvoice, error) {
	var invoice domain.PurchaseInvoice
	if err := r.db.WithContext(ctx).Preload("Items.Item").First(&invoice, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *gormInventoryRepository) DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.PurchaseInvoice{}, "id = ?", id).Error
}
