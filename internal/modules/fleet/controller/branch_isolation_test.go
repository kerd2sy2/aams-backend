package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/fleet/contracts"
	"delivery-backend/internal/modules/fleet/controller"
	"delivery-backend/internal/modules/fleet/domain"
	"delivery-backend/internal/modules/fleet/dto"
	"delivery-backend/internal/modules/fleet/service"
)

type branchIsolationMockVehicleService struct {
	service.VehicleService
	branchA       uuid.UUID
	branchB       uuid.UUID
	vehA          *domain.Vehicle
	vehB          *domain.Vehicle
	oilChangeDone bool
	deletedID     uuid.UUID
	updatedData   *domain.Vehicle
}

func (m *branchIsolationMockVehicleService) GetByID(ctx context.Context, id uuid.UUID) (*dto.VehicleResponse, error) {
	if id == m.vehA.ID {
		return &dto.VehicleResponse{Vehicle: *m.vehA}, nil
	}
	if id == m.vehB.ID {
		return &dto.VehicleResponse{Vehicle: *m.vehB}, nil
	}
	return nil, errors.New("المركبة غير موجودة")
}

func (m *branchIsolationMockVehicleService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateVehicleRequest) (*domain.Vehicle, error) {
	m.updatedData = m.vehA
	return m.vehA, nil
}

func (m *branchIsolationMockVehicleService) Delete(ctx context.Context, id uuid.UUID) error {
	m.deletedID = id
	return nil
}

func (m *branchIsolationMockVehicleService) Create(ctx context.Context, req dto.CreateVehicleRequest) (*domain.Vehicle, error) {
	return &domain.Vehicle{
		ID:          uuid.New(),
		PlateNumber: req.PlateNumber,
		Brand:       req.Brand,
		BranchID:    req.BranchID,
	}, nil
}

func (m *branchIsolationMockVehicleService) RecordOilChange(ctx context.Context, id uuid.UUID) error {
	m.oilChangeDone = true
	return nil
}

func (m *branchIsolationMockVehicleService) GetAll(ctx context.Context, filter dto.VehicleFilter) ([]dto.VehicleResponse, int64, error) {
	var list []dto.VehicleResponse
	if filter.BranchID == nil || *filter.BranchID == m.branchA {
		list = append(list, dto.VehicleResponse{Vehicle: *m.vehA})
	}
	if filter.BranchID == nil || *filter.BranchID == m.branchB {
		list = append(list, dto.VehicleResponse{Vehicle: *m.vehB})
	}
	return list, int64(len(list)), nil
}

func (m *branchIsolationMockVehicleService) GetByPlateNumber(ctx context.Context, plate string) (*domain.Vehicle, error) {
	if plate == m.vehA.PlateNumber {
		return m.vehA, nil
	}
	if plate == m.vehB.PlateNumber {
		return m.vehB, nil
	}
	return nil, errors.New("المركبة غير موجودة")
}

func (m *branchIsolationMockVehicleService) GetLatestKM(ctx context.Context, plate string) (float64, error) {
	return 0, nil
}

func (m *branchIsolationMockVehicleService) GetVehicleByPlate(ctx context.Context, plate string) (*contracts.VehicleSummaryDTO, error) {
	return nil, nil
}

func (m *branchIsolationMockVehicleService) GetVehicleByID(ctx context.Context, id uuid.UUID) (*contracts.VehicleSummaryDTO, error) {
	return nil, nil
}

func (m *branchIsolationMockVehicleService) UpdateOdometer(ctx context.Context, id uuid.UUID, newKM float64) error {
	return nil
}

func (m *branchIsolationMockVehicleService) RecordOilChangeKM(ctx context.Context, id uuid.UUID, km float64) error {
	return nil
}

func setupFleetTestApp(mockSvc *branchIsolationMockVehicleService, role string, branchID *uuid.UUID) (*gin.Engine, *controller.VehicleHandler) {
	h := controller.NewVehicleHandler(mockSvc, nil)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if role != "" {
			c.Set("admin_role", role)
		}
		if branchID != nil {
			c.Set("branch_id", branchID)
		}
		c.Next()
	})

	vehicles := r.Group("/vehicles")
	{
		vehicles.GET("", h.GetAll)
		vehicles.POST("", h.Create)
		vehicles.GET("/check-km", h.CheckKM)
		vehicles.GET("/:id", h.GetByID)
		vehicles.PUT("/:id", h.Update)
		vehicles.DELETE("/:id", h.Delete)
		vehicles.POST("/:id/oil-change", h.RecordOilChange)
	}

	return r, h
}

