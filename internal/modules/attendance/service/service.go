package service

import (
	"context"
	"errors"
	"time"

	"delivery-backend/internal/modules/attendance/contracts"
	"delivery-backend/internal/modules/attendance/domain"
	"delivery-backend/internal/modules/attendance/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeInfo struct {
	ID          uuid.UUID
	Name        string
	NationalID  string
	BranchID    *uuid.UUID
	BranchName  string
	VehicleType string
}

type EmployeeFinder interface {
	FindAllEmployees(ctx context.Context, branchID *uuid.UUID) ([]EmployeeInfo, error)
	FindEmployeeByID(ctx context.Context, id uuid.UUID) (*EmployeeInfo, error)
}

type AttendanceService interface {
	contracts.IAttendanceContract
	GetAttendance(ctx context.Context, date string, branchID *uuid.UUID) ([]domain.AttendanceInfo, error)
	ToggleAttendance(ctx context.Context, adminID uuid.UUID, employeeID uuid.UUID, date string, status string, note string) (*domain.AttendanceInfo, error)
}

type attendanceService struct {
	repo           repository.AttendanceRepository
	employeeFinder EmployeeFinder
}

func NewAttendanceService(repo repository.AttendanceRepository, employeeFinder EmployeeFinder) AttendanceService {
	return &attendanceService{
		repo:           repo,
		employeeFinder: employeeFinder,
	}
}

func (s *attendanceService) GetAttendance(ctx context.Context, date string, branchID *uuid.UUID) ([]domain.AttendanceInfo, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	emps, err := s.employeeFinder.FindAllEmployees(ctx, branchID)
	if err != nil {
		return nil, err
	}

	existingRecords, err := s.repo.FindByDate(ctx, date)
	if err != nil {
		return nil, err
	}

	recordMap := make(map[uuid.UUID]*domain.Attendance)
	for i := range existingRecords {
		recordMap[existingRecords[i].EmployeeID] = &existingRecords[i]
	}

	result := make([]domain.AttendanceInfo, 0, len(emps))
	for _, emp := range emps {
		info := domain.AttendanceInfo{
			EmployeeID:   emp.ID,
			EmployeeName: emp.Name,
			NationalID:   emp.NationalID,
			BranchName:   emp.BranchName,
			VehicleType:  emp.VehicleType,
			Status:       "absent",
		}

		if rec, ok := recordMap[emp.ID]; ok {
			info.Status = rec.Status
			info.Note = rec.Note
		}

		result = append(result, info)
	}

	return result, nil
}

func (s *attendanceService) ToggleAttendance(ctx context.Context, adminID uuid.UUID, employeeID uuid.UUID, date string, status string, note string) (*domain.AttendanceInfo, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	attendance := &domain.Attendance{
		EmployeeID: employeeID,
		Date:       date,
		Status:     status,
		Note:       note,
	}

	if err := s.repo.Upsert(ctx, attendance); err != nil {
		return nil, err
	}

	emp, err := s.employeeFinder.FindEmployeeByID(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	return &domain.AttendanceInfo{
		EmployeeID:   emp.ID,
		EmployeeName: emp.Name,
		NationalID:   emp.NationalID,
		BranchName:   emp.BranchName,
		VehicleType:  emp.VehicleType,
		Status:       status,
		Note:         note,
	}, nil
}

// Implement contracts.IAttendanceContract
func (s *attendanceService) GetEmployeeAttendanceStatus(ctx context.Context, empID uuid.UUID, date string) (*contracts.AttendanceStatusDTO, error) {
	rec, err := s.repo.FindByEmployeeAndDate(ctx, empID, date)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &contracts.AttendanceStatusDTO{
				EmployeeID: empID,
				Date:       date,
				Status:     "absent",
			}, nil
		}
		return nil, err
	}

	return &contracts.AttendanceStatusDTO{
		EmployeeID: rec.EmployeeID,
		Date:       rec.Date,
		Status:     rec.Status,
		Note:       rec.Note,
	}, nil
}

func (s *attendanceService) MarkPresent(ctx context.Context, empID uuid.UUID, date string, note string) error {
	attendance := &domain.Attendance{
		EmployeeID: empID,
		Date:       date,
		Status:     "present",
		Note:       note,
	}
	return s.repo.Upsert(ctx, attendance)
}
