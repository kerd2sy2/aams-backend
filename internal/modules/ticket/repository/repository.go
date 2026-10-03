package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/internal/modules/ticket/domain"
	"delivery-backend/internal/modules/ticket/dto"
)

type SupportTicketRepository interface {
	Create(ctx context.Context, ticket *domain.SupportTicket) error
	Update(ctx context.Context, ticket *domain.SupportTicket) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.SupportTicket, error)
	FindAll(ctx context.Context, filter dto.SupportTicketFilter) ([]domain.SupportTicket, int64, error)
}

type gormSupportTicketRepository struct {
	db *gorm.DB
}

func NewSupportTicketRepository(db *gorm.DB) SupportTicketRepository {
	return &gormSupportTicketRepository{db: db}
}

func (r *gormSupportTicketRepository) Create(ctx context.Context, ticket *domain.SupportTicket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

func (r *gormSupportTicketRepository) Update(ctx context.Context, ticket *domain.SupportTicket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

func (r *gormSupportTicketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.SupportTicket{}, "id = ?", id).Error
}

func (r *gormSupportTicketRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.SupportTicket, error) {
	var ticket domain.SupportTicket
	err := r.db.WithContext(ctx).First(&ticket, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *gormSupportTicketRepository) FindAll(ctx context.Context, filter dto.SupportTicketFilter) ([]domain.SupportTicket, int64, error) {
	var list []domain.SupportTicket
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.SupportTicket{})

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
