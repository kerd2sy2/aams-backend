package inventory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type InventoryItemWithStock struct {
	*InventoryItem
	Quantity       int `json:"quantity"`
	BranchQuantity int `json:"branch_quantity"`
}

type InventoryService interface {
	CreateItem(ctx context.Context, req CreateItemRequest) (*InventoryItem, error)
	UpdateItem(ctx context.Context, id uuid.UUID, req UpdateItemRequest) (*InventoryItem, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
	GetAllItems(ctx context.Context, itemType string, branchID *uuid.UUID) ([]InventoryItemWithStock, error)
	GetItemByID(ctx context.Context, id uuid.UUID) (*InventoryItem, error)
	FindByBarcode(ctx context.Context, barcode string) (*InventoryItem, error)
	AddStock(ctx context.Context, req AddStockRequest, branchID *uuid.UUID) (*InventoryTransaction, error)
	RemoveStock(ctx context.Context, req RemoveStockRequest, branchID *uuid.UUID) (*InventoryTransaction, error)
	GetTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]InventoryTransaction, int64, error)
	DeleteAllTransactions(ctx context.Context) error

	CreatePurchaseInvoice(ctx context.Context, req CreatePurchaseInvoiceRequest, branchID *uuid.UUID, adminName string) (*PurchaseInvoice, error)
	GetPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]PurchaseInvoice, int64, error)
	GetPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*PurchaseInvoice, error)
	DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error
	DispenseOil(ctx context.Context, req DispenseOilRequest, adminName string, branchID *uuid.UUID) ([]*InventoryTransaction, error)

	// Contract implementations
	CheckStockAvailability(ctx context.Context, itemID uuid.UUID, requiredQty int) (bool, error)
	DeductStock(ctx context.Context, itemID uuid.UUID, qty int, employeeID *uuid.UUID, branchID *uuid.UUID, notes string) error
	GetItemByBarcode(ctx context.Context, barcode string) (*StockItemSummaryDTO, error)
	DispenseOilForVehicle(ctx context.Context, plateNumber string, currentKM float64, branchID *uuid.UUID, adminName string) error
}

type inventoryService struct {
	repo InventoryRepository
}

func NewInventoryService(repo InventoryRepository) InventoryService {
	return &inventoryService{repo: repo}
}

