package work

import (
	"context"
	"errors"

	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *workService) UpdateWorkSession(ctx context.Context, sessionID uuid.UUID, req UpdateWorkSessionRequest, adminName string) (*WorkSession, error) {
	session, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, errors.New("جلسة العمل غير موجودة")
	}

	session.IsEditedBySupervisor = true
	session.EditedByName = adminName

	if req.EmployeeID != "" {
		if empUUID, parseErr := uuid.Parse(req.EmployeeID); parseErr == nil && empUUID != uuid.Nil {
			session.EmployeeID = &empUUID
		}
	}

	if st := req.StartTime.Time(); st != nil && !st.IsZero() {
		session.StartTime = *st
	}
	if et := req.EndTime.Time(); et != nil && !et.IsZero() {
		session.EndTime = et
		if session.Status == "ACTIVE" {
			session.Status = "COMPLETED"
		}
	}

	if session.EndTime != nil && !session.EndTime.IsZero() && session.EndTime.Before(session.StartTime) {
		return nil, errors.New("وقت النهاية يجب أن يكون بعد وقت البداية")
	}

	if req.StartKM > 0 {
		session.StartKM = req.StartKM
	}
	if req.EndKM > 0 {
		session.EndKM = req.EndKM
	}
	if session.EndKM >= session.StartKM {
		session.Distance = session.EndKM - session.StartKM
	}
	if req.OrdersCount >= 0 {
		session.OrdersCount = req.OrdersCount
	}
	if req.FuelCost >= 0 {
		session.FuelCost = req.FuelCost
	}
	if req.ApplicationType != "" {
		session.ApplicationType = req.ApplicationType
	}
	if req.Notes != "" {
		session.Notes = req.Notes
	}

	if err := s.repo.Update(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *workService) ReviewSession(ctx context.Context, sessionID uuid.UUID, req ReviewWorkSessionRequest, reviewerID *uuid.UUID, reviewerName string) (*WorkSession, error) {
	session, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, errors.New("جلسة العمل غير موجودة")
	}

	wasReviewed := session.IsReviewed
	session.IsReviewed = req.IsReviewed
	session.ReviewNotes = req.ReviewNotes
	session.ReviewedBy = reviewerID

	// If session is approved/reviewed by supervisor, ensure it is marked COMPLETED if it was ACTIVE
	if req.IsReviewed && (session.Status == "ACTIVE" || session.EndTime == nil) {
		session.Status = "COMPLETED"
		if session.EndTime == nil {
			now := time.Now()
			session.EndTime = &now
		}
	}

	if req.OrdersCount != nil {
		if session.OriginalOrdersCount == 0 && session.OrdersCount > 0 {
			session.OriginalOrdersCount = session.OrdersCount
		}
		if *req.OrdersCount != session.OrdersCount {
			session.IsEditedBySupervisor = true
			if reviewerName != "" {
				session.EditedByName = reviewerName
			}
		}
		session.OrdersCount = *req.OrdersCount
	}
	if req.EndKM != nil {
		if session.OriginalEndKM == 0 && session.EndKM > 0 {
			session.OriginalEndKM = session.EndKM
		}
		if *req.EndKM != session.EndKM {
			session.IsEditedBySupervisor = true
			if reviewerName != "" {
				session.EditedByName = reviewerName
			}
		}
		session.EndKM = *req.EndKM
	}
	if req.StartKM != nil {
		session.StartKM = *req.StartKM
	}
	if req.FuelCost != nil {
		session.FuelCost = *req.FuelCost
	}
	if req.StartTime != nil && req.StartTime.Time() != nil {
		session.StartTime = *req.StartTime.Time()
		session.IsEditedBySupervisor = true
	}
	if req.EndTime != nil && req.EndTime.Time() != nil {
		session.EndTime = req.EndTime.Time()
		session.IsEditedBySupervisor = true
	}
	if req.ApplicationID != nil && *req.ApplicationID != "" {
		session.ApplicationID = *req.ApplicationID
		session.IsEditedBySupervisor = true
		if session.EmployeeID != nil {
			_ = s.empProvider.UpdateEmployeeOnStartWork(ctx, *session.EmployeeID, *req.ApplicationID, "", "")
		}
	}
	if req.ReviewVideoURL != nil && *req.ReviewVideoURL != "" {
		session.ReviewVideoURL = *req.ReviewVideoURL
	}
	if session.EndKM >= session.StartKM {
		session.Distance = session.EndKM - session.StartKM
	}

	if err := s.repo.Update(ctx, session); err != nil {
		return nil, err
	}

	// If session wasn't previously reviewed and is now approved, credit distance and orders to employee, update vehicle KM, and notify
	if !wasReviewed && session.IsReviewed && session.EmployeeID != nil {
		if s.empProvider != nil {
			_ = s.empProvider.UpdateEmployeeOnEndWork(ctx, *session.EmployeeID, session.Distance, session.OrdersCount)
		}
		if s.vehicleProvider != nil && session.MotorcycleNumber != "" {
			_ = s.vehicleProvider.UpdateVehicleKM(ctx, session.MotorcycleNumber, session.EndKM)
		}
		if s.empProvider != nil {
			_ = s.empProvider.SendShiftApprovalNotification(ctx, *session.EmployeeID, session.ID, session.OrdersCount, session.FuelCost, reviewerName)
		}
	}

	return session, nil
}

func (s *workService) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*WorkSession, error) {
	return s.repo.FindByID(ctx, sessionID)
}

func (s *workService) GetActiveSessions(ctx context.Context, branchID *uuid.UUID) ([]WorkSession, error) {
	return s.repo.GetActiveSessions(ctx, branchID)
}

// IWorkContract implementation
func (s *workService) GetActiveSession(ctx context.Context, empID uuid.UUID) (*WorkSessionDTO, error) {
	sess, err := s.repo.FindActiveSessionByEmployeeID(ctx, empID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toContractDTO(sess), nil
}

func (s *workService) GetLastCompletedSession(ctx context.Context, empID uuid.UUID) (*WorkSessionDTO, error) {
	sess, err := s.repo.FindLastCompletedSession(ctx, empID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toContractDTO(sess), nil
}

func (s *workService) CountTodaySessions(ctx context.Context, empID uuid.UUID) (int64, error) {
	return s.repo.CountTodaySessions(ctx, empID)
}

func toContractDTO(s *WorkSession) *WorkSessionDTO {
	if s == nil {
		return nil
	}
	return &WorkSessionDTO{
		ID:               s.ID,
		EmployeeID:       s.EmployeeID,
		StartTime:        s.StartTime,
		EndTime:          s.EndTime,
		StartKM:          s.StartKM,
		EndKM:            s.EndKM,
		Distance:         s.Distance,
		OrdersCount:      s.OrdersCount,
		FuelCost:         s.FuelCost,
		ApplicationID:    s.ApplicationID,
		ApplicationType:  s.ApplicationType,
		VehicleType:      s.VehicleType,
		MotorcycleNumber: s.MotorcycleNumber,
		Status:           s.Status,
	}
}
