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

	"delivery-backend/internal/modules/employee/controller"
	"delivery-backend/internal/modules/employee/domain"
	"delivery-backend/internal/modules/employee/dto"
	"delivery-backend/internal/modules/employee/service"
)

type branchIsolationMockEmployeeService struct {
	service.EmployeeService
	branchA uuid.UUID
	branchB uuid.UUID
	empA    *domain.Employee
	empB    *domain.Employee
}

func (m *branchIsolationMockEmployeeService) FindByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	if id == m.empA.ID {
		return m.empA, nil
	}
	if id == m.empB.ID {
		return m.empB, nil
	}
	return nil, errors.New("not found")
}

func (m *branchIsolationMockEmployeeService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateEmployeeRequest) (*domain.Employee, error) {
	return m.empA, nil
}

func (m *branchIsolationMockEmployeeService) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *branchIsolationMockEmployeeService) BatchSetOilChange(ctx context.Context, req dto.BatchOilSetupRequest) (int, error) {
	return len(req.Entries), nil
}

func (m *branchIsolationMockEmployeeService) Create(ctx context.Context, req dto.CreateEmployeeRequest) (*domain.Employee, error) {
	return &domain.Employee{
		ID:         uuid.New(),
		Name:       req.Name,
		NationalID: req.NationalID,
		BranchID:   req.BranchID,
	}, nil
}

func TestEmployeeHandler_BranchIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchA := uuid.New()
	branchB := uuid.New()

	empA := &domain.Employee{
		ID:         uuid.New(),
		Name:       "Branch A Employee",
		NationalID: "1111111111",
		BranchID:   &branchA,
	}
	empB := &domain.Employee{
		ID:         uuid.New(),
		Name:       "Branch B Employee",
		NationalID: "2222222222",
		BranchID:   &branchB,
	}

	mockSvc := &branchIsolationMockEmployeeService{
		branchA: branchA,
		branchB: branchB,
		empA:    empA,
		empB:    empB,
	}

	h := controller.NewEmployeeHandler(mockSvc, nil)

	r := gin.New()
	// Middleware simulating a supervisor logged in to Branch A
	r.Use(func(c *gin.Context) {
		c.Set("admin_role", "SUPERVISOR")
		c.Set("branch_id", &branchA)
		c.Next()
	})

	r.GET("/employees/:id", h.GetByID)
	r.PUT("/employees/:id", h.Update)
	r.DELETE("/employees/:id", h.Delete)
	r.POST("/employees", h.Create)
	r.POST("/employees/batch-oil-setup", h.BatchSetOilChange)

	// 1. Supervisor reads employee in OWN branch -> Allowed (200 OK)
	t.Run("Supervisor reads own branch employee", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/employees/"+empA.ID.String(), nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 2. Supervisor attempts to read employee in ANOTHER branch by direct UUID -> Blocked (404 Not Found)
	t.Run("Supervisor reading other branch employee is blocked", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/employees/"+empB.ID.String(), nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("IDOR security check failed: expected 404 for cross-branch read, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 3. Supervisor attempts to update employee in ANOTHER branch -> Blocked (404 Not Found)
	t.Run("Supervisor updating other branch employee is blocked", func(t *testing.T) {
		body, _ := json.Marshal(dto.UpdateEmployeeRequest{Name: "Hacked Name"})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/employees/"+empB.ID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("IDOR security check failed: expected 404 for cross-branch update, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 4. Supervisor attempts to delete employee in ANOTHER branch -> Blocked (404 Not Found)
	t.Run("Supervisor deleting other branch employee is blocked", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/employees/"+empB.ID.String(), nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("IDOR security check failed: expected 404 for cross-branch delete, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 5. Supervisor creating an employee has their own branch strictly enforced
	t.Run("Supervisor creating employee enforces supervisor branch", func(t *testing.T) {
		forgedBranch := branchB // trying to create an employee for branch B
		body, _ := json.Marshal(dto.CreateEmployeeRequest{
			Name:       "New Guy",
			NationalID: "3333333333",
			BranchID:   &forgedBranch,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Data struct {
				BranchID *uuid.UUID `json:"branch_id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Data.BranchID == nil || *resp.Data.BranchID != branchA {
			t.Fatalf("expected created employee to be forced to supervisor's branch %s, got %v", branchA, resp.Data.BranchID)
		}
	})

	// 6. Supervisor sets batch oil setup for employee in OWN branch -> Allowed (200 OK)
	t.Run("Supervisor sets batch oil setup for own branch employee", func(t *testing.T) {
		body, _ := json.Marshal(dto.BatchOilSetupRequest{
			Entries: []dto.OilSetupEntry{
				{EmployeeID: empA.ID.String(), LastOilChangeDistance: 1200},
			},
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/employees/batch-oil-setup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 7. Supervisor attempting batch oil setup with an employee from ANOTHER branch is rejected (404 Not Found)
	t.Run("Supervisor batch oil setup with other branch employee is blocked", func(t *testing.T) {
		body, _ := json.Marshal(dto.BatchOilSetupRequest{
			Entries: []dto.OilSetupEntry{
				{EmployeeID: empA.ID.String(), LastOilChangeDistance: 1200},
				{EmployeeID: empB.ID.String(), LastOilChangeDistance: 2500},
			},
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/employees/batch-oil-setup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for cross-branch batch member, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 8. Batch oil setup with duplicate employee IDs is rejected (400 Bad Request)
	t.Run("Batch oil setup with duplicate IDs is blocked", func(t *testing.T) {
		body, _ := json.Marshal(dto.BatchOilSetupRequest{
			Entries: []dto.OilSetupEntry{
				{EmployeeID: empA.ID.String(), LastOilChangeDistance: 1200},
				{EmployeeID: empA.ID.String(), LastOilChangeDistance: 1500},
			},
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/employees/batch-oil-setup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for duplicate ID in batch, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 9. Driver/Employee role attempting batch oil setup is rejected (403 Forbidden)
	t.Run("Driver attempting batch oil setup is blocked", func(t *testing.T) {
		driverEngine := gin.New()
		driverEngine.Use(func(c *gin.Context) {
			c.Set("admin_role", "DRIVER")
			c.Set("is_employee", true)
			c.Set("branch_id", &branchA)
			c.Next()
		})
		driverEngine.POST("/employees/batch-oil-setup", h.BatchSetOilChange)

		body, _ := json.Marshal(dto.BatchOilSetupRequest{
			Entries: []dto.OilSetupEntry{
				{EmployeeID: empA.ID.String(), LastOilChangeDistance: 1200},
			},
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/employees/batch-oil-setup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		driverEngine.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for driver role on batch oil setup, got %d: %s", w.Code, w.Body.String())
		}
	})
}
