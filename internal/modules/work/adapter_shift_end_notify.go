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

type notifyEmpContext struct {
	Name      string
	KeyNum    string
	BikeNum   string
	BranchID  *uuid.UUID
}

func normalizeBikePlateCompact(plate string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(plate), " ", ""))
}

func (a *gormEmployeeAdapter) loadNotifyEmpContext(ctx context.Context, empID uuid.UUID, sessionBike string) (*notifyEmpContext, error) {
	var emp struct {
		Name             string     `gorm:"column:name"`
		KeyNumber        string     `gorm:"column:key_number"`
		EmployeeNumber   string     `gorm:"column:employee_number"`
		MotorcycleNumber string     `gorm:"column:motorcycle_number"`
		BranchID         *uuid.UUID `gorm:"column:branch_id"`
	}
	if err := a.db.WithContext(ctx).Table("employees").Where("id = ? AND deleted_at IS NULL", empID).First(&emp).Error; err != nil {
		return nil, err
	}

	empName := strings.TrimSpace(emp.Name)
	if empName == "" {
		empName = "المندوب"
	}
	bikeNum := strings.TrimSpace(sessionBike)
	if bikeNum == "" {
		bikeNum = strings.TrimSpace(emp.MotorcycleNumber)
	}

	keyNum := strings.TrimSpace(emp.KeyNumber)
	if keyNum == "" && bikeNum != "" {
		var v struct {
			KeyNumber string `gorm:"column:key_number"`
		}
		cleanBike := strings.ReplaceAll(bikeNum, " ", "")
		if err := a.db.WithContext(ctx).Table("vehicles").
			Where("(plate_number = ? OR REPLACE(plate_number, ' ', '') = ?) AND deleted_at IS NULL", bikeNum, cleanBike).
			First(&v).Error; err == nil {
			keyNum = strings.TrimSpace(v.KeyNumber)
		}
	}

	return &notifyEmpContext{
		Name:     empName,
		KeyNum:   keyNum,
		BikeNum:  bikeNum,
		BranchID: emp.BranchID,
	}, nil
}

func (a *gormEmployeeAdapter) notifyBranchSupervisors(
	ctx context.Context,
	branchID *uuid.UUID,
	title string,
	body string,
	notifType string,
	pushData map[string]string,
) {
	var tokens []string
	tokenSet := make(map[string]bool)

	type supervisorRec struct {
		ID        uuid.UUID  `gorm:"column:id"`
		BranchID  *uuid.UUID `gorm:"column:branch_id"`
		PushToken string     `gorm:"column:push_token"`
	}
	var empSupervisors []supervisorRec
	empQuery := a.db.WithContext(ctx).Table("employees").
		Where("(LOWER(job_role) = 'supervisor' OR LOWER(role_name) = 'supervisor') AND deleted_at IS NULL")
	if branchID != nil {
		empQuery = empQuery.Where("branch_id = ?", *branchID)
	} else {
		empQuery = empQuery.Where("branch_id IS NULL")
	}
	_ = empQuery.Find(&empSupervisors).Error

	now := time.Now()
	for _, s := range empSupervisors {
		tok := strings.TrimSpace(s.PushToken)
		if tok != "" && !tokenSet[tok] {
			tokenSet[tok] = true
			tokens = append(tokens, tok)
		}
		_ = a.db.WithContext(ctx).Table("notifications").Create(map[string]interface{}{
			"id":          uuid.New(),
			"employee_id": s.ID,
			"branch_id":   s.BranchID,
			"title":       title,
			"body":        body,
			"type":        notifType,
			"status":      "unread",
			"created_at":  now,
			"updated_at":  now,
		}).Error
	}

	type adminRec struct {
		ID        uuid.UUID  `gorm:"column:id"`
		BranchID  *uuid.UUID `gorm:"column:branch_id"`
		PushToken string     `gorm:"column:push_token"`
		Role      string     `gorm:"column:role"`
	}
	var adminSupervisors []adminRec
	adminQuery := a.db.WithContext(ctx).Table("admins").
		Where("(LOWER(role) = 'supervisor') AND deleted_at IS NULL")
	if branchID != nil {
		adminQuery = adminQuery.Where("branch_id = ?", *branchID)
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
		_ = a.db.WithContext(ctx).Table("notifications").Create(map[string]interface{}{
			"id":         uuid.New(),
			"admin_id":   adm.ID,
			"branch_id":  adm.BranchID,
			"title":      title,
			"body":       body,
			"type":       notifType,
			"status":     "unread",
			"created_at": now,
			"updated_at": now,
		}).Error
	}

	if len(tokens) > 0 {
		legacyService.SendFCMBroadcast(tokens, title, body, pushData)
	}
}

