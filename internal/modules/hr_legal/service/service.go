package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"delivery-backend/internal/modules/hr_legal/contracts"
	"delivery-backend/internal/modules/hr_legal/domain"
	"delivery-backend/internal/modules/hr_legal/dto"
	"delivery-backend/internal/modules/hr_legal/repository"
)

type HRLegalService interface {
	// Investigation
	CreateInvestigation(ctx context.Context, req dto.CreateInvestigationRequest, supervisorID uuid.UUID) (*dto.InvestigationResponse, error)
	UpdateInvestigation(ctx context.Context, id uuid.UUID, req dto.UpdateInvestigationRequest) (*dto.InvestigationResponse, error)
	GetAllInvestigations(ctx context.Context, branchID *uuid.UUID) ([]dto.InvestigationResponse, error)
	GetInvestigationByID(ctx context.Context, id uuid.UUID) (*dto.InvestigationResponse, error)
	ApproveInvestigation(ctx context.Context, id uuid.UUID, adminID uuid.UUID, adminName, adminUsername string) (*dto.InvestigationResponse, error)
	RejectInvestigation(ctx context.Context, id uuid.UUID, adminID uuid.UUID, adminName, adminUsername string) (*dto.InvestigationResponse, error)
	GetPendingInvestigationCount(ctx context.Context) (int64, error)

	// Document
	CreateDocument(ctx context.Context, req dto.CreateEmployeeDocumentRequest) (*domain.EmployeeDocument, error)
	UpdateDocument(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeDocumentRequest) (*domain.EmployeeDocument, error)
	DeleteDocument(ctx context.Context, id uuid.UUID) error
	GetDocumentByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeDocument, error)
	GetAllDocuments(ctx context.Context, filter dto.EmployeeDocumentFilter) ([]domain.EmployeeDocument, int64, error)
	GetExpiringDocuments(ctx context.Context, days int) ([]domain.EmployeeDocument, error)

	// Bank Account
	CreateBankAccount(ctx context.Context, req dto.CreateEmployeeBankAccountRequest) (*domain.EmployeeBankAccount, error)
	UpdateBankAccount(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeBankAccountRequest) (*domain.EmployeeBankAccount, error)
	DeleteBankAccount(ctx context.Context, id uuid.UUID) error
	GetBankAccountByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeBankAccount, error)
	GetAllBankAccounts(ctx context.Context, filter dto.EmployeeBankAccountFilter) ([]domain.EmployeeBankAccount, int64, error)

	// Leave Request
	CreateLeaveRequest(ctx context.Context, req dto.CreateLeaveRequestRequest) (*domain.LeaveRequest, error)
	UpdateLeaveRequestStatus(ctx context.Context, id uuid.UUID, req dto.UpdateLeaveRequestStatusRequest) (*domain.LeaveRequest, error)
	DeleteLeaveRequest(ctx context.Context, id uuid.UUID) error
	GetLeaveRequestByID(ctx context.Context, id uuid.UUID) (*domain.LeaveRequest, error)
	GetAllLeaveRequests(ctx context.Context, filter dto.LeaveRequestFilter) ([]domain.LeaveRequest, int64, error)
	GetPendingLeaveCount(ctx context.Context) (int64, error)

	// Traffic Violation
	CreateViolation(ctx context.Context, req dto.CreateTrafficViolationRequest) (*domain.TrafficViolation, error)
	UpdateViolation(ctx context.Context, id uuid.UUID, req dto.UpdateTrafficViolationRequest) (*domain.TrafficViolation, error)
	DeleteViolation(ctx context.Context, id uuid.UUID) error
	GetViolationByID(ctx context.Context, id uuid.UUID) (*domain.TrafficViolation, error)
	GetAllViolations(ctx context.Context, filter dto.TrafficViolationFilter, adminBranchID *uuid.UUID) ([]domain.TrafficViolation, int64, error)

	// Fuel Log
	CreateFuelLog(ctx context.Context, req dto.CreateFuelLogRequest, adminBranchID *uuid.UUID) (*domain.FuelLog, error)
	UpdateFuelLog(ctx context.Context, id uuid.UUID, req dto.UpdateFuelLogRequest) (*domain.FuelLog, error)
	DeleteFuelLog(ctx context.Context, id uuid.UUID) error
	GetFuelLogByID(ctx context.Context, id uuid.UUID) (*domain.FuelLog, error)
	GetAllFuelLogs(ctx context.Context, filter dto.FuelLogFilter, adminBranchID *uuid.UUID) ([]domain.FuelLog, int64, error)

	// Contract implementations
	CheckEmployeeOnLeave(ctx context.Context, empID uuid.UUID, date time.Time) (*contracts.LeaveCheckDTO, error)
	GetPendingRequestsCount(ctx context.Context) (investigations int64, leaves int64, err error)
}

