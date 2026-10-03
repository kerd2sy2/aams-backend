package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/inventory/dto"
	"delivery-backend/internal/modules/inventory/service"
)

type InventoryHandler struct {
	invService service.InventoryService
}

func NewInventoryHandler(invService service.InventoryService) *InventoryHandler {
	return &InventoryHandler{invService: invService}
}

func (h *InventoryHandler) GetItems(c *gin.Context) {
	itemType := c.Query("type")
	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(*uuid.UUID); ok && val != nil {
			branchID = val
		}
	}

	items, err := h.invService.GetAllItems(c.Request.Context(), itemType, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في جلب الأصناف: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *InventoryHandler) GetItemByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الصنف غير صالح"})
		return
	}

	item, err := h.invService.GetItemByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *InventoryHandler) FindByBarcode(c *gin.Context) {
	barcode := strings.TrimSpace(c.Query("barcode"))
	if barcode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "الباركود مطلوب"})
		return
	}

	item, err := h.invService.FindByBarcode(c.Request.Context(), barcode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الصنف غير موجود"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *InventoryHandler) CreateItem(c *gin.Context) {
	var req dto.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	item, err := h.invService.CreateItem(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *InventoryHandler) UpdateItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الصنف غير صالح"})
		return
	}

	var req dto.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	item, err := h.invService.UpdateItem(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *InventoryHandler) DeleteItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الصنف غير صالح"})
		return
	}

	if err := h.invService.DeleteItem(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الصنف بنجاح"})
}

func (h *InventoryHandler) AddStock(c *gin.Context) {
	var req dto.AddStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(*uuid.UUID); ok && val != nil {
			branchID = val
		}
	}

	tx, err := h.invService.AddStock(c.Request.Context(), req, branchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tx)
}

func (h *InventoryHandler) RemoveStock(c *gin.Context) {
	var req dto.RemoveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(*uuid.UUID); ok && val != nil {
			branchID = val
		}
	}

	tx, err := h.invService.RemoveStock(c.Request.Context(), req, branchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tx)
}

func (h *InventoryHandler) DispenseOil(c *gin.Context) {
	var req dto.DispenseOilRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة", "details": err.Error()})
		return
	}

	adminName := c.GetString("admin_name")
	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(*uuid.UUID); ok && val != nil {
			branchID = val
		}
	}

	txs, err := h.invService.DispenseOil(c.Request.Context(), req, adminName, branchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "تم صرف الزيت بنجاح",
		"transactions": txs,
	})
}

func (h *InventoryHandler) GetTransactions(c *gin.Context) {
	var filter dto.TransactionFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معاملات بحث غير صالحة"})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(*uuid.UUID); ok && val != nil {
			branchID = val
		}
	}
	if filter.BranchID != nil {
		branchID = filter.BranchID
	}

	limit := filter.Limit
	if limit <= 0 && filter.PageSize > 0 {
		limit = filter.PageSize
	}
	if limit <= 0 {
		limit = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	txs, total, err := h.invService.GetTransactions(c.Request.Context(), filter.ItemID, branchID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في جلب الحركات: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  txs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *InventoryHandler) DeleteAllTransactions(c *gin.Context) {
	if err := h.invService.DeleteAllTransactions(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في تصفير سجل الحركات: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم تصفير سجل الحركات بنجاح"})
}

func (h *InventoryHandler) CreatePurchaseInvoice(c *gin.Context) {
	var req dto.CreatePurchaseInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(*uuid.UUID); ok && val != nil {
			branchID = val
		}
	}

	adminName := "Admin"
	if name, exists := c.Get("admin_name"); exists && name != nil {
		adminName = fmt.Sprintf("%v", name)
	}

	invoice, err := h.invService.CreatePurchaseInvoice(c.Request.Context(), req, branchID, adminName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, invoice)
}

func (h *InventoryHandler) GetPurchaseInvoices(c *gin.Context) {
	var filter dto.PurchaseInvoiceFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معاملات بحث غير صالحة"})
		return
	}

	var branchID *uuid.UUID
	if bID, exists := c.Get("branch_id"); exists && bID != nil {
		if val, ok := bID.(*uuid.UUID); ok && val != nil {
			branchID = val
		}
	}
	if filter.BranchID != nil {
		branchID = filter.BranchID
	}

	limit := filter.Limit
	if limit <= 0 && filter.PageSize > 0 {
		limit = filter.PageSize
	}
	if limit <= 0 {
		limit = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	invoices, total, err := h.invService.GetPurchaseInvoices(c.Request.Context(), branchID, filter.Search, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في جلب فواتير المشتريات: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  invoices,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *InventoryHandler) GetPurchaseInvoiceByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفاتورة غير صالح"})
		return
	}

	invoice, err := h.invService.GetPurchaseInvoiceByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الفاتورة غير موجودة"})
		return
	}
	c.JSON(http.StatusOK, invoice)
}

func (h *InventoryHandler) DeletePurchaseInvoice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الفاتورة غير صالح"})
		return
	}

	if err := h.invService.DeletePurchaseInvoice(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تم حذف فاتورة المشتريات بنجاح"})
}
