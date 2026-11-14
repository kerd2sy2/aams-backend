package employee

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	sharedRepo "delivery-backend/internal/repository"
	"delivery-backend/pkg/barcode"
)

type EmployeeService interface {
	IEmployeeContract
	Create(ctx context.Context, req CreateEmployeeRequest) (*Employee, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateEmployeeRequest) (*Employee, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Employee, error)
	FindAll(ctx context.Context, filter EmployeeFilter) ([]Employee, int64, error)
	Search(ctx context.Context, query string, branchID *uuid.UUID) ([]Employee, error)
	GetWorkingEmployees(ctx context.Context, branchID *uuid.UUID) ([]Employee, error)
	GetLocations(ctx context.Context, branchID *uuid.UUID) ([]EmployeeLocationDTO, error)
	UpdateLocation(ctx context.Context, id uuid.UUID, req UpdateLocationRequest) error
	SetPhone(ctx context.Context, id uuid.UUID, phone string) error
	ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, id uuid.UUID, newPassword string) error
	BatchSetOilChange(ctx context.Context, req BatchOilSetupRequest) (int, error)
	GenerateBarcode(nationalID string) (string, error)
	GenerateQRCode(nationalID string) (string, error)
}

type employeeService struct {
	repo      EmployeeRepository
	adminRepo sharedRepo.AdminRepository
}

func NewEmployeeService(repo EmployeeRepository, adminRepo ...sharedRepo.AdminRepository) EmployeeService {
	svc := &employeeService{repo: repo}
	if len(adminRepo) > 0 {
		svc.adminRepo = adminRepo[0]
	}
	return svc
}

func (s *employeeService) Create(ctx context.Context, req CreateEmployeeRequest) (*Employee, error) {
	existing, _ := s.repo.FindByNationalID(ctx, req.NationalID)
	if existing != nil {
		return nil, ErrNationalIDRegistered
	}

	barcodeData, _ := s.GenerateBarcode(req.NationalID)
	qrData, _ := s.GenerateQRCode(req.NationalID)

	// Default password is national ID
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.NationalID), bcrypt.DefaultCost)

	emp := &Employee{
		ID:                       uuid.New(),
		Name:                     req.Name,
		JobRole:                  req.JobRole,
		EmployeeNumber:           req.EmployeeNumber,
		Phone:                    req.Phone,
		PersonalImage:            req.PersonalImage,
		NationalID:               req.NationalID,
		PasswordHash:             string(hashedPwd),
		IqamaExpirationDate:      req.IqamaExpirationDate,
		NationalIDImage:          req.NationalIDImage,
		DrivingLicenseImage:      req.DrivingLicenseImage,
		PassportImage:            req.PassportImage,
		VehicleRegistrationImage: req.VehicleRegistrationImage,
		KeyNumber:                req.KeyNumber,
		MotorcycleNumber:         req.MotorcycleNumber,
		ApplicationID:            req.ApplicationID,
		ApplicationType:          req.ApplicationType,
		VehicleType:              req.VehicleType,
		Shift:                    req.Shift,
		BranchID:                 req.BranchID,
		RoleID:                   req.RoleID,
		RoleName:                 req.RoleName,
		Permissions:              string(req.Permissions),
		Barcode:                  barcodeData,
		QRCode:                   qrData,
		Language:                 "ar",
	}

	if emp.JobRole == "" {
		emp.JobRole = "DRIVER"
	}
	if emp.VehicleType == "" {
		emp.VehicleType = "motorcycle"
	}
	if emp.Shift == "" {
		emp.Shift = "morning"
	}

	if err := s.repo.Create(ctx, emp); err != nil {
		return nil, fmt.Errorf("فشل في حفظ الموظف: %w", err)
	}

	return emp, nil
}

