package operation_video

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OperationVideo stores silent video documentation uploaded by supervisors or employees.
// Retained on disk for 60 days, after which physical file is purged automatically by background cron.
type OperationVideo struct {
	ID           uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	BranchID     *uuid.UUID `gorm:"type:char(36);index" json:"branch_id"`
	EmployeeID   *uuid.UUID `gorm:"type:char(36);index" json:"employee_id"`
	EmployeeName string     `gorm:"type:varchar(150);index" json:"employee_name"`
	OrderNumber  string     `gorm:"type:varchar(100);index" json:"order_number"`
	SessionID    *uuid.UUID `gorm:"type:char(36);index" json:"session_id"`
	VideoURL     string     `gorm:"type:text;not null" json:"video_url"`
	ThumbnailURL string     `gorm:"type:text" json:"thumbnail_url"`
	DurationSec  int        `gorm:"default:0" json:"duration_sec"`
	FileSize     int64      `gorm:"default:0" json:"file_size"`
	Notes        string     `gorm:"type:text" json:"notes"`
	RecordedAt   time.Time  `gorm:"index" json:"recorded_at"`
	ExpiresAt    time.Time  `gorm:"index;not null" json:"expires_at"` // RecordedAt + 60 days
	IsDeleted    bool       `gorm:"default:false;index" json:"is_deleted"`
	DeletedAt    *time.Time `json:"deleted_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (OperationVideo) TableName() string {
	return "operation_videos"
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
