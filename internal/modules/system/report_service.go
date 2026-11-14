package system

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ReportService interface {
	GetReports(ctx context.Context, filter ReportFilter) ([]map[string]interface{}, int64, error)
	ExportReports(ctx context.Context, filter ReportFilter) ([]byte, error)
	GetDailyReport(ctx context.Context, dateStr string, branchID *uuid.UUID) (*DailyReportResponse, error)
	ExportDailyReport(ctx context.Context, dateStr string, branchID *uuid.UUID) ([]byte, error)
}

type reportService struct {
	repo SystemRepository
}

func NewReportService(repo SystemRepository) ReportService {
	return &reportService{repo: repo}
}

func (s *reportService) GetReports(ctx context.Context, filter ReportFilter) ([]map[string]interface{}, int64, error) {
	return s.repo.GetReports(ctx, filter)
}

func (s *reportService) ExportReports(ctx context.Context, filter ReportFilter) ([]byte, error) {
	filter.Limit = 10000
	rows, _, err := s.repo.GetReports(ctx, filter)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.WriteString("\xEF\xBB\xBF") // UTF-8 BOM for Excel
	w := csv.NewWriter(buf)

	_ = w.Write([]string{"المعرف", "المندوب", "الهوية", "الفرع", "التاريخ", "الطلبات", "المسافة (كم)", "تكلفة الوقود", "مدة العمل", "الحالة"})
	for _, r := range rows {
		_ = w.Write([]string{
			fmt.Sprintf("%v", r["id"]),
			fmt.Sprintf("%v", r["employee_name"]),
			fmt.Sprintf("%v", r["national_id"]),
			fmt.Sprintf("%v", r["branch_name"]),
			fmt.Sprintf("%v", r["start_time"]),
			fmt.Sprintf("%v", r["orders_count"]),
			fmt.Sprintf("%v", r["distance"]),
			fmt.Sprintf("%v", r["fuel_cost"]),
			fmt.Sprintf("%v", r["working_duration"]),
			fmt.Sprintf("%v", r["status"]),
		})
	}
	w.Flush()
	return buf.Bytes(), nil
}

func (s *reportService) GetDailyReport(ctx context.Context, dateStr string, branchID *uuid.UUID) (*DailyReportResponse, error) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	return s.repo.GetDailyReport(ctx, dateStr, branchID)
}

func (s *reportService) ExportDailyReport(ctx context.Context, dateStr string, branchID *uuid.UUID) ([]byte, error) {
	rep, err := s.GetDailyReport(ctx, dateStr, branchID)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(buf)

	_ = w.Write([]string{"المندوب", "الفرع", "التطبيق", "الدباب", "الطلبات", "المسافة (كم)", "الوقود", "مدة العمل", "الحالة"})
	for _, e := range rep.Employees {
		_ = w.Write([]string{
			e.EmployeeName,
			e.BranchName,
			e.AppName,
			e.MotorcycleNumber,
			fmt.Sprintf("%d", e.OrdersCount),
			fmt.Sprintf("%.1f", e.Distance),
			fmt.Sprintf("%.1f", e.FuelCost),
			e.WorkingDuration,
			e.Status,
		})
	}
	w.Flush()
	return buf.Bytes(), nil
}
