package system

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *gormSystemRepository) FindNotificationsByAdmin(ctx context.Context, adminID uuid.UUID, branchID *uuid.UUID, status string) ([]Notification, error) {
	var notifs []Notification
	query := r.db.WithContext(ctx)

	if status == "unread" || status == "read" {
		query = query.Where("status = ?", status)
	}

	if branchID != nil {
		query = query.Where(
			"admin_id = ? OR employee_id = ? OR (admin_id IS NULL AND employee_id IS NULL AND (branch_id = ? OR branch_id IS NULL)) OR (branch_id = ? AND type IN ('SHIFT_STARTED','MOTORCYCLE_CHANGED','SHIFT_COMPLETED_PENDING_REVIEW','ADVANCE_REQUEST','MAINTENANCE_REQUEST','INVESTIGATION_REQUEST'))",
			adminID, adminID, branchID, branchID,
		)
	} else {
		query = query.Where(
			"admin_id = ? OR employee_id = ? OR (admin_id IS NULL AND employee_id IS NULL AND branch_id IS NULL)",
			adminID, adminID,
		)
	}

	if err := query.Order("created_at DESC").Limit(200).Find(&notifs).Error; err != nil {
		return nil, err
	}
	return notifs, nil
}

func (r *gormSystemRepository) MarkNotificationAsRead(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Notification{}).Where("id = ?", id).Update("status", "read").Error
}

func (r *gormSystemRepository) MarkAllNotificationsAsRead(ctx context.Context, adminID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Notification{}).
		Where("status = ? AND (admin_id = ? OR employee_id = ? OR (admin_id IS NULL AND employee_id IS NULL))", "unread", adminID, adminID).
		Update("status", "read").Error
}

func (r *gormSystemRepository) CreateNotification(ctx context.Context, notif *Notification) error {
	return r.db.WithContext(ctx).Create(notif).Error
}

func (r *gormSystemRepository) CreateBroadcast(ctx context.Context, broadcast *BroadcastNotification) error {
	return r.db.WithContext(ctx).Create(broadcast).Error
}

func (r *gormSystemRepository) FindBroadcasts(ctx context.Context, branchID *uuid.UUID, limit, offset int) ([]BroadcastNotification, int64, error) {
	var broadcasts []BroadcastNotification
	var total int64

	query := r.db.WithContext(ctx).Model(&BroadcastNotification{})
	if branchID != nil {
		query = query.Where("target = 'ALL' OR branch_id = ?", branchID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 50
	}
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&broadcasts).Error; err != nil {
		return nil, 0, err
	}

	return broadcasts, total, nil
}

func (r *gormSystemRepository) FindBroadcastByID(ctx context.Context, id uuid.UUID) (*BroadcastNotification, error) {
	var b BroadcastNotification
	if err := r.db.WithContext(ctx).First(&b, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormSystemRepository) DeleteBroadcast(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		_ = tx.Delete(&BroadcastVote{}, "broadcast_id = ?", id).Error
		_ = tx.Delete(&BroadcastRead{}, "broadcast_id = ?", id).Error
		return tx.Delete(&BroadcastNotification{}, "id = ?", id).Error
	})
}

