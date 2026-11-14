package hr_legal

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HRLegalRepository interface {
	// Investigation
	CreateInvestigation(ctx context.Context, inv *Investigation) error
	UpdateInvestigation(ctx context.Context, inv *Investigation) error
	DeleteInvestigation(ctx context.Context, id uuid.UUID) error
	FindInvestigations(ctx context.Context, branchID *uuid.UUID) ([]Investigation, error)
	FindInvestigationByID(ctx context.Context, id uuid.UUID) (*Investigation, error)
	CountPendingInvestigations(ctx context.Context) (int64, error)

	// Document
	CreateDocument(ctx context.Context, doc *EmployeeDocument) error
	UpdateDocument(ctx context.Context, doc *EmployeeDocument) error
	DeleteDocument(ctx context.Context, id uuid.UUID) error
	FindDocumentByID(ctx context.Context, id uuid.UUID) (*EmployeeDocument, error)
	FindDocuments(ctx context.Context, filter EmployeeDocumentFilter) ([]EmployeeDocument, int64, error)
	FindExpiringDocuments(ctx context.Context, days int) ([]EmployeeDocument, error)

	// Bank Account
	CreateBankAccount(ctx context.Context, acc *EmployeeBankAccount) error
	UpdateBankAccount(ctx context.Context, acc *EmployeeBankAccount) error
	DeleteBankAccount(ctx context.Context, id uuid.UUID) error
	FindBankAccountByID(ctx context.Context, id uuid.UUID) (*EmployeeBankAccount, error)
	FindBankAccounts(ctx context.Context, filter EmployeeBankAccountFilter) ([]EmployeeBankAccount, int64, error)
	ClearDefaultBankAccount(ctx context.Context, employeeID uuid.UUID) error

	// Leave Request
	CreateLeaveRequest(ctx context.Context, req *LeaveRequest) error
	UpdateLeaveRequest(ctx context.Context, req *LeaveRequest) error
	DeleteLeaveRequest(ctx context.Context, id uuid.UUID) error
	FindLeaveRequestByID(ctx context.Context, id uuid.UUID) (*LeaveRequest, error)
	FindLeaveRequests(ctx context.Context, filter LeaveRequestFilter) ([]LeaveRequest, int64, error)
	CountPendingLeaveRequests(ctx context.Context) (int64, error)
	FindActiveLeave(ctx context.Context, empID uuid.UUID, date string) (*LeaveRequest, error)

	// Traffic Violation
	CreateViolation(ctx context.Context, v *TrafficViolation) error
	UpdateViolation(ctx context.Context, v *TrafficViolation) error
	DeleteViolation(ctx context.Context, id uuid.UUID) error
	FindViolationByID(ctx context.Context, id uuid.UUID) (*TrafficViolation, error)
	FindViolations(ctx context.Context, filter TrafficViolationFilter) ([]TrafficViolation, int64, float64, float64, error)
	GetEmployeeInfo(ctx context.Context, empID uuid.UUID) (*EmployeeInfo, error)
	SaveNotification(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, title, body, notifType string) error

	// Fuel Log
	CreateFuelLog(ctx context.Context, log *FuelLog) error
	UpdateFuelLog(ctx context.Context, log *FuelLog) error
	DeleteFuelLog(ctx context.Context, id uuid.UUID) error
	FindFuelLogByID(ctx context.Context, id uuid.UUID) (*FuelLog, error)
	FindFuelLogs(ctx context.Context, filter FuelLogFilter) ([]FuelLog, int64, error)
}

type gormHRLegalRepository struct {
	db *gorm.DB
}

func NewHRLegalRepository(db *gorm.DB) HRLegalRepository {
	return &gormHRLegalRepository{db: db}
}

func (r *gormHRLegalRepository) likeOp() string {
	if r.db.Dialector.Name() == "postgres" {
		return "ILIKE"
	}
	return "LIKE"
}

// 1. Investigation
func (r *gormHRLegalRepository) CreateInvestigation(ctx context.Context, inv *Investigation) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *gormHRLegalRepository) UpdateInvestigation(ctx context.Context, inv *Investigation) error {
	return r.db.WithContext(ctx).Save(inv).Error
}

func (r *gormHRLegalRepository) DeleteInvestigation(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Investigation{}).Error
}

