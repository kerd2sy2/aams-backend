package repository

import (
	"context"
	"time"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/dto"

	"github.com/google/uuid"
)

type RoleRepository interface {
	FindAll(ctx context.Context) ([]domain.Role, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	FindByName(ctx context.Context, name string) (*domain.Role, error)
	Create(ctx context.Context, role *domain.Role) error
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountUsersByRoleID(ctx context.Context, roleID uuid.UUID) (int64, error)
}

type AdminRepository interface {
	FindByEmail(ctx context.Context, email string) (*domain.Admin, error)
	FindByUsername(ctx context.Context, username string) (*domain.Admin, error)
	FindByPhone(ctx context.Context, phone string) (*domain.Admin, error)
	FindByLogin(ctx context.Context, login string) (*domain.Admin, error)
	FindByGoogleEmail(ctx context.Context, email string) (*domain.Admin, error)
	FindByGoogleID(ctx context.Context, googleID string) (*domain.Admin, error)
	UpdateGoogleLink(ctx context.Context, adminID uuid.UUID, googleEmail, googleID, googleAvatar string) error
	UnlinkGoogle(ctx context.Context, adminID uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error)
	Create(ctx context.Context, admin *domain.Admin) error
	Update(ctx context.Context, admin *domain.Admin) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]domain.Admin, error)
}


type WorkRepository interface {
	CreateSession(ctx context.Context, session *domain.WorkSession) error
	UpdateSession(ctx context.Context, session *domain.WorkSession) error
	FindActiveSessionByEmployeeID(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error)
	FindActiveSessionForToday(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error)
	CountTodaySessions(ctx context.Context, empID uuid.UUID) (int64, error)
	FindLastCompletedSession(ctx context.Context, empID uuid.UUID) (*domain.WorkSession, error)
	FindSessionByID(ctx context.Context, id uuid.UUID) (*domain.WorkSession, error)
	GetReports(ctx context.Context, filter dto.ReportFilter) ([]domain.WorkSession, int64, error)
	GetDailyReport(ctx context.Context, branchID *uuid.UUID, date time.Time) ([]dto.DailyEmployeeRow, error)
	GetDashboardStats(ctx context.Context, branchID *uuid.UUID) (*dto.DashboardResponse, error)
	FindAllActiveEmployeeIDs(ctx context.Context) ([]uuid.UUID, error)
}

type AuditRepository interface {
	CreateLog(ctx context.Context, log *domain.AuditLog) error
	GetLatestLogs(ctx context.Context, limit int) ([]domain.AuditLog, error)
	GetLatestLogsByBranch(ctx context.Context, branchID uuid.UUID, limit int) ([]domain.AuditLog, error)
	FindAll(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	DeleteBulk(ctx context.Context, ids []uuid.UUID) error
	ClearAll(ctx context.Context, branchID *uuid.UUID) error
}

type BranchRepository interface {
	FindAll(ctx context.Context) ([]domain.Branch, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error)
	Create(ctx context.Context, branch *domain.Branch) error
	Update(ctx context.Context, branch *domain.Branch) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type InventoryRepository interface {
	CreateItem(ctx context.Context, item *domain.InventoryItem) error
	UpdateItem(ctx context.Context, item *domain.InventoryItem) error
	DeleteItem(ctx context.Context, id uuid.UUID) error
	FindItemByID(ctx context.Context, id uuid.UUID) (*domain.InventoryItem, error)
	FindByBarcode(ctx context.Context, barcode string) (*domain.InventoryItem, error)
	FindAllItems(ctx context.Context, itemType string) ([]domain.InventoryItem, error)
	CreateTransaction(ctx context.Context, tx *domain.InventoryTransaction) error
	FindTransactions(ctx context.Context, itemID *uuid.UUID, branchID *uuid.UUID, page, limit int) ([]domain.InventoryTransaction, int64, error)
	GetStockByBranch(ctx context.Context, branchID *uuid.UUID) (map[uuid.UUID]int, error) // itemID -> quantity
	GetItemStock(ctx context.Context, itemID uuid.UUID, branchID *uuid.UUID) (int, error)
	FindOilItemsWithStock(ctx context.Context, branchID *uuid.UUID) ([]domain.InventoryItem, error)
	DeleteAllTransactions(ctx context.Context) error
	CreatePurchaseInvoice(ctx context.Context, invoice *domain.PurchaseInvoice, stockTxs []*domain.InventoryTransaction) error
	FindPurchaseInvoices(ctx context.Context, branchID *uuid.UUID, search string, page, limit int) ([]domain.PurchaseInvoice, int64, error)
	FindPurchaseInvoiceByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseInvoice, error)
	DeletePurchaseInvoice(ctx context.Context, id uuid.UUID) error
}

type MaintenanceRepository interface {
	CreateLog(ctx context.Context, log *domain.MaintenanceLog) error
	FindByEmployeeID(ctx context.Context, empID uuid.UUID, limit int) ([]domain.MaintenanceLog, error)
	FindAll(ctx context.Context, page, limit int) ([]domain.MaintenanceLog, int64, error)
	FindLastOilChange(ctx context.Context, empID uuid.UUID) (*domain.MaintenanceLog, error)
}

type EmployeeRepository interface {
	Create(ctx context.Context, emp *domain.Employee) error
	Update(ctx context.Context, emp *domain.Employee) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error)
	FindByNationalID(ctx context.Context, nationalID string) (*domain.Employee, error)
	FindBySearchTerm(ctx context.Context, term string, branchID *uuid.UUID) ([]domain.Employee, error)
	FindAll(ctx context.Context, filter dto.EmployeeFilter) ([]domain.Employee, int64, error)
	CountAll(ctx context.Context, branchID *uuid.UUID) (int64, error)
	UpdateLocation(ctx context.Context, id uuid.UUID, lat, lng float64, isVPN, isMock, outOfZone bool) error
	GetLocations(ctx context.Context, branchID *uuid.UUID) ([]domain.Employee, error)
}

