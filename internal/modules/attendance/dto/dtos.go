package dto

type ToggleAttendanceRequest struct {
	Date   string `json:"date" binding:"required"`
	Status string `json:"status" binding:"required"`
	Note   string `json:"note"`
}
