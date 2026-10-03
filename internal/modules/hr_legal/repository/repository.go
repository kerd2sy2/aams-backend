package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/hr_legal/domain"
	"delivery-backend/internal/modules/hr_legal/dto"
)

type HRLegalRepository interface {
	// Investigation
	CreateInvestigation(ctx context.Context, inv *domain.Investigation) error
	UpdateInvestigation(ctx context.Context, inv *domain.Investigation) error
	FindInvestigations(ctx context.Context, branchID *uuid.UUID) ([]domain.Investigation, error)
	FindInvestigationByID(ctx context.Context, id uuid.UUID) (*domain.Investigation, error)
	CountPendingInvestigations(ctx context.Context) (int64, error)

	// Document
	CreateDocument(ctx context.Context, doc *domain.EmployeeDocument) error
	UpdateDocument(ctx context.Context, doc *domain.EmployeeDocument) error
	DeleteDocument(ctx context.Context, id uuid.UUID) error
	FindDocumentByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeDocument, error)
	FindDocuments(ctx context.Context, filter dto.EmployeeDocumentFilter) ([]domain.EmployeeDocument, int64, error)
	FindExpiringDocuments(ctx context.Context, days int) ([]domain.EmployeeDocument, error)

	// Bank Account
	CreateBankAccount(ctx context.Context, acc *domain.EmployeeBankAccount) error
	UpdateBankAccount(ctx context.Context, acc *domain.EmployeeBankAccount) error
	DeleteBankAccount(ctx context.Context, id uuid.UUID) error
	FindBankAccountByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeBankAccount, error)
	FindBankAccounts(ctx context.Context, filter dto.EmployeeBankAccountFilter) ([]domain.EmployeeBankAccount, int64, error)
	ClearDefaultBankAccount(ctx context.Context, employeeID uuid.UUID) error

	// Leave Request
	CreateLeaveRequest(ctx context.Context, req *domain.LeaveRequest) error
	UpdateLeaveRequest(ctx context.Context, req *domain.LeaveRequest) error
	DeleteLeaveRequest(ctx context.Context, id uuid.UUID) error
	FindLeaveRequestByID(ctx context.Context, id uuid.UUID) (*domain.LeaveRequest, error)
	FindLeaveRequests(ctx context.Context, filter dto.LeaveRequestFilter) ([]domain.LeaveRequest, int64, error)
	CountPendingLeaveRequests(ctx context.Context) (int64, error)
	FindActiveLeave(ctx context.Context, empID uuid.UUID, date string) (*domain.LeaveRequest, error)

	// Traffic Violation
	CreateViolation(ctx context.Context, v *domain.TrafficViolation) error
	UpdateViolation(ctx context.Context, v *domain.TrafficViolation) error
	DeleteViolation(ctx context.Context, id uuid.UUID) error
	FindViolationByID(ctx context.Context, id uuid.UUID) (*domain.TrafficViolation, error)
	FindViolations(ctx context.Context, filter dto.TrafficViolationFilter) ([]domain.TrafficViolation, int64, float64, float64, error)
	GetEmployeeInfo(ctx context.Context, empID uuid.UUID) (*domain.EmployeeInfo, error)
	SaveNotification(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, title, body, notifType string) error

	// Fuel Log
	CreateFuelLog(ctx context.Context, log *domain.FuelLog) error
	UpdateFuelLog(ctx context.Context, log *domain.FuelLog) error
	DeleteFuelLog(ctx context.Context, id uuid.UUID) error
	FindFuelLogByID(ctx context.Context, id uuid.UUID) (*domain.FuelLog, error)
	FindFuelLogs(ctx context.Context, filter dto.FuelLogFilter) ([]domain.FuelLog, int64, error)
}

type gormHRLegalRepository struct {
	db *gorm.DB
}

func NewHRLegalRepository(db *gorm.DB) HRLegalRepository {
	return &gormHRLegalRepository{db: db}
}

// 1. Investigation
func (r *gormHRLegalRepository) CreateInvestigation(ctx context.Context, inv *domain.Investigation) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *gormHRLegalRepository) UpdateInvestigation(ctx context.Context, inv *domain.Investigation) error {
	return r.db.WithContext(ctx).Save(inv).Error
}

func (r *gormHRLegalRepository) FindInvestigations(ctx context.Context, branchID *uuid.UUID) ([]domain.Investigation, error) {
	var list []domain.Investigation
	query := r.db.WithContext(ctx).Order("created_at DESC")
	if branchID != nil {
		query = query.Joins("JOIN employees ON employees.id = investigations.employee_id").
			Where("employees.branch_id = ?", *branchID)
	}
	err := query.Find(&list).Error
	return list, err
}