func (r *gormHRLegalRepository) FindInvestigations(ctx context.Context, branchID *uuid.UUID) ([]Investigation, error) {
	var list []Investigation
	selectCols := `investigations.*,
		COALESCE(NULLIF(investigations.employee_name, ''), emp.name, '') AS employee_name,
		COALESCE(NULLIF(investigations.national_id, ''), emp.national_id, '') AS national_id,
		COALESCE(NULLIF(investigations.supervisor_name, ''), adm.name, sup_emp.name, '') AS supervisor_name`

	query := r.db.WithContext(ctx).
		Table("investigations").
		Select(selectCols).
		Joins("LEFT JOIN employees emp ON emp.id = investigations.employee_id").
		Joins("LEFT JOIN admins adm ON adm.id = investigations.supervisor_id").
		Joins("LEFT JOIN employees sup_emp ON sup_emp.id = investigations.supervisor_id").
		Where("investigations.deleted_at IS NULL").
		Order("investigations.created_at DESC")

	if branchID != nil {
		query = query.Where("(emp.branch_id = ? OR investigations.supervisor_id IN (SELECT id FROM admins WHERE branch_id = ?) OR investigations.supervisor_id IN (SELECT id FROM employees WHERE branch_id = ?))", *branchID, *branchID, *branchID)
	}

	err := query.Find(&list).Error
	return list, err
}

func (r *gormHRLegalRepository) FindInvestigationByID(ctx context.Context, id uuid.UUID) (*Investigation, error) {
	var inv Investigation
	selectCols := `investigations.*,
		COALESCE(NULLIF(investigations.employee_name, ''), emp.name, '') AS employee_name,
		COALESCE(NULLIF(investigations.national_id, ''), emp.national_id, '') AS national_id,
		COALESCE(NULLIF(investigations.supervisor_name, ''), adm.name, sup_emp.name, '') AS supervisor_name`

	err := r.db.WithContext(ctx).
		Table("investigations").
		Select(selectCols).
		Joins("LEFT JOIN employees emp ON emp.id = investigations.employee_id").
		Joins("LEFT JOIN admins adm ON adm.id = investigations.supervisor_id").
		Joins("LEFT JOIN employees sup_emp ON sup_emp.id = investigations.supervisor_id").
		Where("investigations.id = ? AND investigations.deleted_at IS NULL", id).
		First(&inv).Error

	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *gormHRLegalRepository) CountPendingInvestigations(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Investigation{}).
		Where("status = ? AND deleted_at IS NULL", "pending").
		Count(&count).Error
	return count, err
}

// 2. Document
func (r *gormHRLegalRepository) CreateDocument(ctx context.Context, doc *EmployeeDocument) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *gormHRLegalRepository) UpdateDocument(ctx context.Context, doc *EmployeeDocument) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *gormHRLegalRepository) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&EmployeeDocument{}, "id = ?", id).Error
}

func (r *gormHRLegalRepository) FindDocumentByID(ctx context.Context, id uuid.UUID) (*EmployeeDocument, error) {
	var doc EmployeeDocument
	if err := r.db.WithContext(ctx).First(&doc, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *gormHRLegalRepository) FindDocuments(ctx context.Context, filter EmployeeDocumentFilter) ([]EmployeeDocument, int64, error) {
	var list []EmployeeDocument
	var total int64
	db := r.db.WithContext(ctx).Model(&EmployeeDocument{})

	if filter.EmployeeID != nil {
		db = db.Where("employee_id = ?", *filter.EmployeeID)
	}
	if filter.DocType != "" {
		db = db.Where("doc_type = ?", filter.DocType)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		like := r.likeOp()
		db = db.Where(fmt.Sprintf("title %s ? OR doc_number %s ?", like, like), s, s)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.GetEffectiveLimit()
	page := filter.GetEffectivePage()
	offset := (page - 1) * limit

	err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *gormHRLegalRepository) FindExpiringDocuments(ctx context.Context, days int) ([]EmployeeDocument, error) {
	var list []EmployeeDocument
	futureDate := time.Now().AddDate(0, 0, days)
	err := r.db.WithContext(ctx).
		Where("expiry_date IS NOT NULL AND expiry_date <= ? AND status != 'EXPIRED'", futureDate).
		Order("expiry_date ASC").
		Find(&list).Error
	return list, err
}

// 3. Bank Account
func (r *gormHRLegalRepository) CreateBankAccount(ctx context.Context, acc *EmployeeBankAccount) error {
	return r.db.WithContext(ctx).Create(acc).Error
}

func (r *gormHRLegalRepository) UpdateBankAccount(ctx context.Context, acc *EmployeeBankAccount) error {
	return r.db.WithContext(ctx).Save(acc).Error
}

func (r *gormHRLegalRepository) DeleteBankAccount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&EmployeeBankAccount{}, "id = ?", id).Error
}