func TestVehicleHandler_BranchIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchA := uuid.New()
	branchB := uuid.New()

	vehA := &domain.Vehicle{
		ID:          uuid.New(),
		PlateNumber: "1111-AAA",
		Brand:       "Yamaha",
		BranchID:    &branchA,
		CurrentKM:   1000,
	}
	vehB := &domain.Vehicle{
		ID:          uuid.New(),
		PlateNumber: "2222-BBB",
		Brand:       "Honda",
		BranchID:    &branchB,
		CurrentKM:   2000,
	}

	mockSvc := &branchIsolationMockVehicleService{
		branchA: branchA,
		branchB: branchB,
		vehA:    vehA,
		vehB:    vehB,
	}

	// 1. Supervisor reads vehicle from OWN branch -> 200 OK
	t.Run("Supervisor reads own branch vehicle", func(t *testing.T) {
		r, _ := setupFleetTestApp(mockSvc, "SUPERVISOR", &branchA)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/vehicles/"+vehA.ID.String(), nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 2. Cross-branch GET by ID is rejected -> 404 Not Found (anti-IDOR)
	t.Run("Cross-branch GET by ID is rejected with 404", func(t *testing.T) {
		r, _ := setupFleetTestApp(mockSvc, "SUPERVISOR", &branchA)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/vehicles/"+vehB.ID.String(), nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for other branch vehicle, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 3. Cross-branch UPDATE is rejected and does not change data -> 404 Not Found
	t.Run("Cross-branch UPDATE is rejected with 404 and does not mutate", func(t *testing.T) {
		mockSvc.updatedData = nil
		r, _ := setupFleetTestApp(mockSvc, "SUPERVISOR", &branchA)
		body, _ := json.Marshal(dto.UpdateVehicleRequest{Brand: stringPtr("Malicious Edit")})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/vehicles/"+vehB.ID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for cross-branch update, got %d: %s", w.Code, w.Body.String())
		}
		if mockSvc.updatedData != nil {
			t.Fatalf("update must NOT mutate service data")
		}
	})

	// 4. Cross-branch DELETE is rejected and does not delete data -> 404 Not Found
	t.Run("Cross-branch DELETE is rejected with 404 and does not delete", func(t *testing.T) {
		mockSvc.deletedID = uuid.Nil
		r, _ := setupFleetTestApp(mockSvc, "SUPERVISOR", &branchA)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/vehicles/"+vehB.ID.String(), nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for cross-branch delete, got %d: %s", w.Code, w.Body.String())
		}
		if mockSvc.deletedID != uuid.Nil {
			t.Fatalf("delete must NOT execute service deletion")
		}
	})

	// 5. Cross-branch oil-change operation is rejected and creates no record -> 404 Not Found
	t.Run("Cross-branch oil-change is rejected with 404 and creates no record", func(t *testing.T) {
		mockSvc.oilChangeDone = false
		r, _ := setupFleetTestApp(mockSvc, "SUPERVISOR", &branchA)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/vehicles/"+vehB.ID.String()+"/oil-change", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for cross-branch oil change, got %d: %s", w.Code, w.Body.String())
		}
		if mockSvc.oilChangeDone {
			t.Fatalf("oil change must NOT be recorded for vehicle in another branch")
		}
	})

	// 6. A forged branch_id query cannot expand access -> 403 Forbidden
	t.Run("Supervisor forged branch_id query parameter is rejected with 403", func(t *testing.T) {
		r, _ := setupFleetTestApp(mockSvc, "SUPERVISOR", &branchA)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/vehicles?branch_id="+branchB.String(), nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for conflicting query branch_id, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 7. A missing branch context fails closed -> 403 Forbidden
	t.Run("Supervisor missing branch context fails closed with 403", func(t *testing.T) {
		r, _ := setupFleetTestApp(mockSvc, "SUPERVISOR", nil)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/vehicles", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 when branch context is missing, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 8. Creating a vehicle enforces supervisor branch and ignores forged body branch
	t.Run("Creating vehicle enforces supervisor branch", func(t *testing.T) {
		r, _ := setupFleetTestApp(mockSvc, "SUPERVISOR", &branchA)
		forgedBranch := branchB
		body, _ := json.Marshal(dto.CreateVehicleRequest{
			PlateNumber: "9999-XYZ",
			Brand:       "Suzuki",
			BranchID:    &forgedBranch,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/vehicles", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}

		var resp domain.Vehicle
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.BranchID == nil || *resp.BranchID != branchA {
			t.Fatalf("expected vehicle branch to be forced to supervisor's branch %s, got %v", branchA, resp.BranchID)
		}
	})

	// 9. Updating vehicle cannot assign it to an unauthorized branch -> 403 Forbidden
	t.Run("Updating vehicle to another branch is forbidden", func(t *testing.T) {
		r, _ := setupFleetTestApp(mockSvc, "SUPERVISOR", &branchA)
		forgedBranch := branchB
		body, _ := json.Marshal(dto.UpdateVehicleRequest{
			BranchID: &forgedBranch,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/vehicles/"+vehA.ID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 when trying to reassign vehicle branch, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 10. SUPER_ADMIN behavior remains compatible with existing authorization policy (global access)
	t.Run("SUPER_ADMIN can read vehicles across any branch", func(t *testing.T) {
		r, _ := setupFleetTestApp(mockSvc, "SUPER_ADMIN", nil)

		// Read Branch A vehicle
		wA := httptest.NewRecorder()
		reqA, _ := http.NewRequest("GET", "/vehicles/"+vehA.ID.String(), nil)
		r.ServeHTTP(wA, reqA)
		if wA.Code != http.StatusOK {
			t.Fatalf("SUPER_ADMIN expected 200 for branch A, got %d", wA.Code)
		}

		// Read Branch B vehicle
		wB := httptest.NewRecorder()
		reqB, _ := http.NewRequest("GET", "/vehicles/"+vehB.ID.String(), nil)
		r.ServeHTTP(wB, reqB)
		if wB.Code != http.StatusOK {
			t.Fatalf("SUPER_ADMIN expected 200 for branch B, got %d", wB.Code)
		}
	})
}

func stringPtr(s string) *string {
	return &s
}
