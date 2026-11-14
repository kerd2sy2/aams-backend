package hr_legal

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	legacyService "delivery-backend/internal/service"

	"github.com/google/uuid"
)

func (s *hrLegalService) sendViolationNotification(ctx context.Context, v *TrafficViolation, isPayment bool, paymentAmount float64) {
	if v == nil || v.EmployeeID == nil {
		return
	}

	empInfo, err := s.repo.GetEmployeeInfo(ctx, *v.EmployeeID)
	if err != nil || empInfo == nil {
		return
	}

	r := strings.ToLower(v.Reason)
	isPenalty := strings.Contains(r, "جزاء") ||
		strings.Contains(r, "خصم") ||
		strings.Contains(r, "تأخير") ||
		strings.Contains(r, "غياب") ||
		strings.Contains(r, "زي") ||
		strings.Contains(r, "عهدة") ||
		strings.Contains(r, "إهمال") ||
		strings.Contains(r, "سلوك")

	lang := strings.ToLower(strings.TrimSpace(empInfo.Language))

	var title, body string
	if isPayment {
		rem := math.Max(0, v.Amount-v.PaidAmount)
		switch lang {
		case "en":
			title = "Deduction / Payment Notice"
			body = fmt.Sprintf("An amount of %.2f SAR was deducted/paid for (%s). Remaining: %.2f SAR.", paymentAmount, v.Reason, rem)
		case "bn":
			title = "কর্তন ও পরিশোধের বিজ্ঞপ্তি"
			body = fmt.Sprintf("%.2f রিয়াল কর্তন/পরিশোধ করা হয়েছে (%s)। অবশিষ্ট: %.2f রিয়াল।", paymentAmount, v.Reason, rem)
		case "ur":
			title = "کٹوتی اور ادائیگی کا نوٹس"
			body = fmt.Sprintf("(%s) کی مد میں %.2f ریال کٹوتی/ادائیگی کی گئی۔ باقی: %.2f ریال۔", v.Reason, paymentAmount, rem)
		default: // "ar"
			title = "إشعار خصم وسداد"
			body = fmt.Sprintf("تم خصم/سداد دفعة بقيمة %.2f ريال من (%s). المتبقي: %.2f ريال.", paymentAmount, v.Reason, rem)
		}
	} else if isPenalty {
		switch lang {
		case "en":
			title = "New Administrative Penalty"
			body = fmt.Sprintf("An administrative penalty/deduction of %.2f SAR has been recorded for (%s).", v.Amount, v.Reason)
		case "bn":
			title = "নতুন প্রশাসনিক জরিমানা ও কর্তন"
			body = fmt.Sprintf("আপনার উপর %.2f রিয়াল জরিমানা/কর্তন ধার্য করা হয়েছে (%s)।", v.Amount, v.Reason)
		case "ur":
			title = "نیا انتظامی جرمانہ اور کٹوتی"
			body = fmt.Sprintf("آپ پر (%s) کی وجہ سے %.2f ریال کا جرمانہ/کٹوتی عائد کی گئی ہے۔", v.Reason, v.Amount)
		default: // "ar"
			title = "إشعار جزاء وخصم إداري"
			body = fmt.Sprintf("تم تسجيل جزاء/خصم عليك بقيمة %.2f ريال بسبب (%s).", v.Amount, v.Reason)
		}
	} else {
		switch lang {
		case "en":
			title = "New Traffic Violation"
			body = fmt.Sprintf("A traffic violation of %.2f SAR has been recorded for (%s).", v.Amount, v.Reason)
		case "bn":
			title = "নতুন ট্রাফিক জরিমানা"
			body = fmt.Sprintf("আপনার উপর %.2f রিয়াল ট্রাফিক জরিমানা ধার্য করা হয়েছে (%s)।", v.Amount, v.Reason)
		case "ur":
			title = "نئی ٹریفک خلاف ورزی"
			body = fmt.Sprintf("آپ پر (%s) کی وجہ سے %.2f ریال کی ٹریفک خلاف ورزی درج کی گئی۔", v.Reason, v.Amount)
		default: // "ar"
			title = "مخالفة مرورية جديدة"
			body = fmt.Sprintf("تم تسجيل مخالفة مرورية عليك بقيمة %.2f ريال بسبب (%s).", v.Amount, v.Reason)
		}
	}

	// 1. Save in-app notification in DB
	_ = s.repo.SaveNotification(ctx, *v.EmployeeID, v.BranchID, title, body, "VIOLATION")

	// 2. Send push notification to courier's phone
	if empInfo.PushToken != "" {
		token := empInfo.PushToken
		vID := v.ID.String()
		go legacyService.SendFCMBroadcast(
			[]string{token},
			title,
			body,
			map[string]string{
				"type":         "VIOLATION",
				"violation_id": vID,
				"amount":       fmt.Sprintf("%.2f", v.Amount),
			},
		)
	}
}

