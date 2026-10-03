package dto

import (
	"time"

	"github.com/google/uuid"
)

type CustodyExpenseResponse struct {
	ID            uuid.UUID `json:"id"`
	CustodyDayID  uuid.UUID `json:"custody_day_id"`
	Category      string    `json:"category"`
	Amount        float64   `json:"amount"`
	RecipientName string    `json:"recipient_name"`
	CreatedAt     time.Time `json:"created_at"`
}

type CustodyTotals struct {
	Fuel       float64 `json:"fuel"`
	License    float64 `json:"license"`
	SpareParts float64 `json:"spare_parts"`
	Other      float64 `json:"other"`
}

type CustodyDayResponse struct {
	ID             uuid.UUID                `json:"id"`
	BranchID       *uuid.UUID               `json:"branch_id"`
	BranchName     string                   `json:"branch_name"`
	Date           string                   `json:"date"`
	OpeningBalance float64                  `json:"opening_balance"`
	AddedAmount    float64                  `json:"added_amount"`
	CustodyValue   float64                  `json:"custody_value"`
	TotalExpenses  float64                  `json:"total_expenses"`
	ClosingBalance float64                  `json:"closing_balance"`
	Totals         CustodyTotals            `json:"totals"`
	Expenses       []CustodyExpenseResponse `json:"expenses"`
	CreatedAt      time.Time                `json:"created_at"`
}

type CreateCustodyDayRequest struct {
	Date        string     `json:"date" binding:"required"`
	AddedAmount float64    `json:"added_amount"`
	BranchID    *uuid.UUID `json:"branch_id"`
}

type CreateCustodyExpenseRequest struct {
	Category      string  `json:"category" binding:"required"`
	Amount        float64 `json:"amount" binding:"gte=0"`
	RecipientName string  `json:"recipient_name"`
}

type AddCustodyAmountRequest struct {
	CustodyDayID uuid.UUID  `json:"custody_day_id" binding:"required"`
	AddedAmount  float64    `json:"added_amount" binding:"gt=0"`
	BranchID     *uuid.UUID `json:"branch_id"`
}

type CustodyLogFilter struct {
	BranchID   string `form:"branch_id"`
	Date       string `form:"date"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	ActionType string `form:"action_type"`
	CreatedBy  string `form:"created_by"`
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=50"`
	PageSize   int    `form:"page_size"`
}

func (f *CustodyLogFilter) GetEffectiveLimit() int {
	if f.PageSize > 0 {
		return f.PageSize
	}
	if f.Limit > 0 {
		return f.Limit
	}
	return 50
}

func (f *CustodyLogFilter) GetEffectivePage() int {
	if f.Page <= 0 {
		return 1
	}
	return f.Page
}
