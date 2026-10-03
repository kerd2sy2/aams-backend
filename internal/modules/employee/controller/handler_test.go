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

	"delivery-backend/internal/modules/employee/controller"
	"delivery-backend/internal/modules/employee/domain"
	"delivery-backend/internal/modules/employee/dto"
	"delivery-backend/internal/modules/employee/service"
)

type mockEmployeeService struct {
	service.EmployeeService
}

func (m *mockEmployeeService) Create(ctx context.Context, req dto.CreateEmployeeRequest) (*domain.Employee, error) {
	return &domain.Employee{
		ID:         uuid.New(),
		Name:       req.Name,
		NationalID: req.NationalID,
	}, nil
}

func (m *mockEmployeeService) FindAll(ctx context.Context, filter dto.EmployeeFilter) ([]domain.Employee, int64, error) {
	return []domain.Employee{
		{
			ID:         uuid.New(),
			Name:       "Test Emp",
			NationalID: "1122334455",
		},
	}, 1, nil
}

func TestEmployeeHandler_CreateAndGetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := controller.NewEmployeeHandler(&mockEmployeeService{}, nil)

	r := gin.New()
	r.POST("/employees", h.Create)
	r.GET("/employees", h.GetAll)

	// 1. Create
	createReq, _ := json.Marshal(dto.CreateEmployeeRequest{
		Name:       "Ali",
		NationalID: "9988776655",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer(createReq))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	// 2. GetAll
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/employees?page=1&limit=10", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
