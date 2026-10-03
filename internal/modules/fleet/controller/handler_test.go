package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/fleet/contracts"
	"delivery-backend/internal/modules/fleet/controller"
	"delivery-backend/internal/modules/fleet/domain"
	"delivery-backend/internal/modules/fleet/dto"
)

type mockVehicleService struct {
	createFunc func(ctx context.Context, req dto.CreateVehicleRequest) (*domain.Vehicle, error)
	updateFunc func(ctx context.Context, id uuid.UUID, req dto.UpdateVehicleRequest) (*domain.Vehicle, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
	getFunc    func(ctx context.Context, id uuid.UUID) (*dto.VehicleResponse, error)
	plateFunc  func(ctx context.Context, plate string) (*domain.Vehicle, error)
	getAllFunc func(ctx context.Context, filter dto.VehicleFilter) ([]dto.VehicleResponse, int64, error)
}

func (m *mockVehicleService) Create(ctx context.Context, req dto.CreateVehicleRequest) (*domain.Vehicle, error) {
	return m.createFunc(ctx, req)
}
func (m *mockVehicleService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateVehicleRequest) (*domain.Vehicle, error) {
	return m.updateFunc(ctx, id, req)
}
func (m *mockVehicleService) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteFunc(ctx, id)
}
func (m *mockVehicleService) GetByID(ctx context.Context, id uuid.UUID) (*dto.VehicleResponse, error) {
	return m.getFunc(ctx, id)
}
func (m *mockVehicleService) GetByPlateNumber(ctx context.Context, plate string) (*domain.Vehicle, error) {
	return m.plateFunc(ctx, plate)
}
func (m *mockVehicleService) GetAll(ctx context.Context, filter dto.VehicleFilter) ([]dto.VehicleResponse, int64, error) {
	return m.getAllFunc(ctx, filter)
}
func (m *mockVehicleService) GetLatestKM(ctx context.Context, plate string) (float64, error) {
	return 0, nil
}
func (m *mockVehicleService) RecordOilChange(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockVehicleService) GetVehicleByPlate(ctx context.Context, plate string) (*contracts.VehicleSummaryDTO, error) {
	return nil, nil
}
func (m *mockVehicleService) GetVehicleByID(ctx context.Context, id uuid.UUID) (*contracts.VehicleSummaryDTO, error) {
	return nil, nil
}
func (m *mockVehicleService) UpdateOdometer(ctx context.Context, id uuid.UUID, newKM float64) error {
	return nil
}
func (m *mockVehicleService) RecordOilChangeKM(ctx context.Context, id uuid.UUID, km float64) error {
	return nil
}

func TestVehicleHandler_CheckKM(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockVehicleService{
		plateFunc: func(ctx context.Context, plate string) (*domain.Vehicle, error) {
			return &domain.Vehicle{
				PlateNumber:      "7572-ABC",
				CurrentKM:        1500,
				IsOdometerBroken: false,
				VehicleType:      "motorcycle",
				Status:           "AVAILABLE",
			}, nil
		},
	}

	h := controller.NewVehicleHandler(mockSvc, nil)
	r := gin.New()
	r.GET("/api/v1/vehicles/check-km", h.CheckKM)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/vehicles/check-km?plate=7572-ABC", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["current_km"].(float64) != 1500 {
		t.Errorf("expected current_km 1500, got %v", resp["current_km"])
	}
}

func TestVehicleHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	vehicleID := uuid.New()
	mockSvc := &mockVehicleService{
		createFunc: func(ctx context.Context, req dto.CreateVehicleRequest) (*domain.Vehicle, error) {
			return &domain.Vehicle{
				ID:          vehicleID,
				PlateNumber: req.PlateNumber,
				Brand:       req.Brand,
				CurrentKM:   req.CurrentKM,
				Status:      "AVAILABLE",
			}, nil
		},
	}

	h := controller.NewVehicleHandler(mockSvc, nil)
	r := gin.New()
	r.POST("/api/v1/vehicles", h.Create)

	body, _ := json.Marshal(dto.CreateVehicleRequest{
		PlateNumber: "1234-DEF",
		Brand:       "Yamaha",
		CurrentKM:   500,
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}
