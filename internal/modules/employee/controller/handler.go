package controller

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/employee/dto"
	"delivery-backend/internal/modules/employee/service"
)

type StorageProvider interface {
	SaveUploadedFile(fileBytes []byte, filename, folder string) (string, error)
}

type EmployeeHandler struct {
	svc     service.EmployeeService
	storage StorageProvider
}

func NewEmployeeHandler(svc service.EmployeeService, storage StorageProvider) *EmployeeHandler {
	return &EmployeeHandler{
		svc:     svc,
		storage: storage,
	}
}

func employeeBranchID(c *gin.Context) *uuid.UUID {
	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if val, ok := bid.(*uuid.UUID); ok && val != nil {
			return val
		}
		if val, ok := bid.(uuid.UUID); ok && val != uuid.Nil {
			return &val
		}
	}
	if s := c.Query("branch_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			return &id
		}
	}
	return nil
}

func (h *EmployeeHandler) Create(c *gin.Context) {
	var req dto.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	branchID := employeeBranchID(c)
	if branchID != nil {
		req.BranchID = branchID
	}

	emp, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "تم إنشاء الموظف بنجاح",
		"data":    emp,
	})
}

func (h *EmployeeHandler) GetAll(c *gin.Context) {
	var filter dto.EmployeeFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معاملات البحث غير صالحة"})
		return
	}

	branchID := employeeBranchID(c)
	if branchID != nil {
		filter.BranchID = branchID
	}

	emps, total, err := h.svc.FindAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, dto.PaginatedEmployeeResponse{
		Data:       emps,
		Total:      total,
		Page:       filter.Page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func (h *EmployeeHandler) Search(c *gin.Context) {
	q := c.Query("q")
	branchID := employeeBranchID(c)

	emps, err := h.svc.Search(c.Request.Context(), q, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, emps)
}

func (h *EmployeeHandler) GetWorking(c *gin.Context) {
	branchID := employeeBranchID(c)
	emps, err := h.svc.GetWorkingEmployees(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, emps)
}

func (h *EmployeeHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	emp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	c.JSON(http.StatusOK, emp)
}

func (h *EmployeeHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	var req dto.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	emp, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "تم تحديث بيانات الموظف بنجاح",
		"data":    emp,
	})
}

func (h *EmployeeHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الموظف بنجاح"})
}

func (h *EmployeeHandler) BatchSetOilChange(c *gin.Context) {
	var req dto.BatchOilSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	count, err := h.svc.BatchSetOilChange(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("تم تحديث مسافات غيار الزيت لـ %d موظف بنجاح", count),
		"count":   count,
	})
}

func (h *EmployeeHandler) ChangeMyPassword(c *gin.Context) {
	empIDVal, exists := c.Get("employee_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	empID, _ := empIDVal.(uuid.UUID)

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة"})
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), empID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تغيير كلمة المرور بنجاح"})
}

func (h *EmployeeHandler) SetMyPhone(c *gin.Context) {
	empIDVal, exists := c.Get("employee_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	empID, _ := empIDVal.(uuid.UUID)

	var req dto.SetPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رقم الهاتف مطلوب"})
		return
	}

	if err := h.svc.SetPhone(c.Request.Context(), empID, req.Phone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث رقم الهاتف بنجاح"})
}

func (h *EmployeeHandler) SetMyLocation(c *gin.Context) {
	empIDVal, exists := c.Get("employee_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "غير مصرح"})
		return
	}
	empID, _ := empIDVal.(uuid.UUID)

	var req dto.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "إحداثيات الموقع مطلوبة"})
		return
	}

	if err := h.svc.UpdateLocation(c.Request.Context(), empID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث الموقع بنجاح"})
}

func (h *EmployeeHandler) GetLocations(c *gin.Context) {
	branchID := employeeBranchID(c)
	locations, err := h.svc.GetLocations(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": locations})
}

func (h *EmployeeHandler) SetPhone(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	var req dto.SetPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رقم الهاتف مطلوب"})
		return
	}

	if err := h.svc.SetPhone(c.Request.Context(), id, req.Phone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تحديث رقم الهاتف بنجاح"})
}

func (h *EmployeeHandler) ResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	_ = c.ShouldBindJSON(&req)

	if err := h.svc.ResetPassword(c.Request.Context(), id, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم إعادة تعيين كلمة المرور بنجاح"})
}

func (h *EmployeeHandler) GetBarcode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	emp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	if emp.Barcode == "" {
		emp.Barcode, _ = h.svc.GenerateBarcode(emp.NationalID)
	}

	c.JSON(http.StatusOK, gin.H{"barcode": emp.Barcode})
}

func (h *EmployeeHandler) GetQRCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	emp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	if emp.QRCode == "" {
		emp.QRCode, _ = h.svc.GenerateQRCode(emp.NationalID)
	}

	c.JSON(http.StatusOK, gin.H{"qrcode": emp.QRCode})
}

func (h *EmployeeHandler) GetPrintCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح"})
		return
	}

	emp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": emp})
}

func (h *EmployeeHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "الصورة مطلوبة"})
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))
	savePath := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في حفظ الصورة"})
		return
	}

	url := "/uploads/" + filename
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *EmployeeHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "الملف مطلوب"})
		return
	}

	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), strings.ReplaceAll(file.Filename, " ", "_"))
	savePath := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في حفظ الملف"})
		return
	}

	url := "/uploads/" + filename
	c.JSON(http.StatusOK, gin.H{"url": url})
}
