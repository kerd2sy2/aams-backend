package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Investigation model - supports multiple template types
type Investigation struct {
	ID                 uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID         uuid.UUID  `gorm:"type:char(36);index;not null" json:"employee_id"`
	Employee           *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	EmployeeName       string     `gorm:"type:varchar(255)" json:"employee_name"`
	SupervisorID       uuid.UUID  `gorm:"type:char(36);index;not null" json:"supervisor_id"`
	Supervisor         *Admin     `gorm:"foreignKey:SupervisorID" json:"supervisor,omitempty"`
	SupervisorName     string     `gorm:"type:varchar(100)" json:"supervisor_name"`
	NationalID         string     `gorm:"type:varchar(50)" json:"national_id"`
	Type               string     `gorm:"type:varchar(30);default:investigation" json:"type"` // investigation, supervisor_report, advance, internet_advance, absence, custody
	Questions          string     `gorm:"type:text" json:"questions"`                         // JSON array of question strings (for investigation)
	Answers            string     `gorm:"type:text" json:"answers"`                           // JSON array of answer strings (for investigation)
	ReportText         string     `gorm:"type:text" json:"report_text"`                       // Text content (for supervisor_report, absence)
	Images             string     `gorm:"type:text" json:"images"`                            // JSON array of image URLs (for supervisor_report)
	Amount             *float64   `gorm:"type:decimal(10,2)" json:"amount"`                   // Amount (for advance, internet_advance)
	StartDate          *time.Time `json:"start_date"`                                         // Absence start date
	EndDate            *time.Time `json:"end_date"`                                           // Absence end date
	Items              string     `gorm:"type:text" json:"items"`                             // JSON array of items (for custody)
	IsGuilty           *bool      `gorm:"default:false" json:"is_guilty"`
	Notes              string     `gorm:"type:text" json:"notes"`
	DeductionMonth     string     `gorm:"type:varchar(20)" json:"deduction_month"`              // ط´ظ‡ط± ط§ظ„ط®طµظ… ظ…ظ† ط§ظ„ط±ط§طھط¨ (ظ„ط³ظ„ظپط© ط§ظ„ظ†طھ)
	Status             string         `gorm:"type:varchar(20);default:pending;index" json:"status"`
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

// --- Attendance Model ---
type Attendance struct {
	ID         uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:idx_emp_date" json:"employee_id"`
	Employee   *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
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
	EmployeeID     uuid.UUID  `json:"employee_id"`
	EmployeeName   string     `json:"employee_name"`
	NationalID     string     `json:"national_id"`
	BranchName     string     `json:"branch_name"`
	VehicleType    string     `json:"vehicle_type"`
	Status         string     `json:"status"`
	Note           string     `json:"note"`
	HasWorkSession bool       `json:"has_work_session"`
	SessionID      *uuid.UUID `json:"session_id"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
}


// ------------------------------------------------------------------
// 4. EmployeeDocument Model (ط§ظ„ظ…ط³طھظ†ط¯ط§طھ ظˆط§ظ„ط±ط®طµ ظˆط§ظ„ط´ظ‡ط§ط¯ط§طھ)
// ------------------------------------------------------------------
type EmployeeDocument struct {
	ID         uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID uuid.UUID      `gorm:"type:char(36);index;not null" json:"employee_id"`
	Employee   *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	DocType    string         `gorm:"type:varchar(50);not null;index" json:"doc_type"` // PROMISSORY_NOTE, CONTRACT, DRIVING_LICENSE, VEHICLE_REGISTRATION, CRIMINAL_RECORD, MEDICAL_INSURANCE, OTHER
	Title      string         `gorm:"type:varchar(200);not null" json:"title"`
	DocNumber  string         `gorm:"type:varchar(100)" json:"doc_number"`
	FileURL    string         `gorm:"type:varchar(500)" json:"file_url"`
	IssueDate  *time.Time     `json:"issue_date"`
	ExpiryDate *time.Time     `gorm:"index" json:"expiry_date"`
	Status     string         `gorm:"type:varchar(30);default:'VALID';index" json:"status"` // VALID, EXPIRING_SOON, EXPIRED, PENDING_REVIEW
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

// ------------------------------------------------------------------
// 5. EmployeeBankAccount Model (ط§ظ„ط­ط³ط§ط¨ط§طھ ط§ظ„ط¨ظ†ظƒظٹط© ظ„ظ„ظ…ظ†ط§ط¯ظٹط¨)
// ------------------------------------------------------------------
type EmployeeBankAccount struct {
	ID               uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID       uuid.UUID      `gorm:"type:char(36);index;not null" json:"employee_id"`
	Employee         *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
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

// ------------------------------------------------------------------
// 6. LeaveRequest Model (ط·ظ„ط¨ط§طھ ط§ظ„ط¥ط¬ط§ط²ط§طھ)
// ------------------------------------------------------------------
type LeaveRequest struct {
	ID             uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	EmployeeID     uuid.UUID      `gorm:"type:char(36);index;not null" json:"employee_id"`
	Employee       *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	LeaveType      string         `gorm:"type:varchar(50);default:'ANNUAL';index" json:"leave_type"` // ANNUAL, SICK, EMERGENCY, UNPAID
	StartDate      string         `gorm:"type:varchar(10);not null;index" json:"start_date"`         // YYYY-MM-DD
	EndDate        string         `gorm:"type:varchar(10);not null;index" json:"end_date"`           // YYYY-MM-DD
	DaysCount      int            `gorm:"default:1" json:"days_count"`
	Reason         string         `gorm:"type:text" json:"reason"`
	Status         string         `gorm:"type:varchar(30);default:'PENDING';index" json:"status"` // PENDING, APPROVED, REJECTED
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

// ------------------------------------------------------------------
// 7. SupportTicket Model (طھط°ط§ظƒط± ط§ظ„ط¯ط¹ظ… ظˆط§ظ„ط´ظƒط§ظˆظ‰)
// ------------------------------------------------------------------
type SupportTicket struct {
	ID           uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	TicketNumber string         `gorm:"type:varchar(50);index;not null" json:"ticket_number"`
	EmployeeID   *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"`
	Employee     *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Subject      string         `gorm:"type:varchar(255);not null" json:"subject"`
	Category     string         `gorm:"type:varchar(50);default:'OPERATIONAL';index" json:"category"` // OPERATIONAL, FINANCIAL, VEHICLE, APPLICATION, OTHER
	Priority     string         `gorm:"type:varchar(20);default:'MEDIUM';index" json:"priority"`      // LOW, MEDIUM, HIGH, URGENT
	Status       string         `gorm:"type:varchar(30);default:'OPEN';index" json:"status"`          // OPEN, IN_PROGRESS, RESOLVED, CLOSED
	Description  string         `gorm:"type:text;not null" json:"description"`
	Resolution   string         `gorm:"type:text" json:"resolution"`
	BranchID     *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Branch       *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *SupportTicket) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type Notification struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BranchID   *uuid.UUID `gorm:"type:uuid;index" json:"branch_id"`
	AdminID    *uuid.UUID `gorm:"type:uuid;index" json:"admin_id"`
	EmployeeID *uuid.UUID `gorm:"type:uuid;index" json:"employee_id"`
	Title      string     `gorm:"type:varchar(255);not null" json:"title"`
	Body       string     `gorm:"type:text;not null" json:"body"`
	ImageURL   string     `gorm:"type:text" json:"image_url,omitempty"`
	Type       string     `gorm:"type:varchar(50);not null;index" json:"type"`
	Status     string     `gorm:"type:varchar(50);default:'unread';index" json:"status"`
	CreatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	Branch   *Branch   `gorm:"foreignKey:BranchID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"branch,omitempty"`
	Admin    *Admin    `gorm:"foreignKey:AdminID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"admin,omitempty"`
	Employee *Employee `gorm:"foreignKey:EmployeeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"employee,omitempty"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// BroadcastNotification model for mass announcements to delegates and phones
type BroadcastNotification struct {
	ID             uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	Title          string     `gorm:"type:varchar(255);not null" json:"title"`
	Body           string     `gorm:"type:text;not null" json:"body"`
	TitleAr        string     `gorm:"type:varchar(255)" json:"title_ar"`
	TitleEn        string     `gorm:"type:varchar(255)" json:"title_en"`
	TitleBn        string     `gorm:"type:varchar(255)" json:"title_bn"`
	BodyAr         string     `gorm:"type:text" json:"body_ar"`
	BodyEn         string     `gorm:"type:text" json:"body_en"`
	BodyBn         string     `gorm:"type:text" json:"body_bn"`
	ImageURL       string     `gorm:"type:text" json:"image_url"`
	Target         string     `gorm:"type:varchar(50);default:'ALL'" json:"target"` // "ALL", "BRANCH"
	BranchID       *uuid.UUID `gorm:"type:char(36);index" json:"branch_id"`
	Branch         *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	CreatedBy      string     `gorm:"type:varchar(150)" json:"created_by"`
	SentCount      int        `gorm:"default:0" json:"sent_count"`
	HasPoll        bool       `gorm:"default:false" json:"has_poll"`
	PollQuestion   string     `gorm:"type:varchar(255)" json:"poll_question"`
	PollQuestionAr string     `gorm:"type:varchar(255)" json:"poll_question_ar"`
	PollQuestionEn string     `gorm:"type:varchar(255)" json:"poll_question_en"`
	PollQuestionBn string     `gorm:"type:varchar(255)" json:"poll_question_bn"`
	AgreeCount     int        `gorm:"default:0" json:"agree_count"`
	DisagreeCount  int        `gorm:"default:0" json:"disagree_count"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (b *BroadcastNotification) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// BroadcastRead tracks which employees have seen/dismissed a broadcast notification
type BroadcastRead struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	BroadcastID uuid.UUID `gorm:"type:char(36);index;not null" json:"broadcast_id"`
	EmployeeID  uuid.UUID `gorm:"type:char(36);index;not null" json:"employee_id"`
	ReadAt      time.Time `json:"read_at"`
}