func (s *inventoryService) CreateItem(ctx context.Context, req CreateItemRequest) (*InventoryItem, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, ErrItemNameRequired
	}
	if req.Type != "oil" && req.Type != "spare_part" {
		return nil, ErrInvalidItemType
	}

	item := &InventoryItem{
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

func (s *inventoryService) UpdateItem(ctx context.Context, id uuid.UUID, req UpdateItemRequest) (*InventoryItem, error) {
	item, err := s.repo.FindItemByID(ctx, id)
	if err != nil {
		return nil, ErrItemNotFound
	}

	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		item.Name = strings.TrimSpace(*req.Name)
	}
	if req.Type != nil && *req.Type != "" {
		if *req.Type != "oil" && *req.Type != "spare_part" {
			return nil, ErrInvalidItemType
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
		return ErrItemNotFound
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
		qty := stock[items[i].ID]
		result[i] = InventoryItemWithStock{
			InventoryItem:  &items[i],
			Quantity:       qty,
			BranchQuantity: qty,
		}
	}
	return result, nil
}

func (s *inventoryService) GetItemByID(ctx context.Context, id uuid.UUID) (*InventoryItem, error) {
	return s.repo.FindItemByID(ctx, id)
}

func (s *inventoryService) FindByBarcode(ctx context.Context, barcode string) (*InventoryItem, error) {
	return s.repo.FindByBarcode(ctx, strings.TrimSpace(barcode))
}

func (s *inventoryService) AddStock(ctx context.Context, req AddStockRequest, branchID *uuid.UUID) (*InventoryTransaction, error) {
	item, err := s.repo.FindItemByID(ctx, req.ItemID)
	if err != nil {
		return nil, ErrItemNotFound
	}

	effectiveBranch := branchID
	if req.BranchID != nil {
		effectiveBranch = req.BranchID
	}

	tx := &InventoryTransaction{
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

func (s *inventoryService) RemoveStock(ctx context.Context, req RemoveStockRequest, branchID *uuid.UUID) (*InventoryTransaction, error) {
	item, err := s.repo.FindItemByID(ctx, req.ItemID)
	if err != nil {
		return nil, ErrItemNotFound
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

	tx := &InventoryTransaction{
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

func (s *inventoryService) GetTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]InventoryTransaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	return s.repo.FindTransactions(ctx, itemID, branchID, page, limit)
}

func (s *inventoryService) DeleteAllTransactions(ctx context.Context) error {
	return s.repo.DeleteAllTransactions(ctx)
}

func (s *inventoryService) CreatePurchaseInvoice(ctx context.Context, req CreatePurchaseInvoiceRequest, branchID *uuid.UUID, adminName string) (*PurchaseInvoice, error) {
	var subtotal float64
	invoiceItems := make([]PurchaseInvoiceItem, len(req.Items))
	stockTxs := make([]*InventoryTransaction, len(req.Items))

	effectiveBranch := branchID
	if req.BranchID != nil {
		effectiveBranch = req.BranchID
	}

	invoiceID := uuid.New()
	for i, itemReq := range req.Items {
		totalPrice := float64(itemReq.Quantity) * itemReq.UnitPrice
		subtotal += totalPrice

		invoiceItems[i] = PurchaseInvoiceItem{
			ID:         uuid.New(),
			InvoiceID:  invoiceID,
			ItemID:     itemReq.ItemID,
			Quantity:   itemReq.Quantity,
			UnitPrice:  itemReq.UnitPrice,
			TotalPrice: totalPrice,
			Notes:      itemReq.Notes,
		}

		stockTxs[i] = &InventoryTransaction{
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

	invoice := &PurchaseInvoice{
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

func (s *inventoryService) GetPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]PurchaseInvoice, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	return s.repo.FindPurchaseInvoices(ctx, branchID, search, page, limit)
}

func (s *inventoryService) GetPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*PurchaseInvoice, error) {
	return s.repo.FindPurchaseInvoiceByID(ctx, id)
}

func (s *inventoryService) DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePurchaseInvoice(ctx, id)
}

func (s *inventoryService) DispenseOil(ctx context.Context, req DispenseOilRequest, adminName string, branchID *uuid.UUID) ([]*InventoryTransaction, error) {
	effectiveBranch := branchID
	if req.BranchID != nil {
		effectiveBranch = req.BranchID
	}

	// Try to resolve vehicle info if vehicle_id or plate_number is provided
	if (req.VehicleID != nil && *req.VehicleID != uuid.Nil) || req.PlateNumber != "" {
		veh, err := s.repo.FindVehicleByPlateOrID(ctx, req.VehicleID, req.PlateNumber)
		if err == nil && veh != nil {
			if req.PlateNumber == "" {
				req.PlateNumber = veh.PlateNumber
			}
			if req.CurrentKM <= 0 {
				req.CurrentKM = veh.CurrentKM
			}
			if effectiveBranch == nil && veh.BranchID != nil {
				effectiveBranch = veh.BranchID
			}
		}
	}

	var oilItems []InventoryItem
	if req.ItemID != nil && *req.ItemID != uuid.Nil {
		specificItem, err := s.repo.FindItemByID(ctx, *req.ItemID)
		if err == nil && specificItem != nil {
			oilItems = []InventoryItem{*specificItem}
		}
	}
	if len(oilItems) == 0 {
		var err error
		oilItems, err = s.repo.FindOilItemsWithStock(ctx, effectiveBranch)
		if (err != nil || len(oilItems) == 0) && effectiveBranch != nil {
			// Fallback: If branch stock is empty, check central/global oil stock
			oilItems, _ = s.repo.FindOilItemsWithStock(ctx, nil)
		}
		if len(oilItems) == 0 {
			return nil, errors.New("لا يوجد زيت متاح في المخزن. يرجى توريد زيت أو إضافة رصيد أولاً")
		}
	}

	totalOil := 0
	for _, item := range oilItems {
		stock, _ := s.repo.GetItemStock(ctx, item.ID, effectiveBranch)
		if stock <= 0 && effectiveBranch != nil {
			stock, _ = s.repo.GetItemStock(ctx, item.ID, nil)
		}
		totalOil += stock
	}
	if totalOil < req.Quantity {
		return nil, fmt.Errorf("الكمية غير كافية في المخزن. المتاح: %d جركن زيت", totalOil)
	}

	note := strings.TrimSpace(req.Notes)
	if req.PlateNumber != "" {
		plateInfo := fmt.Sprintf("صرف زيت للدباب (%s)", req.PlateNumber)
		if req.CurrentKM > 0 {
			plateInfo += fmt.Sprintf(" - عداد: %.0f كم", req.CurrentKM)
		}
		if note == "" {
			note = plateInfo
		} else if !strings.Contains(note, req.PlateNumber) {
			note = plateInfo + " - " + note
		}
	} else if note == "" {
		note = fmt.Sprintf("صرف زيت بواسطة المشرف %s", adminName)
	}

	var txs []*InventoryTransaction
	remaining := req.Quantity
	for _, item := range oilItems {
		if remaining <= 0 {
			break
		}
		stock, _ := s.repo.GetItemStock(ctx, item.ID, effectiveBranch)
		if stock <= 0 && effectiveBranch != nil {
			stock, _ = s.repo.GetItemStock(ctx, item.ID, nil)
		}
		if stock > 0 {
			deduct := stock
			if deduct > remaining {
				deduct = remaining
			}

			tx := &InventoryTransaction{
				ItemID:     item.ID,
				Type:       "out",
				Quantity:   deduct,
				EmployeeID: req.EmployeeID,
				BranchID:   effectiveBranch,
				Notes:      note,
			}
			if txErr := s.repo.CreateTransaction(ctx, tx); txErr != nil {
				return nil, fmt.Errorf("فشل في تسجيل حركة صرف الزيت: %w", txErr)
			}
			txs = append(txs, tx)
			remaining -= deduct
		}
	}

	if req.PlateNumber != "" {
		_ = s.repo.UpdateVehicleOilKM(ctx, req.PlateNumber, req.CurrentKM)
	}

	return txs, nil
}

func (s *inventoryService) DispenseOilForVehicle(ctx context.Context, plateNumber string, currentKM float64, branchID *uuid.UUID, adminName string) error {
	req := DispenseOilRequest{
		PlateNumber: plateNumber,
		CurrentKM:   currentKM,
		Quantity:    1,
		Notes:       fmt.Sprintf("تغيير زيت للدباب (%s) بواسطة %s", plateNumber, adminName),
		BranchID:    branchID,
	}
	_, err := s.DispenseOil(ctx, req, adminName, branchID)
	return err
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
		return ErrInsufficientStock
	}

	tx := &InventoryTransaction{
		ItemID:     itemID,
		Type:       "out",
		Quantity:   qty,
		EmployeeID: employeeID,
		BranchID:   branchID,
		Notes:      notes,
	}
	return s.repo.CreateTransaction(ctx, tx)
}

func (s *inventoryService) GetItemByBarcode(ctx context.Context, barcode string) (*StockItemSummaryDTO, error) {
	item, err := s.repo.FindByBarcode(ctx, strings.TrimSpace(barcode))
	if err != nil {
		return nil, err
	}
	stock, _ := s.repo.GetItemStock(ctx, item.ID, nil)
	return &StockItemSummaryDTO{
		ID:       item.ID,
		Name:     item.Name,
		Type:     item.Type,
		Barcode:  item.Barcode,
		Quantity: stock,
	}, nil
}
