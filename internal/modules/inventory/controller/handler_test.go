package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/inventory/contracts"
	"delivery-backend/internal/modules/inventory/controller"
	"delivery-backend/internal/modules/inventory/domain"
	"delivery-backend/internal/modules/inventory/dto"
	"delivery-backend/internal/modules/inventory/service"
)

type mockInventoryService struct {
	createItemFunc func(ctx context.Context, req dto.CreateItemRequest) (*domain.InventoryItem, error)
	getItemsFunc   func(ctx context.Context, itemType string, branchID *uuid.UUID) ([]service.InventoryItemWithStock, error)
	addStockFunc   func(ctx context.Context, req dto.AddStockRequest, branchID *uuid.UUID) (*domain.InventoryTransaction, error)
}

func (m *mockInventoryService) CreateItem(ctx context.Context, req dto.CreateItemRequest) (*domain.InventoryItem, error) {
	if m.createItemFunc != nil {
		return m.createItemFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockInventoryService) UpdateItem(ctx context.Context, id uuid.UUID, req dto.UpdateItemRequest) (*domain.InventoryItem, error) {
	return nil, nil
}

func (m *mockInventoryService) DeleteItem(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockInventoryService) GetAllItems(ctx context.Context, itemType string, branchID *uuid.UUID) ([]service.InventoryItemWithStock, error) {
	if m.getItemsFunc != nil {
		return m.getItemsFunc(ctx, itemType, branchID)
	}
	return nil, nil
}

func (m *mockInventoryService) GetItemByID(ctx context.Context, id uuid.UUID) (*domain.InventoryItem, error) {
	return nil, nil
}

func (m *mockInventoryService) FindByBarcode(ctx context.Context, barcode string) (*domain.InventoryItem, error) {
	return nil, nil
}

func (m *mockInventoryService) AddStock(ctx context.Context, req dto.AddStockRequest, branchID *uuid.UUID) (*domain.InventoryTransaction, error) {
	if m.addStockFunc != nil {
		return m.addStockFunc(ctx, req, branchID)
	}
	return nil, nil
}

func (m *mockInventoryService) RemoveStock(ctx context.Context, req dto.RemoveStockRequest, branchID *uuid.UUID) (*domain.InventoryTransaction, error) {
	return nil, nil
}

func (m *mockInventoryService) GetTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]domain.InventoryTransaction, int64, error) {
	return nil, 0, nil
}

func (m *mockInventoryService) DeleteAllTransactions(ctx context.Context) error {
	return nil
}

func (m *mockInventoryService) CreatePurchaseInvoice(ctx context.Context, req dto.CreatePurchaseInvoiceRequest, branchID *uuid.UUID, adminName string) (*domain.PurchaseInvoice, error) {
	return nil, nil
}

func (m *mockInventoryService) GetPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]domain.PurchaseInvoice, int64, error) {
	return nil, 0, nil
}

func (m *mockInventoryService) GetPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseInvoice, error) {
	return nil, nil
}

func (m *mockInventoryService) DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockInventoryService) DispenseOil(ctx context.Context, req dto.DispenseOilRequest, adminName string, branchID *uuid.UUID) ([]*domain.InventoryTransaction, error) {
	return nil, nil
}

func (m *mockInventoryService) CheckStockAvailability(ctx context.Context, itemID uuid.UUID, requiredQty int) (bool, error) {
	return true, nil
}

func (m *mockInventoryService) DeductStock(ctx context.Context, itemID uuid.UUID, qty int, employeeID *uuid.UUID, branchID *uuid.UUID, notes string) error {
	return nil
}

func (m *mockInventoryService) GetItemByBarcode(ctx context.Context, barcode string) (*contracts.StockItemSummaryDTO, error) {
	return nil, nil
}

func TestInventoryHandler_CreateItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	itemID := uuid.New()
	mockSvc := &mockInventoryService{
		createItemFunc: func(ctx context.Context, req dto.CreateItemRequest) (*domain.InventoryItem, error) {
			return &domain.InventoryItem{
				ID:   itemID,
				Name: req.Name,
				Type: req.Type,
			}, nil
		},
	}

	h := controller.NewInventoryHandler(mockSvc)
	r := gin.New()
	r.POST("/api/v1/inventory/items", h.CreateItem)

	body, _ := json.Marshal(dto.CreateItemRequest{
		Name: "Engine Oil 5W30",
		Type: "oil",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/inventory/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestInventoryHandler_GetItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockInventoryService{
		getItemsFunc: func(ctx context.Context, itemType string, branchID *uuid.UUID) ([]service.InventoryItemWithStock, error) {
			return []service.InventoryItemWithStock{
				{
					InventoryItem: &domain.InventoryItem{
						ID:   uuid.New(),
						Name: "Brake Pad",
						Type: "spare_part",
					},
					BranchQuantity: 15,
				},
			}, nil
		},
	}

	h := controller.NewInventoryHandler(mockSvc)
	r := gin.New()
	r.GET("/api/v1/inventory/items", h.GetItems)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/items?type=spare_part", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}
