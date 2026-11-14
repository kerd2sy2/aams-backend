package system

import (
	"context"

	"github.com/google/uuid"
)

type ArchiveService interface {
	GetArchivedItems(ctx context.Context, filter ArchiveFilter) (ArchiveResponseDTO, error)
	Restore(ctx context.Context, itemType string, id uuid.UUID) error
	PermanentDelete(ctx context.Context, itemType string, id uuid.UUID) error
	BulkRestore(ctx context.Context, itemType string, ids []uuid.UUID) error
	BulkPermanentDelete(ctx context.Context, itemType string, ids []uuid.UUID) error
}

type archiveService struct {
	repo SystemRepository
}

func NewArchiveService(repo SystemRepository) ArchiveService {
	return &archiveService{repo: repo}
}

func (s *archiveService) GetArchivedItems(ctx context.Context, filter ArchiveFilter) (ArchiveResponseDTO, error) {
	items, total, stats, err := s.repo.GetArchivedItems(ctx, filter)
	if err != nil {
		return ArchiveResponseDTO{}, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	return ArchiveResponseDTO{
		Data:       items,
		Stats:      stats,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *archiveService) Restore(ctx context.Context, itemType string, id uuid.UUID) error {
	return s.repo.RestoreItem(ctx, itemType, id)
}

func (s *archiveService) PermanentDelete(ctx context.Context, itemType string, id uuid.UUID) error {
	return s.repo.PermanentDeleteItem(ctx, itemType, id)
}

func (s *archiveService) BulkRestore(ctx context.Context, itemType string, ids []uuid.UUID) error {
	return s.repo.BulkRestoreItems(ctx, itemType, ids)
}

func (s *archiveService) BulkPermanentDelete(ctx context.Context, itemType string, ids []uuid.UUID) error {
	return s.repo.BulkPermanentDeleteItems(ctx, itemType, ids)
}