func (r *gormHRLegalRepository) FindBankAccountByID(ctx context.Context, id uuid.UUID) (*EmployeeBankAccount, error) {
	var acc EmployeeBankAccount
	if err := r.db.WithContext(ctx).First(&acc, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *gormHRLegalRepository) FindBankAccounts(ctx context.Context, filter EmployeeBankAccountFilter) ([]EmployeeBankAccount, int64, error) {
	var list []EmployeeBankAccount
	var total int64
	db := r.db.WithContext(ctx).Model(&EmployeeBankAccount{})

	if filter.EmployeeID != nil {
		db = db.Where("employee_id = ?", *filter.EmployeeID)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		like := r.likeOp()
		db = db.Where(fmt.Sprintf("bank_name %s ? OR iban %s ? OR account_owner_name %s ?", like, like, like), s, s, s)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.GetEffectiveLimit()
	page := filter.GetEffectivePage()
	offset := (page - 1) * limit

	err := db.Order("is_default DESC, created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *gormHRLegalRepository) ClearDefaultBankAccount(ctx context.Context, employeeID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&EmployeeBankAccount{}).
		Where("employee_id = ?", employeeID).
		Update("is_default", false).Error
}

// 4. Leave Request
func (r *gormHRLegalRepository) CreateLeaveRequest(ctx context.Context, req *LeaveRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *gormHRLegalRepository) UpdateLeaveRequest(ctx context.Context, req *LeaveRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *gormHRLegalRepository) DeleteLeaveRequest(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&LeaveRequest{}, "id = ?", id).Error
}

func (r *gormHRLegalRepository) FindLeaveRequestByID(ctx context.Context, id uuid.UUID) (*LeaveRequest, error) {
	var req LeaveRequest
	if err := r.db.WithContext(ctx).First(&req, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *gormHRLegalRepository) FindLeaveRequests(ctx context.Context, filter LeaveRequestFilter) ([]LeaveRequest, int64, error) {
	var list []LeaveRequest
	var total int64
	db := r.db.WithContext(ctx).Model(&LeaveRequest{})

	if filter.EmployeeID != nil {
		db = db.Where("employee_id = ?", *filter.EmployeeID)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.LeaveType != "" {
		db = db.Where("leave_type = ?", filter.LeaveType)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.GetEffectiveLimit()
	page := filter.GetEffectivePage()
	offset := (page - 1) * limit

	err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *gormHRLegalRepository) CountPendingLeaveRequests(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&LeaveRequest{}).Where("status = ?", "PENDING").Count(&count).Error
	return count, err
}

func (r *gormHRLegalRepository) FindActiveLeave(ctx context.Context, empID uuid.UUID, date string) (*LeaveRequest, error) {
	var req LeaveRequest
	err := r.db.WithContext(ctx).
		Where("employee_id = ? AND status = 'APPROVED' AND start_date <= ? AND end_date >= ?", empID, date, date).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// 5. Traffic Violation
func (r *gormHRLegalRepository) CreateViolation(ctx context.Context, v *TrafficViolation) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *gormHRLegalRepository) UpdateViolation(ctx context.Context, v *TrafficViolation) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *gormHRLegalRepository) DeleteViolation(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&TrafficViolation{}, "id = ?", id).Error
}

func (r *gormHRLegalRepository) FindViolationByID(ctx context.Context, id uuid.UUID) (*TrafficViolation, error) {
	var v TrafficViolation
	if err := r.db.WithContext(ctx).First(&v, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *gormHRLegalRepository) FindViolations(ctx context.Context, filter TrafficViolationFilter) ([]TrafficViolation, int64, float64, float64, error) {
	var list []TrafficViolation
	var total int64
	var totalAmount float64
	var deductedAmount float64

	db := r.db.WithContext(ctx).Model(&TrafficViolation{})

	if filter.EmployeeID != nil {
		db = db.Where("traffic_violations.employee_id = ?", *filter.EmployeeID)
	} else if filter.BranchID != nil {
		db = db.Where("traffic_violations.branch_id = ?", *filter.BranchID)
	}

	if filter.Status != "" {
		db = db.Where("traffic_violations.status = ?", filter.Status)
	}
	if filter.StartDate != "" {
		db = db.Where("traffic_violations.violation_date >= ?", filter.StartDate+" 00:00:00")
	}
	if filter.EndDate != "" {
		db = db.Where("traffic_violations.violation_date <= ?", filter.EndDate+" 23:59:59")
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		like := r.likeOp()
		db = db.Where(fmt.Sprintf("traffic_violations.violation_number %s ? OR traffic_violations.vehicle_plate %s ? OR traffic_violations.reason %s ?", like, like, like), s, s, s)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, 0, 0, err
	}

	type Sums struct {
		TotalAmount    float64 `gorm:"column:total_amount"`
		DeductedAmount float64 `gorm:"column:deducted_amount"`
	}
	var sums Sums
	_ = db.Session(&gorm.Session{}).
		Select("COALESCE(SUM(amount), 0) AS total_amount, COALESCE(SUM(paid_amount), 0) AS deducted_amount").
		Scan(&sums).Error
	totalAmount = sums.TotalAmount
	deductedAmount = sums.DeductedAmount

	limit := filter.GetEffectiveLimit()
	page := filter.GetEffectivePage()
	offset := (page - 1) * limit

	err := db.Session(&gorm.Session{}).
		Order("traffic_violations.violation_date DESC, traffic_violations.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&list).Error
	return list, total, totalAmount, deductedAmount, err
}

func (r *gormHRLegalRepository) GetEmployeeInfo(ctx context.Context, empID uuid.UUID) (*EmployeeInfo, error) {
	var emp EmployeeInfo
	err := r.db.WithContext(ctx).Table("employees").Where("id = ? AND deleted_at IS NULL", empID).First(&emp).Error
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *gormHRLegalRepository) SaveNotification(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, title, body, notifType string) error {
	return r.db.WithContext(ctx).Table("notifications").Create(map[string]interface{}{
		"id":          uuid.New(),
		"employee_id": empID,
		"branch_id":   branchID,
		"title":       title,
		"body":        body,
		"type":        notifType,
		"status":      "unread",
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}).Error
}

// 6. Fuel Log
func (r *gormHRLegalRepository) CreateFuelLog(ctx context.Context, log *FuelLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *gormHRLegalRepository) UpdateFuelLog(ctx context.Context, log *FuelLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

func (r *gormHRLegalRepository) DeleteFuelLog(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&FuelLog{}, "id = ?", id).Error
}

func (r *gormHRLegalRepository) FindFuelLogByID(ctx context.Context, id uuid.UUID) (*FuelLog, error) {
	var log FuelLog
	if err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *gormHRLegalRepository) FindFuelLogs(ctx context.Context, filter FuelLogFilter) ([]FuelLog, int64, error) {
	var list []FuelLog
	var total int64
	db := r.db.WithContext(ctx).Model(&FuelLog{})

	if filter.BranchID != nil {
		db = db.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.EmployeeID != nil {
		db = db.Where("employee_id = ?", *filter.EmployeeID)
	}
	if filter.Plate != "" {
		db = db.Where("vehicle_plate = ?", filter.Plate)
	}
	if filter.StartDate != "" {
		db = db.Where("fuel_date >= ?", filter.StartDate+" 00:00:00")
	}
	if filter.EndDate != "" {
		db = db.Where("fuel_date <= ?", filter.EndDate+" 23:59:59")
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		like := r.likeOp()
		db = db.Where(fmt.Sprintf("station_name %s ? OR vehicle_plate %s ?", like, like), s, s)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.GetEffectiveLimit()
	page := filter.GetEffectivePage()
	offset := (page - 1) * limit

	err := db.Order("fuel_date DESC, created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}
