package ticket

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SupportTicketRepository interface {
	Create(ctx context.Context, ticket *SupportTicket) error
	Update(ctx context.Context, ticket *SupportTicket) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*SupportTicket, error)
	FindAll(ctx context.Context, filter SupportTicketFilter) ([]SupportTicket, int64, error)
	AppendViolationHistoryStep(ctx context.Context, violationID uuid.UUID, stepType, title, details string) error
}

type gormSupportTicketRepository struct {
	db *gorm.DB
}

func NewSupportTicketRepository(db *gorm.DB) SupportTicketRepository {
	return &gormSupportTicketRepository{db: db}
}

func (r *gormSupportTicketRepository) Create(ctx context.Context, ticket *SupportTicket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

func (r *gormSupportTicketRepository) Update(ctx context.Context, ticket *SupportTicket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

func (r *gormSupportTicketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&SupportTicket{}, "id = ?", id).Error
}

func (r *gormSupportTicketRepository) FindByID(ctx context.Context, id uuid.UUID) (*SupportTicket, error) {
	var ticket SupportTicket
	err := r.db.WithContext(ctx).First(&ticket, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *gormSupportTicketRepository) FindAll(ctx context.Context, filter SupportTicketFilter) ([]SupportTicket, int64, error) {
	var list []SupportTicket
	var total int64

	q := r.db.WithContext(ctx).Model(&SupportTicket{})

	if filter.BranchID != nil {
		q = q.Where("branch_id = ?", filter.BranchID)
	}
	if filter.EmployeeID != nil {
		q = q.Where("employee_id = ?", filter.EmployeeID)
	}
	if filter.Category != "" {
		q = q.Where("category = ?", filter.Category)
	}
	if filter.Priority != "" {
		q = q.Where("priority = ?", filter.Priority)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("ticket_number LIKE ? OR subject LIKE ? OR description LIKE ?", s, s, s)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.PageSize
	if limit < 1 || limit > 500 {
		limit = 50
	}
	offset := (page - 1) * limit

	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

type linkedViolationStep struct {
	Type    string  `json:"type"`
	Title   string  `json:"title"`
	Details string  `json:"details,omitempty"`
	Amount  float64 `json:"amount,omitempty"`
	At      string  `json:"at"`
}

func (r *gormSupportTicketRepository) AppendViolationHistoryStep(ctx context.Context, violationID uuid.UUID, stepType, title, details string) error {
	var row struct {
		HistoryLog string
		Amount     float64
	}
	if err := r.db.WithContext(ctx).
		Table("traffic_violations").
		Select("history_log, amount").
		Where("id = ? AND deleted_at IS NULL", violationID).
		First(&row).Error; err != nil {
		return err
	}

	var steps []linkedViolationStep
	if strings.TrimSpace(row.HistoryLog) != "" {
		_ = json.Unmarshal([]byte(row.HistoryLog), &steps)
	}
	if len(steps) == 0 {
		steps = append(steps, linkedViolationStep{
			Type:    "RECORDED",
			Title:   "تم تسجيل الخصم",
			Amount:  row.Amount,
			At:      time.Now().UTC().Format(time.RFC3339),
		})
	}
	steps = append(steps, linkedViolationStep{
		Type:    stepType,
		Title:   title,
		Details: details,
		At:      time.Now().UTC().Format(time.RFC3339),
	})

	encoded, err := json.Marshal(steps)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Table("traffic_violations").
		Where("id = ?", violationID).
		Update("history_log", string(encoded)).Error
}
