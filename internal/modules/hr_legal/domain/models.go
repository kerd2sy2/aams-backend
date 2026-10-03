package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("record not found")
)

type Investigation struct {
	ID                 uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID         uuid.UUID      `gorm:"type:char(36);index;not null" json:"employee_id"`
	SupervisorID       uuid.UUID      `gorm:"type:char(36);index;not null" json:"supervisor_id"`
	NationalID         string         `gorm:"type:varchar(50)" json:"national_id"`
	Type               string         `gorm:"type:varchar(30);default:investigation" json:"type"` // investigation, supervisor_report, advance, internet_advance, absence, custody
	Questions          string         `gorm:"type:text" json:"questions"`                         // JSON array
	Answers            string         `gorm:"type:text" json:"answers"`                           // JSON array
	ReportText         string         `gorm:"type:text" json:"report_text"`
	Images             string         `gorm:"type:text" json:"images"`                            // JSON array
	Amount             *float64       `gorm:"type:decimal(10,2)" json:"amount"`
	StartDate          *time.Time     `json:"start_date"`
	EndDate            *time.Time     `json:"end_date"`
	Items              string         `gorm:"type:text" json:"items"`
	IsGuilty           *bool          `gorm:"default:false" json:"is_guilty"`
	Notes              string         `gorm:"type:text" json:"notes"`
	DeductionMonth     string         `gorm:"type:varchar(7)" json:"deduction_month"`
	Status             string         `gorm:"type:varchar(30);default:pending" json:"status"`
	ApprovedBy         *uuid.UUID     `gorm:"type:char(36)" json:"approved_by"`
	ApprovedByName     string         `gorm:"type:varchar(100)" json:"approved_by_name"`
	ApprovedByUsername string         `gorm:"type:varchar(50)" json:"approved_by_username"`
	RejectedBy         *uuid.UUID     `gorm:"type:char(36)" json:"rejected_by"`
	RejectedByName     string         `gorm:"type:varchar(100)" json:"rejected_by_name"`
	RejectedByUsername string         `gorm:"type:varchar(50)" json:"rejected_by_username"`
	ApprovedAt         *time.Time     `json:"approved_at"`
	RejectedAt         *time.Time     `json:"rejected_at"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (i *Investigation) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

type EmployeeDocument struct {
	ID         uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID uuid.UUID      `gorm:"type:char(36);index;not null" json:"employee_id"`
	DocType    string         `gorm:"type:varchar(50);not null;index" json:"doc_type"`
	Title      string         `gorm:"type:varchar(200);not null" json:"title"`
	DocNumber  string         `gorm:"type:varchar(100)" json:"doc_number"`
	FileURL    string         `gorm:"type:varchar(500)" json:"file_url"`
	IssueDate  *time.Time     `json:"issue_date"`
	ExpiryDate *time.Time     `gorm:"index" json:"expiry_date"`
	Status     string         `gorm:"type:varchar(30);default:'VALID';index" json:"status"`
	Notes      string         `gorm:"type:text" json:"notes"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (d *EmployeeDocument) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

type EmployeeBankAccount struct {
	ID               uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID       uuid.UUID      `gorm:"type:char(36);index;not null" json:"employee_id"`
	BankName         string         `gorm:"type:varchar(100);not null" json:"bank_name"`
	IBAN             string         `gorm:"type:varchar(50);not null;index" json:"iban"`
	AccountOwnerName string         `gorm:"type:varchar(150);not null" json:"account_owner_name"`
	IsDefault        bool           `gorm:"default:false" json:"is_default"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *EmployeeBankAccount) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type LeaveRequest struct {
	ID             uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID     uuid.UUID      `gorm:"type:char(36);index;not null" json:"employee_id"`
	LeaveType      string         `gorm:"type:varchar(50);default:'ANNUAL';index" json:"leave_type"`
	StartDate      string         `gorm:"type:varchar(10);not null;index" json:"start_date"` // YYYY-MM-DD
	EndDate        string         `gorm:"type:varchar(10);not null;index" json:"end_date"`   // YYYY-MM-DD
	DaysCount      int            `gorm:"default:1" json:"days_count"`
	Reason         string         `gorm:"type:text" json:"reason"`
	Status         string         `gorm:"type:varchar(30);default:'PENDING';index" json:"status"`
	ApprovedByName string         `gorm:"type:varchar(100)" json:"approved_by_name"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (l *LeaveRequest) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

type TrafficViolation struct {
	ID              uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	ViolationNumber string         `gorm:"type:varchar(100);index" json:"violation_number"`
	EmployeeID      *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"`
	VehiclePlate    string         `gorm:"type:varchar(50);index" json:"vehicle_plate"`
	Amount          float64        `gorm:"not null;default:0" json:"amount"`
	PaidAmount      float64        `gorm:"not null;default:0" json:"paid_amount"`
	Reason          string         `gorm:"type:varchar(255);not null" json:"reason"`
	ViolationDate   time.Time      `gorm:"index" json:"violation_date"`
	City            string         `gorm:"type:varchar(100)" json:"city"`
	Status          string         `gorm:"type:varchar(30);default:'RECORDED';index" json:"status"`
	BranchID        *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Notes           string         `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (v *TrafficViolation) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

type FuelLog struct {
	ID              uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID      *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"`
	VehiclePlate    string         `gorm:"type:varchar(50);index" json:"vehicle_plate"`
	ShiftID         *uuid.UUID     `gorm:"type:char(36);index" json:"shift_id"`
	Amount          float64        `gorm:"not null;default:0" json:"amount"`
	Liters          float64        `gorm:"default:0" json:"liters"`
	FuelDate        time.Time      `gorm:"index" json:"fuel_date"`
	StationName     string         `gorm:"type:varchar(150)" json:"station_name"`
	InvoiceImageURL string         `gorm:"type:varchar(500)" json:"invoice_image_url"`
	BranchID        *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Notes           string         `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (f *FuelLog) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

type EmployeeInfo struct {
	ID               uuid.UUID  `gorm:"column:id"`
	Name             string     `gorm:"column:name"`
	Language         string     `gorm:"column:language"`
	PushToken        string     `gorm:"column:push_token"`
	BranchID         *uuid.UUID `gorm:"column:branch_id"`
	MotorcycleNumber string     `gorm:"column:motorcycle_number"`
}

