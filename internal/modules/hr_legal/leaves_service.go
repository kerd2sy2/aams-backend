package hr_legal

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *hrLegalService) CreateLeaveRequest(ctx context.Context, req CreateLeaveRequestRequest) (*LeaveRequest, error) {
	leaveType := "ANNUAL"
	if req.LeaveType != "" {
		leaveType = req.LeaveType
	}
	daysCount := req.DaysCount
	if daysCount <= 0 {
		t1, err1 := time.Parse("2006-01-02", req.StartDate)
		t2, err2 := time.Parse("2006-01-02", req.EndDate)
		if err1 == nil && err2 == nil && !t2.Before(t1) {
			daysCount = int(math.Ceil(t2.Sub(t1).Hours()/24)) + 1
		}
		if daysCount <= 0 {
			daysCount = 1
		}
	}

	leave := &LeaveRequest{
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

func (s *hrLegalService) UpdateLeaveRequestStatus(ctx context.Context, id uuid.UUID, req UpdateLeaveRequestStatusRequest) (*LeaveRequest, error) {
	leave, err := s.repo.FindLeaveRequestByID(ctx, id)
	if err != nil {
		return nil, errors.New("طلب الإجازة غير موجود")
	}

	leave.Status = strings.ToUpper(req.Status)
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

func (s *hrLegalService) GetLeaveRequestByID(ctx context.Context, id uuid.UUID) (*LeaveRequest, error) {
	return s.repo.FindLeaveRequestByID(ctx, id)
}

func (s *hrLegalService) GetAllLeaveRequests(ctx context.Context, filter LeaveRequestFilter) ([]LeaveRequest, int64, error) {
	return s.repo.FindLeaveRequests(ctx, filter)
}

func (s *hrLegalService) GetPendingLeaveCount(ctx context.Context) (int64, error) {
	return s.repo.CountPendingLeaveRequests(ctx)
}
