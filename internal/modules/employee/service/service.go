package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"delivery-backend/internal/modules/employee/contracts"
	"delivery-backend/internal/modules/employee/domain"
	"delivery-backend/internal/modules/employee/dto"
	"delivery-backend/internal/modules/employee/repository"
	"delivery-backend/pkg/barcode"
)

type EmployeeService interface {
	contracts.IEmployeeContract
	Create(ctx context.Context, req dto.CreateEmployeeRequest) (*domain.Employee, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeRequest) (*domain.Employee, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error)
	FindAll(ctx context.Context, filter dto.EmployeeFilter) ([]domain.Employee, int64, error)
	Search(ctx context.Context, query string, branchID *uuid.UUID) ([]domain.Employee, error)
	GetWorkingEmployees(ctx context.Context, branchID *uuid.UUID) ([]domain.Employee, error)
	GetLocations(ctx context.Context, branchID *uuid.UUID) ([]dto.EmployeeLocationDTO, error)
	UpdateLocation(ctx context.Context, id uuid.UUID, req dto.UpdateLocationRequest) error
	SetPhone(ctx context.Context, id uuid.UUID, phone string) error
	ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, id uuid.UUID, newPassword string) error
	BatchSetOilChange(ctx context.Context, req dto.BatchOilSetupRequest) (int, error)
	GenerateBarcode(nationalID string) (string, error)
	GenerateQRCode(nationalID string) (string, error)
}

type employeeService struct {
	repo repository.EmployeeRepository
}

func NewEmployeeService(repo repository.EmployeeRepository) EmployeeService {
	return &employeeService{repo: repo}
}

func (s *employeeService) Create(ctx context.Context, req dto.CreateEmployeeRequest) (*domain.Employee, error) {
	existing, _ := s.repo.FindByNationalID(ctx, req.NationalID)
	if existing != nil {
		return nil, errors.New("رقم الهوية الوطنية أو الإقامة مسجل مسبقاً")
	}

	barcodeData, _ := s.GenerateBarcode(req.NationalID)
	qrData, _ := s.GenerateQRCode(req.NationalID)

	// Default password is national ID
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.NationalID), bcrypt.DefaultCost)

	emp := &domain.Employee{
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

func (s *employeeService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeRequest) (*domain.Employee, error) {
	emp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("الموظف غير موجود")
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
	if req.BranchID != nil {
		emp.BranchID = req.BranchID
	}

	if err := s.repo.Update(ctx, emp); err != nil {
		return nil, fmt.Errorf("فشل في تحديث بيانات الموظف: %w", err)
	}

	return emp, nil
}

func (s *employeeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *employeeService) FindByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *employeeService) FindAll(ctx context.Context, filter dto.EmployeeFilter) ([]domain.Employee, int64, error) {
	return s.repo.FindAll(ctx, filter)
}

func (s *employeeService) Search(ctx context.Context, query string, branchID *uuid.UUID) ([]domain.Employee, error) {
	return s.repo.Search(ctx, query, branchID)
}

func (s *employeeService) GetWorkingEmployees(ctx context.Context, branchID *uuid.UUID) ([]domain.Employee, error) {
	return s.repo.GetWorkingEmployees(ctx, branchID)
}

func (s *employeeService) GetLocations(ctx context.Context, branchID *uuid.UUID) ([]dto.EmployeeLocationDTO, error) {
	return s.repo.GetLocations(ctx, branchID)
}

func (s *employeeService) UpdateLocation(ctx context.Context, id uuid.UUID, req dto.UpdateLocationRequest) error {
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
		return errors.New("الموظف غير موجود")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(emp.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("كلمة المرور الحالية غير صحيحة")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, id, string(hashed))
}

func (s *employeeService) ResetPassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	if newPassword == "" {
		emp, err := s.repo.FindByID(ctx, id)
		if err != nil {
			return errors.New("الموظف غير موجود")
		}
		newPassword = emp.NationalID
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, id, string(hashed))
}

func (s *employeeService) BatchSetOilChange(ctx context.Context, req dto.BatchOilSetupRequest) (int, error) {
	return s.repo.BatchSetOilChange(ctx, req.Entries)
}

func (s *employeeService) GenerateBarcode(nationalID string) (string, error) {
	return barcode.GenerateCode128Base64(nationalID)
}

func (s *employeeService) GenerateQRCode(nationalID string) (string, error) {
	return barcode.GenerateQRCodeBase64(nationalID)
}

// IEmployeeContract implementation
func (s *employeeService) GetEmployee(ctx context.Context, id uuid.UUID) (*contracts.EmployeeDTO, error) {
	emp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toEmployeeDTO(emp), nil
}

func (s *employeeService) FindAll(ctx context.Context, branchID *uuid.UUID) ([]contracts.EmployeeDTO, error) {
	emps, _, err := s.repo.FindAll(ctx, dto.EmployeeFilter{BranchID: branchID, Limit: 1000})
	if err != nil {
		return nil, err
	}
	res := make([]contracts.EmployeeDTO, len(emps))
	for i, e := range emps {
		res[i] = *toEmployeeDTO(&e)
	}
	return res, nil
}

func (s *employeeService) FindByID(ctx context.Context, id uuid.UUID) (*contracts.EmployeeDTO, error) {
	return s.GetEmployee(ctx, id)
}

func (s *employeeService) FindByNationalID(ctx context.Context, nationalID string) (*contracts.EmployeeDTO, error) {
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

func toEmployeeDTO(e *domain.Employee) *contracts.EmployeeDTO {
	if e == nil {
		return nil
	}
	return &contracts.EmployeeDTO{
		ID:                    e.ID,
		Name:                  e.Name,
		JobRole:               e.JobRole,
		EmployeeNumber:        e.EmployeeNumber,
		Phone:                 e.Phone,
		NationalID:            e.NationalID,
		BranchID:              e.BranchID,
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
