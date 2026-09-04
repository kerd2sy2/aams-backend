package dto

import (
	"time"

	"delivery-backend/internal/domain"

	"github.com/google/uuid"
)

// Auth DTOs
type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

type EmployeeInfo struct {
	ID               uuid.UUID  `json:"id"`
	Name             string     `json:"name"`
	NationalID       string     `json:"national_id"`
	MotorcycleNumber string     `json:"motorcycle_number"`
	KeyNumber        string     `json:"key_number"`
	EmployeeNumber   string     `json:"employee_number"`
	JobRole          string     `json:"job_role"`
	PersonalImage    string     `json:"personal_image"`
	ApplicationID    string     `json:"application_id"`
	ApplicationType  string     `json:"application_type"`
	Shift            string     `json:"shift"`
	BranchID         *uuid.UUID `json:"branch_id"`
	BranchName       string     `json:"branch_name,omitempty"`
}

type LoginResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	IsEmployee   bool          `json:"is_employee"`
	Employee     *EmployeeInfo `json:"employee,omitempty"`
	Admin        struct {
		ID             uuid.UUID  `json:"id"`
		Name           string     `json:"name"`
		Email          string     `json:"email"`
		Username       string     `json:"username"`
		Phone          string     `json:"phone"`
		Role           string     `json:"role"`
		RoleID         *uuid.UUID `json:"role_id"`
		Permissions    []string   `json:"permissions"`
		GoogleEmail    string     `json:"google_email,omitempty"`
		GoogleAvatar   string     `json:"google_avatar,omitempty"`
		IsGoogleLinked bool       `json:"is_google_linked"`
		BranchID       *uuid.UUID `json:"branch_id"`
		Branch         *struct {
			ID   uuid.UUID `json:"id"`
			Name string    `json:"name"`
		} `json:"branch,omitempty"`
	} `json:"admin"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type GoogleLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	GoogleID string `json:"google_id"`
	Token    string `json:"token"`
}

type GoogleLinkRequest struct {
	Email    string `json:"email" binding:"required"`
	GoogleID string `json:"google_id"`
	Avatar   string `json:"avatar"`
}

// OTP Management DTOs
type RequestOTPRequest struct {
	NationalID string `json:"national_id" binding:"required"`
	DeviceInfo string `json:"device_info"`
	DeviceUUID string `json:"device_uuid"`
}

type RequestOTPResponse struct {
	Success      bool      `json:"success"`
	Message      string    `json:"message"`
	NationalID   string    `json:"national_id"`
	EmployeeName string    `json:"employee_name"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type VerifyOTPRequest struct {
	NationalID string `json:"national_id" binding:"required"`
	OTPCode    string `json:"otp_code" binding:"required"`
	DeviceUUID string `json:"device_uuid"`
}

