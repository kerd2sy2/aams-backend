package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/internal/modules/work/controller"
	"delivery-backend/internal/modules/work/domain"
	"delivery-backend/internal/modules/work/dto"
	"delivery-backend/internal/modules/work/service"
)

type supervisorMockAuditLogger struct {
	loggedActions int
}

func (m *supervisorMockAuditLogger) LogAction(ctx context.Context, userName, action, details, ip string, branchID *uuid.UUID) error {
	m.loggedActions++
	return nil
}

type supervisorMockWorkService struct {
	service.WorkService
	branchA       uuid.UUID
	branchB       uuid.UUID
	empBranchA    uuid.UUID
	empBranchB    uuid.UUID
	activeSessions map[uuid.UUID]*domain.WorkSession
}

func newSupervisorMockWorkService(branchA, branchB, empA, empB uuid.UUID) *supervisorMockWorkService {
	return &supervisorMockWorkService{
		branchA:        branchA,
		branchB:        branchB,
		empBranchA:     empA,
		empBranchB:     empB,
		activeSessions: make(map[uuid.UUID]*domain.WorkSession),
	}
}

func (m *supervisorMockWorkService) SupervisorStartShift(
	ctx context.Context,
	req dto.SupervisorStartShiftRequest,
	supervisorID uuid.UUID,
	supervisorName, supervisorRole string,
	supervisorBranchID *uuid.UUID,
) (*domain.WorkSession, error) {
	targetID := supervisorID
	if req.EmployeeID != "" {
		parsed, err := uuid.Parse(req.EmployeeID)
		if err != nil {
			return nil, err
		}
		targetID = parsed
		// Branch isolation check
		if supervisorRole != "SUPER_ADMIN" {
			if targetID == m.empBranchB && supervisorBranchID != nil && *supervisorBranchID == m.branchA {
				return nil, service.ErrCrossBranchForbidden
			}
		}
	}

	if _, exists := m.activeSessions[targetID]; exists {
		return nil, service.ErrActiveSessionExists
	}

	sess := &domain.WorkSession{
		ID:                   uuid.New(),
		EmployeeID:           &targetID,
		StartTime:            time.Now(),
		StartKM:              req.StartKM,
		Status:               "ACTIVE",
		IsEditedBySupervisor: true,
		EditedByName:         supervisorName,
		ReviewedBy:           &supervisorID,
	}
	m.activeSessions[targetID] = sess
	return sess, nil
}

func (m *supervisorMockWorkService) SupervisorEndShift(
	ctx context.Context,
	req dto.SupervisorEndShiftRequest,
	supervisorID uuid.UUID,
	supervisorName, supervisorRole string,
	supervisorBranchID *uuid.UUID,
) (*domain.WorkSession, error) {
	targetID := supervisorID
	if req.EmployeeID != "" {
		parsed, err := uuid.Parse(req.EmployeeID)
		if err != nil {
			return nil, err
		}
		targetID = parsed
		// Branch isolation check
		if supervisorRole != "SUPER_ADMIN" {
			if targetID == m.empBranchB && supervisorBranchID != nil && *supervisorBranchID == m.branchA {
				return nil, service.ErrCrossBranchForbidden
			}
		}
	}

	sess, exists := m.activeSessions[targetID]
	if !exists {
		return nil, service.ErrNoActiveSession
	}

	delete(m.activeSessions, targetID)
	now := time.Now()
	sess.EndTime = &now
	sess.EndKM = req.EndKM
	sess.OrdersCount = req.OrdersCount
	sess.FuelCost = req.FuelCost
	sess.Status = "COMPLETED"
	sess.IsReviewed = true
	sess.ReviewedBy = &supervisorID
	sess.IsEditedBySupervisor = true
	sess.EditedByName = supervisorName
	return sess, nil
}