type hrLegalService struct {
	repo repository.HRLegalRepository
}

func NewHRLegalService(repo repository.HRLegalRepository) HRLegalService {
	return &hrLegalService{repo: repo}
}

// ------------------------------------------------------------------
// 1. Investigation
// ------------------------------------------------------------------
func (s *hrLegalService) toInvestigationResponse(inv *domain.Investigation) *dto.InvestigationResponse {
	var questions, answers, items, images []string
	_ = json.Unmarshal([]byte(inv.Questions), &questions)
	_ = json.Unmarshal([]byte(inv.Answers), &answers)
	_ = json.Unmarshal([]byte(inv.Items), &items)
	_ = json.Unmarshal([]byte(inv.Images), &images)

	g := false
	if inv.IsGuilty != nil {
		g = *inv.IsGuilty
	}

	return &dto.InvestigationResponse{
		ID:                 inv.ID,
		EmployeeID:         inv.EmployeeID,
		NationalID:         inv.NationalID,
		SupervisorID:       inv.SupervisorID,
		Type:               inv.Type,
		Questions:          questions,
		Answers:            answers,
		ReportText:         inv.ReportText,
		Images:             images,
		Amount:             inv.Amount,
		StartDate:          inv.StartDate,
		EndDate:            inv.EndDate,
		Items:              items,
		IsGuilty:           g,
		Notes:              inv.Notes,
		DeductionMonth:     inv.DeductionMonth,
		Status:             inv.Status,
		ApprovedByName:     inv.ApprovedByName,
		ApprovedByUsername: inv.ApprovedByUsername,
		RejectedByName:     inv.RejectedByName,
		RejectedByUsername: inv.RejectedByUsername,
		ApprovedAt:         inv.ApprovedAt,
		RejectedAt:         inv.RejectedAt,
		CreatedAt:          inv.CreatedAt,
	}
}

func (s *hrLegalService) CreateInvestigation(ctx context.Context, req dto.CreateInvestigationRequest, supervisorID uuid.UUID) (*dto.InvestigationResponse, error) {
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, errors.New("معرف الموظف غير صالح")
	}

	questionsJSON, _ := json.Marshal(req.Questions)
	answersJSON, _ := json.Marshal(req.Answers)
	itemsJSON, _ := json.Marshal(req.Items)
	imagesJSON, _ := json.Marshal(req.Images)

	invType := req.Type
	if invType == "" {
		invType = "investigation"
	}

	var startD, endD *time.Time
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startD = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endD = &t
		}
	}

	inv := &domain.Investigation{
		ID:             uuid.New(),
		EmployeeID:     empID,
		SupervisorID:   supervisorID,
		Type:           invType,
		Questions:      string(questionsJSON),
		Answers:        string(answersJSON),
		ReportText:     req.ReportText,
		Images:         string(imagesJSON),
		Amount:         req.Amount,
		StartDate:      startD,
		EndDate:        endD,
		Items:          string(itemsJSON),
		IsGuilty:       &req.IsGuilty,
		Notes:          req.Notes,
		DeductionMonth: req.DeductionMonth,
		Status:         "pending",
	}

	if err := s.repo.CreateInvestigation(ctx, inv); err != nil {
		return nil, err
	}
	return s.toInvestigationResponse(inv), nil
}