type OTPListQuery struct {
	Status string `form:"status"`
	Search string `form:"search"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

// Role Management DTOs
type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	IsSystem    bool      `json:"is_system"`
	UsersCount  int64     `json:"users_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	DisplayName string   `json:"display_name" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type UpdateRoleRequest struct {
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type PermissionItemDTO struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type PermissionGroupDTO struct {
	Group       string              `json:"group"`
	Label       string              `json:"label"`
	Permissions []PermissionItemDTO `json:"permissions"`
}

// Admin Management DTOs
type CreateAdminRequest struct {
	Name        string     `json:"name" binding:"required"`
	Email       string     `json:"email" binding:"required,email"`
	Username    string     `json:"username" binding:"required,min=3,max=50"`
	Phone       string     `json:"phone"`
	Password    string     `json:"password" binding:"required,min=8"`
	Role        string     `json:"role"`
	RoleID      *uuid.UUID `json:"role_id"`
	Permissions []string   `json:"permissions"`
	BranchID    *uuid.UUID `json:"branch_id"`
}

type UpdateAdminRequest struct {
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`
	Phone       string     `json:"phone"`
	Password    string     `json:"password"`
	Role        string     `json:"role"`
	RoleID      *uuid.UUID `json:"role_id"`
	Permissions []string   `json:"permissions"`
	BranchID    *uuid.UUID `json:"branch_id"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}


// Employee DTOs
type CreateEmployeeRequest struct {
	Name                string     `json:"name" binding:"required"`
	JobRole             string     `json:"job_role"`
	EmployeeNumber      string     `json:"employee_number"`
	Phone               string     `json:"phone"`
	PersonalImage       string     `json:"personal_image"`
	NationalID          string     `json:"national_id" binding:"required"`
	IqamaExpirationDate *string    `json:"iqama_expiration_date"`
	NationalIDImage          string     `json:"national_id_image"`
	DrivingLicenseImage      string     `json:"driving_license_image"`
	PassportImage            string     `json:"passport_image"`
	VehicleRegistrationImage string     `json:"vehicle_registration_image"`
	KeyNumber                string     `json:"key_number"`
	MotorcycleNumber         string     `json:"motorcycle_number"`
	ApplicationID            string     `json:"application_id"`
	ApplicationType          string     `json:"application_type"`
	VehicleType              string     `json:"vehicle_type"`
	Shift                    string     `json:"shift"`
	BranchID                 *uuid.UUID `json:"branch_id"`
}

type UpdateEmployeeRequest struct {
	Name                     string     `json:"name"`
	JobRole                  string     `json:"job_role"`
	EmployeeNumber           string     `json:"employee_number"`
	Phone                    string     `json:"phone"`
	PersonalImage            string     `json:"personal_image"`
	NationalID               string     `json:"national_id"`
	IqamaExpirationDate      *string    `json:"iqama_expiration_date"`
	NationalIDImage          string     `json:"national_id_image"`
	DrivingLicenseImage      string     `json:"driving_license_image"`
	PassportImage            string     `json:"passport_image"`
	VehicleRegistrationImage string     `json:"vehicle_registration_image"`
	KeyNumber                string     `json:"key_number"`
	MotorcycleNumber         string     `json:"motorcycle_number"`
	ApplicationID            string     `json:"application_id"`
	ApplicationType          string     `json:"application_type"`
	VehicleType              string     `json:"vehicle_type"`
	Shift                    string     `json:"shift"`
	BranchID                 *uuid.UUID `json:"branch_id"`
}

type SetPhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type UpdateLocationRequest struct {
	Latitude  float64  `json:"latitude" binding:"required"`
	Longitude float64  `json:"longitude" binding:"required"`
	Speed     *float64 `json:"speed"`
	Heading   *float64 `json:"heading"`
}

type EmployeeLocationDTO struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	JobRole          string   `json:"job_role"`
	EmployeeNumber   string   `json:"employee_number"`
	Phone            string   `json:"phone"`
	PersonalImage    string   `json:"personal_image"`
	NationalID       string   `json:"national_id"`
	KeyNumber        string   `json:"key_number"`
	MotorcycleNumber string   `json:"motorcycle_number"`
	ApplicationType  string   `json:"application_type"`
	Shift            string   `json:"shift"`
	BranchID         *string  `json:"branch_id"`
	BranchName       string   `json:"branch_name"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	LastLocationAt   *string  `json:"last_location_at"`
	IsShiftActive    bool     `json:"is_shift_active"`
	ActiveSessionID  *string  `json:"active_session_id"`
}

type EmployeeFilter struct {
	Search          string     `form:"search"`
	ApplicationID   string     `form:"application_id"`
	ApplicationType string     `form:"application_type"`
	BranchID        *uuid.UUID `form:"branch_id"`
	Page            int        `form:"page,default=1"`
	Limit           int        `form:"limit,default=10"`
	SortBy          string     `form:"sort_by,default=created_at"`
	Order           string     `form:"order,default=desc"`
}

type PaginatedEmployeeResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// Work Session DTOs
type StartWorkRequest struct {
	EmployeeID       string  `json:"employee_id" binding:"required,uuid"`
	StartKM          float64 `json:"start_km" binding:"gte=0"`             // not required — 0 is valid when odometer is broken
	StartKMImage     string  `json:"start_km_image"`    // صورة عداد البداية
	ApplicationID    string  `json:"application_id"`
	ApplicationType  string  `json:"application_type"`
	VehicleType      string  `json:"vehicle_type"`      // override for this shift
	MotorcycleNumber string  `json:"motorcycle_number"` // رقم الدباب لهذا الشفت (قد يختلف عن المسجل)
	Notes            string  `json:"notes"`
}

type EndWorkRequest struct {
	EmployeeID      string  `json:"employee_id" binding:"required,uuid"`
	EndKM           float64 `json:"end_km" binding:"gte=0"`               // not required — 0 is valid when odometer is broken
	EndKMImage      string  `json:"end_km_image"`      // صورة عداد النهاية
	OrdersCount     int     `json:"orders_count"`
	FuelCost        float64 `json:"fuel_cost"`
	ApplicationID   string   `json:"application_id"`
	ApplicationType string   `json:"application_type"`
	Notes           string   `json:"notes"`
	IsReviewed      *bool    `json:"is_reviewed"`
	ReviewNotes     string   `json:"review_notes"`
}

type ReviewWorkSessionRequest struct {
	IsReviewed  bool     `json:"is_reviewed"`
	ReviewNotes string   `json:"review_notes"`
	OrdersCount *int     `json:"orders_count,omitempty"`
	EndKM       *float64 `json:"end_km,omitempty"`
	StartKM     *float64 `json:"start_km,omitempty"`
	FuelCost    *float64 `json:"fuel_cost,omitempty"`
}

type UpdateWorkSessionRequest struct {
	EmployeeID      string     `json:"employee_id"`
	StartKM         float64    `json:"start_km"`
	StartKMImage    string     `json:"start_km_image"`
	EndKM           float64    `json:"end_km"`
	EndKMImage      string     `json:"end_km_image"`
	OrdersCount     int        `json:"orders_count"`
	FuelCost        float64    `json:"fuel_cost"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	ApplicationType string     `json:"application_type"`
	Notes           string     `json:"notes"`
}