func (s *employeeService) Update(ctx context.Context, id uuid.UUID, req UpdateEmployeeRequest) (*Employee, error) {
	emp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrEmployeeNotFound
	}

	if req.NationalID != "" && req.NationalID != emp.NationalID {
		existing, _ := s.repo.FindByNationalID(ctx, req.NationalID)
		if existing != nil && existing.ID != emp.ID {
			return nil, errors.New("رقم الهوية مسجل لموظف آخر")
		}
		emp.NationalID = req.NationalID
		emp.Barcode, _ = s.GenerateBarcode(req.NationalID)
		emp.QRCode, _ = s.GenerateQRCode(req.NationalID)
	}

	if req.Name != "" {
		emp.Name = req.Name
	}
	if req.JobRole != "" {
		emp.JobRole = req.JobRole
	}
	if req.EmployeeNumber != "" {
		emp.EmployeeNumber = req.EmployeeNumber
	}
	if req.Phone != "" {
		emp.Phone = req.Phone
	}
	if req.PersonalImage != "" {
		emp.PersonalImage = req.PersonalImage
	}
	if req.IqamaExpirationDate != nil {
		emp.IqamaExpirationDate = req.IqamaExpirationDate
	}
	if req.NationalIDImage != "" {
		emp.NationalIDImage = req.NationalIDImage
	}
	if req.DrivingLicenseImage != "" {
		emp.DrivingLicenseImage = req.DrivingLicenseImage
	}
	if req.PassportImage != "" {
		emp.PassportImage = req.PassportImage
	}
	if req.VehicleRegistrationImage != "" {
		emp.VehicleRegistrationImage = req.VehicleRegistrationImage
	}
	if req.KeyNumber != "" {
		emp.KeyNumber = req.KeyNumber
	}
	if req.MotorcycleNumber != "" {
		emp.MotorcycleNumber = req.MotorcycleNumber
	}
	if req.ApplicationID != "" {
		emp.ApplicationID = req.ApplicationID
	}
	if req.ApplicationType != "" {
		emp.ApplicationType = req.ApplicationType
	}
	if req.VehicleType != "" {
		emp.VehicleType = req.VehicleType
	}
	if req.Shift != "" {
		emp.Shift = req.Shift
	}
	if req.ClearBranch {
		emp.BranchID = nil
	} else if req.BranchID != nil {
		emp.BranchID = req.BranchID
	}
	if req.ClearRole || (req.RoleID == nil && (req.Permissions == "" || req.Permissions == "[]")) {
		emp.RoleID = nil
		emp.RoleName = ""
		emp.Permissions = ""
	} else if req.RoleID != nil {
		emp.RoleID = req.RoleID
		emp.RoleName = req.RoleName
		emp.Permissions = string(req.Permissions)
	} else if req.Permissions != "" && req.Permissions != "[]" {
		emp.Permissions = string(req.Permissions)
	}

	if err := s.repo.Update(ctx, emp); err != nil {
		return nil, fmt.Errorf("فشل في تحديث بيانات الموظف: %w", err)
	}

	// Bidirectional sync: keep any Admin linked to this Employee aligned
	if s.adminRepo != nil {
		if linkedAdmin, _ := s.adminRepo.FindByEmployeeID(ctx, emp.ID); linkedAdmin != nil {
			linkedAdmin.RoleID = emp.RoleID
			linkedAdmin.Permissions = emp.Permissions
			linkedAdmin.PermissionsConfigured = (emp.Permissions != "")
			if emp.JobRole != "" {
				linkedAdmin.Role = emp.JobRole
			}
			linkedAdmin.BranchID = emp.BranchID
			_ = s.adminRepo.Update(ctx, linkedAdmin)
		}
	}

	return emp, nil
}

func (s *employeeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *employeeService) FindByID(ctx context.Context, id uuid.UUID) (*Employee, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *employeeService) FindAll(ctx context.Context, filter EmployeeFilter) ([]Employee, int64, error) {
	return s.repo.FindAll(ctx, filter)
}

func (s *employeeService) Search(ctx context.Context, query string, branchID *uuid.UUID) ([]Employee, error) {
	return s.repo.Search(ctx, query, branchID)
}

func (s *employeeService) GetWorkingEmployees(ctx context.Context, branchID *uuid.UUID) ([]Employee, error) {
	return s.repo.GetWorkingEmployees(ctx, branchID)
}

func (s *employeeService) GetLocations(ctx context.Context, branchID *uuid.UUID) ([]EmployeeLocationDTO, error) {
	return s.repo.GetLocations(ctx, branchID)
}

func (s *employeeService) UpdateLocation(ctx context.Context, id uuid.UUID, req UpdateLocationRequest) error {
	isVPN := false
	if req.IsVPN != nil {
		isVPN = *req.IsVPN
	}
	isMock := false
	if req.IsMockLocation != nil {
		isMock = *req.IsMockLocation
	}
	return s.repo.UpdateLocation(ctx, id, req.Latitude, req.Longitude, isVPN, isMock, false)
}

func (s *employeeService) SetPhone(ctx context.Context, id uuid.UUID, phone string) error {
	return s.repo.UpdatePhone(ctx, id, phone)
}

