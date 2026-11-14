package system

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReportHandler struct {
	svc ReportService
}

func NewReportHandler(svc ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

func (h *ReportHandler) GetReports(c *gin.Context) {
	var filter ReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "فلاتر غير صالحة"})
		return
	}

	// Strict IDOR Prevention & Field Ownership Enforcement:
	// If caller is a field employee/driver/delegate, strictly bind filter to their verified identity
	isEmp, _ := c.Get("is_employee")
	role := strings.ToUpper(c.GetString("admin_role"))
	if isEmp == true || role == "DRIVER" || role == "EMPLOYEE" || role == "DELEGATE" {
		if authEmpID, exists := c.Get("employee_id"); exists {
			if aUUID, ok := authEmpID.(uuid.UUID); ok && aUUID != uuid.Nil {
				filter.EmployeeID = aUUID.String()
				filter.BranchID = nil // Avoid branch filter restricting employee's records
			}
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "غير مصرح"})
			return
		}
	} else {
		branchID := getBranchID(c)
		if branchID != nil {
			filter.BranchID = branchID
		}
	}

	reports, total, err := h.svc.GetReports(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب التقارير"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  reports,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

func (h *ReportHandler) ExportReports(c *gin.Context) {
	var filter ReportFilter
	_ = c.ShouldBindQuery(&filter)

	branchID := getBranchID(c)
	if branchID != nil {
		filter.BranchID = branchID
	}

	csvData, err := h.svc.ExportReports(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تصدير التقارير"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=work_reports_%s.csv", time.Now().Format("2006-01-02")))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvData)
}

func (h *ReportHandler) GetDailyReport(c *gin.Context) {
	dateStr := c.Query("date")
	branchID := getBranchID(c)

	rep, err := h.svc.GetDailyReport(c.Request.Context(), dateStr, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل جلب التقرير اليومي"})
		return
	}

	c.JSON(http.StatusOK, rep)
}

func (h *ReportHandler) ExportDailyReport(c *gin.Context) {
	dateStr := c.Query("date")
	branchID := getBranchID(c)

	csvData, err := h.svc.ExportDailyReport(c.Request.Context(), dateStr, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل تصدير التقرير اليومي"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=daily_report_%s.csv", dateStr))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvData)
}
