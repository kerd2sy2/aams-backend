package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/inventory/domain"
	"delivery-backend/internal/modules/inventory/dto"
	"delivery-backend/internal/modules/inventory/service"
)

type mockInventoryRepo struct {
	items        map[uuid.UUID]*domain.InventoryItem
	transactions []*domain.InventoryTransaction
}

func newMockInventoryRepo() *mockInventoryRepo {
	return &mockInventoryRepo{
		items:        make(map[uuid.UUID]*domain.InventoryItem),
		transactions: make([]*domain.InventoryTransaction, 0),
	}
}

func (m *mockInventoryRepo) CreateItem(ctx context.Context, item *domain.InventoryItem) error {
	m.items[item.ID] = item
	return nil
}
func (m *mockInventoryRepo) UpdateItem(ctx context.Context, item *domain.InventoryItem) error {
	m.items[item.ID] = item
	return nil
}
func (m *mockInventoryRepo) DeleteItem(ctx context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}
func (m *mockInventoryRepo) FindItemByID(ctx context.Context, id uuid.UUID) (*domain.InventoryItem, error) {
	it, ok := m.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return it, nil
}
func (m *mockInventoryRepo) FindByBarcode(ctx context.Context, barcode string) (*domain.InventoryItem, error) {
	for _, it := range m.items {
		if it.Barcode == barcode {
			return it, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (m *mockInventoryRepo) FindAllItems(ctx context.Context, itemType string) ([]domain.InventoryItem, error) {
	var list []domain.InventoryItem
	for _, it := range m.items {
		if itemType == "" || it.Type == itemType {
			list = append(list, *it)
		}
	}
	return list, nil
}
func (m *mockInventoryRepo) GetStockByBranch(ctx context.Context, branchID *uuid.UUID) (map[uuid.UUID]int, error) {
	res := make(map[uuid.UUID]int)
	for _, tx := range m.transactions {
		if tx.Type == "in" {
			res[tx.ItemID] += tx.Quantity
		} else {
			res[tx.ItemID] -= tx.Quantity
		}
	}
	return res, nil
}
func (m *mockInventoryRepo) GetItemStock(ctx context.Context, itemID uuid.UUID, branchID *uuid.UUID) (int, error) {
	total := 0
	for _, tx := range m.transactions {
		if tx.ItemID == itemID {
			if tx.Type == "in" {
				total += tx.Quantity
			} else {
				total -= tx.Quantity
			}
		}
	}
	return total, nil
}
func (m *mockInventoryRepo) FindOilItemsWithStock(ctx context.Context, branchID *uuid.UUID) ([]domain.InventoryItem, error) {
	return nil, nil
}
func (m *mockInventoryRepo) CreateTransaction(ctx context.Context, tx *domain.InventoryTransaction) error {
	m.transactions = append(m.transactions, tx)
	return nil
}
func (m *mockInventoryRepo) FindTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]domain.InventoryTransaction, int64, error) {
	var list []domain.InventoryTransaction
	for _, tx := range m.transactions {
		list = append(list, *tx)
	}
	return list, int64(len(list)), nil
}
func (m *mockInventoryRepo) DeleteAllTransactions(ctx context.Context) error {
	m.transactions = nil
	return nil
}
func (m *mockInventoryRepo) CreatePurchaseInvoice(ctx context.Context, invoice *domain.PurchaseInvoice, stockTxs []*domain.InventoryTransaction) error {
	m.transactions = append(m.transactions, stockTxs...)
	return nil
}
func (m *mockInventoryRepo) FindPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]domain.PurchaseInvoice, int64, error) {
	return nil, 0, nil
}
func (m *mockInventoryRepo) FindPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseInvoice, error) {
	return nil, nil
}
func (m *mockInventoryRepo) DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestInventoryService_AddAndRemoveStock(t *testing.T) {
	repo := newMockInventoryRepo()
	svc := service.NewInventoryService(repo)

	item, err := svc.CreateItem(context.Background(), dto.CreateItemRequest{
		Name:     "زيت بترومين 20W50",
		Type:     "oil",
		Unit:     "جركن",
		Barcode:  "OIL-1001",
		MinAlert: 10,
	})
	if err != nil {
		t.Fatalf("expected create item to succeed, got %v", err)
	}

	// 1. Add Stock (+50)
	_, err = svc.AddStock(context.Background(), dto.AddStockRequest{
		ItemID:   item.ID,
		Quantity: 50,
		Notes:    "وارد مستودع",
	}, nil)
	if err != nil {
		t.Fatalf("expected add stock to succeed, got %v", err)
	}

	avail, err := svc.CheckStockAvailability(context.Background(), item.ID, 30)
	if err != nil || !avail {
		t.Errorf("expected 30 items available out of 50")
	}

	// 2. Remove Stock (-20)
	_, err = svc.RemoveStock(context.Background(), dto.RemoveStockRequest{
		ItemID:   item.ID,
		Quantity: 20,
		Notes:    "صرف لصيانة دورية",
	}, nil)
	if err != nil {
		t.Fatalf("expected remove stock to succeed, got %v", err)
	}

	// 3. Check remaining stock (50 - 20 = 30)
	availRemaining, _ := svc.CheckStockAvailability(context.Background(), item.ID, 31)
	if availRemaining {
		t.Errorf("expected 31 items NOT to be available when only 30 remain")
	}
}
