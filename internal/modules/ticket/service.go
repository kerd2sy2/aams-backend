package ticket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"delivery-backend/internal/shared/safe"
)

type TicketHistoryStep struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Details string `json:"details,omitempty"`
	By      string `json:"by,omitempty"`
	At      string `json:"at"`
}

func appendTicketHistory(rawLog string, step TicketHistoryStep) string {
	var steps []TicketHistoryStep
	if strings.TrimSpace(rawLog) != "" {
		_ = json.Unmarshal([]byte(rawLog), &steps)
	}
	steps = append(steps, step)
	b, err := json.Marshal(steps)
	if err != nil {
		return rawLog
	}
	return string(b)
}

func ticketStatusArabic(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "IN_PROGRESS":
		return "قيد المعالجة"
	case "RESOLVED":
		return "تم الحل"
	case "CLOSED":
		return "مغلقة"
	default:
		return "مفتوحة"
	}
}

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

var ErrComplaintCooldown = errors.New("يمكنك تقديم شكوى واحدة فقط كل 24 ساعة")

type SupportTicketService interface {
	Create(ctx context.Context, req CreateSupportTicketRequest, adminBranchID *uuid.UUID, enforceCooldown bool) (*SupportTicket, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateSupportTicketRequest, adminName string) (*SupportTicket, error)
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

func (s *supportTicketService) Create(ctx context.Context, req CreateSupportTicketRequest, adminBranchID *uuid.UUID, enforceCooldown bool) (*SupportTicket, error) {
	if enforceCooldown && req.ViolationID == nil && req.EmployeeID != nil {
		if latest, err := s.repo.FindLatestGeneralComplaintByEmployee(ctx, *req.EmployeeID); err == nil && latest != nil {
			elapsed := time.Since(latest.CreatedAt)
			if elapsed < 24*time.Hour {
				remaining := 24*time.Hour - elapsed
				hours := int(remaining.Hours())
				mins := int(remaining.Minutes()) % 60
				if hours > 0 {
					return nil, fmt.Errorf("%w — المتبقي %d ساعة و %d دقيقة", ErrComplaintCooldown, hours, mins)
				}
				if mins < 1 {
					mins = 1
				}
				return nil, fmt.Errorf("%w — المتبقي %d دقيقة", ErrComplaintCooldown, mins)
			}
		}
	}

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

	initialHistory := appendTicketHistory("", TicketHistoryStep{
		Type:    "CREATED",
		Title:   "تم تقديم الشكوى",
		Details: req.Subject,
		At:      time.Now().UTC().Format(time.RFC3339),
	})

	ticket := &SupportTicket{
		ID:           uuid.New(),
		TicketNumber: ticketNum,
		EmployeeID:   req.EmployeeID,
		ViolationID:  req.ViolationID,
		Subject:      req.Subject,
		Category:     category,
		Priority:     priority,
		Status:       "OPEN",
		Description:  req.Description,
		HistoryLog:   initialHistory,
		BranchID:     branchID,
	}

	if err := s.repo.Create(ctx, ticket); err != nil {
		return nil, err
	}

	if req.ViolationID != nil {
		_ = s.repo.AppendViolationHistoryStep(
			ctx,
			*req.ViolationID,
			"DISPUTED",
			"تم تقديم اعتراض",
			fmt.Sprintf("تذكرة %s", ticketNum),
		)
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

func (s *supportTicketService) Update(ctx context.Context, id uuid.UUID, req UpdateSupportTicketRequest, adminName string) (*SupportTicket, error) {
	ticket, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("التذكرة غير موجودة")
	}

	oldResolution := strings.TrimSpace(ticket.Resolution)
	oldStatus := strings.TrimSpace(ticket.Status)
	actorName := strings.TrimSpace(adminName)
	if actorName == "" {
		actorName = "الإدارة"
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

	now := time.Now().UTC()
	newResolution := strings.TrimSpace(ticket.Resolution)
	if newResolution != "" && newResolution != oldResolution {
		ticket.RepliedByName = actorName
		ticket.RepliedAt = &now
		ticket.HistoryLog = appendTicketHistory(ticket.HistoryLog, TicketHistoryStep{
			Type:    "REPLY",
			Title:   "رد الإدارة",
			Details: newResolution,
			By:      actorName,
			At:      now.Format(time.RFC3339),
		})
	}
	if req.Status != nil && strings.TrimSpace(*req.Status) != "" && strings.TrimSpace(*req.Status) != oldStatus {
		ticket.HistoryLog = appendTicketHistory(ticket.HistoryLog, TicketHistoryStep{
			Type:    "STATUS",
			Title:   fmt.Sprintf("تم تغيير الحالة إلى %s", ticketStatusArabic(*req.Status)),
			By:      actorName,
			At:      now.Format(time.RFC3339),
		})
	}

	if err := s.repo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	if newResolution != "" && newResolution != oldResolution {
		if ticket.ViolationID != nil {
			_ = s.repo.AppendViolationHistoryStep(
				ctx,
				*ticket.ViolationID,
				"TICKET_REPLY",
				"رد الإدارة على الاعتراض",
				newResolution,
			)
		}
		if ticket.EmployeeID != nil && s.notifContract != nil {
			empID := *ticket.EmployeeID
			tNum := ticket.TicketNumber
			reply := newResolution
			notifTitle := "رد الإدارة على شكواك"
			notifType := "COMPLAINT"
			if ticket.ViolationID != nil {
				notifTitle = "رد الإدارة على اعتراضك"
				notifType = "VIOLATION"
			}
			safe.Go(func() {
				s.notifContract.SendAsyncNotification(context.Background(), PushNotificationDTO{
					TargetUserID: empID,
					Title:        notifTitle,
					Body:         fmt.Sprintf("تذكرة %s (%s): %s", tNum, actorName, reply),
					Data: map[string]string{
						"ticket_id": ticket.ID.String(),
						"type":      notifType,
					},
				})
			})
		}
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