func (s *hrLegalService) UpdateInvestigation(ctx context.Context, id uuid.UUID, req dto.UpdateInvestigationRequest) (*dto.InvestigationResponse, error) {
	inv, err := s.repo.FindInvestigationByID(ctx, id)
	if err != nil {
		return nil, errors.New("سجل التحقيق غير موجود")
	}

	if req.Type != "" {
		inv.Type = req.Type
	}
	if req.Questions != nil {
		b, _ := json.Marshal(req.Questions)
		inv.Questions = string(b)
	}
	if req.Answers != nil {
		b, _ := json.Marshal(req.Answers)
		inv.Answers = string(b)
	}
	if req.ReportText != "" {
		inv.ReportText = req.ReportText
	}
	if req.Images != nil {
		b, _ := json.Marshal(req.Images)
		inv.Images = string(b)
	}
	if req.Amount != nil {
		inv.Amount = req.Amount
	}
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			inv.StartDate = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			inv.EndDate = &t
		}
	}
	if req.Items != nil {
		b, _ := json.Marshal(req.Items)
		inv.Items = string(b)
	}
	inv.IsGuilty = &req.IsGuilty
	if req.Notes != "" {
		inv.Notes = req.Notes
	}
	if req.DeductionMonth != "" {
		inv.DeductionMonth = req.DeductionMonth
	}

	if err := s.repo.UpdateInvestigation(ctx, inv); err != nil {
		return nil, err
	}
	return s.toInvestigationResponse(inv), nil
}

func (s *hrLegalService) GetAllInvestigations(ctx context.Context, branchID *uuid.UUID) ([]dto.InvestigationResponse, error) {
	list, err := s.repo.FindInvestigations(ctx, branchID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.InvestigationResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toInvestigationResponse(&list[i]))
	}
	return result, nil
}

func (s *hrLegalService) GetInvestigationByID(ctx context.Context, id uuid.UUID) (*dto.InvestigationResponse, error) {
	inv, err := s.repo.FindInvestigationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toInvestigationResponse(inv), nil
}

func (s *hrLegalService) ApproveInvestigation(ctx context.Context, id uuid.UUID, adminID uuid.UUID, adminName, adminUsername string) (*dto.InvestigationResponse, error) {
	inv, err := s.repo.FindInvestigationByID(ctx, id)
	if err != nil {
		return nil, errors.New("التحقيق غير موجود")
	}

	now := time.Now()
	inv.Status = "approved"
	inv.ApprovedBy = &adminID
	inv.ApprovedByName = adminName
	inv.ApprovedByUsername = adminUsername
	inv.ApprovedAt = &now

	if err := s.repo.UpdateInvestigation(ctx, inv); err != nil {
		return nil, err
	}
	return s.toInvestigationResponse(inv), nil
}

func (s *hrLegalService) RejectInvestigation(ctx context.Context, id uuid.UUID, adminID uuid.UUID, adminName, adminUsername string) (*dto.InvestigationResponse, error) {
	inv, err := s.repo.FindInvestigationByID(ctx, id)
	if err != nil {
		return nil, errors.New("التحقيق غير موجود")
	}

	now := time.Now()
	inv.Status = "rejected"
	inv.RejectedBy = &adminID
	inv.RejectedByName = adminName
	inv.RejectedByUsername = adminUsername
	inv.RejectedAt = &now

	if err := s.repo.UpdateInvestigation(ctx, inv); err != nil {
		return nil, err
	}
	return s.toInvestigationResponse(inv), nil
}

func (s *hrLegalService) GetPendingInvestigationCount(ctx context.Context) (int64, error) {
	return s.repo.CountPendingInvestigations(ctx)
}

