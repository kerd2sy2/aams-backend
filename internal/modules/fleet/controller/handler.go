package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/fleet/contracts"
	"delivery-backend/internal/modules/fleet/dto"
	"delivery-backend/internal/modules/fleet/service"
)

type VehicleHandler struct {
	vehicleService service.VehicleService
	auditService   contracts.IAuditContract
}

func NewVehicleHandler(vehicleService service.VehicleService, auditService contracts.IAuditContract) *VehicleHandler {
	return &VehicleHandler{
		vehicleService: vehicleService,
		auditService:   auditService,
	}
}

func (h *VehicleHandler) Create(c *gin.Context) {
	var req dto.CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if b, ok := bid.(*uuid.UUID); ok && b != nil {
			req.BranchID = b
		}
	}

	vehicle, err := h.vehicleService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.auditService != nil {
		adminName, _ := c.Get("admin_name")
		adminStr := "System"
		if adminName != nil {
			adminStr = adminName.(string)
		}
		_ = h.auditService.LogAction(c.Request.Context(), adminStr, "إضافة دباب/مركبة", fmt.Sprintf("تمت إضافة الدباب رقم اللوحة %s", vehicle.PlateNumber), c.ClientIP(), vehicle.BranchID)
	}

	c.JSON(http.StatusCreated, vehicle)
}

func (h *VehicleHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المركبة غير صالح"})
		return
	}

	var req dto.UpdateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "بيانات غير صالحة: " + err.Error()})
		return
	}

	vehicle, err := h.vehicleService.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.auditService != nil {
		adminName, _ := c.Get("admin_name")
		adminStr := "System"
		if adminName != nil {
			adminStr = adminName.(string)
		}
		_ = h.auditService.LogAction(c.Request.Context(), adminStr, "تعديل دباب/مركبة", fmt.Sprintf("تم تعديل بيانات الدباب رقم اللوحة %s", vehicle.PlateNumber), c.ClientIP(), vehicle.BranchID)
	}

	c.JSON(http.StatusOK, vehicle)
}

func (h *VehicleHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المركبة غير صالح"})
		return
	}

	if err := h.vehicleService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في حذف المركبة"})
		return
	}

	if h.auditService != nil {
		adminName, _ := c.Get("admin_name")
		adminStr := "System"
		if adminName != nil {
			adminStr = adminName.(string)
		}
		_ = h.auditService.LogAction(c.Request.Context(), adminStr, "حذف دباب/مركبة", fmt.Sprintf("تم حذف المركبة معرف %s", idStr), c.ClientIP(), nil)
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم حذف المركبة بنجاح"})
}

func (h *VehicleHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المركبة غير صالح"})
		return
	}

	vehicle, err := h.vehicleService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "المركبة غير موجودة"})
		return
	}

	c.JSON(http.StatusOK, vehicle)
}

func (h *VehicleHandler) GetAll(c *gin.Context) {
	var filter dto.VehicleFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معاملات بحث غير صالحة"})
		return
	}

	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if b, ok := bid.(*uuid.UUID); ok && b != nil {
			filter.BranchID = b
		}
	}

	vehicles, total, err := h.vehicleService.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في جلب قائمة المركبات"})
		return
	}

	limit := filter.Limit
	if limit <= 0 && filter.PageSize > 0 {
		limit = filter.PageSize
	}
	if limit <= 0 {
		limit = 50
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  vehicles,
		"total": total,
		"page":  filter.Page,
		"limit": limit,
	})
}

func (h *VehicleHandler) CheckKM(c *gin.Context) {
	plate := strings.TrimSpace(c.Query("plate"))
	if plate == "" {
		plate = strings.TrimSpace(c.Query("motorcycle_number"))
	}
	if plate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "رقم اللوحة مطلوب"})
		return
	}

	vehicle, err := h.vehicleService.GetByPlateNumber(c.Request.Context(), plate)
	if err == nil && vehicle != nil {
		c.JSON(http.StatusOK, gin.H{
			"plate_number":       plate,
			"current_km":         vehicle.CurrentKM,
			"is_odometer_broken": vehicle.IsOdometerBroken,
			"vehicle_type":       vehicle.VehicleType,
			"status":             vehicle.Status,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plate_number":       plate,
		"current_km":         0,
		"is_odometer_broken": false,
		"vehicle_type":       "motorcycle",
		"status":             "AVAILABLE",
	})
}

func (h *VehicleHandler) RecordOilChange(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "معرف المركبة غير صالح"})
		return
	}

	if err := h.vehicleService.RecordOilChange(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "تم تسجيل تغيير الزيت بنجاح"})
}
