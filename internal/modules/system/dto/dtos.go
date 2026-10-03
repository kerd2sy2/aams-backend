package dto

import (
	"time"

	"github.com/google/uuid"
)

// Dashboard DTOs
type DashboardStatsResponse struct {
	TotalEmployees    int64              `json:"total_employees"`
	ActiveEmployees   int64              `json:"active_employees"`
	TotalMotorcycles  int64              `json:"total_motorcycles"`
	ActiveMotorcycles int64              `json:"active_motorcycles"`
	TodayOrders       int64              `json:"today_orders"`
	TodayDistance     float64            `json:"today_distance"`
	TodayFuelCost     float64            `json:"today_fuel_cost"`
	AvgWorkingHours   float64            `json:"avg_working_hours"`
	DistanceChart     []ChartDataPoint   `json:"distance_chart"`
	OrdersChart       []ChartDataPoint   `json:"orders_chart"`
	FuelCostChart     []ChartDataPoint   `json:"fuel_cost_chart"`
	LatestActivities  []AuditLogResponse `json:"latest_activities"`
}

type ChartDataPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// Report DTOs
type ReportFilter struct {
	StartDate     string     `form:"start_date"`
	EndDate       string     `form:"end_date"`
	EmployeeID    string     `form:"employee_id"`
	ApplicationID string     `form:"application_id"`
	BranchID      *uuid.UUID `form:"branch_id"`
	IsReviewed    *bool      `form:"is_reviewed"`
	Page          int        `form:"page,default=1"`
	Limit         int        `form:"limit,default=50"`
}

type DailyAppSummary struct {
	AppType     string  `json:"app_type"`
	AppName     string  `json:"app_name"`
	TotalOrders int     `json:"total_orders"`
	TotalKM     float64 `json:"total_km"`
	TotalFuel   float64 `json:"total_fuel"`
	Count       int     `json:"count"`
}

type DailyEmployeeRow struct {
	EmployeeID       string  `json:"employee_id"`
	EmployeeName     string  `json:"employee_name"`
	BranchName       string  `json:"branch_name"`
	KeyNumber        string  `json:"key_number"`
	AppType          string  `json:"app_type"`
	AppName          string  `json:"app_name"`
	MotorcycleNumber string  `json:"motorcycle_number"`
	OrdersCount      int     `json:"orders_count"`
	Distance         float64 `json:"distance"`
	FuelCost         float64 `json:"fuel_cost"`
	WorkingDuration  string  `json:"working_duration"`
	Status           string  `json:"status"`
	ReviewNotes      string  `json:"review_notes"`
	IsReviewed       bool    `json:"is_reviewed"`
}

type DailyReportResponse struct {
	Date          string             `json:"date"`
	TotalOrders   int                `json:"total_orders"`
	TotalKM       float64            `json:"total_km"`
	TotalFuel     float64            `json:"total_fuel"`
	TotalSessions int                `json:"total_sessions"`
	AppSummaries  []DailyAppSummary  `json:"app_summaries"`
	Employees     []DailyEmployeeRow `json:"employees"`
}

