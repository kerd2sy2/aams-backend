package work

import (
	"context"

	"strings"
	"time"

	"github.com/google/uuid"

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
	SendShiftApprovalNotification(ctx context.Context, empID uuid.UUID, sessionID uuid.UUID, ordersCount int, fuelCost float64, supervisorName string) error
	SendShiftStartedNotificationToSupervisor(ctx context.Context, empID uuid.UUID, session *WorkSession, previousMotorcycle string) error
	SendShiftEndedNotificationToSupervisor(ctx context.Context, empID uuid.UUID, session *WorkSession) error
}

type VehicleData struct {
	ID                uuid.UUID
	PlateNumber       string
	VehicleType       string
	CurrentKM         float64
	LastOilChangeKM   float64
	LastOilChangeDate *time.Time
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
	EnsureVehicleExists(ctx context.Context, plateNumber string, vehicleType string, branchID *uuid.UUID, startKM float64) (*VehicleData, error)
}

type StorageProvider interface {
	SaveBase64Image(data, folder string) (string, error)
}

type WorkService interface {
	IWorkContract
	StartWork(ctx context.Context, req StartWorkRequest) (*WorkSession, error)
	EndWork(ctx context.Context, req EndWorkRequest, reviewerID *uuid.UUID, reviewerName string, isSupervisor bool) (*WorkSession, error)
	UpdateWorkSession(ctx context.Context, sessionID uuid.UUID, req UpdateWorkSessionRequest, adminName string) (*WorkSession, error)
	ReviewSession(ctx context.Context, sessionID uuid.UUID, req ReviewWorkSessionRequest, reviewerID *uuid.UUID, reviewerName string) (*WorkSession, error)
	GetLastSessionOrVehicleKM(ctx context.Context, empID uuid.UUID, motorcycleNumber string) (*LastKMResponse, error)
	ScanPlateImage(ctx context.Context, imageBase64 string) (map[string]interface{}, error)
	CheckOilChange(ctx context.Context, empID uuid.UUID, motorcycleNumber string) (*OilChangeCheckResponse, error)
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*WorkSession, error)
	GetActiveSessions(ctx context.Context, branchID *uuid.UUID) ([]WorkSession, error)
	SupervisorStartShift(ctx context.Context, req SupervisorStartShiftRequest, supervisorID uuid.UUID, supervisorName, supervisorRole string, supervisorBranchID *uuid.UUID) (*WorkSession, error)
	SupervisorEndShift(ctx context.Context, req SupervisorEndShiftRequest, supervisorID uuid.UUID, supervisorName, supervisorRole string, supervisorBranchID *uuid.UUID) (*WorkSession, error)
	SaveManualEntries(ctx context.Context, req ManualWorkEntryRequest, reviewerID *uuid.UUID, reviewerName string) ([]WorkSession, error)
}

type workService struct {
	repo            WorkRepository
	empProvider     ExternalEmployeeProvider
	vehicleProvider ExternalVehicleProvider
	storage         StorageProvider
}

func NewWorkService(
	repo WorkRepository,
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