// InvestigationRepository interface
type InvestigationRepository interface {
	Create(ctx context.Context, investigation *domain.Investigation) error
	Update(ctx context.Context, investigation *domain.Investigation) error
	FindAll(ctx context.Context) ([]domain.Investigation, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Investigation, error)
	CountPendingApprovals(ctx context.Context) (int64, error)
}

// AttendanceRepository interface
type AttendanceRepository interface {
	Upsert(ctx context.Context, attendance *domain.Attendance) error
	FindByDate(ctx context.Context, date string) ([]domain.Attendance, error)
	FindByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date string) (*domain.Attendance, error)
	DeleteByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date string) error
}

// CustodyRepository interface
type CustodyRepository interface {
	CreateDay(ctx context.Context, day *domain.CustodyDay) error
	UpdateDay(ctx context.Context, day *domain.CustodyDay) error
	FindDayByID(ctx context.Context, id uuid.UUID) (*domain.CustodyDay, error)
	FindDayByDate(ctx context.Context, branchID *uuid.UUID, date string) (*domain.CustodyDay, error)
	FindLastDay(ctx context.Context, branchID *uuid.UUID) (*domain.CustodyDay, error)
	FindAll(ctx context.Context, branchID *uuid.UUID) ([]domain.CustodyDay, error)
	CreateExpense(ctx context.Context, expense *domain.CustodyExpense) error
	DeleteExpense(ctx context.Context, id uuid.UUID) error
	FindExpenseByID(ctx context.Context, id uuid.UUID) (*domain.CustodyExpense, error)
	CreateLog(ctx context.Context, log *domain.CustodyLog) error
	FindLogs(ctx context.Context, filter dto.CustodyLogFilter) ([]domain.CustodyLog, int64, error)
	FindLogByID(ctx context.Context, id uuid.UUID) (*domain.CustodyLog, error)
	DeleteLog(ctx context.Context, id uuid.UUID) error
}

// SettingRepository interface for key-value app settings
type SettingRepository interface {
	GetAll(ctx context.Context) ([]domain.AppSetting, error)
	GetByKey(ctx context.Context, key string) (*domain.AppSetting, error)
	Upsert(ctx context.Context, setting *domain.AppSetting) error
	Update(ctx context.Context, setting *domain.AppSetting) error
	Create(ctx context.Context, setting *domain.AppSetting) error
}

// VehicleRepository interface for motorcycle & car assets (الثوابت / الدبابات)
type VehicleRepository interface {
	Create(ctx context.Context, vehicle *domain.Vehicle) error
	Update(ctx context.Context, vehicle *domain.Vehicle) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Vehicle, error)
	FindByPlateNumber(ctx context.Context, plateNumber string) (*domain.Vehicle, error)
	FindByPlateNumberUnscoped(ctx context.Context, plateNumber string) (*domain.Vehicle, error)
	RestoreVehicle(ctx context.Context, vehicle *domain.Vehicle) error
	FindAll(ctx context.Context, filter dto.VehicleFilter) ([]domain.Vehicle, int64, error)
	CountAll(ctx context.Context, branchID *uuid.UUID) (int64, error)
	UpdateOdometer(ctx context.Context, plateNumber string, newKM float64, distance float64) error
	RecordOilChange(ctx context.Context, id uuid.UUID, currentKM float64) error
	FindLatestVehicleKM(ctx context.Context, plateNumber string) (float64, error)
	SetVehicleStatus(ctx context.Context, plateNumber string, status string) error
}

// ------------------------------------------------------------------
// 1. FuelLogRepository
// ------------------------------------------------------------------
type FuelLogRepository interface {
	Create(ctx context.Context, log *domain.FuelLog) error
	Update(ctx context.Context, log *domain.FuelLog) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.FuelLog, error)
	FindAll(ctx context.Context, filter dto.FuelLogFilter) ([]domain.FuelLog, int64, error)
	GetFuelStats(ctx context.Context, branchID *uuid.UUID, startDate, endDate string) (totalCost float64, totalLiters float64, totalLogs int64, err error)
}

