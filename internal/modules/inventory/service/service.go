package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/inventory/contracts"
	"delivery-backend/internal/modules/inventory/domain"
	"delivery-backend/internal/modules/inventory/dto"
	"delivery-backend/internal/modules/inventory/repository"
)

type InventoryItemWithStock struct {
	*domain.InventoryItem
	BranchQuantity int `json:"branch_quantity"`
}

type InventoryService interface {
	CreateItem(ctx context.Context, req dto.CreateItemRequest) (*domain.InventoryItem, error)
	UpdateItem(ctx context.Context, id uuid.UUID, req dto.UpdateItemRequest) (*domain.InventoryItem, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
	GetAllItems(ctx context.Context, itemType string, branchID *uuid.UUID) ([]InventoryItemWithStock, error)
	GetItemByID(ctx context.Context, id uuid.UUID) (*domain.InventoryItem, error)
	FindByBarcode(ctx context.Context, barcode string) (*domain.InventoryItem, error)
	AddStock(ctx context.Context, req dto.AddStockRequest, branchID *uuid.UUID) (*domain.InventoryTransaction, error)
	RemoveStock(ctx context.Context, req dto.RemoveStockRequest, branchID *uuid.UUID) (*domain.InventoryTransaction, error)
	GetTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]domain.InventoryTransaction, int64, error)
	DeleteAllTransactions(ctx context.Context) error

	CreatePurchaseInvoice(ctx context.Context, req dto.CreatePurchaseInvoiceRequest, branchID *uuid.UUID, adminName string) (*domain.PurchaseInvoice, error)
	GetPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]domain.PurchaseInvoice, int64, error)
	GetPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseInvoice, error)
	DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error
	DispenseOil(ctx context.Context, req dto.DispenseOilRequest, adminName string, branchID *uuid.UUID) ([]*domain.InventoryTransaction, error)

	// Contract implementations
	CheckStockAvailability(ctx context.Context, itemID uuid.UUID, requiredQty int) (bool, error)
	DeductStock(ctx context.Context, itemID uuid.UUID, qty int, employeeID *uuid.UUID, branchID *uuid.UUID, notes string) error
	GetItemByBarcode(ctx context.Context, barcode string) (*contracts.StockItemSummaryDTO, error)
}

type inventoryService struct {
	repo repository.InventoryRepository
}

func NewInventoryService(repo repository.InventoryRepository) InventoryService {
	return &inventoryService{repo: repo}
}

func (s *inventoryService) CreateItem(ctx context.Context, req dto.CreateItemRequest) (*domain.InventoryItem, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("اسم الصنف مطلوب")
	}
	if req.Type != "oil" && req.Type != "spare_part" {
		return nil, errors.New("نوع الصنف يجب أن يكون 'oil' أو 'spare_part'")
	}

	item := &domain.InventoryItem{
		Name:     strings.TrimSpace(req.Name),
		Type:     req.Type,
		Unit:     req.Unit,
		Barcode:  strings.TrimSpace(req.Barcode),
		MinAlert: req.MinAlert,
		Notes:    req.Notes,
	}
	if item.MinAlert == 0 {
		item.MinAlert = 5
	}

	if err := s.repo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("فشل في إضافة الصنف: %w", err)
	}
	return item, nil
}

func (s *inventoryService) UpdateItem(ctx context.Context, id uuid.UUID, req dto.UpdateItemRequest) (*domain.InventoryItem, error) {
	item, err := s.repo.FindItemByID(ctx, id)
	if err != nil {
		return nil, errors.New("الصنف غير موجود")
	}

	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		item.Name = strings.TrimSpace(*req.Name)
	}
	if req.Type != nil && *req.Type != "" {
		if *req.Type != "oil" && *req.Type != "spare_part" {
			return nil, errors.New("نوع الصنف يجب أن يكون 'oil' أو 'spare_part'")
		}
		item.Type = *req.Type
	}
	if req.Unit != nil {
		item.Unit = *req.Unit
	}
	if req.Barcode != nil {
		item.Barcode = strings.TrimSpace(*req.Barcode)
	}
	if req.MinAlert != nil && *req.MinAlert > 0 {
		item.MinAlert = *req.MinAlert
	}
	if req.Notes != nil {
		item.Notes = *req.Notes
	}

	if err := s.repo.UpdateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("فشل في تحديث الصنف: %w", err)
	}
	return item, nil
}

func (s *inventoryService) DeleteItem(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.FindItemByID(ctx, id)
	if err != nil {
		return errors.New("الصنف غير موجود")
	}
	return s.repo.DeleteItem(ctx, id)
}

