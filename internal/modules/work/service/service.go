package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/work/contracts"
	"delivery-backend/internal/modules/work/domain"
	"delivery-backend/internal/modules/work/dto"
	"delivery-backend/internal/modules/work/repository"
)

type EmployeeData struct {
	ID                    uuid.UUID
	Name                  string
	NationalID            string
	BranchID              *uuid.UUID
	VehicleType           string
	MotorcycleNumber      string
	ApplicationID         string
	ApplicationType       string
	TotalDistance         float64
	LastOilChangeDistance float64
	LastOilChangeDate     *time.Time
}

type ExternalEmployeeProvider interface {
	GetEmployee(ctx context.Context, id uuid.UUID) (*EmployeeData, error)
	UpdateEmployeeOnStartWork(ctx context.Context, id uuid.UUID, appID, appType, motorcycleNumber string) error
	UpdateEmployeeOnEndWork(ctx context.Context, id uuid.UUID, addedDistance float64, totalOrders int) error
}

type VehicleData struct {
	ID                uuid.UUID
	PlateNumber       string
	VehicleType       string
	CurrentKM         float64
	LastOilChangeKM   float64
	TotalDistance     float64
	IsOdometerBroken  bool
	RegistrationImage string
	Status            string
}

type ExternalVehicleProvider interface {
	GetVehicleLastKM(ctx context.Context, plateNumber string) (float64, error)
	GetVehicleInfo(ctx context.Context, plateNumber string) (*VehicleData, error)
	UpdateVehicleKM(ctx context.Context, plateNumber string, km float64) error
	HasActiveMaintenance(ctx context.Context, motorcycleNumber string) (bool, error)
}

type StorageProvider interface {
	SaveBase64Image(data, folder string) (string, error)
}

type WorkService interface {
	contracts.IWorkContract
	StartWork(ctx context.Context, req dto.StartWorkRequest) (*domain.WorkSession, error)
	EndWork(ctx context.Context, req dto.EndWorkRequest, reviewerID *uuid.UUID, reviewerName string, isSupervisor bool) (*domain.WorkSession, error)
	UpdateWorkSession(ctx context.Context, sessionID uuid.UUID, req dto.UpdateWorkSessionRequest, adminName string) (*domain.WorkSession, error)
	ReviewSession(ctx context.Context, sessionID uuid.UUID, req dto.ReviewWorkSessionRequest, reviewerID *uuid.UUID, reviewerName string) (*domain.WorkSession, error)
	GetLastSessionOrVehicleKM(ctx context.Context, empID uuid.UUID, motorcycleNumber string) (*dto.LastKMResponse, error)
	ScanPlateImage(ctx context.Context, imageBase64 string) (map[string]interface{}, error)
	CheckOilChange(ctx context.Context, empID uuid.UUID, motorcycleNumber string) (*dto.OilChangeCheckResponse, error)
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*domain.WorkSession, error)
	GetActiveSessions(ctx context.Context, branchID *uuid.UUID) ([]domain.WorkSession, error)
}

type workService struct {
	repo            repository.WorkRepository
	empProvider     ExternalEmployeeProvider
	vehicleProvider ExternalVehicleProvider
	storage         StorageProvider
}

func NewWorkService(
	repo repository.WorkRepository,
	empProvider ExternalEmployeeProvider,
	vehicleProvider ExternalVehicleProvider,
	storage StorageProvider,
) WorkService {
	return &workService{
		repo:            repo,
		empProvider:     empProvider,
		vehicleProvider: vehicleProvider,
		storage:         storage,
	}
}

func oilChangeInterval(vehicleType string) float64 {
	if strings.EqualFold(vehicleType, "car") {
		return 10000
	}
	return 950 // motorcycle default
}

