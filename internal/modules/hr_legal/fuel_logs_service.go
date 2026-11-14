package hr_legal

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

func (s *hrLegalService) CreateFuelLog(ctx context.Context, req CreateFuelLogRequest, adminBranchID *uuid.UUID) (*FuelLog, error) {
	branchID := req.BranchID
	if branchID == nil && adminBranchID != nil {
		branchID = adminBranchID
	}

	var fuelDate time.Time
	if req.FuelDate != "" {
		if t, err := time.Parse("2006-01-02", req.FuelDate); err == nil {
			fuelDate = t
		} else {
			fuelDate = time.Now()
		}
	} else {
		fuelDate = time.Now()
	}

	log := &FuelLog{
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

func (s *hrLegalService) UpdateFuelLog(ctx context.Context, id uuid.UUID, req UpdateFuelLogRequest) (*FuelLog, error) {
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

func (s *hrLegalService) GetFuelLogByID(ctx context.Context, id uuid.UUID) (*FuelLog, error) {
	return s.repo.FindFuelLogByID(ctx, id)
}

func (s *hrLegalService) GetAllFuelLogs(ctx context.Context, filter FuelLogFilter, adminBranchID *uuid.UUID) ([]FuelLog, int64, error) {
	if adminBranchID != nil {
		filter.BranchID = adminBranchID
	}
	return s.repo.FindFuelLogs(ctx, filter)
}

// ------------------------------------------------------------------
// Contract implementations
// ------------------------------------------------------------------
func (s *hrLegalService) CheckEmployeeOnLeave(ctx context.Context, empID uuid.UUID, date time.Time) (*LeaveCheckDTO, error) {
	dStr := date.Format("2006-01-02")
	lr, err := s.repo.FindActiveLeave(ctx, empID, dStr)
	if err != nil || lr == nil {
		return &LeaveCheckDTO{
			EmployeeID: empID,
			IsOnLeave:  false,
		}, nil
	}

	t1, _ := time.Parse("2006-01-02", lr.StartDate)
	t2, _ := time.Parse("2006-01-02", lr.EndDate)

	return &LeaveCheckDTO{
		EmployeeID: empID,
		IsOnLeave:  true,
		LeaveType:  lr.LeaveType,
		StartDate:  t1,
		EndDate:    t2,
	}, nil
}

func (s *hrLegalService) GetPendingRequestsCount(ctx context.Context) (investigations int64, leaves int64, err error) {
	invCount, err1 := s.repo.CountPendingInvestigations(ctx)
	leaveCount, err2 := s.repo.CountPendingLeaveRequests(ctx)
	if err1 != nil {
		return 0, 0, err1
	}
	if err2 != nil {
		return 0, 0, err2
	}
	return invCount, leaveCount, nil
}