// Dashboard DTOs
type DashboardResponse struct {
	TotalEmployees    int64              `json:"total_employees"`
	TodayEmployees    int64              `json:"today_employees"`
	WorkingEmployees  int64              `json:"working_employees"`
	FinishedEmployees int64              `json:"finished_employees"`
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

type WorkSessionDetailResponse struct {
	ID                   uuid.UUID  `json:"id"`
	EmployeeID           uuid.UUID  `json:"employee_id"`
	EmployeeName         string     `json:"employee_name"`
	PersonalImage        string     `json:"personal_image"`
	NationalID           string     `json:"national_id"`
	KeyNumber            string     `json:"key_number"`
	BranchName           string     `json:"branch_name"`
	StartTime            time.Time  `json:"start_time"`
	EndTime              *time.Time `json:"end_time"`
	WorkingDuration      string     `json:"working_duration"`
	StartKM              float64    `json:"start_km"`
	StartKMImage         string     `json:"start_km_image"`
	EndKM                float64    `json:"end_km"`
	EndKMImage           string     `json:"end_km_image"`
	Distance             float64    `json:"distance"`
	OrdersCount          int        `json:"orders_count"`
	FuelCost             float64    `json:"fuel_cost"`
	ApplicationID        string     `json:"application_id"`
	ApplicationType      string     `json:"application_type"`
	MotorcycleNumber     string     `json:"motorcycle_number"`
	IsReviewed           bool       `json:"is_reviewed"`
	ReviewNotes          string     `json:"review_notes"`
	IsEditedBySupervisor bool       `json:"is_edited_by_supervisor"`
	EditedByName         string     `json:"edited_by_name"`
	OriginalOrdersCount  int        `json:"original_orders_count"`
	OriginalEndKM        float64    `json:"original_end_km"`
	OriginalStartKM      float64    `json:"original_start_km"`
	Notes                string     `json:"notes"`
	Status               string     `json:"status"`
}

type AuditLogResponse struct {
	ID        uuid.UUID `json:"id"`
	AdminName string    `json:"admin_name"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

// Branch DTOs
type CreateBranchRequest struct {
	Name string `json:"name" binding:"required"`
}

type BranchResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	EmployeeCount int64     `json:"employee_count"`
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

// Daily Report DTOs
type DailyAppSummary struct {
	AppType     string  `json:"app_type"`
	AppName     string  `json:"app_name"`
	TotalOrders int     `json:"total_orders"`
	TotalKM     float64 `json:"total_km"`
	TotalFuel   float64 `json:"total_fuel"`
	Count       int     `json:"count"`
}

type DailyEmployeeRow struct {
	EmployeeID    string  `json:"employee_id"`
	EmployeeName  string  `json:"employee_name"`
	BranchName    string  `json:"branch_name"`
	KeyNumber     string  `json:"key_number"`
	AppType       string  `json:"app_type"`
	AppName       string  `json:"app_name"`
	SessionsCount int     `json:"sessions_count"`
	TotalKM       float64 `json:"total_km"`
	TotalOrders   int     `json:"total_orders"`
	TotalFuel     float64 `json:"total_fuel"`
}

type DailyReportResponse struct {
	Rows         []DailyEmployeeRow `json:"rows"`
	TotalOrders  int                `json:"total_orders"`
	TotalKM      float64            `json:"total_km"`
	TotalFuel    float64            `json:"total_fuel"`
	AppSummaries []DailyAppSummary  `json:"app_summaries"`
}

// Inventory DTOs
type CreateInventoryItemRequest struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"` // "oil" or "spare_part"
	Unit     string `json:"unit"`
	Barcode  string `json:"barcode"`
	Quantity int    `json:"quantity"`
	MinAlert int    `json:"min_alert"`
	Notes    string `json:"notes"`
}

type UpdateInventoryItemRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Unit     string `json:"unit"`
	Barcode  string `json:"barcode"`
	Quantity int    `json:"quantity"`
	MinAlert int    `json:"min_alert"`
	Notes    string `json:"notes"`
}