// ------------------------------------------------------------------
// 2. TrafficViolationRepository
// ------------------------------------------------------------------
type TrafficViolationRepository interface {
	Create(ctx context.Context, v *domain.TrafficViolation) error
	Update(ctx context.Context, v *domain.TrafficViolation) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.TrafficViolation, error)
	FindAll(ctx context.Context, filter dto.TrafficViolationFilter) ([]domain.TrafficViolation, int64, error)
	GetViolationStats(ctx context.Context, branchID *uuid.UUID) (totalAmount float64, deductedAmount float64, totalCount int64, err error)
}

// ------------------------------------------------------------------
// 3. MaintenanceRequestRepository
// ------------------------------------------------------------------
type MaintenanceRequestRepository interface {
	Create(ctx context.Context, req *domain.MaintenanceRequest) error
	Update(ctx context.Context, req *domain.MaintenanceRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.MaintenanceRequest, error)
	FindAll(ctx context.Context, filter dto.MaintenanceRequestFilter) ([]domain.MaintenanceRequest, int64, error)
}

// ------------------------------------------------------------------
// 4. EmployeeDocumentRepository
// ------------------------------------------------------------------
type EmployeeDocumentRepository interface {
	Create(ctx context.Context, doc *domain.EmployeeDocument) error
	Update(ctx context.Context, doc *domain.EmployeeDocument) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeDocument, error)
	FindAll(ctx context.Context, filter dto.EmployeeDocumentFilter) ([]domain.EmployeeDocument, int64, error)
	FindExpiringSoon(ctx context.Context, days int) ([]domain.EmployeeDocument, error)
}

// ------------------------------------------------------------------
// 5. EmployeeBankAccountRepository
// ------------------------------------------------------------------
type EmployeeBankAccountRepository interface {
	Create(ctx context.Context, acc *domain.EmployeeBankAccount) error
	Update(ctx context.Context, acc *domain.EmployeeBankAccount) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeBankAccount, error)
	FindAll(ctx context.Context, filter dto.EmployeeBankAccountFilter) ([]domain.EmployeeBankAccount, int64, error)
}

// ------------------------------------------------------------------
// 6. LeaveRequestRepository
// ------------------------------------------------------------------
type LeaveRequestRepository interface {
	Create(ctx context.Context, leave *domain.LeaveRequest) error
	Update(ctx context.Context, leave *domain.LeaveRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.LeaveRequest, error)
	FindAll(ctx context.Context, filter dto.LeaveRequestFilter) ([]domain.LeaveRequest, int64, error)
}

// ------------------------------------------------------------------
// 7. SupportTicketRepository
// ------------------------------------------------------------------
type SupportTicketRepository interface {
	Create(ctx context.Context, ticket *domain.SupportTicket) error
	Update(ctx context.Context, ticket *domain.SupportTicket) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.SupportTicket, error)
	FindAll(ctx context.Context, filter dto.SupportTicketFilter) ([]domain.SupportTicket, int64, error)
}





type NotificationRepository interface {
	FindUnreadByAdmin(ctx context.Context, adminID uuid.UUID, branchID *uuid.UUID) ([]domain.Notification, error)
	FindAllByAdmin(ctx context.Context, adminID uuid.UUID, branchID *uuid.UUID, status string) ([]domain.Notification, error)
	MarkAsRead(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, adminID uuid.UUID) error
	Create(ctx context.Context, notif *domain.Notification) error
	FindByEmployeeAndTypeAndDate(ctx context.Context, empID uuid.UUID, notifType string, date string) (*domain.Notification, error)
}

// ------------------------------------------------------------------
// 8. ArchiveRepository (سجل الأرشيف والمحذوفات)
// ------------------------------------------------------------------
type ArchiveRepository interface {
	GetArchivedItems(ctx context.Context, filter dto.ArchiveFilter) ([]dto.ArchivedItemDTO, int64, dto.ArchiveStatsDTO, error)
	Restore(ctx context.Context, itemType string, id uuid.UUID) error
	PermanentDelete(ctx context.Context, itemType string, id uuid.UUID) error
	BulkRestore(ctx context.Context, itemType string, ids []uuid.UUID) error
	BulkPermanentDelete(ctx context.Context, itemType string, ids []uuid.UUID) error
}

// ------------------------------------------------------------------
// 9. OTPRepository (رموز التحقق OTP وتوثيق الأجهزة)
// ------------------------------------------------------------------
type OTPRepository interface {
	Create(ctx context.Context, otp *domain.OTPRequest) error
	FindActiveByNationalID(ctx context.Context, nationalID string) (*domain.OTPRequest, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.OTPRequest, error)
	FindAll(ctx context.Context, query dto.OTPListQuery) ([]domain.OTPRequest, int64, error)
	MarkVerified(ctx context.Context, id uuid.UUID) error
	Cancel(ctx context.Context, id uuid.UUID) error
	InvalidatePreviousPending(ctx context.Context, nationalID string) error
}

