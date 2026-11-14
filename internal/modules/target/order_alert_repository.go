package target

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *gormTargetRepository) CreateDailyOrdersBatch(ctx context.Context, orders []DailyOrder) error {
	if len(orders) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(orders, 100).Error
}

func (r *gormTargetRepository) DeleteOrdersByDateAndApp(ctx context.Context, orderDate, appName string, identifierID *uuid.UUID, driverID uuid.UUID) error {
	q := r.db.WithContext(ctx).Where("order_date = ? AND driver_id = ? AND app_name = ?", orderDate, driverID, appName)
	if identifierID != nil {
		q = q.Where("identifier_id = ?", *identifierID)
	} else {
		q = q.Where("identifier_id IS NULL")
	}
	return q.Delete(&DailyOrder{}).Error
}

func (r *gormTargetRepository) CheckDuplicates(ctx context.Context, orderDate string, identifierID *uuid.UUID, driverID uuid.UUID, appName string) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&DailyOrder{}).
		Where("order_date = ? AND driver_id = ? AND app_name = ?", orderDate, driverID, appName)
	if identifierID != nil {
		q = q.Where("identifier_id = ?", *identifierID)
	} else {
		q = q.Where("identifier_id IS NULL")
	}
	err := q.Count(&count).Error
	return count > 0, err
}

func (r *gormTargetRepository) GetDailyOrders(ctx context.Context, orderDate string, identifierID *uuid.UUID) ([]DailyOrder, error) {
	var orders []DailyOrder
	q := r.db.WithContext(ctx).Preload("Identifier").Preload("Driver").Where("order_date = ?", orderDate)
	if identifierID != nil {
		q = q.Where("identifier_id = ?", *identifierID)
	}
	err := q.Order("orders_count DESC").Find(&orders).Error
	return orders, err
}

func (r *gormTargetRepository) GetOrdersForMonth(ctx context.Context, monthPrefix string) ([]DailyOrder, error) {
	var orders []DailyOrder
	err := r.db.WithContext(ctx).Preload("Identifier").Preload("Driver").
		Where("order_date LIKE ?", monthPrefix+"%").
		Order("order_date ASC, created_at ASC, id ASC").Find(&orders).Error
	return orders, err
}

func (r *gormTargetRepository) GetOrdersForIdentifierMonth(ctx context.Context, identifierID uuid.UUID, monthPrefix string) ([]DailyOrder, error) {
	var orders []DailyOrder
	err := r.db.WithContext(ctx).Preload("Driver").
		Where("identifier_id = ? AND order_date LIKE ?", identifierID, monthPrefix+"%").
		Order("order_date ASC").Find(&orders).Error
	return orders, err
}

func (r *gormTargetRepository) GetOrdersSummaryByDateRange(ctx context.Context, startDate, endDate string) (int, error) {
	var total int
	row := r.db.WithContext(ctx).Model(&DailyOrder{}).
		Select("COALESCE(SUM(orders_count), 0)").
		Where("order_date >= ? AND order_date <= ?", startDate, endDate).
		Row()
	err := row.Scan(&total)
	return total, err
}

func (r *gormTargetRepository) CreateImportBatch(ctx context.Context, batch *ImportBatch) error {
	return r.db.WithContext(ctx).Create(batch).Error
}

func (r *gormTargetRepository) UpdateImportBatch(ctx context.Context, batch *ImportBatch) error {
	return r.db.WithContext(ctx).Save(batch).Error
}

func (r *gormTargetRepository) ListImportBatches(ctx context.Context, limit int) ([]ImportBatch, error) {
	var list []ImportBatch
	if limit <= 0 {
		limit = 30
	}
	err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&list).Error
	return list, err
}