func (s *employeeService) ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	emp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return ErrEmployeeNotFound
	}

	cleanOld := strings.TrimSpace(oldPassword)
	cleanNew := strings.TrimSpace(newPassword)

	if cleanNew == "" || len(cleanNew) < 4 {
		return errors.New("كلمة المرور الجديدة يجب ألا تقل عن 4 خانات")
	}

	if emp.PasswordHash != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(emp.PasswordHash), []byte(cleanOld)); err != nil {
			return ErrOldPasswordIncorrect
		}
	} else {
		// Default password: last 4 digits of Phone or National ID / Iqama
		matched := false
		if emp.Phone != "" {
			phoneDigits := extractASCIIOnlyDigits(emp.Phone)
			if len(phoneDigits) >= 4 && cleanOld == phoneDigits[len(phoneDigits)-4:] {
				matched = true
			}
		}
		if !matched && emp.NationalID != "" {
			natDigits := extractASCIIOnlyDigits(emp.NationalID)
			if len(natDigits) >= 4 && cleanOld == natDigits[len(natDigits)-4:] {
				matched = true
			} else if cleanOld == emp.NationalID {
				matched = true
			}
		}
		if !matched {
			return errors.New("كلمة المرور الحالية غير صحيحة (كلمة المرور الافتراضية هي آخر 4 أرقام من رقم الجوال)")
		}
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(cleanNew), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, id, string(hashed))
}

func (s *employeeService) ResetPassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	if newPassword == "" {
		emp, err := s.repo.FindByID(ctx, id)
		if err != nil {
			return ErrEmployeeNotFound
		}
		newPassword = emp.NationalID
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, id, string(hashed))
}

func (s *employeeService) BatchSetOilChange(ctx context.Context, req BatchOilSetupRequest) (int, error) {
	return s.repo.BatchSetOilChange(ctx, req.Entries)
}

func (s *employeeService) GenerateBarcode(nationalID string) (string, error) {
	return barcode.GenerateCode128Base64(nationalID)
}

func (s *employeeService) GenerateQRCode(nationalID string) (string, error) {
	return barcode.GenerateQRCodeBase64(nationalID)
}

// IEmployeeContract implementation
func (s *employeeService) GetEmployee(ctx context.Context, id uuid.UUID) (*EmployeeDTO, error) {
	emp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toEmployeeDTO(emp), nil
}

func (s *employeeService) FindByNationalID(ctx context.Context, nationalID string) (*EmployeeDTO, error) {
	emp, err := s.repo.FindByNationalID(ctx, nationalID)
	if err != nil {
		return nil, err
	}
	return toEmployeeDTO(emp), nil
}

func (s *employeeService) UpdateOnStartWork(ctx context.Context, id uuid.UUID, appID, appType, motorcycleNumber string) error {
	emp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	emp.IsWorking = true
	if appID != "" {
		emp.ApplicationID = appID
	}
	if appType != "" {
		emp.ApplicationType = appType
	}
	if motorcycleNumber != "" {
		emp.MotorcycleNumber = motorcycleNumber
	}
	return s.repo.Update(ctx, emp)
}

func (s *employeeService) UpdateOnEndWork(ctx context.Context, id uuid.UUID, addedDistance float64, totalOrders int) error {
	emp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	emp.IsWorking = false
	emp.TotalDistance += addedDistance
	return s.repo.Update(ctx, emp)
}

func toEmployeeDTO(e *Employee) *EmployeeDTO {
	if e == nil {
		return nil
	}
	branchName := ""
	if e.Branch != nil {
		branchName = e.Branch.Name
	}
	return &EmployeeDTO{
		ID:                    e.ID,
		Name:                  e.Name,
		JobRole:               e.JobRole,
		EmployeeNumber:        e.EmployeeNumber,
		Phone:                 e.Phone,
		NationalID:            e.NationalID,
		BranchID:              e.BranchID,
		BranchName:            branchName,
		VehicleType:           e.VehicleType,
		MotorcycleNumber:      e.MotorcycleNumber,
		ApplicationID:         e.ApplicationID,
		ApplicationType:       e.ApplicationType,
		TotalDistance:         e.TotalDistance,
		LastOilChangeDistance: e.LastOilChangeDistance,
		Latitude:              e.Latitude,
		Longitude:             e.Longitude,
	}
}

func extractASCIIOnlyDigits(input string) string {
	var b strings.Builder
	for _, r := range input {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
