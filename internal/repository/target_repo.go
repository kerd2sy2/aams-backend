package repository

import (
	"context"
	"strings"
	"time"

	"delivery-backend/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TargetRepository interface {
	// Identifiers
	FindIdentifierByID(ctx context.Context, id uuid.UUID) (*domain.Identifier, error)
	FindIdentifierByName(ctx context.Context, name string) (*domain.Identifier, error)
	FindIdentifierByNameAndApp(ctx context.Context, name, appName string) (*domain.Identifier, error)
	FindOrCreateIdentifier(ctx context.Context, name, appName string) (*domain.Identifier, error)
	ListIdentifiers(ctx context.Context, search string, isActive *bool) ([]domain.Identifier, error)
	CreateIdentifier(ctx context.Context, ident *domain.Identifier) error
	UpdateIdentifier(ctx context.Context, ident *domain.Identifier) error
	DeleteIdentifier(ctx context.Context, id uuid.UUID) error
	DeleteAllIdentifiers(ctx context.Context) error

	// Drivers
	FindDriverByID(ctx context.Context, id uuid.UUID) (*domain.Driver, error)
	FindDriverByName(ctx context.Context, name string) (*domain.Driver, error)
	FindOrCreateDriver(ctx context.Context, name string) (*domain.Driver, error)
	ListDrivers(ctx context.Context, search string) ([]domain.Driver, error)

	// Identifier-Driver Links
	LinkDriverToIdentifier(ctx context.Context, identifierID, driverID uuid.UUID, orderDate string) error
	GetDriversForIdentifier(ctx context.Context, identifierID uuid.UUID) ([]domain.Driver, error)

	// Daily Orders
	CreateDailyOrdersBatch(ctx context.Context, orders []domain.DailyOrder) error
	DeleteOrdersByDateAndApp(ctx context.Context, orderDate, appName string, identifierID *uuid.UUID, driverID uuid.UUID) error
	CheckDuplicates(ctx context.Context, orderDate string, identifierID *uuid.UUID, driverID uuid.UUID, appName string) (bool, error)
	GetDailyOrders(ctx context.Context, orderDate string, identifierID *uuid.UUID) ([]domain.DailyOrder, error)
	GetOrdersForMonth(ctx context.Context, monthPrefix string) ([]domain.DailyOrder, error)
	GetOrdersForIdentifierMonth(ctx context.Context, identifierID uuid.UUID, monthPrefix string) ([]domain.DailyOrder, error)
	GetOrdersSummaryByDateRange(ctx context.Context, startDate, endDate string) (int, error)

	// Batches
	CreateImportBatch(ctx context.Context, batch *domain.ImportBatch) error
	UpdateImportBatch(ctx context.Context, batch *domain.ImportBatch) error
	ListImportBatches(ctx context.Context, limit int) ([]domain.ImportBatch, error)
	FindImportBatchByID(ctx context.Context, id uuid.UUID) (*domain.ImportBatch, error)
	DeleteImportBatch(ctx context.Context, id uuid.UUID) error
	DeleteOrdersByDate(ctx context.Context, orderDate string) error

	// Alerts
	CreateTargetAlert(ctx context.Context, alert *domain.TargetAlert) error
	ListTargetAlerts(ctx context.Context, alertDate string, unresolvedOnly bool) ([]domain.TargetAlert, error)
	ResolveAlert(ctx context.Context, id uuid.UUID) error
	ResolveAllAlerts(ctx context.Context, branch string, alertDate string) error

	// Settings
	GetTargetSetting(ctx context.Context, key string) (string, error)
	SetTargetSetting(ctx context.Context, key, val string) error
	UpdateAllIdentifiersTargets(ctx context.Context, monthlyTarget, dailyTarget int) error
}

type gormTargetRepository struct {
	db *gorm.DB
}

func NewTargetRepository(db *gorm.DB) TargetRepository {
	return &gormTargetRepository{db: db}
}

