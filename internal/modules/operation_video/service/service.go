package service

import (
	"context"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"delivery-backend/internal/domain"
	"delivery-backend/internal/modules/operation_video/dto"
	"delivery-backend/internal/modules/operation_video/repository"
)

type OperationVideoService interface {
	SaveVideo(ctx context.Context, video *domain.OperationVideo) (*dto.OperationVideoItem, error)
	ListVideos(ctx context.Context, filter dto.OperationVideoFilter) ([]dto.OperationVideoItem, int64, error)
	GetVideo(ctx context.Context, id uuid.UUID) (*dto.OperationVideoItem, error)
	UpdateVideo(ctx context.Context, id uuid.UUID, orderNumber, notes string) error
	DeleteVideo(ctx context.Context, id uuid.UUID, uploadBaseDir string) error
	PurgeExpiredVideos(ctx context.Context, uploadBaseDir string) (int, error)
}


type operationVideoService struct {
	repo repository.OperationVideoRepository
}

func NewOperationVideoService(repo repository.OperationVideoRepository) OperationVideoService {
	return &operationVideoService{repo: repo}
}

func (s *operationVideoService) SaveVideo(ctx context.Context, video *domain.OperationVideo) (*dto.OperationVideoItem, error) {
	if video.RecordedAt.IsZero() {
		video.RecordedAt = time.Now()
	}
	if video.ExpiresAt.IsZero() {
		video.ExpiresAt = video.RecordedAt.AddDate(0, 0, 60)
	}

	if err := s.repo.Create(ctx, video); err != nil {
		return nil, err
	}

	return toDTO(video), nil
}

func (s *operationVideoService) ListVideos(ctx context.Context, filter dto.OperationVideoFilter) ([]dto.OperationVideoItem, int64, error) {
	videos, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	items := make([]dto.OperationVideoItem, len(videos))
	for i, v := range videos {
		items[i] = *toDTO(&v)
	}

	return items, total, nil
}

func (s *operationVideoService) GetVideo(ctx context.Context, id uuid.UUID) (*dto.OperationVideoItem, error) {
	video, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDTO(video), nil
}

func (s *operationVideoService) UpdateVideo(ctx context.Context, id uuid.UUID, orderNumber, notes string) error {
	updates := map[string]interface{}{}
	if strings.TrimSpace(orderNumber) != "" {
		updates["order_number"] = strings.TrimSpace(orderNumber)
	}
	if strings.TrimSpace(notes) != "" {
		updates["notes"] = strings.TrimSpace(notes)
	}
	if len(updates) == 0 {
		return nil
	}
	return s.repo.Update(ctx, id, updates)
}

func (s *operationVideoService) DeleteVideo(ctx context.Context, id uuid.UUID, uploadBaseDir string) error {
	video, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Purge physical file from disk if present
	if video.VideoURL != "" && uploadBaseDir != "" {
		relPath := strings.TrimPrefix(video.VideoURL, "/uploads/")
		relPath = strings.TrimPrefix(relPath, "uploads/")
		fullPath := filepath.Join(uploadBaseDir, relPath)
		if _, err := os.Stat(fullPath); err == nil {
			_ = os.Remove(fullPath)
		}
	}

	return s.repo.Delete(ctx, id)
}

// PurgeExpiredVideos finds all videos where expires_at <= NOW() and removes the physical files
func (s *operationVideoService) PurgeExpiredVideos(ctx context.Context, uploadBaseDir string) (int, error) {
	now := time.Now()
	expiredList, err := s.repo.FindExpiredVideos(ctx, now, 300)
	if err != nil {
		return 0, err
	}

	if len(expiredList) == 0 {
		return 0, nil
	}

	purgedCount := 0
	for _, v := range expiredList {
		// Remove physical video from disk
		if v.VideoURL != "" && uploadBaseDir != "" {
			relPath := strings.TrimPrefix(v.VideoURL, "/uploads/")
			relPath = strings.TrimPrefix(relPath, "uploads/")
			fullPath := filepath.Join(uploadBaseDir, relPath)
			if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
				log.Printf("[Auto-Purge] Warning: could not delete video file %s: %v", fullPath, err)
			}
		}

		if err := s.repo.MarkDeleted(ctx, v.ID); err == nil {
			purgedCount++
		}
	}

	log.Printf("[Auto-Purge] Purged %d expired operation videos older than 60 days", purgedCount)
	return purgedCount, nil
}

func toDTO(v *domain.OperationVideo) *dto.OperationVideoItem {
	var bID, eID, sID *string
	if v.BranchID != nil {
		s := v.BranchID.String()
		bID = &s
	}
	if v.EmployeeID != nil {
		s := v.EmployeeID.String()
		eID = &s
	}
	if v.SessionID != nil {
		s := v.SessionID.String()
		sID = &s
	}

	daysLeft := int(math.Max(0, math.Ceil(time.Until(v.ExpiresAt).Hours()/24)))

	return &dto.OperationVideoItem{
		ID:           v.ID.String(),
		BranchID:     bID,
		EmployeeID:   eID,
		EmployeeName: v.EmployeeName,
		OrderNumber:  v.OrderNumber,
		SessionID:    sID,
		VideoURL:     v.VideoURL,
		ThumbnailURL: v.ThumbnailURL,
		DurationSec:  v.DurationSec,
		FileSize:     v.FileSize,
		Notes:        v.Notes,
		RecordedAt:   v.RecordedAt,
		ExpiresAt:    v.ExpiresAt,
		DaysLeft:     daysLeft,
		IsDeleted:    v.IsDeleted,
		CreatedAt:    v.CreatedAt,
	}
}