type InventoryTransactionRequest struct {
	ItemID     string  `json:"item_id" binding:"required,uuid"`
	Type       string  `json:"type" binding:"required"` // "in" or "out"
	Quantity   int     `json:"quantity" binding:"required,gte=1"`
	EmployeeID *string `json:"employee_id"`
	Notes      string  `json:"notes"`
}

type DispenseOilRequest struct {
	EmployeeID string `json:"employee_id" binding:"required,uuid"`
	Quantity   int    `json:"quantity" binding:"required,gte=1"` // number of oil containers
	Notes      string `json:"notes"`
}

// InventoryItemWithStock combines a global inventory item with branch-specific stock
type InventoryItemWithStock struct {
	*domain.InventoryItem
	BranchQuantity int `json:"branch_quantity"` // الكمية المتاحة في الفرع الحالي
}

// Purchase Invoice DTOs
type PurchaseInvoiceItemRequest struct {
	ItemID    string  `json:"item_id" binding:"required,uuid"`
	Quantity  int     `json:"quantity" binding:"required,gte=1"`
	UnitPrice float64 `json:"unit_price" binding:"gte=0"`
	Notes     string  `json:"notes"`
}

type CreatePurchaseInvoiceRequest struct {
	InvoiceNumber string                       `json:"invoice_number"`
	SupplierName  string                       `json:"supplier_name" binding:"required"`
	InvoiceDate   *time.Time                   `json:"invoice_date"`
	Subtotal      float64                      `json:"subtotal"`
	Discount      float64                      `json:"discount"`
	TaxRate       float64                      `json:"tax_rate"`
	TaxAmount     float64                      `json:"tax_amount"`
	TotalAmount   float64                      `json:"total_amount"`
	Notes         string                       `json:"notes"`
	Items         []PurchaseInvoiceItemRequest `json:"items" binding:"required,min=1,dive"`
}

// Maintenance DTOs
type MaintenanceLogResponse struct {
	ID           string    `json:"id"`
	EmployeeID   string    `json:"employee_id"`
	EmployeeName string    `json:"employee_name"`
	Type         string    `json:"type"`
	Details      string    `json:"details"`
	DistanceAt   float64   `json:"distance_at"`
	Cost         float64   `json:"cost"`
	AdminName    string    `json:"admin_name"`
	CreatedAt    time.Time `json:"created_at"`
}

type OilChangeCheckResponse struct {
	NeedsOilChange    bool    `json:"needs_oil_change"`
	TotalDistance     float64 `json:"total_distance"`
	DistanceSinceOil  float64 `json:"distance_since_oil"`
	OilChangeInterval float64 `json:"oil_change_interval"` // 950 للموتوسيكل، 10000 للسيارة
	VehicleType       string  `json:"vehicle_type"`        // "car" أو "motorcycle"
}

// Batch oil change setup
type OilSetupEntry struct {
	EmployeeID            string  `json:"employee_id" binding:"required,uuid"`
	LastOilChangeDistance float64 `json:"last_oil_change_distance"`
}