// ------------------------------------------------------------------
// 2. Document
// ------------------------------------------------------------------
func (s *hrLegalService) CreateDocument(ctx context.Context, req dto.CreateEmployeeDocumentRequest) (*domain.EmployeeDocument, error) {
	status := "VALID"
	if req.Status != "" {
		status = req.Status
	}

	var issueDate, expiryDate *time.Time
	if req.IssueDate != nil && *req.IssueDate != "" {
		if t, err := time.Parse("2006-01-02", *req.IssueDate); err == nil {
			issueDate = &t
		}
	}
	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		if t, err := time.Parse("2006-01-02", *req.ExpiryDate); err == nil {
			expiryDate = &t
			if t.Before(time.Now()) {
				status = "EXPIRED"
			}
		}
	}

	doc := &domain.EmployeeDocument{
		ID:         uuid.New(),
		EmployeeID: req.EmployeeID,
		DocType:    req.DocType,
		Title:      req.Title,
		DocNumber:  req.DocNumber,
		FileURL:    req.FileURL,
		IssueDate:  issueDate,
		ExpiryDate: expiryDate,
		Status:     status,
		Notes:      req.Notes,
	}

	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, err
	}
	return s.repo.FindDocumentByID(ctx, doc.ID)
}

func (s *hrLegalService) UpdateDocument(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeDocumentRequest) (*domain.EmployeeDocument, error) {
	doc, err := s.repo.FindDocumentByID(ctx, id)
	if err != nil {
		return nil, errors.New("المستند غير موجود")
	}

	if req.DocType != nil {
		doc.DocType = *req.DocType
	}
	if req.Title != nil {
		doc.Title = *req.Title
	}
	if req.DocNumber != nil {
		doc.DocNumber = *req.DocNumber
	}
	if req.FileURL != nil {
		doc.FileURL = *req.FileURL
	}
	if req.IssueDate != nil && *req.IssueDate != "" {
		if t, err := time.Parse("2006-01-02", *req.IssueDate); err == nil {
			doc.IssueDate = &t
		}
	}
	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		if t, err := time.Parse("2006-01-02", *req.ExpiryDate); err == nil {
			doc.ExpiryDate = &t
		}
	}
	if req.Status != nil {
		doc.Status = *req.Status
	}
	if req.Notes != nil {
		doc.Notes = *req.Notes
	}

	if err := s.repo.UpdateDocument(ctx, doc); err != nil {
		return nil, err
	}
	return s.repo.FindDocumentByID(ctx, id)
}

func (s *hrLegalService) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteDocument(ctx, id)
}

func (s *hrLegalService) GetDocumentByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeDocument, error) {
	return s.repo.FindDocumentByID(ctx, id)
}

func (s *hrLegalService) GetAllDocuments(ctx context.Context, filter dto.EmployeeDocumentFilter) ([]domain.EmployeeDocument, int64, error) {
	return s.repo.FindDocuments(ctx, filter)
}

func (s *hrLegalService) GetExpiringDocuments(ctx context.Context, days int) ([]domain.EmployeeDocument, error) {
	if days <= 0 {
		days = 30
	}
	return s.repo.FindExpiringDocuments(ctx, days)
}

// ------------------------------------------------------------------
// 3. Bank Account
// ------------------------------------------------------------------
func (s *hrLegalService) CreateBankAccount(ctx context.Context, req dto.CreateEmployeeBankAccountRequest) (*domain.EmployeeBankAccount, error) {
	if req.IsDefault {
		_ = s.repo.ClearDefaultBankAccount(ctx, req.EmployeeID)
	}

	acc := &domain.EmployeeBankAccount{
		ID:               uuid.New(),
		EmployeeID:       req.EmployeeID,
		BankName:         req.BankName,
		IBAN:             req.IBAN,
		AccountOwnerName: req.AccountOwnerName,
		IsDefault:        req.IsDefault,
	}

	if err := s.repo.CreateBankAccount(ctx, acc); err != nil {
		return nil, err
	}
	return s.repo.FindBankAccountByID(ctx, acc.ID)
}