func (s *inventoryService) GetAllItems(ctx context.Context, itemType string, branchID *uuid.UUID) ([]InventoryItemWithStock, error) {
	items, err := s.repo.FindAllItems(ctx, itemType)
	if err != nil {
		return nil, err
	}

	stock, err := s.repo.GetStockByBranch(ctx, branchID)
	if err != nil {
		return nil, err
	}

	result := make([]InventoryItemWithStock, len(items))
	for i := range items {
		result[i] = InventoryItemWithStock{
			InventoryItem:  &items[i],
			BranchQuantity: stock[items[i].ID],
		}
	}
	return result, nil
}

func (s *inventoryService) GetItemByID(ctx context.Context, id uuid.UUID) (*domain.InventoryItem, error) {
	return s.repo.FindItemByID(ctx, id)
}

func (s *inventoryService) FindByBarcode(ctx context.Context, barcode string) (*domain.InventoryItem, error) {
	return s.repo.FindByBarcode(ctx, barcode)
}

func (s *inventoryService) AddStock(ctx context.Context, req dto.AddStockRequest, branchID *uuid.UUID) (*domain.InventoryTransaction, error) {
	item, err := s.repo.FindItemByID(ctx, req.ItemID)
	if err != nil {
		return nil, errors.New("الصنف غير موجود")
	}

	effectiveBranch := branchID
	if req.BranchID != nil {
		effectiveBranch = req.BranchID
	}

	tx := &domain.InventoryTransaction{
		ItemID:   req.ItemID,
		Type:     "in",
		Quantity: req.Quantity,
		BranchID: effectiveBranch,
		Notes:    req.Notes,
	}

	if err := s.repo.CreateTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("فشل في تسجيل الحركة: %w", err)
	}

	tx.Item = item
	return tx, nil
}

func (s *inventoryService) RemoveStock(ctx context.Context, req dto.RemoveStockRequest, branchID *uuid.UUID) (*domain.InventoryTransaction, error) {
	item, err := s.repo.FindItemByID(ctx, req.ItemID)
	if err != nil {
		return nil, errors.New("الصنف غير موجود")
	}

	effectiveBranch := branchID
	if req.BranchID != nil {
		effectiveBranch = req.BranchID
	}

	currentStock, err := s.repo.GetItemStock(ctx, req.ItemID, effectiveBranch)
	if err != nil {
		return nil, fmt.Errorf("فشل في التحقق من المخزون: %w", err)
	}
	if currentStock < req.Quantity {
		return nil, fmt.Errorf("الكمية غير كافية في الفرع. المتاح: %d %s", currentStock, item.Unit)
	}

	tx := &domain.InventoryTransaction{
		ItemID:     req.ItemID,
		Type:       "out",
		Quantity:   req.Quantity,
		EmployeeID: req.EmployeeID,
		BranchID:   effectiveBranch,
		Notes:      req.Notes,
	}

	if err := s.repo.CreateTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("فشل في تسجيل الحركة: %w", err)
	}

	tx.Item = item
	return tx, nil
}

func (s *inventoryService) GetTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]domain.InventoryTransaction, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.FindTransactions(ctx, itemID, branchID, page, limit)
}

func (s *inventoryService) DeleteAllTransactions(ctx context.Context) error {
	return s.repo.DeleteAllTransactions(ctx)
}

func (s *inventoryService) CreatePurchaseInvoice(ctx context.Context, req dto.CreatePurchaseInvoiceRequest, branchID *uuid.UUID, adminName string) (*domain.PurchaseInvoice, error) {
	var subtotal float64
	invoiceItems := make([]domain.PurchaseInvoiceItem, len(req.Items))
	stockTxs := make([]*domain.InventoryTransaction, len(req.Items))

	effectiveBranch := branchID
	if req.BranchID != nil {
		effectiveBranch = req.BranchID
	}

	invoiceID := uuid.New()
	for i, itemReq := range req.Items {
		totalPrice := float64(itemReq.Quantity) * itemReq.UnitPrice
		subtotal += totalPrice

		invoiceItems[i] = domain.PurchaseInvoiceItem{
			ID:         uuid.New(),
			InvoiceID:  invoiceID,
			ItemID:     itemReq.ItemID,
			Quantity:   itemReq.Quantity,
			UnitPrice:  itemReq.UnitPrice,
			TotalPrice: totalPrice,
			Notes:      itemReq.Notes,
		}

		stockTxs[i] = &domain.InventoryTransaction{
			ID:        uuid.New(),
			ItemID:    itemReq.ItemID,
			Type:      "in",
			Quantity:  itemReq.Quantity,
			BranchID:  effectiveBranch,
			Notes:     fmt.Sprintf("وارد فاتورة مشتريات رقم %s", req.InvoiceNumber),
			CreatedAt: req.InvoiceDate,
		}
	}

	taxAmount := (subtotal - req.Discount) * (req.TaxRate / 100.0)
	if taxAmount < 0 {
		taxAmount = 0
	}
	totalAmount := subtotal - req.Discount + taxAmount

	invoice := &domain.PurchaseInvoice{
		ID:            invoiceID,
		InvoiceNumber: strings.TrimSpace(req.InvoiceNumber),
		SupplierName:  strings.TrimSpace(req.SupplierName),
		InvoiceDate:   req.InvoiceDate,
		Subtotal:      subtotal,
		Discount:      req.Discount,
		TaxRate:       req.TaxRate,
		TaxAmount:     taxAmount,
		TotalAmount:   totalAmount,
		BranchID:      effectiveBranch,
		CreatedByName: adminName,
		Notes:         req.Notes,
		Items:         invoiceItems,
	}

	if err := s.repo.CreatePurchaseInvoice(ctx, invoice, stockTxs); err != nil {
		return nil, fmt.Errorf("فشل في حفظ فاتورة المشتريات: %w", err)
	}

	return invoice, nil
}