type BatchOilSetupRequest struct {
	Entries []OilSetupEntry `json:"entries" binding:"required,dive"`
}

// Investigation
type CreateInvestigationRequest struct {
	EmployeeID     string   `json:"employee_id" binding:"required,uuid"`
	Type           string   `json:"type"` // investigation, supervisor_report, advance, internet_advance, absence, custody
	Questions      []string `json:"questions"`
	Answers        []string `json:"answers"`
	ReportText     string   `json:"report_text"`
	Images         []string `json:"images"`
	Amount         *float64 `json:"amount"`
	StartDate      string   `json:"start_date"`
	EndDate        string   `json:"end_date"`
	Items          []string `json:"items"`
	IsGuilty       bool     `json:"is_guilty"`
	Notes          string   `json:"notes"`
	DeductionMonth string   `json:"deduction_month"`
}

type UpdateInvestigationRequest struct {
	EmployeeID     string   `json:"employee_id" binding:"required,uuid"`
	Type           string   `json:"type"`
	Questions      []string `json:"questions"`
	Answers        []string `json:"answers"`
	ReportText     string   `json:"report_text"`
	Images         []string `json:"images"`
	Amount         *float64 `json:"amount"`
	StartDate      string   `json:"start_date"`
	EndDate        string   `json:"end_date"`
	Items          []string `json:"items"`
	IsGuilty       bool     `json:"is_guilty"`
	Notes          string   `json:"notes"`
	DeductionMonth string   `json:"deduction_month"`
}

type InvestigationResponse struct {
	ID                 uuid.UUID  `json:"id"`
	EmployeeID         uuid.UUID  `json:"employee_id"`
	EmployeeName       string     `json:"employee_name"`
	NationalID         string     `json:"national_id"`
	SupervisorID       uuid.UUID  `json:"supervisor_id"`
	SupervisorName     string     `json:"supervisor_name"`
	Type               string     `json:"type"`
	Questions          []string   `json:"questions"`
	Answers            []string   `json:"answers"`
	ReportText         string     `json:"report_text"`
	Images             []string   `json:"images"`
	Amount             *float64   `json:"amount"`
	StartDate          *time.Time `json:"start_date"`
	EndDate            *time.Time `json:"end_date"`
	Items              []string   `json:"items"`
	IsGuilty           bool       `json:"is_guilty"`
	Notes              string     `json:"notes"`
	DeductionMonth     string     `json:"deduction_month"`
	Status             string     `json:"status"`
	ApprovedByName     string     `json:"approved_by_name"`
	ApprovedByUsername string     `json:"approved_by_username"`
	RejectedByName     string     `json:"rejected_by_name"`
	RejectedByUsername string     `json:"rejected_by_username"`
	ApprovedAt         *time.Time `json:"approved_at"`
	RejectedAt         *time.Time `json:"rejected_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

// UpdateInvestigationApprovalRequest يُستخدم لموافقة أو رفض السلفة وسلفة النت
type UpdateInvestigationApprovalRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
}

// Custody DTOs
type CustodyExpenseResponse struct {
	ID            uuid.UUID `json:"id"`
	CustodyDayID  uuid.UUID `json:"custody_day_id"`
	Category      string    `json:"category"`
	Amount        float64   `json:"amount"`
	RecipientName string    `json:"recipient_name"`
	CreatedAt     time.Time `json:"created_at"`
}

type CustodyTotals struct {
	Fuel       float64 `json:"fuel"`
	License    float64 `json:"license"`
	SpareParts float64 `json:"spare_parts"`
	Other      float64 `json:"other"`
}

type CustodyDayResponse struct {
	ID             uuid.UUID                `json:"id"`
	BranchID       *uuid.UUID               `json:"branch_id"`
	BranchName     string                   `json:"branch_name"`
	Date           string                   `json:"date"`
	OpeningBalance float64                  `json:"opening_balance"`
	AddedAmount    float64                  `json:"added_amount"`
	CustodyValue   float64                  `json:"custody_value"`
	TotalExpenses  float64                  `json:"total_expenses"`
	ClosingBalance float64                  `json:"closing_balance"`
	Totals         CustodyTotals            `json:"totals"`
	Expenses       []CustodyExpenseResponse `json:"expenses"`
	CreatedAt      time.Time                `json:"created_at"`
}

type CreateCustodyDayRequest struct {
	Date        string     `json:"date" binding:"required"`
	AddedAmount float64    `json:"added_amount"`
	BranchID    *uuid.UUID `json:"branch_id"`
}

type CreateCustodyExpenseRequest struct {
	Category      string  `json:"category" binding:"required"`
	Amount        float64 `json:"amount" binding:"gte=0"`
	RecipientName string  `json:"recipient_name"`
}

type AddCustodyAmountRequest struct {
	CustodyDayID uuid.UUID  `json:"custody_day_id" binding:"required"`
	AddedAmount  float64    `json:"added_amount" binding:"gt=0"`
	BranchID     *uuid.UUID `json:"branch_id"`
}

type CustodyLogFilter struct {
	BranchID   string `form:"branch_id"`
	Date       string `form:"date"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	ActionType string `form:"action_type"`
	CreatedBy  string `form:"created_by"`
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=50"`
}