func (r *gormTargetRepository) FindIdentifierByID(ctx context.Context, id uuid.UUID) (*domain.Identifier, error) {
	var ident domain.Identifier
	if err := r.db.WithContext(ctx).Preload("Drivers").First(&ident, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &ident, nil
}

func (r *gormTargetRepository) FindIdentifierByName(ctx context.Context, name string) (*domain.Identifier, error) {
	var ident domain.Identifier
	if err := r.db.WithContext(ctx).Where("TRIM(LOWER(name)) = TRIM(LOWER(?))", name).First(&ident).Error; err != nil {
		return nil, err
	}
	return &ident, nil
}

func (r *gormTargetRepository) FindIdentifierByNameAndApp(ctx context.Context, name, appName string) (*domain.Identifier, error) {
	var ident domain.Identifier
	name = strings.TrimSpace(name)
	appName = strings.TrimSpace(appName)
	q := r.db.WithContext(ctx).Where("TRIM(LOWER(name)) = TRIM(LOWER(?))", name)
	if appName != "" {
		q = q.Where("TRIM(LOWER(COALESCE(app_name, ''))) = TRIM(LOWER(?))", appName)
	} else {
		q = q.Where("app_name = '' OR app_name IS NULL")
	}
	if err := q.First(&ident).Error; err != nil {
		return nil, err
	}
	return &ident, nil
}

func (r *gormTargetRepository) FindOrCreateIdentifier(ctx context.Context, name, appName string) (*domain.Identifier, error) {
	name = strings.TrimSpace(name)
	appName = strings.TrimSpace(appName)
	ident, err := r.FindIdentifierByNameAndApp(ctx, name, appName)
	if err == nil && ident != nil {
		return ident, nil
	}

	newIdent := domain.Identifier{
		Name:          name,
		AppName:       appName,
		MonthlyTarget: 460,
		DailyTarget:   18,
		IsActive:      true,
	}
	if err := r.db.WithContext(ctx).Create(&newIdent).Error; err != nil {
		// Double check concurrency race
		return r.FindIdentifierByNameAndApp(ctx, name, appName)
	}
	return &newIdent, nil
}

func (r *gormTargetRepository) ListIdentifiers(ctx context.Context, search string, isActive *bool) ([]domain.Identifier, error) {
	var list []domain.Identifier
	q := r.db.WithContext(ctx).Model(&domain.Identifier{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("name ILIKE ? OR code ILIKE ? OR app_name ILIKE ?", like, like, like)
	}
	if isActive != nil {
		q = q.Where("is_active = ?", *isActive)
	}
	if err := q.Order("name ASC, app_name ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormTargetRepository) CreateIdentifier(ctx context.Context, ident *domain.Identifier) error {
	return r.db.WithContext(ctx).Create(ident).Error
}

func (r *gormTargetRepository) UpdateIdentifier(ctx context.Context, ident *domain.Identifier) error {
	return r.db.WithContext(ctx).Save(ident).Error
}

func (r *gormTargetRepository) DeleteIdentifier(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Identifier{}, "id = ?", id).Error
}

func (r *gormTargetRepository) DeleteAllIdentifiers(ctx context.Context) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		_ = tx.Exec("DELETE FROM target_alerts").Error
		_ = tx.Exec("DELETE FROM daily_orders").Error
		_ = tx.Exec("DELETE FROM identifier_drivers").Error
		_ = tx.Exec("DELETE FROM import_batches").Error
		return tx.Unscoped().Where("1 = 1").Delete(&domain.Identifier{}).Error
	})
}

