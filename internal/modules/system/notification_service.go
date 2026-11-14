package system

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type NotificationService interface {
	INotificationSender
	GetMyNotifications(ctx context.Context, adminID uuid.UUID, branchID *uuid.UUID, status string) ([]Notification, error)
	MarkAllAsRead(ctx context.Context, adminID uuid.UUID) error
	MarkAsRead(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error
	SendBroadcast(ctx context.Context, req CreateBroadcastRequest, createdBy string) (*BroadcastNotification, error)
	GetBroadcasts(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]BroadcastNotification, int64, error)
	DeleteBroadcast(ctx context.Context, id uuid.UUID) error
	GetBroadcastVotes(ctx context.Context, broadcastID uuid.UUID) ([]BroadcastVoteItemDTO, error)
	GetEmployeeBroadcasts(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time, limit int) ([]BroadcastItemDTO, error)
	GetEmployeeUnreadBroadcasts(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) ([]BroadcastItemDTO, error)
	MarkEmployeeBroadcastRead(ctx context.Context, empID uuid.UUID, broadcastID uuid.UUID) error
	MarkAllEmployeeBroadcastsRead(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) error
	SubmitVote(ctx context.Context, empID uuid.UUID, broadcastID uuid.UUID, req SubmitPollVoteRequest) error
	SaveEmployeePushToken(ctx context.Context, empID uuid.UUID, req PushTokenRequest) error
}

type notificationService struct {
	repo SystemRepository
}

func NewNotificationService(repo SystemRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) SendNotification(ctx context.Context, branchID, adminID, employeeID *uuid.UUID, title, body, notifType string) error {
	notif := &Notification{
		BranchID:   branchID,
		AdminID:    adminID,
		EmployeeID: employeeID,
		Title:      title,
		Body:       body,
		Type:       notifType,
		Status:     "unread",
		CreatedAt:  time.Now(),
	}
	return s.repo.CreateNotification(ctx, notif)
}

func (s *notificationService) GetMyNotifications(ctx context.Context, adminID uuid.UUID, branchID *uuid.UUID, status string) ([]Notification, error) {
	return s.repo.FindNotificationsByAdmin(ctx, adminID, branchID, status)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, adminID uuid.UUID) error {
	return s.repo.MarkAllNotificationsAsRead(ctx, adminID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error {
	return s.repo.MarkNotificationAsRead(ctx, id, adminID)
}

func (s *notificationService) SendBroadcast(ctx context.Context, req CreateBroadcastRequest, createdBy string) (*BroadcastNotification, error) {
	broadcast := &BroadcastNotification{
		Title:          req.Title,
		Body:           req.Body,
		TitleAr:        req.TitleAr,
		TitleEn:        req.TitleEn,
		TitleBn:        req.TitleBn,
		BodyAr:         req.BodyAr,
		BodyEn:         req.BodyEn,
		BodyBn:         req.BodyBn,
		ImageURL:       req.ImageURL,
		Target:         req.Target,
		BranchID:       req.BranchID,
		CreatedBy:      createdBy,
		HasPoll:        req.HasPoll,
		PollQuestion:   req.PollQuestion,
		PollQuestionAr: req.PollQuestionAr,
		PollQuestionEn: req.PollQuestionEn,
		PollQuestionBn: req.PollQuestionBn,
		CreatedAt:      time.Now(),
	}
	if broadcast.Target == "" {
		broadcast.Target = "ALL"
	}
	err := s.repo.CreateBroadcast(ctx, broadcast)
	return broadcast, err
}

func (s *notificationService) GetBroadcasts(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]BroadcastNotification, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	return s.repo.FindBroadcasts(ctx, branchID, limit, offset)
}

func (s *notificationService) DeleteBroadcast(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteBroadcast(ctx, id)
}

func (s *notificationService) GetBroadcastVotes(ctx context.Context, broadcastID uuid.UUID) ([]BroadcastVoteItemDTO, error) {
	return s.repo.FindVotesForBroadcast(ctx, broadcastID)
}

func (s *notificationService) GetEmployeeBroadcasts(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time, limit int) ([]BroadcastItemDTO, error) {
	return s.repo.FindBroadcastsForEmployee(ctx, empID, branchID, registeredAt, limit)
}

func (s *notificationService) GetEmployeeUnreadBroadcasts(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) ([]BroadcastItemDTO, error) {
	return s.repo.FindUnreadBroadcastsForEmployee(ctx, empID, branchID, registeredAt)
}

func (s *notificationService) MarkEmployeeBroadcastRead(ctx context.Context, empID uuid.UUID, broadcastID uuid.UUID) error {
	return s.repo.MarkEmployeeBroadcastRead(ctx, empID, broadcastID)
}

func (s *notificationService) MarkAllEmployeeBroadcastsRead(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) error {
	return s.repo.MarkAllEmployeeBroadcastsRead(ctx, empID, branchID, registeredAt)
}

func (s *notificationService) SubmitVote(ctx context.Context, empID uuid.UUID, broadcastID uuid.UUID, req SubmitPollVoteRequest) error {
	vote := &BroadcastVote{
		BroadcastID: broadcastID,
		EmployeeID:  empID,
		Response:    req.Response,
		Reason:      req.Reason,
		CreatedAt:   time.Now(),
	}
	return s.repo.SubmitVote(ctx, vote)
}

func (s *notificationService) SaveEmployeePushToken(ctx context.Context, empID uuid.UUID, req PushTokenRequest) error {
	return s.repo.UpdateEmployeePushToken(ctx, empID, req.PushToken, req.Language)
}
