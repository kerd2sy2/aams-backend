package work

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

)

func (s *workService) StartWork(ctx context.Context, req StartWorkRequest) (*WorkSession, error) {
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, errors.New("معرف الموظف غير صالح")
	}

	emp, err := s.empProvider.GetEmployee(ctx, empID)
	if err != nil {
		return nil, errors.New("الموظف غير موجود")
	}

	activeSession, _ := s.repo.FindActiveSessionByEmployeeID(ctx, empID)
	if activeSession != nil {
		return nil, errors.New("الموظف لديه شفت عمل نشط بالفعل حالياً")
	}

	motorcycleNumber := emp.MotorcycleNumber
	if req.MotorcycleNumber != "" {
		motorcycleNumber = req.MotorcycleNumber
	}

	vehicleType := emp.VehicleType
	if req.VehicleType != "" {
		vehicleType = req.VehicleType
	}

	// 1. Vehicle-based Oil Check: Calculate oil based on motorcycle rather than employee
	var vehicleInfo *VehicleData
	if motorcycleNumber != "" && s.vehicleProvider != nil {
		vehicleInfo, _ = s.vehicleProvider.EnsureVehicleExists(ctx, motorcycleNumber, vehicleType, emp.BranchID, req.StartKM)
	}

	if vehicleInfo != nil {
		if vehicleInfo.VehicleType != "" {
			vehicleType = vehicleInfo.VehicleType
		}
		interval := oilChangeInterval(vehicleType)

		isBroken := vehicleInfo.IsOdometerBroken ||
			strings.Contains(motorcycleNumber, "7569") ||
			strings.Contains(motorcycleNumber, "8043") ||
			strings.Contains(vehicleInfo.PlateNumber, "7569") ||
			strings.Contains(vehicleInfo.PlateNumber, "8043")

		if isBroken {
			if vehicleInfo.LastOilChangeDate != nil {
				days := time.Since(*vehicleInfo.LastOilChangeDate).Hours() / 24.0
				if days >= 5.0 {
					return nil, fmt.Errorf("يجب تغيير زيت الدباب (%s) أولاً! مر %.0f أيام منذ آخر تغيير زيت (الحد الأقصى 5 أيام)", motorcycleNumber, math.Floor(days))
				}
			} else {
				return nil, fmt.Errorf("يجب تغيير زيت الدباب (%s) أولاً! لم يتم تسجيل تاريخ آخر تغيير زيت", motorcycleNumber)
			}
		} else {
			currKM := vehicleInfo.CurrentKM
			if req.StartKM > currKM {
				currKM = req.StartKM
			}

			var distanceSinceOil float64
			if vehicleInfo.LastOilChangeKM > 0 || vehicleInfo.LastOilChangeDate != nil {
				distanceSinceOil = currKM - vehicleInfo.LastOilChangeKM
				if distanceSinceOil < 0 {
					distanceSinceOil = 0
				}
			} else {
				// Vehicle has no prior oil change recorded in system.
				// If it has tracked distance in AAMS, use that; otherwise this initial shift establishes the baseline.
				if vehicleInfo.TotalDistance > 0 {
					distanceSinceOil = vehicleInfo.TotalDistance
				} else {
					distanceSinceOil = 0
					vehicleInfo.LastOilChangeKM = req.StartKM
					if s.vehicleProvider != nil && motorcycleNumber != "" {
						_ = s.vehicleProvider.UpdateVehicleKM(ctx, motorcycleNumber, req.StartKM)
					}
				}
			}

			if distanceSinceOil >= interval {
				return nil, fmt.Errorf("يجب تغيير زيت الدباب (%s) أولاً! المسافة المقطوعة منذ آخر تغيير زيت: %.0f كم (الحد الأقصى %.0f كم)", motorcycleNumber, distanceSinceOil, interval)
			}
		}
	} else if motorcycleNumber == "" {
		// Fallback for employee with no bike assigned
		var distanceSinceOil float64
		if emp.LastOilChangeDistance > 0 {
			distanceSinceOil = emp.TotalDistance - emp.LastOilChangeDistance
			if distanceSinceOil < 0 {
				distanceSinceOil = 0
			}
		} else {
			distanceSinceOil = 0
		}
		interval := oilChangeInterval(vehicleType)
		if distanceSinceOil >= interval {
			return nil, fmt.Errorf("يجب تغيير الزيت أولاً! المسافة المقطوعة منذ آخر تغيير زيت: %.0f كم (الحد الأقصى %.0f كم)", distanceSinceOil, interval)
		}
	}

	startImg := req.StartKMImage
	if s.storage != nil && strings.HasPrefix(startImg, "data:image") {
		if savedUrl, err := s.storage.SaveBase64Image(startImg, "odometer"); err == nil && savedUrl != "" {
			startImg = savedUrl
		}
	}

	startPlateImg := req.StartPlateImage
	if s.storage != nil && strings.HasPrefix(startPlateImg, "data:image") {
		if savedUrl, err := s.storage.SaveBase64Image(startPlateImg, "plates"); err == nil && savedUrl != "" {
			startPlateImg = savedUrl
		}
	}

	appID := emp.ApplicationID
	if req.ApplicationID != "" {
		appID = req.ApplicationID
	}
	appType := emp.ApplicationType
	if req.ApplicationType != "" {
		appType = req.ApplicationType
	}

	session := &WorkSession{
		ID:               uuid.New(),
		EmployeeID:       &empID,
		StartTime:        time.Now(),
		StartKM:          req.StartKM,
		StartKMImage:     startImg,
		StartPlateImage:  startPlateImg,
		ApplicationID:    appID,
		ApplicationType:  appType,
		VehicleType:      vehicleType,
		MotorcycleNumber: motorcycleNumber,
		Notes:            req.Notes,
		Status:           "ACTIVE",
		OriginalStartKM:  req.StartKM,
	}

	err = s.repo.ExecInTx(ctx, func(txRepo WorkRepository) error {
		if err := txRepo.Create(ctx, session); err != nil {
			return err
		}
		if s.empProvider != nil {
			return s.empProvider.UpdateEmployeeOnStartWork(ctx, empID, appID, appType, motorcycleNumber)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("فشل في بدء الشفت: %w", err)
	}

	return session, nil
}

func (s *workService) EndWork(ctx context.Context, req EndWorkRequest, reviewerID *uuid.UUID, reviewerName string, isSupervisor bool) (*WorkSession, error) {
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, errors.New("معرف الموظف غير صالح")
	}

	session, err := s.repo.FindActiveSessionByEmployeeID(ctx, empID)
	if err != nil {
		return nil, errors.New("لا يوجد شفت نشط لهذا الموظف")
	}

	if req.EndKM > 0 && req.EndKM < session.StartKM {
		return nil, errors.New("عداد النهاية لا يمكن أن يكون أقل من عداد البداية")
	}

	distance := 0.0
	if req.EndKM >= session.StartKM {
		distance = req.EndKM - session.StartKM
	}

	now := time.Now()
	session.EndTime = &now
	session.EndKM = req.EndKM
	session.Distance = distance
	session.OrdersCount = req.OrdersCount
	session.FuelCost = req.FuelCost
	session.Status = "COMPLETED"
	session.OriginalEndKM = req.EndKM
	session.OriginalOrdersCount = req.OrdersCount

	if req.ApplicationID != "" {
		session.ApplicationID = req.ApplicationID
	}
	if req.ApplicationType != "" {
		session.ApplicationType = req.ApplicationType
	}
	if req.Notes != "" {
		session.Notes = req.Notes
	}

	endImg := req.EndKMImage
	if s.storage != nil && strings.HasPrefix(endImg, "data:image") {
		if savedUrl, err := s.storage.SaveBase64Image(endImg, "odometer"); err == nil && savedUrl != "" {
			endImg = savedUrl
		}
	}
	session.EndKMImage = endImg

	if isSupervisor {
		session.IsReviewed = true
		session.ReviewedBy = reviewerID
		session.ReviewNotes = req.ReviewNotes
	}

	err = s.repo.ExecInTx(ctx, func(txRepo WorkRepository) error {
		if err := txRepo.Update(ctx, session); err != nil {
			return err
		}
		if s.empProvider != nil {
			distanceToAdd := 0.0
			ordersToAdd := 0
			if session.IsReviewed {
				distanceToAdd = distance
				ordersToAdd = req.OrdersCount
			}
			if err := s.empProvider.UpdateEmployeeOnEndWork(ctx, empID, distanceToAdd, ordersToAdd); err != nil {
				return err
			}
		}
		if s.vehicleProvider != nil && session.MotorcycleNumber != "" {
			_ = s.vehicleProvider.UpdateVehicleKM(ctx, session.MotorcycleNumber, req.EndKM)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("فشل في إنهاء الشفت: %w", err)
	}

	if isSupervisor && session.EmployeeID != nil && s.empProvider != nil {
		_ = s.empProvider.SendShiftApprovalNotification(ctx, *session.EmployeeID, session.ID, session.OrdersCount, session.FuelCost, reviewerName)
	} else if !isSupervisor && session.EmployeeID != nil && s.empProvider != nil {
		_ = s.empProvider.SendShiftEndedNotificationToSupervisor(ctx, *session.EmployeeID, session)
	}

	return session, nil
}

func (s *workService) SupervisorStartShift(
	ctx context.Context,
	req SupervisorStartShiftRequest,
	supervisorID uuid.UUID,
	supervisorName, supervisorRole string,
	supervisorBranchID *uuid.UUID,
) (*WorkSession, error) {
	var targetEmpID *uuid.UUID

	if strings.TrimSpace(req.EmployeeID) != "" {
		empUUID, err := uuid.Parse(req.EmployeeID)
		if err != nil {
			return nil, errors.New("معرف الموظف غير صالح")
		}
		if s.empProvider != nil {
			emp, err := s.empProvider.GetEmployee(ctx, empUUID)
			if err != nil || emp == nil {
				return nil, errors.New("الموظف غير موجود")
			}
			if !strings.EqualFold(supervisorRole, "SUPER_ADMIN") {
				if supervisorBranchID == nil || emp.BranchID == nil || *emp.BranchID != *supervisorBranchID {
					return nil, ErrCrossBranchForbidden
				}
			}
		}
		targetEmpID = &empUUID
	} else {
		targetEmpID = &supervisorID
	}

	startImg := req.StartKMImage
	if s.storage != nil && strings.HasPrefix(startImg, "data:image") {
		if savedUrl, err := s.storage.SaveBase64Image(startImg, "odometer"); err == nil && savedUrl != "" {
			startImg = savedUrl
		}
	}

	startPlateImg := req.StartPlateImage
	if s.storage != nil && strings.HasPrefix(startPlateImg, "data:image") {
		if savedUrl, err := s.storage.SaveBase64Image(startPlateImg, "plates"); err == nil && savedUrl != "" {
			startPlateImg = savedUrl
		}
	}

	session := &WorkSession{
		ID:                   uuid.New(),
		EmployeeID:           targetEmpID,
		StartTime:            time.Now(),
		StartKM:              req.StartKM,
		StartKMImage:         startImg,
		StartPlateImage:      startPlateImg,
		ApplicationID:        req.ApplicationID,
		ApplicationType:      req.ApplicationType,
		VehicleType:          req.VehicleType,
		MotorcycleNumber:     req.MotorcycleNumber,
		IsEditedBySupervisor: true,
		EditedByName:         supervisorName,
		ReviewedBy:           &supervisorID,
		ReviewNotes:          req.OverrideReason,
		Notes:                req.Notes,
		Status:               "ACTIVE",
		OriginalStartKM:      req.StartKM,
	}

	if req.MotorcycleNumber != "" && s.vehicleProvider != nil {
		_, _ = s.vehicleProvider.EnsureVehicleExists(ctx, req.MotorcycleNumber, req.VehicleType, supervisorBranchID, req.StartKM)
	}

	err := s.repo.ExecInTx(ctx, func(txRepo WorkRepository) error {
		active, _ := txRepo.FindActiveSessionByEmployeeID(ctx, *targetEmpID)
		if active != nil {
			return ErrActiveSessionExists
		}
		if err := txRepo.Create(ctx, session); err != nil {
			return err
		}
		if strings.TrimSpace(req.EmployeeID) != "" && s.empProvider != nil {
			return s.empProvider.UpdateEmployeeOnStartWork(ctx, *targetEmpID, req.ApplicationID, req.ApplicationType, req.MotorcycleNumber)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *workService) SupervisorEndShift(
	ctx context.Context,
	req SupervisorEndShiftRequest,
	supervisorID uuid.UUID,
	supervisorName, supervisorRole string,
	supervisorBranchID *uuid.UUID,
) (*WorkSession, error) {
	var targetEmpID *uuid.UUID

	if strings.TrimSpace(req.EmployeeID) != "" {
		empUUID, err := uuid.Parse(req.EmployeeID)
		if err != nil {
			return nil, errors.New("معرف الموظف غير صالح")
		}
		if s.empProvider != nil {
			emp, err := s.empProvider.GetEmployee(ctx, empUUID)
			if err != nil || emp == nil {
				return nil, errors.New("الموظف غير موجود")
			}
			if !strings.EqualFold(supervisorRole, "SUPER_ADMIN") {
				if supervisorBranchID == nil || emp.BranchID == nil || *emp.BranchID != *supervisorBranchID {
					return nil, ErrCrossBranchForbidden
				}
			}
		}
		targetEmpID = &empUUID
	} else {
		targetEmpID = &supervisorID
	}

	endImg := req.EndKMImage
	if s.storage != nil && strings.HasPrefix(endImg, "data:image") {
		if savedUrl, err := s.storage.SaveBase64Image(endImg, "odometer"); err == nil && savedUrl != "" {
			endImg = savedUrl
		}
	}

	var updatedSession *WorkSession

	err := s.repo.ExecInTx(ctx, func(txRepo WorkRepository) error {
		session, err := txRepo.FindActiveSessionByEmployeeIDForUpdate(ctx, *targetEmpID)
		if err != nil || session == nil {
			return ErrNoActiveSession
		}

		if req.EndKM > 0 && req.EndKM < session.StartKM {
			return errors.New("عداد النهاية لا يمكن أن يكون أقل من عداد البداية")
		}

		distance := 0.0
		if req.EndKM >= session.StartKM {
			distance = req.EndKM - session.StartKM
		}

		now := time.Now()
		session.EndTime = &now
		session.EndKM = req.EndKM
		session.EndKMImage = endImg
		session.Distance = distance
		session.OrdersCount = req.OrdersCount
		session.FuelCost = req.FuelCost
		session.Status = "COMPLETED"
		session.IsReviewed = true
		session.ReviewedBy = &supervisorID
		session.IsEditedBySupervisor = true
		session.EditedByName = supervisorName
		if req.OverrideReason != "" {
			session.ReviewNotes = req.OverrideReason
		}
		if req.ReviewNotes != "" {
			if session.ReviewNotes != "" {
				session.ReviewNotes += " | " + req.ReviewNotes
			} else {
				session.ReviewNotes = req.ReviewNotes
			}
		}
		if req.Notes != "" {
			session.Notes = req.Notes
		}
		session.OriginalEndKM = req.EndKM
		session.OriginalOrdersCount = req.OrdersCount

		if err := txRepo.Update(ctx, session); err != nil {
			return err
		}

		if strings.TrimSpace(req.EmployeeID) != "" && s.empProvider != nil {
			if err := s.empProvider.UpdateEmployeeOnEndWork(ctx, *targetEmpID, distance, req.OrdersCount); err != nil {
				return err
			}
		}

		if s.vehicleProvider != nil && session.MotorcycleNumber != "" && req.EndKM > 0 {
			_ = s.vehicleProvider.UpdateVehicleKM(ctx, session.MotorcycleNumber, req.EndKM)
		}

		updatedSession = session
		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedSession, nil
}
