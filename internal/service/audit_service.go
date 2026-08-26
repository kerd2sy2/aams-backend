package service

import (
	"context"
	"log"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/repository"

	"github.com/google/uuid"
)

type AuditService interface {
	LogAction(ctx context.Context, adminName, action, details, ipAddress string, branchID *uuid.UUID) error
	GetLogs(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error)
	DeleteLog(ctx context.Context, id uuid.UUID) error
	BulkDeleteLogs(ctx context.Context, ids []uuid.UUID) error
	ClearLogs(ctx context.Context, branchID *uuid.UUID) error
}

type auditService struct {
	auditRepo repository.AuditRepository
}

func NewAuditService(auditRepo repository.AuditRepository) AuditService {
	return &auditService{auditRepo: auditRepo}
}

func (s *auditService) LogAction(ctx context.Context, adminName, action, details, ipAddress string, branchID *uuid.UUID) error {
	auditLog := &domain.AuditLog{
		ID:        uuid.New(),
		AdminName: adminName,
		Action:    action,
		Details:   details,
		IPAddress: ipAddress,
		BranchID:  branchID,
	}
	err := s.auditRepo.CreateLog(ctx, auditLog)
	if err != nil {
		log.Printf("WARNING: Failed to write audit log [%s by %s]: %v", action, adminName, err)
	}
	return err
}

func (s *auditService) GetLogs(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	return s.auditRepo.FindAll(ctx, branchID, page, limit)
}

func (s *auditService) DeleteLog(ctx context.Context, id uuid.UUID) error {
	return s.auditRepo.DeleteByID(ctx, id)
}

func (s *auditService) BulkDeleteLogs(ctx context.Context, ids []uuid.UUID) error {
	return s.auditRepo.DeleteBulk(ctx, ids)
}

func (s *auditService) ClearLogs(ctx context.Context, branchID *uuid.UUID) error {
	return s.auditRepo.ClearAll(ctx, branchID)
}