func (r *BroadcastRead) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// BroadcastVote records employee responses to poll questions (موافق / معترض)
type BroadcastVote struct {
	ID          uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	BroadcastID uuid.UUID  `gorm:"type:char(36);index;not null" json:"broadcast_id"`
	EmployeeID  uuid.UUID  `gorm:"type:char(36);index;not null" json:"employee_id"`
	Employee    *Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Response    string     `gorm:"type:varchar(20);not null" json:"response"` // "AGREE" or "DISAGREE"
	Reason      string     `gorm:"type:text" json:"reason"`                  // optional reason
	CreatedAt   time.Time  `json:"created_at"`
}

func (v *BroadcastVote) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

// OperationVideo stores silent video documentation uploaded by supervisors or employees.
// Retained on disk for 60 days, after which physical file is purged automatically by background cron.
type OperationVideo struct {
	ID           uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	BranchID     *uuid.UUID     `gorm:"type:char(36);index" json:"branch_id"`
	Branch       *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	EmployeeID   *uuid.UUID     `gorm:"type:char(36);index" json:"employee_id"`
	Employee     *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	EmployeeName string         `gorm:"type:varchar(150);index" json:"employee_name"`
	OrderNumber  string         `gorm:"type:varchar(100);index" json:"order_number"`
	SessionID    *uuid.UUID     `gorm:"type:char(36);index" json:"session_id"`
	VideoURL     string         `gorm:"type:text;not null" json:"video_url"`
	ThumbnailURL string         `gorm:"type:text" json:"thumbnail_url"`
	DurationSec  int            `gorm:"default:0" json:"duration_sec"`
	FileSize     int64          `gorm:"default:0" json:"file_size"`
	Notes        string         `gorm:"type:text" json:"notes"`
	RecordedAt   time.Time      `gorm:"index" json:"recorded_at"`
	ExpiresAt    time.Time      `gorm:"index;not null" json:"expires_at"` // RecordedAt + 60 days
	IsDeleted    bool           `gorm:"default:false;index" json:"is_deleted"`
	DeletedAt    *time.Time     `json:"deleted_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

func (o *OperationVideo) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	if o.RecordedAt.IsZero() {
		o.RecordedAt = time.Now()
	}
	if o.ExpiresAt.IsZero() {
		o.ExpiresAt = o.RecordedAt.AddDate(0, 0, 60)
	}
	return nil
}