// Vehicle DTOs (الدبابات والمركبات)
type CreateVehicleRequest struct {
	PlateNumber      string     `json:"plate_number" binding:"required"`
	VehicleType      string     `json:"vehicle_type"`
	Brand            string     `json:"brand"`
	ModelYear        string     `json:"model_year"`
	KeyNumber        string     `json:"key_number"`
	CurrentKM        float64    `json:"current_km"`
	LastOilChangeKM  float64    `json:"last_oil_change_km"`
	IsOdometerBroken bool       `json:"is_odometer_broken"`
	BranchID         *uuid.UUID `json:"branch_id"`
	Notes            string     `json:"notes"`
}

type UpdateVehicleRequest struct {
	PlateNumber      *string    `json:"plate_number"`
	VehicleType      *string    `json:"vehicle_type"`
	Brand            *string    `json:"brand"`
	ModelYear        *string    `json:"model_year"`
	KeyNumber        *string    `json:"key_number"`
	CurrentKM        *float64   `json:"current_km"`
	LastOilChangeKM  *float64   `json:"last_oil_change_km"`
	IsOdometerBroken *bool      `json:"is_odometer_broken"`
	Status           *string    `json:"status"`
	BranchID         *uuid.UUID `json:"branch_id"`
	Notes            *string    `json:"notes"`
}

type VehicleFilter struct {
	BranchID    *uuid.UUID `form:"branch_id"`
	VehicleType string     `form:"vehicle_type"`
	Status      string     `form:"status"`
	Search      string     `form:"search"`
	Page        int        `form:"page,default=1"`
	Limit       int        `form:"limit,default=50"`
}

type VehicleResponse struct {
	domain.Vehicle
	NeedsOilChange bool    `json:"needs_oil_change"`
	RemainingOilKM float64 `json:"remaining_oil_km"`
	CurrentDriver  *string `json:"current_driver,omitempty"`
}

// ------------------------------------------------------------------
// 1. FuelLog DTOs (سجلات الوقود)
// ------------------------------------------------------------------
type CreateFuelLogRequest struct {
	EmployeeID      *uuid.UUID `json:"employee_id"`
	VehiclePlate    string     `json:"vehicle_plate"`
	ShiftID         *uuid.UUID `json:"shift_id"`
	Amount          float64    `json:"amount" binding:"required"`
	Liters          float64    `json:"liters"`
	FuelDate        string     `json:"fuel_date"` // YYYY-MM-DD
	StationName     string     `json:"station_name"`
	InvoiceImageURL string     `json:"invoice_image_url"`
	BranchID        *uuid.UUID `json:"branch_id"`
	Notes           string     `json:"notes"`
}

type UpdateFuelLogRequest struct {
	EmployeeID      *uuid.UUID `json:"employee_id"`
	VehiclePlate    *string    `json:"vehicle_plate"`
	Amount          *float64   `json:"amount"`
	Liters          *float64   `json:"liters"`
	FuelDate        *string    `json:"fuel_date"`
	StationName     *string    `json:"station_name"`
	InvoiceImageURL *string    `json:"invoice_image_url"`
	Notes           *string    `json:"notes"`
}

type FuelLogFilter struct {
	BranchID   *uuid.UUID `form:"branch_id"`
	EmployeeID *uuid.UUID `form:"employee_id"`
	Plate      string     `form:"plate"`
	StartDate  string     `form:"start_date"`
	EndDate    string     `form:"end_date"`
	Search     string     `form:"search"`
	Page       int        `form:"page,default=1"`
	Limit      int        `form:"limit,default=50"`
}

