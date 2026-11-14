package system

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AuditService interface {
	IAuditLogger
	GetLogs(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]AuditLog, int64, error)
	ClearLogs(ctx context.Context, branchID *uuid.UUID) error
	BulkDeleteLogs(ctx context.Context, ids []uuid.UUID) error
	DeleteLog(ctx context.Context, id uuid.UUID) error
	LogAction(ctx context.Context, adminName, action, details, ipAddress string, branchID *uuid.UUID) error
}

type auditService struct {
	repo SystemRepository
}

func NewAuditService(repo SystemRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) Log(ctx context.Context, adminName, action, details, ipAddress string, branchID *uuid.UUID) error {
	return s.LogAction(ctx, adminName, action, details, ipAddress, branchID)
}

func (s *auditService) LogAction(ctx context.Context, adminName, action, details, ipAddress string, branchID *uuid.UUID) error {
	entry := &AuditLog{
		AdminName: adminName,
		Action:    action,
		Details:   details,
		IPAddress: ipAddress,
		BranchID:  branchID,
		CreatedAt: time.Now(),
	}
	return s.repo.CreateAuditLog(ctx, entry)
}

func (s *auditService) GetLogs(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]AuditLog, int64, error) {
	return s.repo.FindAuditLogs(ctx, branchID, page, limit)
}

func (s *auditService) ClearLogs(ctx context.Context, branchID *uuid.UUID) error {
	return s.repo.ClearAuditLogs(ctx, branchID)
}

func (s *auditService) BulkDeleteLogs(ctx context.Context, ids []uuid.UUID) error {
	return s.repo.BulkDeleteAuditLogs(ctx, ids)
}

func (s *auditService) DeleteLog(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteAuditLog(ctx, id)
}
