package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Attendance model
type Attendance struct {
	ID         uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:idx_emp_date" json:"employee_id"`
	Date       string    `gorm:"type:varchar(10);not null;uniqueIndex:idx_emp_date" json:"date"`
	Status     string    `gorm:"type:varchar(20);not null;default:present" json:"status"` // present / absent
	Note       string    `gorm:"type:text" json:"note"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (a *Attendance) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.Status == "" {
		a.Status = "present"
	}
	return nil
}

type AttendanceInfo struct {
	EmployeeID     uuid.UUID `json:"employee_id"`
	EmployeeName   string    `json:"employee_name"`
	NationalID     string    `json:"national_id"`
	BranchName     string    `json:"branch_name"`
	VehicleType    string    `json:"vehicle_type"`
	Status         string    `json:"status"`
	Note           string    `json:"note"`
	HasWorkSession bool      `json:"has_work_session"`
}
