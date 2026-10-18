package dto

import "time"

type OperationVideoItem struct {
	ID           string    `json:"id"`
	BranchID     *string   `json:"branch_id,omitempty"`
	EmployeeID   *string   `json:"employee_id,omitempty"`
	EmployeeName string    `json:"employee_name"`
	OrderNumber  string    `json:"order_number"`
	SessionID    *string   `json:"session_id,omitempty"`
	VideoURL     string    `json:"video_url"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	DurationSec  int       `json:"duration_sec"`
	FileSize     int64     `json:"file_size"`
	Notes        string    `json:"notes"`
	RecordedAt   time.Time `json:"recorded_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	DaysLeft     int       `json:"days_left"`
	IsDeleted    bool      `json:"is_deleted"`
	CreatedAt    time.Time `json:"created_at"`
}

type OperationVideoFilter struct {
	Date         string `form:"date"`          // YYYY-MM-DD
	EmployeeName string `form:"employee_name"` // Substring search
	OrderNumber  string `form:"order_number"`
	BranchID     string `form:"branch_id"`
	Limit        int    `form:"limit"`
	Offset       int    `form:"offset"`
}