func (r *gormTargetRepository) FindDriverByID(ctx context.Context, id uuid.UUID) (*domain.Driver, error) {
	var d domain.Driver
	if err := r.db.WithContext(ctx).First(&d, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *gormTargetRepository) FindDriverByName(ctx context.Context, name string) (*domain.Driver, error) {
	var d domain.Driver
	if err := r.db.WithContext(ctx).Where("TRIM(LOWER(name)) = TRIM(LOWER(?))", name).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *gormTargetRepository) FindOrCreateDriver(ctx context.Context, name string) (*domain.Driver, error) {
	d, err := r.FindDriverByName(ctx, name)
	if err == nil && d != nil {
		return d, nil
	}

	newDriver := domain.Driver{
		Name:     name,
		IsActive: true,
	}
	if err := r.db.WithContext(ctx).Create(&newDriver).Error; err != nil {
		return r.FindDriverByName(ctx, name)
	}
	return &newDriver, nil
}

func (r *gormTargetRepository) ListDrivers(ctx context.Context, search string) ([]domain.Driver, error) {
	var list []domain.Driver
	q := r.db.WithContext(ctx).Model(&domain.Driver{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("name ILIKE ?", like)
	}
	if err := q.Order("name ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *gormTargetRepository) LinkDriverToIdentifier(ctx context.Context, identifierID, driverID uuid.UUID, orderDate string) error {
	var link domain.IdentifierDriver
	err := r.db.WithContext(ctx).Where("identifier_id = ? AND driver_id = ?", identifierID, driverID).First(&link).Error
	if err == nil {
		// Update last seen date
		if orderDate > link.LastSeenDate {
			link.LastSeenDate = orderDate
			link.IsActive = true
			return r.db.WithContext(ctx).Save(&link).Error
		}
		return nil
	}

	newLink := domain.IdentifierDriver{
		IdentifierID:  identifierID,
		DriverID:      driverID,
		IsActive:      true,
		FirstSeenDate: orderDate,
		LastSeenDate:  orderDate,
	}
	return r.db.WithContext(ctx).Create(&newLink).Error
}

func (r *gormTargetRepository) GetDriversForIdentifier(ctx context.Context, identifierID uuid.UUID) ([]domain.Driver, error) {
	var drivers []domain.Driver
	err := r.db.WithContext(ctx).
		Table("drivers").
		Joins("JOIN identifier_drivers ON identifier_drivers.driver_id = drivers.id").
		Where("identifier_drivers.identifier_id = ?", identifierID).
		Find(&drivers).Error
	return drivers, err
}

func (r *gormTargetRepository) CreateDailyOrdersBatch(ctx context.Context, orders []domain.DailyOrder) error {
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
	return q.Delete(&domain.DailyOrder{}).Error
}

func (r *gormTargetRepository) CheckDuplicates(ctx context.Context, orderDate string, identifierID *uuid.UUID, driverID uuid.UUID, appName string) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&domain.DailyOrder{}).
		Where("order_date = ? AND driver_id = ? AND app_name = ?", orderDate, driverID, appName)
	if identifierID != nil {
		q = q.Where("identifier_id = ?", *identifierID)
	} else {
		q = q.Where("identifier_id IS NULL")
	}
	err := q.Count(&count).Error
	return count > 0, err
}

func (r *gormTargetRepository) GetDailyOrders(ctx context.Context, orderDate string, identifierID *uuid.UUID) ([]domain.DailyOrder, error) {
	var orders []domain.DailyOrder
	q := r.db.WithContext(ctx).Preload("Identifier").Preload("Driver").Where("order_date = ?", orderDate)
	if identifierID != nil {
		q = q.Where("identifier_id = ?", *identifierID)
	}
	err := q.Order("orders_count DESC").Find(&orders).Error
	return orders, err
}

func (r *gormTargetRepository) GetOrdersForMonth(ctx context.Context, monthPrefix string) ([]domain.DailyOrder, error) {
	var orders []domain.DailyOrder
	// monthPrefix format: "YYYY-MM"
	err := r.db.WithContext(ctx).Preload("Identifier").Preload("Driver").
		Where("order_date LIKE ?", monthPrefix+"%").
		Order("order_date ASC, created_at ASC, id ASC").Find(&orders).Error
	return orders, err
}

func (r *gormTargetRepository) GetOrdersForIdentifierMonth(ctx context.Context, identifierID uuid.UUID, monthPrefix string) ([]domain.DailyOrder, error) {
	var orders []domain.DailyOrder
	err := r.db.WithContext(ctx).Preload("Driver").
		Where("identifier_id = ? AND order_date LIKE ?", identifierID, monthPrefix+"%").
		Order("order_date ASC").Find(&orders).Error
	return orders, err
}

func (r *gormTargetRepository) GetOrdersSummaryByDateRange(ctx context.Context, startDate, endDate string) (int, error) {
	var total int
	row := r.db.WithContext(ctx).Model(&domain.DailyOrder{}).
		Select("COALESCE(SUM(orders_count), 0)").
		Where("order_date >= ? AND order_date <= ?", startDate, endDate).
		Row()
	err := row.Scan(&total)
	return total, err
}

func (r *gormTargetRepository) CreateImportBatch(ctx context.Context, batch *domain.ImportBatch) error {
	return r.db.WithContext(ctx).Create(batch).Error
}

func (r *gormTargetRepository) UpdateImportBatch(ctx context.Context, batch *domain.ImportBatch) error {
	return r.db.WithContext(ctx).Save(batch).Error
}

func (r *gormTargetRepository) ListImportBatches(ctx context.Context, limit int) ([]domain.ImportBatch, error) {
	var list []domain.ImportBatch
	if limit <= 0 {
		limit = 30
	}
	err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&list).Error
	return list, err
}

