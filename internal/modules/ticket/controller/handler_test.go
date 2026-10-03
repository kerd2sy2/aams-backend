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

	"delivery-backend/internal/modules/ticket/controller"
	"delivery-backend/internal/modules/ticket/domain"
	"delivery-backend/internal/modules/ticket/dto"
)

type mockTicketService struct {
	createFunc func(ctx context.Context, req dto.CreateSupportTicketRequest, branchID *uuid.UUID) (*domain.SupportTicket, error)
	updateFunc func(ctx context.Context, id uuid.UUID, req dto.UpdateSupportTicketRequest) (*domain.SupportTicket, error)
	deleteFunc func(ctx context.Context, id uuid.UUID) error
	getFunc    func(ctx context.Context, id uuid.UUID) (*domain.SupportTicket, error)
	getAllFunc func(ctx context.Context, filter dto.SupportTicketFilter, branchID *uuid.UUID) ([]domain.SupportTicket, int64, error)
}

func (m *mockTicketService) Create(ctx context.Context, req dto.CreateSupportTicketRequest, branchID *uuid.UUID) (*domain.SupportTicket, error) {
	return m.createFunc(ctx, req, branchID)
}

func (m *mockTicketService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateSupportTicketRequest) (*domain.SupportTicket, error) {
	return m.updateFunc(ctx, id, req)
}

func (m *mockTicketService) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteFunc(ctx, id)
}

func (m *mockTicketService) GetByID(ctx context.Context, id uuid.UUID) (*domain.SupportTicket, error) {
	return m.getFunc(ctx, id)
}

func (m *mockTicketService) GetAll(ctx context.Context, filter dto.SupportTicketFilter, branchID *uuid.UUID) ([]domain.SupportTicket, int64, error) {
	return m.getAllFunc(ctx, filter, branchID)
}

func TestSupportTicketHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ticketID := uuid.New()
	mockSvc := &mockTicketService{
		createFunc: func(ctx context.Context, req dto.CreateSupportTicketRequest, branchID *uuid.UUID) (*domain.SupportTicket, error) {
			return &domain.SupportTicket{
				ID:           ticketID,
				TicketNumber: "TCK-999999",
				Subject:      req.Subject,
				Category:     req.Category,
				Status:       "OPEN",
				Description:  req.Description,
			}, nil
		},
	}

	h := controller.NewSupportTicketHandler(mockSvc)
	r := gin.New()
	r.POST("/api/v1/tickets", h.Create)

	body, _ := json.Marshal(dto.CreateSupportTicketRequest{
		Subject:     "مشكلة في التطبيق",
		Category:    "APPLICATION",
		Description: "التطبيق يغلق تلقائيا",
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/tickets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp domain.SupportTicket
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.TicketNumber != "TCK-999999" {
		t.Errorf("expected ticket number TCK-999999, got %s", resp.TicketNumber)
	}
}