func (s *workService) StartWork(ctx context.Context, req dto.StartWorkRequest) (*domain.WorkSession, error) {
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
		vehicleInfo, _ = s.vehicleProvider.GetVehicleInfo(ctx, motorcycleNumber)
	}

	if vehicleInfo != nil {
		if vehicleInfo.VehicleType != "" {
			vehicleType = vehicleInfo.VehicleType
		}
		interval := oilChangeInterval(vehicleType)

		currKM := vehicleInfo.CurrentKM
		if req.StartKM > currKM {
			currKM = req.StartKM
		}

		distanceSinceOil := currKM - vehicleInfo.LastOilChangeKM
		if distanceSinceOil < 0 {
			distanceSinceOil = 0
		}

		if distanceSinceOil >= interval {
			return nil, fmt.Errorf("يجب تغيير زيت الدباب (%s) أولاً! المسافة المقطوعة منذ آخر تغيير زيت: %.0f كم (الحد الأقصى %.0f كم)", motorcycleNumber, distanceSinceOil, interval)
		}
	} else if motorcycleNumber == "" {
		// Fallback for employee with no bike assigned
		distanceSinceOil := emp.TotalDistance - emp.LastOilChangeDistance
		if distanceSinceOil < 0 {
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

	session := &domain.WorkSession{
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

	err = s.repo.ExecInTx(ctx, func(txRepo repository.WorkRepository) error {
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

func (s *workService) EndWork(ctx context.Context, req dto.EndWorkRequest, reviewerID *uuid.UUID, reviewerName string, isSupervisor bool) (*domain.WorkSession, error) {
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

	err = s.repo.ExecInTx(ctx, func(txRepo repository.WorkRepository) error {
		if err := txRepo.Update(ctx, session); err != nil {
			return err
		}
		if s.empProvider != nil {
			ordersToAdd := 0
			if session.IsReviewed {
				ordersToAdd = req.OrdersCount
			}
			if err := s.empProvider.UpdateEmployeeOnEndWork(ctx, empID, distance, ordersToAdd); err != nil {
				return err
			}
		}
		if s.vehicleProvider != nil && session.MotorcycleNumber != "" && req.EndKM > 0 {
			_ = s.vehicleProvider.UpdateVehicleKM(ctx, session.MotorcycleNumber, req.EndKM)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("فشل في إنهاء الشفت: %w", err)
	}

	return session, nil
}

func (s *workService) UpdateWorkSession(ctx context.Context, sessionID uuid.UUID, req dto.UpdateWorkSessionRequest, adminName string) (*domain.WorkSession, error) {
	session, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, errors.New("جلسة العمل غير موجودة")
	}

	session.IsEditedBySupervisor = true
	session.EditedByName = adminName

	if req.StartKM > 0 {
		session.StartKM = req.StartKM
	}
	if req.EndKM > 0 {
		session.EndKM = req.EndKM
	}
	if session.EndKM >= session.StartKM {
		session.Distance = session.EndKM - session.StartKM
	}
	if req.OrdersCount > 0 {
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

func (s *workService) ReviewSession(ctx context.Context, sessionID uuid.UUID, req dto.ReviewWorkSessionRequest, reviewerID *uuid.UUID, reviewerName string) (*domain.WorkSession, error) {
	session, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, errors.New("جلسة العمل غير موجودة")
	}

	wasReviewed := session.IsReviewed
	session.IsReviewed = req.IsReviewed
	session.ReviewNotes = req.ReviewNotes
	session.ReviewedBy = reviewerID

	if req.OrdersCount != nil {
		session.OrdersCount = *req.OrdersCount
	}
	if req.EndKM != nil {
		session.EndKM = *req.EndKM
	}
	if req.StartKM != nil {
		session.StartKM = *req.StartKM
	}
	if req.FuelCost != nil {
		session.FuelCost = *req.FuelCost
	}
	if session.EndKM >= session.StartKM {
		session.Distance = session.EndKM - session.StartKM
	}

	if err := s.repo.Update(ctx, session); err != nil {
		return nil, err
	}

	// If session wasn't previously reviewed and is now approved, credit orders to employee
	if !wasReviewed && session.IsReviewed && s.empProvider != nil && session.EmployeeID != nil && session.OrdersCount > 0 {
		_ = s.empProvider.UpdateEmployeeOnEndWork(ctx, *session.EmployeeID, 0, session.OrdersCount)
	}

	return session, nil
}

func (s *workService) GetLastSessionOrVehicleKM(ctx context.Context, empID uuid.UUID, motorcycleNumber string) (*dto.LastKMResponse, error) {
	resp := &dto.LastKMResponse{}

	lastSession, _ := s.repo.FindLastCompletedSession(ctx, empID)
	if lastSession != nil {
		resp.LastKM = lastSession.EndKM
		resp.LastEndKM = lastSession.EndKM
		resp.LastStartKM = lastSession.StartKM
		resp.MotorcycleNumber = lastSession.MotorcycleNumber
	}

	targetBike := motorcycleNumber
	if targetBike == "" && lastSession != nil {
		targetBike = lastSession.MotorcycleNumber
	}

	if targetBike != "" && s.vehicleProvider != nil {
		vInfo, err := s.vehicleProvider.GetVehicleInfo(ctx, targetBike)
		if err == nil && vInfo != nil {
			resp.VehicleLastKM = vInfo.CurrentKM
			resp.LastKM = vInfo.CurrentKM
			resp.LastEndKM = vInfo.CurrentKM
			resp.RegistrationImage = vInfo.RegistrationImage
			resp.IsOdometerBroken = vInfo.IsOdometerBroken

			interval := oilChangeInterval(vInfo.VehicleType)
			distanceSinceOil := vInfo.CurrentKM - vInfo.LastOilChangeKM
			if distanceSinceOil < 0 {
				distanceSinceOil = 0
			}
			remaining := interval - distanceSinceOil
			if remaining < 0 {
				remaining = 0
			}
			resp.DistanceSinceOil = distanceSinceOil
			resp.RemainingOilKM = remaining
			resp.NeedsOilChange = distanceSinceOil >= interval

			if lastSession != nil && lastSession.MotorcycleNumber != targetBike {
				resp.IsDifferentBike = true
			}
			if lastSession != nil && resp.VehicleLastKM > lastSession.EndKM && lastSession.EndKM > 0 {
				resp.HasGap = true
				resp.GapKM = resp.VehicleLastKM - lastSession.EndKM
			}
		}
	}

	return resp, nil
}

func (s *workService) ScanPlateImage(ctx context.Context, imageBase64 string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"success": true,
		"message": "plate scanned successfully",
	}, nil
}

func (s *workService) CheckOilChange(ctx context.Context, empID uuid.UUID, motorcycleNumber string) (*dto.OilChangeCheckResponse, error) {
	emp, err := s.empProvider.GetEmployee(ctx, empID)
	if err != nil {
		return nil, errors.New("الموظف غير موجود")
	}

	bike := motorcycleNumber
	if bike == "" {
		bike = emp.MotorcycleNumber
	}

	var vehicleInfo *VehicleData
	if bike != "" && s.vehicleProvider != nil {
		vehicleInfo, _ = s.vehicleProvider.GetVehicleInfo(ctx, bike)
	}

	if vehicleInfo != nil {
		vType := vehicleInfo.VehicleType
		if vType == "" {
			vType = emp.VehicleType
		}
		interval := oilChangeInterval(vType)
		distanceSinceOil := vehicleInfo.CurrentKM - vehicleInfo.LastOilChangeKM
		if distanceSinceOil < 0 {
			distanceSinceOil = 0
		}
		needsOil := distanceSinceOil >= interval
		remaining := interval - distanceSinceOil
		if remaining < 0 {
			remaining = 0
		}
		pct := (distanceSinceOil / interval) * 100
		if pct > 100 {
			pct = 100
		}

		hasMaintenance := false
		if s.vehicleProvider != nil {
			hasMaintenance, _ = s.vehicleProvider.HasActiveMaintenance(ctx, bike)
		}

		return &dto.OilChangeCheckResponse{
			NeedsOilChange:          needsOil,
			CurrentDistance:         vehicleInfo.CurrentKM,
			LastOilChangeDistance:   vehicleInfo.LastOilChangeKM,
			DistanceSinceOil:        distanceSinceOil,
			RemainingDistance:       remaining,
			Percentage:              pct,
			Interval:                interval,
			VehicleType:             vType,
			HasActiveMaintenanceReq: hasMaintenance,
			LastOilChangeDate:       "",
		}, nil
	}

	// Fallback to employee profile if no vehicle is found
	interval := oilChangeInterval(emp.VehicleType)
	distanceSinceOil := emp.TotalDistance - emp.LastOilChangeDistance
	if distanceSinceOil < 0 {
		distanceSinceOil = 0
	}
	needsOil := distanceSinceOil >= interval
	remaining := interval - distanceSinceOil
	if remaining < 0 {
		remaining = 0
	}
	pct := (distanceSinceOil / interval) * 100
	if pct > 100 {
		pct = 100
	}

	hasMaintenance := false
	if bike != "" && s.vehicleProvider != nil {
		hasMaintenance, _ = s.vehicleProvider.HasActiveMaintenance(ctx, bike)
	}

	lastDate := ""
	if emp.LastOilChangeDate != nil {
		lastDate = emp.LastOilChangeDate.Format("2006-01-02")
	}

	return &dto.OilChangeCheckResponse{
		NeedsOilChange:          needsOil,
		CurrentDistance:         emp.TotalDistance,
		LastOilChangeDistance:   emp.LastOilChangeDistance,
		DistanceSinceOil:        distanceSinceOil,
		RemainingDistance:       remaining,
		Percentage:              pct,
		Interval:                interval,
		VehicleType:             emp.VehicleType,
		HasActiveMaintenanceReq: hasMaintenance,
		LastOilChangeDate:       lastDate,
	}, nil
}

func (s *workService) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*domain.WorkSession, error) {
	return s.repo.FindByID(ctx, sessionID)
}

func (s *workService) GetActiveSessions(ctx context.Context, branchID *uuid.UUID) ([]domain.WorkSession, error) {
	return s.repo.GetActiveSessions(ctx, branchID)
}

// IWorkContract implementation
func (s *workService) GetActiveSession(ctx context.Context, empID uuid.UUID) (*contracts.WorkSessionDTO, error) {
	sess, err := s.repo.FindActiveSessionByEmployeeID(ctx, empID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toContractDTO(sess), nil
}

func (s *workService) GetLastCompletedSession(ctx context.Context, empID uuid.UUID) (*contracts.WorkSessionDTO, error) {
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

func toContractDTO(s *domain.WorkSession) *contracts.WorkSessionDTO {
	if s == nil {
		return nil
	}
	return &contracts.WorkSessionDTO{
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