func (r *gormTargetRepository) FindImportBatchByID(ctx context.Context, id uuid.UUID) (*domain.ImportBatch, error) {
	var b domain.ImportBatch
	if err := r.db.WithContext(ctx).First(&b, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *gormTargetRepository) DeleteImportBatch(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var b domain.ImportBatch
		if err := tx.First(&b, "id = ?", id).Error; err == nil && b.OrderDate != "" {
			_ = tx.Where("alert_date = ?", b.OrderDate).Delete(&domain.TargetAlert{}).Error
		}
		if err := tx.Where("import_batch_id = ?", id).Delete(&domain.DailyOrder{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.ImportBatch{}, "id = ?", id).Error
	})
}

func (r *gormTargetRepository) DeleteOrdersByDate(ctx context.Context, orderDate string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_date = ?", orderDate).Delete(&domain.DailyOrder{}).Error; err != nil {
			return err
		}
		_ = tx.Where("alert_date = ?", orderDate).Delete(&domain.TargetAlert{}).Error
		return tx.Where("order_date = ?", orderDate).Delete(&domain.ImportBatch{}).Error
	})
}

func (r *gormTargetRepository) CreateTargetAlert(ctx context.Context, alert *domain.TargetAlert) error {
	// Don't create duplicate alerts for same identifier and date
	var count int64
	r.db.WithContext(ctx).Model(&domain.TargetAlert{}).
		Where("identifier_id = ? AND alert_date = ?", alert.IdentifierID, alert.AlertDate).
		Count(&count)
	if count > 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *gormTargetRepository) ListTargetAlerts(ctx context.Context, alertDate string, unresolvedOnly bool) ([]domain.TargetAlert, error) {
	var alerts []domain.TargetAlert
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
	return r.db.WithContext(ctx).Model(&domain.TargetAlert{}).Where("id = ?", id).Update("is_resolved", true).Error
}

func (r *gormTargetRepository) ResolveAllAlerts(ctx context.Context, branch string, alertDate string) error {
	branch = strings.TrimSpace(branch)
	if branch != "" && branch != "all" && branch != "الكل" {
		subQuery := r.db.WithContext(ctx).Model(&domain.Identifier{}).Select("id").Where("branch = ?", branch)
		q := r.db.WithContext(ctx).Model(&domain.TargetAlert{}).Where("is_resolved = false").Where("identifier_id IN (?)", subQuery)
		if alertDate != "" {
			q = q.Where("alert_date = ?", alertDate)
		}
		return q.Update("is_resolved", true).Error
	}

	q := r.db.WithContext(ctx).Model(&domain.TargetAlert{}).Where("is_resolved = false")
	if alertDate != "" {
		q = q.Where("alert_date = ?", alertDate)
	}
	return q.Update("is_resolved", true).Error
}

func (r *gormTargetRepository) GetTargetSetting(ctx context.Context, key string) (string, error) {
	var s domain.TargetSetting
	err := r.db.WithContext(ctx).Where("setting_key = ?", key).First(&s).Error
	if err != nil {
		return "", err
	}
	return s.SettingValue, nil
}

func (r *gormTargetRepository) SetTargetSetting(ctx context.Context, key, val string) error {
	var s domain.TargetSetting
	if err := r.db.WithContext(ctx).Where("setting_key = ?", key).First(&s).Error; err == nil {
		s.SettingValue = val
		s.UpdatedAt = time.Now()
		return r.db.WithContext(ctx).Save(&s).Error
	}
	newSetting := domain.TargetSetting{
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
	return r.db.WithContext(ctx).Model(&domain.Identifier{}).Where("is_active = ?", true).Updates(updates).Error
}