func (s *inventoryService) GetPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]domain.PurchaseInvoice, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.FindPurchaseInvoices(ctx, branchID, search, page, limit)
}

func (s *inventoryService) GetPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseInvoice, error) {
	return s.repo.FindPurchaseInvoiceByID(ctx, id)
}

func (s *inventoryService) DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePurchaseInvoice(ctx, id)
}

func (s *inventoryService) DispenseOil(ctx context.Context, req dto.DispenseOilRequest, adminName string, branchID *uuid.UUID) ([]*domain.InventoryTransaction, error) {
	oilItems, err := s.repo.FindOilItemsWithStock(ctx, branchID)
	if err != nil || len(oilItems) == 0 {
		return nil, errors.New("لا يوجد زيت متاح في المخزن. يرجى إضافة زيت أولاً")
	}

	totalOil := 0
	for _, item := range oilItems {
		stock, _ := s.repo.GetItemStock(ctx, item.ID, branchID)
		totalOil += stock
	}
	if totalOil < req.Quantity {
		return nil, fmt.Errorf("الكمية غير كافية. المتاح: %d جركن زيت", totalOil)
	}

	var txs []*domain.InventoryTransaction
	remaining := req.Quantity
	for _, item := range oilItems {
		if remaining <= 0 {
			break
		}
		stock, _ := s.repo.GetItemStock(ctx, item.ID, branchID)
		if stock > 0 {
			deduct := stock
			if deduct > remaining {
				deduct = remaining
			}

			note := req.Notes
			if note == "" {
				note = fmt.Sprintf("صرف زيت للمندوب - %d جركن", deduct)
			}

			tx := &domain.InventoryTransaction{
				ItemID:     item.ID,
				Type:       "out",
				Quantity:   deduct,
				EmployeeID: &req.EmployeeID,
				BranchID:   branchID,
				Notes:      note,
			}
			if txErr := s.repo.CreateTransaction(ctx, tx); txErr != nil {
				return nil, fmt.Errorf("فشل في تسجيل حركة صرف الزيت: %w", txErr)
			}
			txs = append(txs, tx)
			remaining -= deduct
		}
	}

	return txs, nil
}

// ---------------------------------------------------------
// Contract Implementation for other modules
// ---------------------------------------------------------

func (s *inventoryService) CheckStockAvailability(ctx context.Context, itemID uuid.UUID, requiredQty int) (bool, error) {
	qty, err := s.repo.GetItemStock(ctx, itemID, nil)
	if err != nil {
		return false, err
	}
	return qty >= requiredQty, nil
}

func (s *inventoryService) DeductStock(ctx context.Context, itemID uuid.UUID, qty int, employeeID *uuid.UUID, branchID *uuid.UUID, notes string) error {
	currentStock, err := s.repo.GetItemStock(ctx, itemID, branchID)
	if err != nil {
		return err
	}
	if currentStock < qty {
		return domain.ErrInsufficientStock
	}

	tx := &domain.InventoryTransaction{
		ItemID:     itemID,
		Type:       "out",
		Quantity:   qty,
		EmployeeID: employeeID,
		BranchID:   branchID,
		Notes:      notes,
	}
	return s.repo.CreateTransaction(ctx, tx)
}

func (s *inventoryService) GetItemByBarcode(ctx context.Context, barcode string) (*contracts.StockItemSummaryDTO, error) {
	item, err := s.repo.FindByBarcode(ctx, barcode)
	if err != nil {
		return nil, err
	}
	stock, _ := s.repo.GetItemStock(ctx, item.ID, nil)
	return &contracts.StockItemSummaryDTO{
		ID:       item.ID,
		Name:     item.Name,
		Type:     item.Type,
		Barcode:  item.Barcode,
		Quantity: stock,
	}, nil
}