func (s *hrLegalService) UpdateBankAccount(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeBankAccountRequest) (*domain.EmployeeBankAccount, error) {
	acc, err := s.repo.FindBankAccountByID(ctx, id)
	if err != nil {
		return nil, errors.New("الحساب البنكي غير موجود")
	}

	if req.BankName != nil {
		acc.BankName = *req.BankName
	}
	if req.IBAN != nil {
		acc.IBAN = *req.IBAN
	}
	if req.AccountOwnerName != nil {
		acc.AccountOwnerName = *req.AccountOwnerName
	}
	if req.IsDefault != nil && *req.IsDefault {
		_ = s.repo.ClearDefaultBankAccount(ctx, acc.EmployeeID)
		acc.IsDefault = true
	} else if req.IsDefault != nil {
		acc.IsDefault = *req.IsDefault
	}

	if err := s.repo.UpdateBankAccount(ctx, acc); err != nil {
		return nil, err
	}
	return s.repo.FindBankAccountByID(ctx, id)
}

func (s *hrLegalService) DeleteBankAccount(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteBankAccount(ctx, id)
}

func (s *hrLegalService) GetBankAccountByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeBankAccount, error) {
	return s.repo.FindBankAccountByID(ctx, id)
}

func (s *hrLegalService) GetAllBankAccounts(ctx context.Context, filter dto.EmployeeBankAccountFilter) ([]domain.EmployeeBankAccount, int64, error) {
	return s.repo.FindBankAccounts(ctx, filter)
}

// ------------------------------------------------------------------
// 4. Leave Request
// ------------------------------------------------------------------
func (s *hrLegalService) CreateLeaveRequest(ctx context.Context, req dto.CreateLeaveRequestRequest) (*domain.LeaveRequest, error) {
	leaveType := "ANNUAL"
	if req.LeaveType != "" {
		leaveType = req.LeaveType
	}
	daysCount := req.DaysCount
	if daysCount <= 0 {
		daysCount = 1
	}

	leave := &domain.LeaveRequest{
		ID:         uuid.New(),
		EmployeeID: req.EmployeeID,
		LeaveType:  leaveType,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		DaysCount:  daysCount,
		Reason:     req.Reason,
		Status:     "PENDING",
	}

	if err := s.repo.CreateLeaveRequest(ctx, leave); err != nil {
		return nil, err
	}
	return s.repo.FindLeaveRequestByID(ctx, leave.ID)
}

func (s *hrLegalService) UpdateLeaveRequestStatus(ctx context.Context, id uuid.UUID, req dto.UpdateLeaveRequestStatusRequest) (*domain.LeaveRequest, error) {
	leave, err := s.repo.FindLeaveRequestByID(ctx, id)
	if err != nil {
		return nil, errors.New("طلب الإجازة غير موجود")
	}

	leave.Status = req.Status
	if req.ApprovedByName != "" {
		leave.ApprovedByName = req.ApprovedByName
	}

	if err := s.repo.UpdateLeaveRequest(ctx, leave); err != nil {
		return nil, err
	}
	return s.repo.FindLeaveRequestByID(ctx, id)
}

func (s *hrLegalService) DeleteLeaveRequest(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteLeaveRequest(ctx, id)
}

func (s *hrLegalService) GetLeaveRequestByID(ctx context.Context, id uuid.UUID) (*domain.LeaveRequest, error) {
	return s.repo.FindLeaveRequestByID(ctx, id)
}

func (s *hrLegalService) GetAllLeaveRequests(ctx context.Context, filter dto.LeaveRequestFilter) ([]domain.LeaveRequest, int64, error) {
	return s.repo.FindLeaveRequests(ctx, filter)
}

func (s *hrLegalService) GetPendingLeaveCount(ctx context.Context) (int64, error) {
	return s.repo.CountPendingLeaveRequests(ctx)
}

