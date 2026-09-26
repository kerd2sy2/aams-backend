package dto

import (
	"github.com/google/uuid"
	"time"
)

type NotificationResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	ImageURL  string    `json:"image_url,omitempty"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateBroadcastRequest struct {
	Title          string     `json:"title"`
	Body           string     `json:"body"`
	TitleAr        string     `json:"title_ar"`
	TitleEn        string     `json:"title_en"`
	TitleBn        string     `json:"title_bn"`
	BodyAr         string     `json:"body_ar"`
	BodyEn         string     `json:"body_en"`
	BodyBn         string     `json:"body_bn"`
	ImageURL       string     `json:"image_url"`
	Target         string     `json:"target"` // "ALL" or "BRANCH"
	BranchID       *uuid.UUID `json:"branch_id"`
	HasPoll        bool       `json:"has_poll"`
	PollQuestion   string     `json:"poll_question"`
	PollQuestionAr string     `json:"poll_question_ar"`
	PollQuestionEn string     `json:"poll_question_en"`
	PollQuestionBn string     `json:"poll_question_bn"`
}

type BroadcastItemDTO struct {
	ID             uuid.UUID  `json:"id"`
	Title          string     `json:"title"`
	Body           string     `json:"body"`
	TitleAr        string     `json:"title_ar,omitempty"`
	TitleEn        string     `json:"title_en,omitempty"`
	TitleBn        string     `json:"title_bn,omitempty"`
	BodyAr         string     `json:"body_ar,omitempty"`
	BodyEn         string     `json:"body_en,omitempty"`
	BodyBn         string     `json:"body_bn,omitempty"`
	ImageURL       string     `json:"image_url"`
	Target         string     `json:"target"`
	BranchID       *uuid.UUID `json:"branch_id"`
	BranchName     string     `json:"branch_name,omitempty"`
	CreatedBy      string     `json:"created_by"`
	SentCount      int        `json:"sent_count"`
	HasPoll        bool       `json:"has_poll"`
	PollQuestion   string     `json:"poll_question"`
	PollQuestionAr string     `json:"poll_question_ar,omitempty"`
	PollQuestionEn string     `json:"poll_question_en,omitempty"`
	PollQuestionBn string     `json:"poll_question_bn,omitempty"`
	AgreeCount     int        `json:"agree_count"`
	DisagreeCount  int        `json:"disagree_count"`
	CreatedAt      time.Time  `json:"created_at"`
	IsRead         bool       `json:"is_read"`
	UserVote       string     `json:"user_vote,omitempty"` // "AGREE" or "DISAGREE"
}

type SubmitPollVoteRequest struct {
	Response string `json:"response" binding:"required,oneof=AGREE DISAGREE"`
	Reason   string `json:"reason"`
}

type BroadcastVoteItemDTO struct {
	ID             uuid.UUID `json:"id"`
	EmployeeID     uuid.UUID `json:"employee_id"`
	EmployeeName   string    `json:"employee_name"`
	EmployeeNumber string    `json:"employee_number"`
	NationalID     string    `json:"national_id"`
	Phone          string    `json:"phone"`
	Response       string    `json:"response"` // "AGREE" or "DISAGREE"
	Reason         string    `json:"reason"`
	CreatedAt      time.Time `json:"created_at"`
}

type PushTokenRequest struct {
	PushToken  string `json:"push_token"`
	DeviceUUID string `json:"device_uuid"`
	Language   string `json:"language"`
}
