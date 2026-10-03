package contracts

import (
	"context"

	"github.com/google/uuid"
)

type TargetSummaryDTO struct {
	TotalIdentifiers int `json:"total_identifiers"`
	TotalOrders      int `json:"total_orders"`
	ActiveAlerts     int `json:"active_alerts"`
}

type ITargetContract interface {
	GetSummary(ctx context.Context, month string) (*TargetSummaryDTO, error)
	GetIdentifierOrders(ctx context.Context, identifierID uuid.UUID, month string) (int, error)
}
