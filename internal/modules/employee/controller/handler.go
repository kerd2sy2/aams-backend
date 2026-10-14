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
	"delivery-backend/pkg/middleware"
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


func (h *EmployeeHandler) Create(c *gin.Context) {
	var req dto.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if !scope.IsGlobal {
		req.BranchID = scope.BranchID
	} else if req.BranchID == nil && scope.BranchID != nil {
		req.BranchID = scope.BranchID
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

	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if !scope.IsGlobal {
		filter.BranchID = scope.BranchID
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
	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	var branchFilter *uuid.UUID
	if !scope.IsGlobal {
		branchFilter = scope.BranchID
	} else if scope.BranchID != nil {
		branchFilter = scope.BranchID
	}

	emps, err := h.svc.Search(c.Request.Context(), q, branchFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, emps)
}

func (h *EmployeeHandler) GetWorking(c *gin.Context) {
	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	var branchFilter *uuid.UUID
	if !scope.IsGlobal {
		branchFilter = scope.BranchID
	} else if scope.BranchID != nil {
		branchFilter = scope.BranchID
	}

	emps, err := h.svc.GetWorkingEmployees(c.Request.Context(), branchFilter)
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

	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	emp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil || !scope.AllowsBranch(emp.BranchID) {
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

	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	existingEmp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil || !scope.AllowsBranch(existingEmp.BranchID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	var req dto.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	if !scope.IsGlobal && req.BranchID != nil && !scope.AllowsBranch(req.BranchID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح بنقل الموظف إلى فرع آخر"})
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

	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	existingEmp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil || !scope.AllowsBranch(existingEmp.BranchID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف الموظف بنجاح"})
}

func (h *EmployeeHandler) BatchSetOilChange(c *gin.Context) {
	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if isEmpVal, ok := c.Get("is_employee"); ok {
		if b, okBool := isEmpVal.(bool); okBool && b {
			c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح للموظفين بإجراء هذا التعديل"})
			return
		}
	}
	if strings.EqualFold(scope.Role, "DRIVER") || strings.EqualFold(scope.Role, "EMPLOYEE") {
		c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح للسائقين بإجراء هذا التعديل"})
		return
	}

	var req dto.BatchOilSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("بيانات غير صالحة: %v", err)})
		return
	}

	if len(req.Entries) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "قائمة الموظفين لا يمكن أن تكون فارغة"})
		return
	}

	// Pre-validate all entries fail-closed before performing any write operations
	seenIDs := make(map[uuid.UUID]bool, len(req.Entries))
	for _, entry := range req.Entries {
		id, err := uuid.Parse(entry.EmployeeID)
		if err != nil || id == uuid.Nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "معرف الموظف غير صالح في القائمة"})
			return
		}
		if entry.LastOilChangeDistance < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "مسافة غيار الزيت لا يمكن أن تكون سالبة"})
			return
		}
		if seenIDs[id] {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("تكرار معرف الموظف %s في نفس الطلب غير مسموح به", entry.EmployeeID)})
			return
		}
		seenIDs[id] = true

		emp, err := h.svc.FindByID(c.Request.Context(), id)
		if err != nil || !scope.AllowsBranch(emp.BranchID) {
			c.JSON(http.StatusNotFound, gin.H{"error": "أحد الموظفين في القائمة غير موجود"})
			return
		}
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
	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	var branchFilter *uuid.UUID
	if !scope.IsGlobal {
		branchFilter = scope.BranchID
	} else if scope.BranchID != nil {
		branchFilter = scope.BranchID
	}
	locations, err := h.svc.GetLocations(c.Request.Context(), branchFilter)
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

	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	existingEmp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil || !scope.AllowsBranch(existingEmp.BranchID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
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

	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	existingEmp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil || !scope.AllowsBranch(existingEmp.BranchID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
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

	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	emp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil || !scope.AllowsBranch(emp.BranchID) {
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

	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	emp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil || !scope.AllowsBranch(emp.BranchID) {
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

	scope, err := middleware.ResolveBranchScope(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	emp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil || !scope.AllowsBranch(emp.BranchID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "الموظف غير موجود"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": emp})
}

func (h *EmployeeHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		file, err = c.FormFile("file")
	}
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
