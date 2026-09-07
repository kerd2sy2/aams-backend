package dto

import "github.com/google/uuid"

// ParsedExcelRow represents a validated row from Excel
type ParsedExcelRow struct {
	Serial        string `json:"serial"`
	Identifier    string `json:"identifier"`
	App           string `json:"app"`
	Branch        string `json:"branch,omitempty"` // "1", "2", إلخ
	DriverName    string `json:"driver_name"`
	NinjaOrders   int    `json:"ninja_orders"`
	KeetaOrders   int    `json:"keeta_orders"`
	ToyoOrders    int    `json:"toyo_orders"`
	TotalOrders   int    `json:"total_orders"`
	PlateNumber   string `json:"plate_number"`
	Notes         string `json:"notes"`
	IsDuplicate   bool   `json:"is_duplicate"`
	ExistingCount int    `json:"existing_count,omitempty"`
}

// ExcelImportPreviewResponse returns preview data before database commit
type ExcelImportPreviewResponse struct {
	FileName              string           `json:"file_name"`
	OrderDate             string           `json:"order_date"` // YYYY-MM-DD
	TotalRows             int              `json:"total_rows"`
	TotalOrders           int              `json:"total_orders"`
	IdentifiersCount      int              `json:"identifiers_count"`
	Identifiers           []string         `json:"identifiers"`
	DriversCount          int              `json:"drivers_count"`
	Drivers               []string         `json:"drivers"`
	DuplicatesCount       int              `json:"duplicates_count"`
	HasDuplicates         bool             `json:"has_duplicates"`
	EmptyIdentifiersCount int              `json:"empty_identifiers_count"`
	Warnings              []string         `json:"warnings,omitempty"`
	Rows                  []ParsedExcelRow `json:"rows"`
}

// ConfirmImportRequest is sent by Admin to save previewed rows
type ConfirmImportRequest struct {
	FileName            string           `json:"file_name" binding:"required"`
	OrderDate           string           `json:"order_date" binding:"required"` // YYYY-MM-DD
	DeduplicationAction string           `json:"deduplication_action"`          // IGNORE_DUPLICATES, REPLACE_DUPLICATES, CANCEL
	Rows                []ParsedExcelRow `json:"rows" binding:"required"`
}

// ConfirmImportResponse summary of save operation
type ConfirmImportResponse struct {
	BatchID             uuid.UUID `json:"batch_id"`
	ImportedOrdersCount int       `json:"imported_orders_count"`
	SkippedCount        int       `json:"skipped_count"`
	ReplacedCount       int       `json:"replaced_count"`
	TotalOrders         int       `json:"total_orders"`
	Message             string    `json:"message"`
}

// IdentifierPerformanceDTO represents metrics and prediction for a single identifier
type IdentifierPerformanceDTO struct {
	ID                       uuid.UUID `json:"id"`
	Name                     string    `json:"name"`
	AppName                  string    `json:"app_name"`
	Branch                   string    `json:"branch,omitempty"` // "1", "2", إلخ
	Code                     string    `json:"code,omitempty"`
	TodayOrders              int       `json:"today_orders"`
	WeekOrders               int       `json:"week_orders"`
	MonthOrders              int       `json:"month_orders"`
	MonthlyTarget            int       `json:"monthly_target"`
	AchievementPercent       float64   `json:"achievement_percent"`
	DailyAverage             float64   `json:"daily_average"`
	DailyRequired            float64   `json:"daily_required"`
	RemainingDays            int       `json:"remaining_days"`
	Status                   string    `json:"status"` // ON_TRACK, AT_RISK, BEHIND_TARGET, TARGET_ACHIEVED
	ProjectedMonthlyOrders   int       `json:"projected_monthly_orders"`
	EstimatedAchievementDate string    `json:"estimated_achievement_date,omitempty"` // YYYY-MM-DD or "N/A"
	IsQualified              bool      `json:"is_qualified"`
	IsActive                 bool      `json:"is_active"`
}

// DriverContributionDTO breakdown of orders by driver within an identifier
type DriverContributionDTO struct {
	DriverID    uuid.UUID      `json:"driver_id"`
	DriverName  string         `json:"driver_name"`
	Orders      int            `json:"orders"`
	Percentage  float64        `json:"percentage"`
	DailyOrders map[string]int `json:"daily_orders,omitempty"` // map[YYYY-MM-DD]ordersCount
	DaysActive  int            `json:"days_active,omitempty"`
}

// DayTrendDTO daily order counts for charts
type DayTrendDTO struct {
	Date   string `json:"date"`   // YYYY-MM-DD
	Day    int    `json:"day"`    // 1-31
	Orders int    `json:"orders"` // total orders on this day
	Target int    `json:"target"` // daily target baseline
}

// IdentifierDetailsDTO full details page for an identifier
type IdentifierDetailsDTO struct {
	Performance        IdentifierPerformanceDTO `json:"performance"`
	DriversBreakdown   []DriverContributionDTO  `json:"drivers_breakdown"`
	AppsBreakdown      map[string]int           `json:"apps_breakdown"`
	DailyTimeline      []DayTrendDTO            `json:"daily_timeline"`
	ActiveDriversCount int                      `json:"active_drivers_count"`
}

// DriverPerformanceDTO driver view across all identifiers
type DriverPerformanceDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Phone       string    `json:"phone,omitempty"`
	Branch      string    `json:"branch,omitempty"`
	MonthOrders int       `json:"month_orders"`
	TodayOrders int       `json:"today_orders"`
	Identifiers []string  `json:"identifiers"`
	Apps        []string  `json:"apps"`
}

// TargetDashboardSummaryDTO top-level stats and charts for dashboard
type TargetDashboardSummaryDTO struct {
	TotalIdentifiers int                        `json:"total_identifiers"`
	TargetAchieved   int                        `json:"target_achieved"`
	OnTrack          int                        `json:"on_track"`
	AtRisk           int                        `json:"at_risk"`
	BehindTarget     int                        `json:"behind_target"`
	TotalMonthOrders int                        `json:"total_month_orders"`
	TodayTotalOrders int                        `json:"today_total_orders"`
	DailyTrend       []DayTrendDTO              `json:"daily_trend"`
	TopIdentifiers   []IdentifierPerformanceDTO `json:"top_identifiers"`
	RecentAlerts     []TargetAlertDTO           `json:"recent_alerts"`
	CurrentMonth     string                     `json:"current_month"` // YYYY-MM
	DaysElapsed      int                        `json:"days_elapsed"`
	TotalDaysInMonth int                        `json:"total_days_in_month"`
	RemainingDays    int                        `json:"remaining_days"`
}

// TargetAlertDTO alert item
type TargetAlertDTO struct {
	ID             uuid.UUID `json:"id"`
	IdentifierID   uuid.UUID `json:"identifier_id"`
	IdentifierName string    `json:"identifier_name"`
	AlertDate      string    `json:"alert_date"`
	TargetOrders   int       `json:"target_orders"`
	ActualOrders   int       `json:"actual_orders"`
	Deficit        int       `json:"deficit"`
	IsResolved     bool      `json:"is_resolved"`
	CreatedAt      string    `json:"created_at"`
}

// TargetSettingsDTO target configuration
type TargetSettingsDTO struct {
	DefaultMonthlyTarget int `json:"default_monthly_target"`
	DefaultDailyTarget   int `json:"default_daily_target"`
}