func (r *gormHRLegalRepository) FindInvestigationByID(ctx context.Context, id uuid.UUID) (*domain.Investigation, error) {
	var inv domain.Investigation
	if err := r.db.WithContext(ctx).First(&inv, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *gormHRLegalRepository) CountPendingInvestigations(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Investigation{}).Where("status = ?", "pending").Count(&count).Error
	return count, err
}

// 2. Document
func (r *gormHRLegalRepository) CreateDocument(ctx context.Context, doc *domain.EmployeeDocument) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *gormHRLegalRepository) UpdateDocument(ctx context.Context, doc *domain.EmployeeDocument) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *gormHRLegalRepository) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.EmployeeDocument{}, "id = ?", id).Error
}

func (r *gormHRLegalRepository) FindDocumentByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeDocument, error) {
	var doc domain.EmployeeDocument
	if err := r.db.WithContext(ctx).First(&doc, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *gormHRLegalRepository) FindDocuments(ctx context.Context, filter dto.EmployeeDocumentFilter) ([]domain.EmployeeDocument, int64, error) {
	var list []domain.EmployeeDocument
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.EmployeeDocument{})

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
		db = db.Where("title ILIKE ? OR doc_number ILIKE ?", s, s)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.GetEffectiveLimit()
	page := filter.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *gormHRLegalRepository) FindExpiringDocuments(ctx context.Context, days int) ([]domain.EmployeeDocument, error) {
	var list []domain.EmployeeDocument
	futureDate := time.Now().AddDate(0, 0, days)
	err := r.db.WithContext(ctx).
		Where("expiry_date IS NOT NULL AND expiry_date <= ? AND status != 'EXPIRED'", futureDate).
		Order("expiry_date ASC").
		Find(&list).Error
	return list, err
}

// 3. Bank Account
func (r *gormHRLegalRepository) CreateBankAccount(ctx context.Context, acc *domain.EmployeeBankAccount) error {
	return r.db.WithContext(ctx).Create(acc).Error
}

func (r *gormHRLegalRepository) UpdateBankAccount(ctx context.Context, acc *domain.EmployeeBankAccount) error {
	return r.db.WithContext(ctx).Save(acc).Error
}

func (r *gormHRLegalRepository) DeleteBankAccount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.EmployeeBankAccount{}, "id = ?", id).Error
}

func (r *gormHRLegalRepository) FindBankAccountByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeBankAccount, error) {
	var acc domain.EmployeeBankAccount
	if err := r.db.WithContext(ctx).First(&acc, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *gormHRLegalRepository) FindBankAccounts(ctx context.Context, filter dto.EmployeeBankAccountFilter) ([]domain.EmployeeBankAccount, int64, error) {
	var list []domain.EmployeeBankAccount
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.EmployeeBankAccount{})

	if filter.EmployeeID != nil {
		db = db.Where("employee_id = ?", *filter.EmployeeID)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		db = db.Where("bank_name ILIKE ? OR iban ILIKE ? OR account_owner_name ILIKE ?", s, s, s)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.GetEffectiveLimit()
	page := filter.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	err := db.Order("is_default DESC, created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *gormHRLegalRepository) ClearDefaultBankAccount(ctx context.Context, employeeID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.EmployeeBankAccount{}).
		Where("employee_id = ?", employeeID).
		Update("is_default", false).Error
}

// 4. Leave Request
func (r *gormHRLegalRepository) CreateLeaveRequest(ctx context.Context, req *domain.LeaveRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *gormHRLegalRepository) UpdateLeaveRequest(ctx context.Context, req *domain.LeaveRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *gormHRLegalRepository) DeleteLeaveRequest(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.LeaveRequest{}, "id = ?", id).Error
}

func (r *gormHRLegalRepository) FindLeaveRequestByID(ctx context.Context, id uuid.UUID) (*domain.LeaveRequest, error) {
	var req domain.LeaveRequest
	if err := r.db.WithContext(ctx).First(&req, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *gormHRLegalRepository) FindLeaveRequests(ctx context.Context, filter dto.LeaveRequestFilter) ([]domain.LeaveRequest, int64, error) {
	var list []domain.LeaveRequest
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.LeaveRequest{})

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
	page := filter.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *gormHRLegalRepository) CountPendingLeaveRequests(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.LeaveRequest{}).Where("status = ?", "PENDING").Count(&count).Error
	return count, err
}

