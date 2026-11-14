package work

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	
	legacyService "delivery-backend/internal/service"
)

func (a *gormEmployeeAdapter) SendShiftEndedNotificationToSupervisor(ctx context.Context, empID uuid.UUID, session *WorkSession) error {
	sessionCopy := *session
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("🚨 [WORK MODULE] Recovered panic in SendShiftEndedNotificationToSupervisor: %v", rec)
			}
		}()

		bgCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		var emp struct {
			Name             string     `gorm:"column:name"`
			KeyNumber        string     `gorm:"column:key_number"`
			EmployeeNumber   string     `gorm:"column:employee_number"`
			MotorcycleNumber string     `gorm:"column:motorcycle_number"`
			BranchID         *uuid.UUID `gorm:"column:branch_id"`
		}
		if err := a.db.WithContext(bgCtx).Table("employees").Where("id = ? AND deleted_at IS NULL", empID).First(&emp).Error; err != nil {
			return
		}

		empName := strings.TrimSpace(emp.Name)
		if empName == "" {
			empName = "المندوب"
		}
		bikeNum := strings.TrimSpace(sessionCopy.MotorcycleNumber)
		if bikeNum == "" {
			bikeNum = strings.TrimSpace(emp.MotorcycleNumber)
		}

		keyNum := strings.TrimSpace(emp.KeyNumber)
		if keyNum == "" && bikeNum != "" {
			var v struct {
				KeyNumber string `gorm:"column:key_number"`
			}
			cleanBike := strings.ReplaceAll(bikeNum, " ", "")
			if err := a.db.WithContext(bgCtx).Table("vehicles").
				Where("(plate_number = ? OR REPLACE(plate_number, ' ', '') = ?) AND deleted_at IS NULL", bikeNum, cleanBike).
				First(&v).Error; err == nil {
				keyNum = strings.TrimSpace(v.KeyNumber)
			}
		}

		title := "إنهاء وردية مندوب"
		var body string
		if keyNum != "" {
			body = fmt.Sprintf("أنهى المندوب %s - مفتاح رقم %s", empName, keyNum)
		} else if bikeNum != "" {
			body = fmt.Sprintf("أنهى المندوب %s - دباب رقم %s", empName, bikeNum)
		} else {
			body = fmt.Sprintf("أنهى المندوب %s الشفت وبانتظار الاعتماد", empName)
		}

		var tokens []string
		tokenSet := make(map[string]bool)

		type supervisorRec struct {
			ID        uuid.UUID  `gorm:"column:id"`
			BranchID  *uuid.UUID `gorm:"column:branch_id"`
			PushToken string     `gorm:"column:push_token"`
		}
		var empSupervisors []supervisorRec
		empQuery := a.db.WithContext(bgCtx).Table("employees").
			Where("(LOWER(job_role) = 'supervisor' OR LOWER(role_name) = 'supervisor') AND deleted_at IS NULL AND push_token IS NOT NULL AND push_token != ''")
		if emp.BranchID != nil {
			empQuery = empQuery.Where("branch_id = ?", *emp.BranchID)
		} else {
			empQuery = empQuery.Where("branch_id IS NULL")
		}
		_ = empQuery.Find(&empSupervisors).Error

		for _, s := range empSupervisors {
			tok := strings.TrimSpace(s.PushToken)
			if tok != "" && !tokenSet[tok] {
				tokenSet[tok] = true
				tokens = append(tokens, tok)
			}
			_ = a.db.WithContext(bgCtx).Table("notifications").Create(map[string]interface{}{
				"id":          uuid.New(),
				"employee_id": s.ID,
				"branch_id":   s.BranchID,
				"title":       title,
				"body":        body,
				"type":        "SHIFT_COMPLETED_PENDING_REVIEW",
				"status":      "unread",
				"created_at":  time.Now(),
				"updated_at":  time.Now(),
			}).Error
		}

		type adminRec struct {
			ID        uuid.UUID  `gorm:"column:id"`
			BranchID  *uuid.UUID `gorm:"column:branch_id"`
			PushToken string     `gorm:"column:push_token"`
			Role      string     `gorm:"column:role"`
		}
		var adminSupervisors []adminRec
		adminQuery := a.db.WithContext(bgCtx).Table("admins").
			Where("(LOWER(role) = 'supervisor') AND deleted_at IS NULL AND push_token IS NOT NULL AND push_token != ''")
		if emp.BranchID != nil {
			adminQuery = adminQuery.Where("branch_id = ?", *emp.BranchID)
		} else {
			adminQuery = adminQuery.Where("branch_id IS NULL")
		}
		_ = adminQuery.Find(&adminSupervisors).Error

		for _, adm := range adminSupervisors {
			tok := strings.TrimSpace(adm.PushToken)
			if tok != "" && !tokenSet[tok] {
				tokenSet[tok] = true
				tokens = append(tokens, tok)
			}
			_ = a.db.WithContext(bgCtx).Table("notifications").Create(map[string]interface{}{
				"id":         uuid.New(),
				"admin_id":   adm.ID,
				"branch_id":  adm.BranchID,
				"title":      title,
				"body":       body,
				"type":       "SHIFT_COMPLETED_PENDING_REVIEW",
				"status":     "unread",
				"created_at": time.Now(),
				"updated_at": time.Now(),
			}).Error
		}

		if len(tokens) > 0 {
			sid := sessionCopy.ID.String()
			legacyService.SendFCMBroadcast(tokens, title, body, map[string]string{
				"type":         "SHIFT_COMPLETED_PENDING_REVIEW",
				"employee_id":  empID.String(),
				"employeeName": empName,
				"keyNumber":    keyNum,
				"key_number":   keyNum,
				"bikeNumber":   bikeNum,
				"sessionId":    sid,
				"session_id":   sid,
				"ordersCount":  fmt.Sprintf("%d", sessionCopy.OrdersCount),
				"distance":     fmt.Sprintf("%.1f", sessionCopy.Distance),
			})
		}
	}()

	return nil
}
