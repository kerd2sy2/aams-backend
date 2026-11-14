package target

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func (s *targetService) ListAlerts(ctx context.Context, date string, unresolvedOnly bool, branch string) ([]TargetAlertDTO, error) {
	alerts, err := s.targetRepo.ListTargetAlerts(ctx, date, unresolvedOnly)
	if err != nil {
		return nil, err
	}

	branch = strings.TrimSpace(branch)

	result := make([]TargetAlertDTO, 0)
	for _, a := range alerts {
		if branch != "" && branch != "all" && branch != "الكل" {
			if a.Identifier != nil && a.Identifier.Branch != "" && a.Identifier.Branch != branch {
				continue
			}
		}

		identName := "معرف"
		if a.Identifier != nil {
			identName = a.Identifier.Name
		}
		result = append(result, TargetAlertDTO{
			ID:             a.ID,
			IdentifierID:   a.IdentifierID,
			IdentifierName: identName,
			AlertDate:      a.AlertDate,
			TargetOrders:   a.TargetOrders,
			ActualOrders:   a.ActualOrders,
			Deficit:        a.Deficit,
			IsResolved:     a.IsResolved,
			CreatedAt:      a.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	return result, nil
}

func (s *targetService) ResolveAlert(ctx context.Context, id uuid.UUID) error {
	return s.targetRepo.ResolveAlert(ctx, id)
}

func (s *targetService) ResolveAllAlerts(ctx context.Context, branch string, date string) error {
	return s.targetRepo.ResolveAllAlerts(ctx, branch, date)
}

func (s *targetService) GetTargetSettings(ctx context.Context) (*TargetSettingsDTO, error) {
	mVal, _ := s.targetRepo.GetTargetSetting(ctx, "DEFAULT_MONTHLY_TARGET")
	dVal, _ := s.targetRepo.GetTargetSetting(ctx, "DEFAULT_DAILY_TARGET")

	mTarget, _ := strconv.Atoi(mVal)
	if mTarget <= 0 {
		mTarget = 460
	}
	dTarget, _ := strconv.Atoi(dVal)
	if dTarget <= 0 {
		dTarget = 18
	}

	return &TargetSettingsDTO{
		DefaultMonthlyTarget: mTarget,
		DefaultDailyTarget:   dTarget,
	}, nil
}

func (s *targetService) UpdateTargetSettings(ctx context.Context, monthlyTarget, dailyTarget int) error {
	if monthlyTarget > 0 {
		_ = s.targetRepo.SetTargetSetting(ctx, "DEFAULT_MONTHLY_TARGET", strconv.Itoa(monthlyTarget))
	}
	if dailyTarget > 0 {
		_ = s.targetRepo.SetTargetSetting(ctx, "DEFAULT_DAILY_TARGET", strconv.Itoa(dailyTarget))
	}
	if monthlyTarget > 0 || dailyTarget > 0 {
		_ = s.targetRepo.UpdateAllIdentifiersTargets(ctx, monthlyTarget, dailyTarget)
	}
	return nil
}