// ------------------------------------------------------------------
// 5. Traffic Violation
// ------------------------------------------------------------------
func (s *hrLegalService) CreateViolation(ctx context.Context, req dto.CreateTrafficViolationRequest) (*domain.TrafficViolation, error) {
	status := "RECORDED"
	if req.Status != "" {
		status = req.Status
	}

	var paidAmount float64
	if req.PaidAmount != nil {
		paidAmount = *req.PaidAmount
	}

	var violationDate time.Time
	if req.ViolationDate != "" {
		if t, err := time.Parse("2006-01-02", req.ViolationDate); err == nil {
			violationDate = t
		}
	}

	v := &domain.TrafficViolation{
		ID:              uuid.New(),
		ViolationNumber: req.ViolationNumber,
		EmployeeID:      req.EmployeeID,
		VehiclePlate:    req.VehiclePlate,
		Amount:          req.Amount,
		PaidAmount:      paidAmount,
		Reason:          req.Reason,
		ViolationDate:   violationDate,
		City:            req.City,
		Status:          status,
		BranchID:        req.BranchID,
		Notes:           req.Notes,
	}

	if err := s.repo.CreateViolation(ctx, v); err != nil {
		return nil, err
	}
	return s.repo.FindViolationByID(ctx, v.ID)
}

func (s *hrLegalService) UpdateViolation(ctx context.Context, id uuid.UUID, req dto.UpdateTrafficViolationRequest) (*domain.TrafficViolation, error) {
	v, err := s.repo.FindViolationByID(ctx, id)
	if err != nil {
		return nil, errors.New("المخالفة غير موجودة")
	}

	if req.ViolationNumber != nil {
		v.ViolationNumber = *req.ViolationNumber
	}
	if req.EmployeeID != nil {
		v.EmployeeID = req.EmployeeID
	}
	if req.VehiclePlate != nil {
		v.VehiclePlate = *req.VehiclePlate
	}
	if req.Amount != nil {
		v.Amount = *req.Amount
	}
	if req.PaidAmount != nil {
		v.PaidAmount = *req.PaidAmount
	}
	if req.AddPayment != nil {
		v.PaidAmount += *req.AddPayment
		if v.PaidAmount >= v.Amount {
			v.Status = "PAID"
		} else if v.PaidAmount > 0 {
			v.Status = "PARTIAL"
		}
	}
	if req.Reason != nil {
		v.Reason = *req.Reason
	}
	if req.City != nil {
		v.City = *req.City
	}
	if req.Status != nil {
		v.Status = *req.Status
	}
	if req.Notes != nil {
		v.Notes = *req.Notes
	}

	if err := s.repo.UpdateViolation(ctx, v); err != nil {
		return nil, err
	}
	return s.repo.FindViolationByID(ctx, id)
}

func (s *hrLegalService) DeleteViolation(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteViolation(ctx, id)
}

func (s *hrLegalService) GetViolationByID(ctx context.Context, id uuid.UUID) (*domain.TrafficViolation, error) {
	return s.repo.FindViolationByID(ctx, id)
}

func (s *hrLegalService) GetAllViolations(ctx context.Context, filter dto.TrafficViolationFilter, adminBranchID *uuid.UUID) ([]domain.TrafficViolation, int64, error) {
	if adminBranchID != nil {
		filter.BranchID = adminBranchID
	}
	return s.repo.FindViolations(ctx, filter)
}

