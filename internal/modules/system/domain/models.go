package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLog model
type AuditLog struct {
	ID        uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	AdminName string     `gorm:"type:varchar(100);not null" json:"admin_name"`
	Action    string     `gorm:"type:varchar(100);not null;index" json:"action"`
	Details   string     `gorm:"type:text" json:"details"`
	IPAddress string     `gorm:"type:varchar(50)" json:"ip_address"`
	BranchID  *uuid.UUID `gorm:"type:char(36);index" json:"branch_id"`
	CreatedAt time.Time  `gorm:"index" json:"created_at"`
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// AppSetting key-value model
type AppSetting struct {
	ID        uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Key       string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`
	Value     string         `gorm:"type:text" json:"value"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *AppSetting) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// Notification model
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
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// BroadcastNotification model
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

// BroadcastRead tracks which employees have seen a broadcast
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

// BroadcastVote records employee responses to poll questions
type BroadcastVote struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	BroadcastID uuid.UUID `gorm:"type:char(36);index;not null" json:"broadcast_id"`
	EmployeeID  uuid.UUID `gorm:"type:char(36);index;not null" json:"employee_id"`
	Response    string    `gorm:"type:varchar(20);not null" json:"response"` // "AGREE" or "DISAGREE"
	Reason      string    `gorm:"type:text" json:"reason"`                  // optional reason
	CreatedAt   time.Time `json:"created_at"`
}

func (v *BroadcastVote) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}