func (r *gormSystemRepository) FindBroadcastsForEmployee(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time, limit int) ([]BroadcastItemDTO, error) {
	if limit <= 0 {
		limit = 50
	}

	query := r.db.WithContext(ctx).Model(&BroadcastNotification{})
	if branchID != nil {
		query = query.Where("target = 'ALL' OR branch_id = ?", branchID)
	} else {
		query = query.Where("target = 'ALL'")
	}

	if !registeredAt.IsZero() {
		query = query.Where("created_at >= ?", registeredAt.Add(-2*time.Minute))
	}

	var broadcasts []BroadcastNotification
	if err := query.Order("created_at DESC").Limit(limit).Find(&broadcasts).Error; err != nil {
		return nil, err
	}

	if len(broadcasts) == 0 {
		return []BroadcastItemDTO{}, nil
	}

	var readIDs []uuid.UUID
	r.db.WithContext(ctx).Model(&BroadcastRead{}).Where("employee_id = ?", empID).Pluck("broadcast_id", &readIDs)
	readMap := make(map[uuid.UUID]bool, len(readIDs))
	for _, id := range readIDs {
		readMap[id] = true
	}

	var votes []BroadcastVote
	r.db.WithContext(ctx).Where("employee_id = ?", empID).Find(&votes)
	voteMap := make(map[uuid.UUID]string, len(votes))
	for _, v := range votes {
		voteMap[v.BroadcastID] = v.Response
	}

	result := make([]BroadcastItemDTO, len(broadcasts))
	for i, b := range broadcasts {
		result[i] = BroadcastItemDTO{
			ID:             b.ID,
			Title:          b.Title,
			Body:           b.Body,
			TitleAr:        b.TitleAr,
			TitleEn:        b.TitleEn,
			TitleBn:        b.TitleBn,
			BodyAr:         b.BodyAr,
			BodyEn:         b.BodyEn,
			BodyBn:         b.BodyBn,
			ImageURL:       b.ImageURL,
			Target:         b.Target,
			BranchID:       b.BranchID,
			CreatedBy:      b.CreatedBy,
			SentCount:      b.SentCount,
			HasPoll:        b.HasPoll,
			PollQuestion:   b.PollQuestion,
			PollQuestionAr: b.PollQuestionAr,
			PollQuestionEn: b.PollQuestionEn,
			PollQuestionBn: b.PollQuestionBn,
			AgreeCount:     b.AgreeCount,
			DisagreeCount:  b.DisagreeCount,
			CreatedAt:      b.CreatedAt,
			IsRead:         readMap[b.ID],
			UserVote:       voteMap[b.ID],
		}
	}

	return result, nil
}

func (r *gormSystemRepository) FindUnreadBroadcastsForEmployee(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) ([]BroadcastItemDTO, error) {
	all, err := r.FindBroadcastsForEmployee(ctx, empID, branchID, registeredAt, 100)
	if err != nil {
		return nil, err
	}

	unread := make([]BroadcastItemDTO, 0)
	for _, b := range all {
		if !b.IsRead {
			unread = append(unread, b)
		}
	}
	return unread, nil
}

func (r *gormSystemRepository) MarkEmployeeBroadcastRead(ctx context.Context, empID uuid.UUID, broadcastID uuid.UUID) error {
	var count int64
	r.db.WithContext(ctx).Model(&BroadcastRead{}).Where("employee_id = ? AND broadcast_id = ?", empID, broadcastID).Count(&count)
	if count > 0 {
		return nil
	}

	rec := BroadcastRead{
		EmployeeID:  empID,
		BroadcastID: broadcastID,
		ReadAt:      time.Now(),
	}
	return r.db.WithContext(ctx).Create(&rec).Error
}

func (r *gormSystemRepository) MarkAllEmployeeBroadcastsRead(ctx context.Context, empID uuid.UUID, branchID *uuid.UUID, registeredAt time.Time) error {
	broadcasts, err := r.FindBroadcastsForEmployee(ctx, empID, branchID, registeredAt, 200)
	if err != nil {
		return err
	}

	for _, b := range broadcasts {
		if !b.IsRead {
			_ = r.MarkEmployeeBroadcastRead(ctx, empID, b.ID)
		}
	}
	return nil
}

