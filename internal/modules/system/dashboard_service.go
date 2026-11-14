package system

import (
	"context"

	"github.com/google/uuid"
)

type DashboardService interface {
	GetStats(ctx context.Context, branchID *uuid.UUID) (*DashboardStatsResponse, error)
}

type dashboardService struct {
	repo SystemRepository
}

func NewDashboardService(repo SystemRepository) DashboardService {
	return &dashboardService{repo: repo}
}

func (s *dashboardService) GetStats(ctx context.Context, branchID *uuid.UUID) (*DashboardStatsResponse, error) {
	return s.repo.GetDashboardStats(ctx, branchID)
}
