package contracts

import (
	"context"

	"github.com/google/uuid"
)

type CustodySummaryDTO struct {
	DayID          uuid.UUID  `json:"day_id"`
	BranchID       *uuid.UUID `json:"branch_id"`
	Date           string     `json:"date"`
	OpeningBalance float64    `json:"opening_balance"`
	AddedAmount    float64    `json:"added_amount"`
	TotalExpenses  float64    `json:"total_expenses"`
	ClosingBalance float64    `json:"closing_balance"`
}

type ICustodyContract interface {
	GetDaySummary(ctx context.Context, branchID *uuid.UUID, date string) (*CustodySummaryDTO, error)
}