func (a *gormEmployeeAdapter) SendShiftStartedNotificationToSupervisor(ctx context.Context, empID uuid.UUID, session *WorkSession, previousMotorcycle string) error {
	sessionCopy := *session
	prevBike := strings.TrimSpace(previousMotorcycle)
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("🚨 [WORK MODULE] Recovered panic in SendShiftStartedNotificationToSupervisor: %v", rec)
			}
		}()

		bgCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		info, err := a.loadNotifyEmpContext(bgCtx, empID, sessionCopy.MotorcycleNumber)
		if err != nil || info == nil {
			return
		}

		sid := sessionCopy.ID.String()
		startTitle := "بدء وردية مندوب"
		var startBody string
		if info.KeyNum != "" && info.BikeNum != "" {
			startBody = fmt.Sprintf("بدأ المندوب %s الوردية - مفتاح %s | دباب %s", info.Name, info.KeyNum, info.BikeNum)
		} else if info.KeyNum != "" {
			startBody = fmt.Sprintf("بدأ المندوب %s الوردية - مفتاح رقم %s", info.Name, info.KeyNum)
		} else if info.BikeNum != "" {
			startBody = fmt.Sprintf("بدأ المندوب %s الوردية - دباب رقم %s", info.Name, info.BikeNum)
		} else {
			startBody = fmt.Sprintf("بدأ المندوب %s الوردية الآن", info.Name)
		}

		a.notifyBranchSupervisors(bgCtx, info.BranchID, startTitle, startBody, "SHIFT_STARTED", map[string]string{
			"type":         "SHIFT_STARTED",
			"employee_id":  empID.String(),
			"employeeName": info.Name,
			"keyNumber":    info.KeyNum,
			"key_number":   info.KeyNum,
			"bikeNumber":   info.BikeNum,
			"sessionId":    sid,
			"session_id":   sid,
		})

		currCompact := normalizeBikePlateCompact(sessionCopy.MotorcycleNumber)
		prevCompact := normalizeBikePlateCompact(prevBike)
		if prevCompact != "" && currCompact != "" && prevCompact != currCompact {
			bikeTitle := "تغيير دباب مندوب"
			bikeBody := fmt.Sprintf("قام المندوب %s بتغيير الدباب من %s إلى %s", info.Name, prevBike, info.BikeNum)
			a.notifyBranchSupervisors(bgCtx, info.BranchID, bikeTitle, bikeBody, "MOTORCYCLE_CHANGED", map[string]string{
				"type":                 "MOTORCYCLE_CHANGED",
				"employee_id":          empID.String(),
				"employeeName":         info.Name,
				"keyNumber":            info.KeyNum,
				"key_number":           info.KeyNum,
				"bikeNumber":           info.BikeNum,
				"previous_motorcycle":  prevBike,
				"new_motorcycle":       info.BikeNum,
				"sessionId":            sid,
				"session_id":           sid,
			})
		}
	}()

	return nil
}

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

		info, err := a.loadNotifyEmpContext(bgCtx, empID, sessionCopy.MotorcycleNumber)
		if err != nil || info == nil {
			return
		}

		title := "إنهاء وردية مندوب"
		var body string
		if info.KeyNum != "" {
			body = fmt.Sprintf("أنهى المندوب %s - مفتاح رقم %s", info.Name, info.KeyNum)
		} else if info.BikeNum != "" {
			body = fmt.Sprintf("أنهى المندوب %s - دباب رقم %s", info.Name, info.BikeNum)
		} else {
			body = fmt.Sprintf("أنهى المندوب %s الشفت وبانتظار الاعتماد", info.Name)
		}

		sid := sessionCopy.ID.String()
		a.notifyBranchSupervisors(bgCtx, info.BranchID, title, body, "SHIFT_COMPLETED_PENDING_REVIEW", map[string]string{
			"type":         "SHIFT_COMPLETED_PENDING_REVIEW",
			"employee_id":  empID.String(),
			"employeeName": info.Name,
			"keyNumber":    info.KeyNum,
			"key_number":   info.KeyNum,
			"bikeNumber":   info.BikeNum,
			"sessionId":    sid,
			"session_id":   sid,
			"ordersCount":  fmt.Sprintf("%d", sessionCopy.OrdersCount),
			"distance":     fmt.Sprintf("%.1f", sessionCopy.Distance),
		})
	}()

	return nil
}