func (s *hrLegalService) CreateViolation(ctx context.Context, req CreateTrafficViolationRequest) (*TrafficViolation, error) {
	status := "RECORDED"
	if req.Status != "" {
		status = req.Status
	}

	var paidAmount float64
	if req.PaidAmount != nil {
		paidAmount = *req.PaidAmount
	}

	var violationDate time.Time
	if req.ViolationDate != "" {
		if t, err := time.Parse("2006-01-02", req.ViolationDate); err == nil {
			violationDate = t
		}
	}

	branchID := req.BranchID
	if branchID == nil && req.EmployeeID != nil {
		if empInfo, err := s.repo.GetEmployeeInfo(ctx, *req.EmployeeID); err == nil && empInfo != nil {
			branchID = empInfo.BranchID
		}
	}

	v := &TrafficViolation{
		ID:              uuid.New(),
		ViolationNumber: req.ViolationNumber,
		EmployeeID:      req.EmployeeID,
		VehiclePlate:    req.VehiclePlate,
		Amount:          req.Amount,
		PaidAmount:      paidAmount,
		Reason:          req.Reason,
		ViolationDate:   violationDate,
		City:            req.City,
		Status:          status,
		BranchID:        branchID,
		Notes:           req.Notes,
	}

	if err := s.repo.CreateViolation(ctx, v); err != nil {
		return nil, err
	}

	created, err := s.repo.FindViolationByID(ctx, v.ID)
	if err == nil && created != nil {
		s.sendViolationNotification(ctx, created, false, 0)
	}
	return created, err
}

func (s *hrLegalService) UpdateViolation(ctx context.Context, id uuid.UUID, req UpdateTrafficViolationRequest) (*TrafficViolation, error) {
	v, err := s.repo.FindViolationByID(ctx, id)
	if err != nil {
		return nil, errors.New("المخالفة غير موجودة")
	}

	var paymentAdded float64

	if req.ViolationNumber != nil {
		v.ViolationNumber = *req.ViolationNumber
	}
	if req.EmployeeID != nil {
		v.EmployeeID = req.EmployeeID
		if v.BranchID == nil {
			if empInfo, err := s.repo.GetEmployeeInfo(ctx, *req.EmployeeID); err == nil && empInfo != nil {
				v.BranchID = empInfo.BranchID
			}
		}
	}
	if req.VehiclePlate != nil {
		v.VehiclePlate = *req.VehiclePlate
	}
	if req.Amount != nil {
		v.Amount = *req.Amount
	}
	if req.PaidAmount != nil {
		if *req.PaidAmount > v.PaidAmount {
			paymentAdded = *req.PaidAmount - v.PaidAmount
		}
		v.PaidAmount = *req.PaidAmount
	}
	if req.AddPayment != nil && *req.AddPayment > 0 {
		v.PaidAmount += *req.AddPayment
		paymentAdded = *req.AddPayment
	}
	if req.Reason != nil {
		v.Reason = *req.Reason
	}
	if req.ViolationDate != nil && *req.ViolationDate != "" {
		if t, err := time.Parse("2006-01-02", *req.ViolationDate); err == nil {
			v.ViolationDate = t
		}
	}
	if req.City != nil {
		v.City = *req.City
	}
	if req.Status != nil {
		v.Status = *req.Status
		if (*req.Status == "DEDUCTED" || *req.Status == "PAID") && v.PaidAmount < v.Amount {
			paymentAdded = v.Amount - v.PaidAmount
			v.PaidAmount = v.Amount
		} else if *req.Status == "RECORDED" && req.PaidAmount == nil {
			v.PaidAmount = 0
		}
	}
	if req.Notes != nil {
		v.Notes = *req.Notes
	}

	if err := s.repo.UpdateViolation(ctx, v); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindViolationByID(ctx, id)
	if err == nil && updated != nil && paymentAdded > 0 {
		s.sendViolationNotification(ctx, updated, true, paymentAdded)
	}
	return updated, err
}

func (s *hrLegalService) DeleteViolation(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteViolation(ctx, id)
}

func (s *hrLegalService) GetViolationByID(ctx context.Context, id uuid.UUID) (*TrafficViolation, error) {
	return s.repo.FindViolationByID(ctx, id)
}

func (s *hrLegalService) GetAllViolations(ctx context.Context, filter TrafficViolationFilter, adminBranchID *uuid.UUID) ([]TrafficViolation, int64, float64, float64, error) {
	if filter.EmployeeID != nil {
		// When querying for an employee, do not restrict by adminBranchID so courier sees all their records
		filter.BranchID = nil
	} else if adminBranchID != nil {
		filter.BranchID = adminBranchID
	}
	return s.repo.FindViolations(ctx, filter)
}
