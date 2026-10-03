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

	"delivery-backend/internal/modules/attendance/controller"
	"delivery-backend/internal/modules/attendance/domain"
	"delivery-backend/internal/modules/attendance/dto"
	"delivery-backend/internal/modules/attendance/service"
)

type mockAttendanceService struct {
	service.AttendanceService
}

func (m *mockAttendanceService) GetAttendance(ctx context.Context, date string, branchID *uuid.UUID) ([]domain.AttendanceInfo, error) {
	return []domain.AttendanceInfo{
		{
			EmployeeID:   uuid.New(),
			EmployeeName: "Test Driver",
			NationalID:   "1020304050",
			BranchName:   "Riyadh",
			VehicleType:  "car",
			Status:       "present",
		},
	}, nil
}

func (m *mockAttendanceService) ToggleAttendance(ctx context.Context, adminID uuid.UUID, employeeID uuid.UUID, date string, status string, note string) (*domain.AttendanceInfo, error) {
	return &domain.AttendanceInfo{
		EmployeeID:   employeeID,
		EmployeeName: "Test Driver",
		Status:       status,
		Note:         note,
	}, nil
}

func TestAttendanceHandler_GetAttendance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := controller.NewAttendanceHandler(&mockAttendanceService{})

	r := gin.New()
	r.GET("/attendance", h.GetAttendance)

	req, _ := http.NewRequest("GET", "/attendance?date=2026-10-03", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestAttendanceHandler_ToggleAttendance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := controller.NewAttendanceHandler(&mockAttendanceService{})

	r := gin.New()
	r.POST("/attendance/:employee_id", h.ToggleAttendance)

	empID := uuid.New()
	reqBody, _ := json.Marshal(dto.ToggleAttendanceRequest{
		Date:   "2026-10-03",
		Status: "present",
		Note:   "Good",
	})
	req, _ := http.NewRequest("POST", "/attendance/"+empID.String(), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}