// ------------------------------------------------------------------
// 2. TrafficViolation DTOs (المخالفات المرورية)
// ------------------------------------------------------------------
type CreateTrafficViolationRequest struct {
	ViolationNumber string     `json:"violation_number"`
	EmployeeID      *uuid.UUID `json:"employee_id"`
	VehiclePlate    string     `json:"vehicle_plate"`
	Amount          float64    `json:"amount" binding:"required"`
	Reason          string     `json:"reason" binding:"required"`
	ViolationDate   string     `json:"violation_date"`
	City            string     `json:"city"`
	Status          string     `json:"status"` // RECORDED, DEDUCTED, DISPUTED, PAID
	BranchID        *uuid.UUID `json:"branch_id"`
	Notes           string     `json:"notes"`
}

type UpdateTrafficViolationRequest struct {
	ViolationNumber *string    `json:"violation_number"`
	EmployeeID      *uuid.UUID `json:"employee_id"`
	VehiclePlate    *string    `json:"vehicle_plate"`
	Amount          *float64   `json:"amount"`
	Reason          *string    `json:"reason"`
	ViolationDate   *string    `json:"violation_date"`
	City            *string    `json:"city"`
	Status          *string    `json:"status"`
	Notes           *string    `json:"notes"`
}

type TrafficViolationFilter struct {
	BranchID   *uuid.UUID `form:"branch_id"`
	EmployeeID *uuid.UUID `form:"employee_id"`
	Status     string     `form:"status"`
	Search     string     `form:"search"`
	StartDate  string     `form:"start_date"`
	EndDate    string     `form:"end_date"`
	Page       int        `form:"page,default=1"`
	Limit      int        `form:"limit,default=50"`
}

// ------------------------------------------------------------------
// 3. MaintenanceRequest DTOs (طلبات الصيانة)
// ------------------------------------------------------------------
type CreateMaintenanceRequestRequest struct {
	VehiclePlate     string     `json:"vehicle_plate" binding:"required"`
	EmployeeID       *uuid.UUID `json:"employee_id"`
	IssueDescription string     `json:"issue_description" binding:"required"`
	Priority         string     `json:"priority"` // LOW, MEDIUM, HIGH, URGENT
	EstimatedCost    float64    `json:"estimated_cost"`
	ActualCost       float64    `json:"actual_cost"`
	WorkshopName     string     `json:"workshop_name"`
	Status           string     `json:"status"`
	BranchID         *uuid.UUID `json:"branch_id"`
	Notes            string     `json:"notes"`
}

type UpdateMaintenanceRequestRequest struct {
	VehiclePlate     *string    `json:"vehicle_plate"`
	EmployeeID       *uuid.UUID `json:"employee_id"`
	IssueDescription *string    `json:"issue_description"`
	Priority         *string    `json:"priority"`
	EstimatedCost    *float64   `json:"estimated_cost"`
	ActualCost       *float64   `json:"actual_cost"`
	WorkshopName     *string    `json:"workshop_name"`
	Status           *string    `json:"status"`
	Notes            *string    `json:"notes"`
}

type MaintenanceRequestFilter struct {
	BranchID *uuid.UUID `form:"branch_id"`
	Plate    string     `form:"plate"`
	Priority string     `form:"priority"`
	Status   string     `form:"status"`
	Search   string     `form:"search"`
	Page     int        `form:"page,default=1"`
	Limit    int        `form:"limit,default=50"`
}

// ------------------------------------------------------------------
// 4. EmployeeDocument DTOs (المستندات والرخص)
// ------------------------------------------------------------------
type CreateEmployeeDocumentRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" binding:"required"`
	DocType    string    `json:"doc_type" binding:"required"`
	Title      string    `json:"title" binding:"required"`
	DocNumber  string    `json:"doc_number"`
	FileURL    string    `json:"file_url"`
	IssueDate  *string   `json:"issue_date"`
	ExpiryDate *string   `json:"expiry_date"`
	Status     string    `json:"status"`
	Notes      string    `json:"notes"`
}

