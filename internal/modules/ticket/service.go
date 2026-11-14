package ticket

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"delivery-backend/internal/shared/safe"
)

type IEmployeeContract interface {
	GetEmployeeSummary(ctx context.Context, id uuid.UUID) (*EmployeeSummaryDTO, error)
}

type INotificationContract interface {
	SendAsyncNotification(ctx context.Context, notif PushNotificationDTO)
}

type EmployeeSummaryDTO struct {
	ID       uuid.UUID  `json:"id"`
	Name     string     `json:"name"`
	BranchID *uuid.UUID `json:"branch_id"`
}

type PushNotificationDTO struct {
	TargetUserID uuid.UUID         `json:"target_user_id"`
	Title        string            `json:"title"`
	Body         string            `json:"body"`
	Data         map[string]string `json:"data"`
}

type SupportTicketService interface {
	Create(ctx context.Context, req CreateSupportTicketRequest, adminBranchID *uuid.UUID) (*SupportTicket, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateSupportTicketRequest) (*SupportTicket, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*SupportTicket, error)
	GetAll(ctx context.Context, filter SupportTicketFilter, adminBranchID *uuid.UUID) ([]SupportTicket, int64, error)
}

type supportTicketService struct {
	repo          SupportTicketRepository
	empContract   IEmployeeContract
	notifContract INotificationContract
}

func NewSupportTicketService(
	repo SupportTicketRepository,
	empContract IEmployeeContract,
	notifContract INotificationContract,
) SupportTicketService {
	return &supportTicketService{
		repo:          repo,
		empContract:   empContract,
		notifContract: notifContract,
	}
}

func (s *supportTicketService) Create(ctx context.Context, req CreateSupportTicketRequest, adminBranchID *uuid.UUID) (*SupportTicket, error) {
	category := "OPERATIONAL"
	if req.Category != "" {
		category = req.Category
	}
	priority := "MEDIUM"
	if req.Priority != "" {
		priority = req.Priority
	}

	branchID := req.BranchID
	if branchID == nil && adminBranchID != nil {
		branchID = adminBranchID
	}

	if branchID == nil && req.EmployeeID != nil && s.empContract != nil {
		if emp, err := s.empContract.GetEmployeeSummary(ctx, *req.EmployeeID); err == nil && emp != nil && emp.BranchID != nil {
			branchID = emp.BranchID
		}
	}

	ticketNum := fmt.Sprintf("TCK-%d", time.Now().Unix()%1000000)

	ticket := &SupportTicket{
		ID:           uuid.New(),
		TicketNumber: ticketNum,
		EmployeeID:   req.EmployeeID,
		Subject:      req.Subject,
		Category:     category,
		Priority:     priority,
		Status:       "OPEN",
		Description:  req.Description,
		BranchID:     branchID,
	}

	if err := s.repo.Create(ctx, ticket); err != nil {
		return nil, err
	}

	if req.EmployeeID != nil && s.notifContract != nil {
		empID := *req.EmployeeID
		tNum := ticket.TicketNumber
		subj := ticket.Subject
		safe.Go(func() {
			s.notifContract.SendAsyncNotification(context.Background(), PushNotificationDTO{
				TargetUserID: empID,
				Title:        "تذكرة دعم جديدة",
				Body:         fmt.Sprintf("تم فتح التذكرة رقم %s: %s", tNum, subj),
				Data: map[string]string{
					"ticket_id":     ticket.ID.String(),
					"ticket_number": tNum,
				},
			})
		})
	}

	return s.repo.FindByID(ctx, ticket.ID)
}

func (s *supportTicketService) Update(ctx context.Context, id uuid.UUID, req UpdateSupportTicketRequest) (*SupportTicket, error) {
	ticket, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("التذكرة غير موجودة")
	}

	if req.Subject != nil {
		ticket.Subject = *req.Subject
	}
	if req.Category != nil {
		ticket.Category = *req.Category
	}
	if req.Priority != nil {
		ticket.Priority = *req.Priority
	}
	if req.Status != nil {
		ticket.Status = *req.Status
	}
	if req.Description != nil {
		ticket.Description = *req.Description
	}
	if req.Resolution != nil {
		ticket.Resolution = *req.Resolution
	}

	if err := s.repo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	if req.Status != nil && (*req.Status == "RESOLVED" || *req.Status == "CLOSED") && ticket.EmployeeID != nil && s.notifContract != nil {
		empID := *ticket.EmployeeID
		tNum := ticket.TicketNumber
		st := *req.Status
		safe.Go(func() {
			s.notifContract.SendAsyncNotification(context.Background(), PushNotificationDTO{
				TargetUserID: empID,
				Title:        "تحديث حالة التذكرة",
				Body:         fmt.Sprintf("تم تغيير حالة تذكرتك %s إلى %s", tNum, st),
				Data: map[string]string{
					"ticket_id": ticket.ID.String(),
					"status":    st,
				},
			})
		})
	}

	return s.repo.FindByID(ctx, id)
}

func (s *supportTicketService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *supportTicketService) GetByID(ctx context.Context, id uuid.UUID) (*SupportTicket, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *supportTicketService) GetAll(ctx context.Context, filter SupportTicketFilter, adminBranchID *uuid.UUID) ([]SupportTicket, int64, error) {
	if adminBranchID != nil {
		filter.BranchID = adminBranchID
	}
	return s.repo.FindAll(ctx, filter)
}