func (r *gormHRLegalRepository) FindActiveLeave(ctx context.Context, empID uuid.UUID, date string) (*domain.LeaveRequest, error) {
	var req domain.LeaveRequest
	err := r.db.WithContext(ctx).
		Where("employee_id = ? AND status = 'APPROVED' AND start_date <= ? AND end_date >= ?", empID, date, date).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// 5. Traffic Violation
func (r *gormHRLegalRepository) CreateViolation(ctx context.Context, v *domain.TrafficViolation) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *gormHRLegalRepository) UpdateViolation(ctx context.Context, v *domain.TrafficViolation) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *gormHRLegalRepository) DeleteViolation(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.TrafficViolation{}, "id = ?", id).Error
}

func (r *gormHRLegalRepository) FindViolationByID(ctx context.Context, id uuid.UUID) (*domain.TrafficViolation, error) {
	var v domain.TrafficViolation
	if err := r.db.WithContext(ctx).First(&v, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *gormHRLegalRepository) FindViolations(ctx context.Context, filter dto.TrafficViolationFilter) ([]domain.TrafficViolation, int64, float64, float64, error) {
	var list []domain.TrafficViolation
	var total int64
	var totalAmount float64
	var deductedAmount float64

	db := r.db.WithContext(ctx).Model(&domain.TrafficViolation{})

	if filter.EmployeeID != nil {
		var emp struct {
			MotorcycleNumber string `gorm:"column:motorcycle_number"`
		}
		if err := r.db.WithContext(ctx).Table("employees").Select("motorcycle_number").Where("id = ? AND deleted_at IS NULL", *filter.EmployeeID).First(&emp).Error; err == nil && strings.TrimSpace(emp.MotorcycleNumber) != "" {
			db = db.Where("traffic_violations.employee_id = ? OR (traffic_violations.employee_id IS NULL AND traffic_violations.vehicle_plate = ?)", *filter.EmployeeID, strings.TrimSpace(emp.MotorcycleNumber))
		} else {
			db = db.Where("traffic_violations.employee_id = ?", *filter.EmployeeID)
		}
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
		db = db.Where("traffic_violations.violation_number ILIKE ? OR traffic_violations.vehicle_plate ILIKE ? OR traffic_violations.reason ILIKE ?", s, s, s)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, 0, 0, err
	}

	type Sums struct {
		TotalAmount    float64 `gorm:"column:total_amount"`
		DeductedAmount float64 `gorm:"column:deducted_amount"`
	}
	var sums Sums
	_ = db.Select("COALESCE(SUM(amount), 0) AS total_amount, COALESCE(SUM(paid_amount), 0) AS deducted_amount").Scan(&sums).Error
	totalAmount = sums.TotalAmount
	deductedAmount = sums.DeductedAmount

	limit := filter.GetEffectiveLimit()
	page := filter.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	err := db.Order("traffic_violations.violation_date DESC, traffic_violations.created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, totalAmount, deductedAmount, err
}

func (r *gormHRLegalRepository) GetEmployeeInfo(ctx context.Context, empID uuid.UUID) (*domain.EmployeeInfo, error) {
	var emp domain.EmployeeInfo
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
func (r *gormHRLegalRepository) CreateFuelLog(ctx context.Context, log *domain.FuelLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *gormHRLegalRepository) UpdateFuelLog(ctx context.Context, log *domain.FuelLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

func (r *gormHRLegalRepository) DeleteFuelLog(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.FuelLog{}, "id = ?", id).Error
}

func (r *gormHRLegalRepository) FindFuelLogByID(ctx context.Context, id uuid.UUID) (*domain.FuelLog, error) {
	var log domain.FuelLog
	if err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *gormHRLegalRepository) FindFuelLogs(ctx context.Context, filter dto.FuelLogFilter) ([]domain.FuelLog, int64, error) {
	var list []domain.FuelLog
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.FuelLog{})

	if filter.BranchID != nil {
		db = db.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.EmployeeID != nil {
		db = db.Where("employee_id = ?", *filter.EmployeeID)
	}
	if filter.Plate != "" {
		db = db.Where("vehicle_plate = ?", filter.Plate)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		db = db.Where("station_name ILIKE ? OR vehicle_plate ILIKE ?", s, s)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.GetEffectiveLimit()
	page := filter.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	err := db.Order("fuel_date DESC, created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}