func (r *gormTargetRepository) FindImportBatchByID(ctx context.Context, id uuid.UUID) (*ImportBatch, error) {
	var b ImportBatch
	if err := r.db.WithContext(ctx).First(&b, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormTargetRepository) DeleteImportBatch(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var b ImportBatch
		if err := tx.First(&b, "id = ?", id).Error; err == nil && b.OrderDate != "" {
			_ = tx.Where("alert_date = ?", b.OrderDate).Delete(&TargetAlert{}).Error
		}
		if err := tx.Where("import_batch_id = ?", id).Delete(&DailyOrder{}).Error; err != nil {
			return err
		}
		return tx.Delete(&ImportBatch{}, "id = ?", id).Error
	})
}

func (r *gormTargetRepository) DeleteOrdersByDate(ctx context.Context, orderDate string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_date = ?", orderDate).Delete(&DailyOrder{}).Error; err != nil {
			return err
		}
		_ = tx.Where("alert_date = ?", orderDate).Delete(&TargetAlert{}).Error
		return tx.Where("order_date = ?", orderDate).Delete(&ImportBatch{}).Error
	})
}

func (r *gormTargetRepository) CreateTargetAlert(ctx context.Context, alert *TargetAlert) error {
	var count int64
	r.db.WithContext(ctx).Model(&TargetAlert{}).
		Where("identifier_id = ? AND alert_date = ?", alert.IdentifierID, alert.AlertDate).
		Count(&count)
	if count > 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *gormTargetRepository) ListTargetAlerts(ctx context.Context, alertDate string, unresolvedOnly bool) ([]TargetAlert, error) {
	var alerts []TargetAlert
	q := r.db.WithContext(ctx).Preload("Identifier")
	if alertDate != "" {
		q = q.Where("alert_date = ?", alertDate)
	}
	if unresolvedOnly {
		q = q.Where("is_resolved = false")
	}
	err := q.Order("alert_date DESC, deficit DESC").Find(&alerts).Error
	return alerts, err
}

func (r *gormTargetRepository) ResolveAlert(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&TargetAlert{}).Where("id = ?", id).Update("is_resolved", true).Error
}

func (r *gormTargetRepository) ResolveAllAlerts(ctx context.Context, branch string, alertDate string) error {
	branch = strings.TrimSpace(branch)
	if branch != "" && branch != "all" && branch != "الكل" {
		subQuery := r.db.WithContext(ctx).Model(&Identifier{}).Select("id").Where("branch = ?", branch)
		q := r.db.WithContext(ctx).Model(&TargetAlert{}).Where("is_resolved = false").Where("identifier_id IN (?)", subQuery)
		if alertDate != "" {
			q = q.Where("alert_date = ?", alertDate)
		}
		return q.Update("is_resolved", true).Error
	}

	q := r.db.WithContext(ctx).Model(&TargetAlert{}).Where("is_resolved = false")
	if alertDate != "" {
		q = q.Where("alert_date = ?", alertDate)
	}
	return q.Update("is_resolved", true).Error
}

func (r *gormTargetRepository) GetTargetSetting(ctx context.Context, key string) (string, error) {
	var s TargetSetting
	err := r.db.WithContext(ctx).Where("setting_key = ?", key).First(&s).Error
	if err != nil {
		return "", err
	}
	return s.SettingValue, nil
}

func (r *gormTargetRepository) SetTargetSetting(ctx context.Context, key, val string) error {
	var s TargetSetting
	if err := r.db.WithContext(ctx).Where("setting_key = ?", key).First(&s).Error; err == nil {
		s.SettingValue = val
		s.UpdatedAt = time.Now()
		return r.db.WithContext(ctx).Save(&s).Error
	}
	newSetting := TargetSetting{
		SettingKey:   key,
		SettingValue: val,
		UpdatedAt:    time.Now(),
	}
	return r.db.WithContext(ctx).Create(&newSetting).Error
}

func (r *gormTargetRepository) UpdateAllIdentifiersTargets(ctx context.Context, monthlyTarget, dailyTarget int) error {
	updates := map[string]interface{}{}
	if monthlyTarget > 0 {
		updates["monthly_target"] = monthlyTarget
	}
	if dailyTarget > 0 {
		updates["daily_target"] = dailyTarget
	}
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(&Identifier{}).Where("is_active = ?", true).Updates(updates).Error
}
