package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrDayNotFound     = errors.New("custody day not found")
	ErrExpenseNotFound = errors.New("custody expense not found")
	ErrLogNotFound     = errors.New("custody log not found")
)

type CustodyDay struct {
	ID             uuid.UUID        `gorm:"type:char(36);primary_key" json:"id"`
	BranchID       *uuid.UUID       `gorm:"type:char(36);uniqueIndex:idx_custody_day_branch_date" json:"branch_id"`
	Date           string           `gorm:"type:varchar(10);not null;uniqueIndex:idx_custody_day_branch_date" json:"date"`
	OpeningBalance float64          `gorm:"default:0" json:"opening_balance"`
	AddedAmount    float64          `gorm:"default:0" json:"added_amount"`
	ClosingBalance float64          `gorm:"default:0" json:"closing_balance"`
	Expenses       []CustodyExpense `gorm:"foreignKey:CustodyDayID;constraint:OnDelete:CASCADE" json:"expenses,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

func (c *CustodyDay) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type CustodyExpense struct {
	ID                uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	CustodyDayID      uuid.UUID  `gorm:"type:char(36);index;not null" json:"custody_day_id"`
	Category          string     `gorm:"type:varchar(20);not null;index" json:"category"` // fuel, license, spare_parts, other
	Amount            float64    `gorm:"default:0" json:"amount"`
	RecipientName     string     `gorm:"type:varchar(150)" json:"recipient_name"`
	CreatedByID       *uuid.UUID `gorm:"type:char(36);index" json:"created_by_id"`
	CreatedByName     string     `gorm:"type:varchar(100)" json:"created_by_name"`
	CreatedByUsername string     `gorm:"type:varchar(50)" json:"created_by_username"`
	CreatedAt         time.Time  `json:"created_at"`
}

func (e *CustodyExpense) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

type CustodyLog struct {
	ID            uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	BranchID      *uuid.UUID `gorm:"type:char(36);index" json:"branch_id"`
	CustodyDayID  uuid.UUID  `gorm:"type:char(36);index" json:"custody_day_id"`
	Date          string     `gorm:"type:varchar(10);not null;index" json:"date"`
	ActionType    string     `gorm:"type:varchar(30);not null;index" json:"action_type"` // ADD_CUSTODY, ADD_EXPENSE, DELETE_EXPENSE
	Category      string     `gorm:"type:varchar(30)" json:"category"`                   // fuel, license, spare_parts, other, custody
	Amount        float64    `gorm:"default:0" json:"amount"`
	Description   string     `gorm:"type:varchar(255)" json:"description"`
	RecipientName string     `gorm:"type:varchar(150)" json:"recipient_name"`
	AdminID       *uuid.UUID `gorm:"type:char(36);index" json:"admin_id"`
	AdminName     string     `gorm:"type:varchar(100)" json:"admin_name"`
	AdminUsername string     `gorm:"type:varchar(50)" json:"admin_username"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (l *CustodyLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}