type UpdateEmployeeDocumentRequest struct {
	DocType    *string `json:"doc_type"`
	Title      *string `json:"title"`
	DocNumber  *string `json:"doc_number"`
	FileURL    *string `json:"file_url"`
	IssueDate  *string `json:"issue_date"`
	ExpiryDate *string `json:"expiry_date"`
	Status     *string `json:"status"`
	Notes      *string `json:"notes"`
}

type EmployeeDocumentFilter struct {
	EmployeeID *uuid.UUID `form:"employee_id"`
	DocType    string     `form:"doc_type"`
	Status     string     `form:"status"`
	Search     string     `form:"search"`
	Page       int        `form:"page,default=1"`
	Limit      int        `form:"limit,default=50"`
}

// ------------------------------------------------------------------
// 5. EmployeeBankAccount DTOs (الحسابات البنكية)
// ------------------------------------------------------------------
type CreateEmployeeBankAccountRequest struct {
	EmployeeID       uuid.UUID `json:"employee_id" binding:"required"`
	BankName         string    `json:"bank_name" binding:"required"`
	IBAN             string    `json:"iban" binding:"required"`
	AccountOwnerName string    `json:"account_owner_name" binding:"required"`
	IsDefault        bool      `json:"is_default"`
}

type UpdateEmployeeBankAccountRequest struct {
	BankName         *string `json:"bank_name"`
	IBAN             *string `json:"iban"`
	AccountOwnerName *string `json:"account_owner_name"`
	IsDefault        *bool   `json:"is_default"`
}

type EmployeeBankAccountFilter struct {
	EmployeeID *uuid.UUID `form:"employee_id"`
	Search     string     `form:"search"`
	Page       int        `form:"page,default=1"`
	Limit      int        `form:"limit,default=50"`
}

// ------------------------------------------------------------------
// 6. LeaveRequest DTOs (طلبات الإجازات)
// ------------------------------------------------------------------
type CreateLeaveRequestRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" binding:"required"`
	LeaveType  string    `json:"leave_type"`
	StartDate  string    `json:"start_date" binding:"required"`
	EndDate    string    `json:"end_date" binding:"required"`
	DaysCount  int       `json:"days_count"`
	Reason     string    `json:"reason"`
}

type UpdateLeaveRequestStatusRequest struct {
	Status         string `json:"status" binding:"required"` // APPROVED, REJECTED, PENDING
	ApprovedByName string `json:"approved_by_name"`
}

type LeaveRequestFilter struct {
	EmployeeID *uuid.UUID `form:"employee_id"`
	Status     string     `form:"status"`
	LeaveType  string     `form:"leave_type"`
	Search     string     `form:"search"`
	Page       int        `form:"page,default=1"`
	Limit      int        `form:"limit,default=50"`
}

// ------------------------------------------------------------------
// 7. SupportTicket DTOs (تذاكر الدعم والشكاوى)
// ------------------------------------------------------------------
type CreateSupportTicketRequest struct {
	EmployeeID  *uuid.UUID `json:"employee_id"`
	Subject     string     `json:"subject" binding:"required"`
	Category    string     `json:"category"`
	Priority    string     `json:"priority"`
	Description string     `json:"description" binding:"required"`
	BranchID    *uuid.UUID `json:"branch_id"`
}

type UpdateSupportTicketRequest struct {
	Subject     *string `json:"subject"`
	Category    *string `json:"category"`
	Priority    *string `json:"priority"`
	Status      *string `json:"status"`
	Description *string `json:"description"`
	Resolution  *string `json:"resolution"`
}

type SupportTicketFilter struct {
	BranchID   *uuid.UUID `form:"branch_id"`
	EmployeeID *uuid.UUID `form:"employee_id"`
	Category   string     `form:"category"`
	Priority   string     `form:"priority"`
	Status     string     `form:"status"`
	Search     string     `form:"search"`
	Page       int        `form:"page,default=1"`
	Limit      int        `form:"limit,default=50"`
}

// ------------------------------------------------------------------
// 8. Archive & Trash DTOs (سجل الأرشيف والمحذوفات)
// ------------------------------------------------------------------
type ArchivedItemDTO struct {
	ID         uuid.UUID  `json:"id"`
	Type       string     `json:"type"` // "employees", "vehicles", "branches", "documents", "work_sessions"
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
	Type     string     `form:"type"` // "all", "employees", "vehicles", "branches", "documents", "work_sessions"
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