func setupSupervisorShiftApp(
	mockSvc service.WorkService,
	auditLogger *supervisorMockAuditLogger,
	role string,
	branchID *uuid.UUID,
	adminID *uuid.UUID,
	isEmployee bool,
) *gin.Engine {
	h := controller.NewWorkHandler(mockSvc, auditLogger)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if role != "" {
			c.Set("admin_role", role)
		}
		if branchID != nil {
			c.Set("branch_id", branchID)
		}
		if adminID != nil {
			c.Set("admin_id", *adminID)
			c.Set("admin_name", "Test Supervisor")
		}
		c.Set("is_employee", isEmployee)
		c.Next()
	})

	group := r.Group("/api/v1/work/supervisor")
	{
		group.POST("/start-shift", h.SupervisorStartShift)
		group.POST("/end-shift", h.SupervisorEndShift)
	}
	return r
}

func TestSupervisorShift_SecurityAndIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchA := uuid.New()
	branchB := uuid.New()
	empA := uuid.New()
	empB := uuid.New()
	supervisorID := uuid.New()

	// 1. Supervisor starts shift for delegate in OWN branch -> 201 Created & audit logged
	t.Run("Supervisor starts shift for own branch delegate", func(t *testing.T) {
		mockSvc := newSupervisorMockWorkService(branchA, branchB, empA, empB)
		audit := &supervisorMockAuditLogger{}
		r := setupSupervisorShiftApp(mockSvc, audit, "SUPERVISOR", &branchA, &supervisorID, false)

		body, _ := json.Marshal(dto.SupervisorStartShiftRequest{
			EmployeeID:     empA.String(),
			StartKM:        500,
			OverrideReason: "Delegate device broken",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/work/supervisor/start-shift", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		if audit.loggedActions != 1 {
			t.Fatalf("expected audit log to be recorded, got %d", audit.loggedActions)
		}
	})

	// 2. Supervisor attempting to start shift for delegate in ANOTHER branch -> 404 Not Found (anti-IDOR)
	t.Run("Supervisor attempting start shift for other branch delegate is blocked with 404", func(t *testing.T) {
		mockSvc := newSupervisorMockWorkService(branchA, branchB, empA, empB)
		audit := &supervisorMockAuditLogger{}
		r := setupSupervisorShiftApp(mockSvc, audit, "SUPERVISOR", &branchA, &supervisorID, false)

		body, _ := json.Marshal(dto.SupervisorStartShiftRequest{
			EmployeeID:     empB.String(),
			StartKM:        500,
			OverrideReason: "Attempted cross branch start",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/work/supervisor/start-shift", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for cross branch delegate start, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 3. Supervisor attempting to start shift when active shift already exists -> 409 Conflict
	t.Run("Supervisor attempting to start shift when active shift exists is blocked with 409", func(t *testing.T) {
		mockSvc := newSupervisorMockWorkService(branchA, branchB, empA, empB)
		audit := &supervisorMockAuditLogger{}
		r := setupSupervisorShiftApp(mockSvc, audit, "SUPERVISOR", &branchA, &supervisorID, false)

		// First start succeeds
		mockSvc.activeSessions[empA] = &domain.WorkSession{ID: uuid.New(), Status: "ACTIVE"}

		body, _ := json.Marshal(dto.SupervisorStartShiftRequest{
			EmployeeID: empA.String(),
			StartKM:    600,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/work/supervisor/start-shift", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409 conflict for duplicate active shift, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 4. Supervisor ends shift for own branch delegate -> 200 OK & audit logged
	t.Run("Supervisor ends shift for own branch delegate", func(t *testing.T) {
		mockSvc := newSupervisorMockWorkService(branchA, branchB, empA, empB)
		mockSvc.activeSessions[empA] = &domain.WorkSession{ID: uuid.New(), StartKM: 500, Status: "ACTIVE"}
		audit := &supervisorMockAuditLogger{}
		r := setupSupervisorShiftApp(mockSvc, audit, "SUPERVISOR", &branchA, &supervisorID, false)

		body, _ := json.Marshal(dto.SupervisorEndShiftRequest{
			EmployeeID:     empA.String(),
			EndKM:          580,
			OrdersCount:    12,
			FuelCost:       25,
			OverrideReason: "Manual shift closure",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/work/supervisor/end-shift", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if audit.loggedActions != 1 {
			t.Fatalf("expected audit log to be recorded, got %d", audit.loggedActions)
		}
	})

	// 5. Supervisor attempting to end shift when no active shift exists -> 404 Not Found
	t.Run("Supervisor ending shift when none exists returns 404", func(t *testing.T) {
		mockSvc := newSupervisorMockWorkService(branchA, branchB, empA, empB)
		audit := &supervisorMockAuditLogger{}
		r := setupSupervisorShiftApp(mockSvc, audit, "SUPERVISOR", &branchA, &supervisorID, false)

		body, _ := json.Marshal(dto.SupervisorEndShiftRequest{
			EmployeeID: empA.String(),
			EndKM:      580,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/work/supervisor/end-shift", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for ending nonexistent shift, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 6. Supervisor attempting to end shift for delegate in ANOTHER branch -> 404 Not Found
	t.Run("Supervisor ending shift for delegate in another branch returns 404", func(t *testing.T) {
		mockSvc := newSupervisorMockWorkService(branchA, branchB, empA, empB)
		audit := &supervisorMockAuditLogger{}
		r := setupSupervisorShiftApp(mockSvc, audit, "SUPERVISOR", &branchA, &supervisorID, false)

		body, _ := json.Marshal(dto.SupervisorEndShiftRequest{
			EmployeeID: empB.String(),
			EndKM:      580,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/work/supervisor/end-shift", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for cross branch delegate end, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 7. Driver / Courier role attempting to call supervisor shift endpoint -> 403 Forbidden
	t.Run("Driver role attempting supervisor shift endpoint is rejected with 403", func(t *testing.T) {
		mockSvc := newSupervisorMockWorkService(branchA, branchB, empA, empB)
		audit := &supervisorMockAuditLogger{}
		r := setupSupervisorShiftApp(mockSvc, audit, "DRIVER", &branchA, &supervisorID, true)

		body, _ := json.Marshal(dto.SupervisorStartShiftRequest{
			EmployeeID: empA.String(),
			StartKM:    500,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/work/supervisor/start-shift", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for driver attempting supervisor shift, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 8. Supervisor without branch context fails closed -> 403 Forbidden
	t.Run("Supervisor without branch context fails closed with 403", func(t *testing.T) {
		mockSvc := newSupervisorMockWorkService(branchA, branchB, empA, empB)
		audit := &supervisorMockAuditLogger{}
		r := setupSupervisorShiftApp(mockSvc, audit, "SUPERVISOR", nil, &supervisorID, false)

		body, _ := json.Marshal(dto.SupervisorStartShiftRequest{
			EmployeeID: empA.String(),
			StartKM:    500,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/work/supervisor/start-shift", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 when branch context missing, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 9. Supervisor starts own shift (omitted employee_id) -> 201 Created
	t.Run("Supervisor starts own shift", func(t *testing.T) {
		mockSvc := newSupervisorMockWorkService(branchA, branchB, empA, empB)
		audit := &supervisorMockAuditLogger{}
		r := setupSupervisorShiftApp(mockSvc, audit, "SUPERVISOR", &branchA, &supervisorID, false)

		body, _ := json.Marshal(dto.SupervisorStartShiftRequest{
			StartKM: 100,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/work/supervisor/start-shift", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 for supervisor own shift, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 10. Supervisor ends own shift -> 200 OK
	t.Run("Supervisor ends own shift", func(t *testing.T) {
		mockSvc := newSupervisorMockWorkService(branchA, branchB, empA, empB)
		mockSvc.activeSessions[supervisorID] = &domain.WorkSession{ID: uuid.New(), StartKM: 100, Status: "ACTIVE"}
		audit := &supervisorMockAuditLogger{}
		r := setupSupervisorShiftApp(mockSvc, audit, "SUPERVISOR", &branchA, &supervisorID, false)

		body, _ := json.Marshal(dto.SupervisorEndShiftRequest{
			EndKM: 120,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/work/supervisor/end-shift", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for ending supervisor own shift, got %d: %s", w.Code, w.Body.String())
		}
	})
}