func (r *gormSystemRepository) SubmitVote(ctx context.Context, vote *BroadcastVote) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing BroadcastVote
		err := tx.Where("broadcast_id = ? AND employee_id = ?", vote.BroadcastID, vote.EmployeeID).First(&existing).Error
		if err == nil {
			return errors.New("تم التصويت مسبقاً")
		}

		if err := tx.Create(vote).Error; err != nil {
			return err
		}

		switch vote.Response {
		case "AGREE":
			return tx.Model(&BroadcastNotification{}).Where("id = ?", vote.BroadcastID).
				Update("agree_count", gorm.Expr("agree_count + 1")).Error
		case "DISAGREE":
			return tx.Model(&BroadcastNotification{}).Where("id = ?", vote.BroadcastID).
				Update("disagree_count", gorm.Expr("disagree_count + 1")).Error
		}
		return nil
	})
}

func (r *gormSystemRepository) FindVotesForBroadcast(ctx context.Context, broadcastID uuid.UUID) ([]BroadcastVoteItemDTO, error) {
	var votes []struct {
		ID             uuid.UUID `json:"id"`
		EmployeeID     uuid.UUID `json:"employee_id"`
		EmployeeName   string    `json:"employee_name"`
		EmployeeNumber string    `json:"employee_number"`
		NationalID     string    `json:"national_id"`
		Phone          string    `json:"phone"`
		Response       string    `json:"response"`
		Reason         string    `json:"reason"`
		CreatedAt      time.Time `json:"created_at"`
	}

	err := r.db.WithContext(ctx).
		Table("broadcast_votes").
		Select("broadcast_votes.id, broadcast_votes.employee_id, employees.name as employee_name, employees.employee_number, employees.national_id, employees.phone, broadcast_votes.response, broadcast_votes.reason, broadcast_votes.created_at").
		Joins("LEFT JOIN employees ON employees.id = broadcast_votes.employee_id").
		Where("broadcast_votes.broadcast_id = ?", broadcastID).
		Order("broadcast_votes.created_at DESC").
		Scan(&votes).Error

	if err != nil {
		return nil, err
	}

	result := make([]BroadcastVoteItemDTO, len(votes))
	for i, v := range votes {
		result[i] = BroadcastVoteItemDTO{
			ID:             v.ID,
			EmployeeID:     v.EmployeeID,
			EmployeeName:   v.EmployeeName,
			EmployeeNumber: v.EmployeeNumber,
			NationalID:     v.NationalID,
			Phone:          v.Phone,
			Response:       v.Response,
			Reason:         v.Reason,
			CreatedAt:      v.CreatedAt,
		}
	}
	return result, nil
}

func (r *gormSystemRepository) UpdateEmployeePushToken(ctx context.Context, empID uuid.UUID, token, language string) error {
	cleanToken := strings.TrimSpace(token)
	updates := map[string]interface{}{}
	likeOp := r.likeOperator()

	if cleanToken != "" {
		updates["push_token"] = cleanToken

		// Ensure token uniqueness: remove this physical device's token from any other employee or admin
		parts := strings.Split(cleanToken, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				likeCond := fmt.Sprintf("id != ? AND push_token %s ?", likeOp)
				_ = r.db.WithContext(ctx).Table("employees").
					Where(likeCond, empID, "%"+trimmed+"%").
					Update("push_token", nil).Error
				_ = r.db.WithContext(ctx).Table("admins").
					Where(likeCond, empID, "%"+trimmed+"%").
					Update("push_token", nil).Error
			}
		}
	} else {
		// Empty token passed (e.g. employee logged out)
		updates["push_token"] = nil
	}

	cleanLang := strings.ToLower(strings.TrimSpace(language))
	if cleanLang != "" {
		updates["language"] = cleanLang
	}
	if len(updates) == 0 {
		return nil
	}
	_ = r.db.WithContext(ctx).Table("employees").Where("id = ?", empID).Updates(updates).Error
	adminUpdates := map[string]interface{}{}
	if pt, exists := updates["push_token"]; exists {
		adminUpdates["push_token"] = pt
	}
	if len(adminUpdates) > 0 {
		_ = r.db.WithContext(ctx).Table("admins").Where("id = ? OR employee_id = ?", empID, empID).Updates(adminUpdates).Error
	}
	return nil
}