// ------------------------------------------------------------------
// 6. Fuel Log
// ------------------------------------------------------------------
func (s *hrLegalService) CreateFuelLog(ctx context.Context, req dto.CreateFuelLogRequest, adminBranchID *uuid.UUID) (*domain.FuelLog, error) {
	branchID := req.BranchID
	if branchID == nil && adminBranchID != nil {
		branchID = adminBranchID
	}

	var fuelDate time.Time
	if req.FuelDate != "" {
		if t, err := time.Parse("2006-01-02", req.FuelDate); err == nil {
			fuelDate = t
		}
	} else {
		fuelDate = time.Now()
	}

	log := &domain.FuelLog{
		ID:              uuid.New(),
		EmployeeID:      req.EmployeeID,
		VehiclePlate:    req.VehiclePlate,
		ShiftID:         req.ShiftID,
		Amount:          req.Amount,
		Liters:          req.Liters,
		FuelDate:        fuelDate,
		StationName:     req.StationName,
		InvoiceImageURL: req.InvoiceImageURL,
		BranchID:        branchID,
		Notes:           req.Notes,
	}

	if err := s.repo.CreateFuelLog(ctx, log); err != nil {
		return nil, err
	}
	return s.repo.FindFuelLogByID(ctx, log.ID)
}

func (s *hrLegalService) UpdateFuelLog(ctx context.Context, id uuid.UUID, req dto.UpdateFuelLogRequest) (*domain.FuelLog, error) {
	log, err := s.repo.FindFuelLogByID(ctx, id)
	if err != nil {
		return nil, errors.New("سجل الوقود غير موجود")
	}

	if req.EmployeeID != nil {
		log.EmployeeID = req.EmployeeID
	}
	if req.VehiclePlate != nil {
		log.VehiclePlate = *req.VehiclePlate
	}
	if req.Amount != nil {
		log.Amount = *req.Amount
	}
	if req.Liters != nil {
		log.Liters = *req.Liters
	}
	if req.FuelDate != nil && *req.FuelDate != "" {
		if t, err := time.Parse("2006-01-02", *req.FuelDate); err == nil {
			log.FuelDate = t
		}
	}
	if req.StationName != nil {
		log.StationName = *req.StationName
	}
	if req.InvoiceImageURL != nil {
		log.InvoiceImageURL = *req.InvoiceImageURL
	}
	if req.Notes != nil {
		log.Notes = *req.Notes
	}

	if err := s.repo.UpdateFuelLog(ctx, log); err != nil {
		return nil, err
	}
	return s.repo.FindFuelLogByID(ctx, id)
}

func (s *hrLegalService) DeleteFuelLog(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteFuelLog(ctx, id)
}

func (s *hrLegalService) GetFuelLogByID(ctx context.Context, id uuid.UUID) (*domain.FuelLog, error) {
	return s.repo.FindFuelLogByID(ctx, id)
}

func (s *hrLegalService) GetAllFuelLogs(ctx context.Context, filter dto.FuelLogFilter, adminBranchID *uuid.UUID) ([]domain.FuelLog, int64, error) {
	if adminBranchID != nil {
		filter.BranchID = adminBranchID
	}
	return s.repo.FindFuelLogs(ctx, filter)
}

// ------------------------------------------------------------------
// Contract Implementations
// ------------------------------------------------------------------
func (s *hrLegalService) CheckEmployeeOnLeave(ctx context.Context, empID uuid.UUID, date time.Time) (*contracts.LeaveCheckDTO, error) {
	dateStr := date.Format("2006-01-02")
	leave, err := s.repo.FindActiveLeave(ctx, empID, dateStr)
	if err != nil || leave == nil {
		return &contracts.LeaveCheckDTO{
			EmployeeID: empID,
			IsOnLeave:  false,
		}, nil
	}
	start, _ := time.Parse("2006-01-02", leave.StartDate)
	end, _ := time.Parse("2006-01-02", leave.EndDate)
	return &contracts.LeaveCheckDTO{
		EmployeeID: empID,
		IsOnLeave:  true,
		LeaveType:  leave.LeaveType,
		StartDate:  start,
		EndDate:    end,
	}, nil
}

func (s *hrLegalService) GetPendingRequestsCount(ctx context.Context) (int64, int64, error) {
	invCount, err := s.repo.CountPendingInvestigations(ctx)
	if err != nil {
		return 0, 0, err
	}
	leaveCount, err := s.repo.CountPendingLeaveRequests(ctx)
	if err != nil {
		return 0, 0, err
	}
	return invCount, leaveCount, nil
}