// Audit Log DTOs
type AuditLogResponse struct {
	ID        uuid.UUID `json:"id"`
	AdminName string    `json:"admin_name"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

type BulkDeleteAuditLogsRequest struct {
	IDs []uuid.UUID `json:"ids" binding:"required"`
}

// Settings DTOs
type AppSettingsResponse struct {
	SiteName string `json:"site_name"`
	LogoURL  string `json:"logo_url"`
}

type UpdateAppSettingsRequest struct {
	SiteName string `json:"site_name"`
	LogoURL  string `json:"logo_url"`
}

type UpdateSettingRequest struct {
	Value string `json:"value" binding:"required"`
}

// Notification DTOs
type NotificationResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	ImageURL  string    `json:"image_url,omitempty"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateBroadcastRequest struct {
	Title          string     `json:"title"`
	Body           string     `json:"body"`
	TitleAr        string     `json:"title_ar"`
	TitleEn        string     `json:"title_en"`
	TitleBn        string     `json:"title_bn"`
	BodyAr         string     `json:"body_ar"`
	BodyEn         string     `json:"body_en"`
	BodyBn         string     `json:"body_bn"`
	ImageURL       string     `json:"image_url"`
	Target         string     `json:"target"` // "ALL" or "BRANCH"
	BranchID       *uuid.UUID `json:"branch_id"`
	HasPoll        bool       `json:"has_poll"`
	PollQuestion   string     `json:"poll_question"`
	PollQuestionAr string     `json:"poll_question_ar"`
	PollQuestionEn string     `json:"poll_question_en"`
	PollQuestionBn string     `json:"poll_question_bn"`
}

type BroadcastItemDTO struct {
	ID             uuid.UUID  `json:"id"`
	Title          string     `json:"title"`
	Body           string     `json:"body"`
	TitleAr        string     `json:"title_ar,omitempty"`
	TitleEn        string     `json:"title_en,omitempty"`
	TitleBn        string     `json:"title_bn,omitempty"`
	BodyAr         string     `json:"body_ar,omitempty"`
	BodyEn         string     `json:"body_en,omitempty"`
	BodyBn         string     `json:"body_bn,omitempty"`
	ImageURL       string     `json:"image_url"`
	Target         string     `json:"target"`
	BranchID       *uuid.UUID `json:"branch_id"`
	BranchName     string     `json:"branch_name,omitempty"`
	CreatedBy      string     `json:"created_by"`
	SentCount      int        `json:"sent_count"`
	HasPoll        bool       `json:"has_poll"`
	PollQuestion   string     `json:"poll_question"`
	PollQuestionAr string     `json:"poll_question_ar,omitempty"`
	PollQuestionEn string     `json:"poll_question_en,omitempty"`
	PollQuestionBn string     `json:"poll_question_bn,omitempty"`
	AgreeCount     int        `json:"agree_count"`
	DisagreeCount  int        `json:"disagree_count"`
	CreatedAt      time.Time  `json:"created_at"`
	IsRead         bool       `json:"is_read"`
	UserVote       string     `json:"user_vote,omitempty"` // "AGREE" or "DISAGREE"
}

type SubmitPollVoteRequest struct {
	Response string `json:"response" binding:"required,oneof=AGREE DISAGREE"`
	Reason   string `json:"reason"`
}

type BroadcastVoteItemDTO struct {
	ID             uuid.UUID `json:"id"`
	EmployeeID     uuid.UUID `json:"employee_id"`
	EmployeeName   string    `json:"employee_name"`
	EmployeeNumber string    `json:"employee_number"`
	NationalID     string    `json:"national_id"`
	Phone          string    `json:"phone"`
	Response       string    `json:"response"` // "AGREE" or "DISAGREE"
	Reason         string    `json:"reason"`
	CreatedAt      time.Time `json:"created_at"`
}

type PushTokenRequest struct {
	PushToken  string `json:"push_token"`
	DeviceUUID string `json:"device_uuid"`
	Language   string `json:"language"`
}

// Archive DTOs
type ArchivedItemDTO struct {
	ID         uuid.UUID  `json:"id"`
	Type       string     `json:"type"`
	TypeName   string     `json:"type_name"`
	Title      string     `json:"title"`
	Subtitle   string     `json:"subtitle"`
	Details    string     `json:"details"`
	BranchID   *uuid.UUID `json:"branch_id,omitempty"`
	BranchName string     `json:"branch_name,omitempty"`
	ArchivedAt time.Time  `json:"archived_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

type ArchiveFilter struct {
	Type     string     `form:"type"`
	Search   string     `form:"search"`
	BranchID *uuid.UUID `form:"branch_id"`
	Page     int        `form:"page,default=1"`
	Limit    int        `form:"limit,default=50"`
}

type ArchiveStatsDTO struct {
	TotalEmployees    int64 `json:"total_employees"`
	TotalVehicles     int64 `json:"total_vehicles"`
	TotalBranches     int64 `json:"total_branches"`
	TotalDocuments    int64 `json:"total_documents"`
	TotalWorkSessions int64 `json:"total_work_sessions"`
	TotalLeaves       int64 `json:"total_leaves"`
	TotalMaintenance  int64 `json:"total_maintenance"`
	TotalViolations   int64 `json:"total_violations"`
	TotalTickets      int64 `json:"total_tickets"`
	GrandTotal        int64 `json:"grand_total"`
}

type ArchiveResponseDTO struct {
	Data       []ArchivedItemDTO `json:"data"`
	Stats      ArchiveStatsDTO   `json:"stats"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

type RestoreArchiveRequest struct {
	Type string    `json:"type" binding:"required"`
	ID   uuid.UUID `json:"id" binding:"required"`
}

type PermanentDeleteRequest struct {
	Type string    `json:"type" binding:"required"`
	ID   uuid.UUID `json:"id" binding:"required"`
}

type BulkArchiveRequest struct {
	Type string      `json:"type" binding:"required"`
	IDs  []uuid.UUID `json:"ids" binding:"required"`
}
