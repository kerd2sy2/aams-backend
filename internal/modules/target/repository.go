package target

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TargetRepository interface {
	// Identifiers
	FindIdentifierByID(ctx context.Context, id uuid.UUID) (*Identifier, error)
	FindIdentifierByName(ctx context.Context, name string) (*Identifier, error)
	FindIdentifierByNameAndApp(ctx context.Context, name, appName string) (*Identifier, error)
	FindOrCreateIdentifier(ctx context.Context, name, appName string) (*Identifier, error)
	ListIdentifiers(ctx context.Context, search string, isActive *bool) ([]Identifier, error)
	CreateIdentifier(ctx context.Context, ident *Identifier) error
	UpdateIdentifier(ctx context.Context, ident *Identifier) error
	FindIdentifierByIDOrCode(ctx context.Context, idOrCode string) (*Identifier, error)
	LinkEmployeeToIdentifierCode(ctx context.Context, idOrCode string, employeeID *string) error
	DeleteIdentifier(ctx context.Context, id uuid.UUID) error
	DeleteAllIdentifiers(ctx context.Context) error

	// Drivers
	FindDriverByID(ctx context.Context, id uuid.UUID) (*Driver, error)
	FindDriverByName(ctx context.Context, name string) (*Driver, error)
	FindOrCreateDriver(ctx context.Context, name string) (*Driver, error)
	ListDrivers(ctx context.Context, search string) ([]Driver, error)

	// Identifier-Driver Links
	LinkDriverToIdentifier(ctx context.Context, identifierID, driverID uuid.UUID, orderDate string) error
	GetDriversForIdentifier(ctx context.Context, identifierID uuid.UUID) ([]Driver, error)

	// Daily Orders
	CreateDailyOrdersBatch(ctx context.Context, orders []DailyOrder) error
	DeleteOrdersByDateAndApp(ctx context.Context, orderDate, appName string, identifierID *uuid.UUID, driverID uuid.UUID) error
	CheckDuplicates(ctx context.Context, orderDate string, identifierID *uuid.UUID, driverID uuid.UUID, appName string) (bool, error)
	GetDailyOrders(ctx context.Context, orderDate string, identifierID *uuid.UUID) ([]DailyOrder, error)
	GetOrdersForMonth(ctx context.Context, monthPrefix string) ([]DailyOrder, error)
	GetOrdersForIdentifierMonth(ctx context.Context, identifierID uuid.UUID, monthPrefix string) ([]DailyOrder, error)
	GetOrdersSummaryByDateRange(ctx context.Context, startDate, endDate string) (int, error)

	// Batches
	CreateImportBatch(ctx context.Context, batch *ImportBatch) error
	UpdateImportBatch(ctx context.Context, batch *ImportBatch) error
	ListImportBatches(ctx context.Context, limit int) ([]ImportBatch, error)
	FindImportBatchByID(ctx context.Context, id uuid.UUID) (*ImportBatch, error)
	DeleteImportBatch(ctx context.Context, id uuid.UUID) error
	DeleteOrdersByDate(ctx context.Context, orderDate string) error

	// Alerts
	CreateTargetAlert(ctx context.Context, alert *TargetAlert) error
	ListTargetAlerts(ctx context.Context, alertDate string, unresolvedOnly bool) ([]TargetAlert, error)
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

func (r *gormTargetRepository) FindIdentifierByID(ctx context.Context, id uuid.UUID) (*Identifier, error) {
	var ident Identifier
	if err := r.db.WithContext(ctx).Preload("Drivers").First(&ident, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &ident, nil
}

func (r *gormTargetRepository) FindIdentifierByName(ctx context.Context, name string) (*Identifier, error) {
	var ident Identifier
	if err := r.db.WithContext(ctx).Where("TRIM(LOWER(name)) = TRIM(LOWER(?))", name).First(&ident).Error; err != nil {
		return nil, err
	}
	return &ident, nil
}

func (r *gormTargetRepository) FindIdentifierByNameAndApp(ctx context.Context, name, appName string) (*Identifier, error) {
	var ident Identifier
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

func (r *gormTargetRepository) FindOrCreateIdentifier(ctx context.Context, name, appName string) (*Identifier, error) {
	name = strings.TrimSpace(name)
	appName = strings.TrimSpace(appName)
	ident, err := r.FindIdentifierByNameAndApp(ctx, name, appName)
	if err == nil && ident != nil {
		return ident, nil
	}

	newIdent := Identifier{
		Name:          name,
		AppName:       appName,
		MonthlyTarget: 460,
		DailyTarget:   18,
		IsActive:      true,
	}
	if err := r.db.WithContext(ctx).Create(&newIdent).Error; err != nil {
		return r.FindIdentifierByNameAndApp(ctx, name, appName)
	}
	return &newIdent, nil
}

func (r *gormTargetRepository) FindIdentifierByIDOrCode(ctx context.Context, idOrCode string) (*Identifier, error) {
	clean := strings.TrimSpace(idOrCode)
	if clean == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var ident Identifier
	if parsed, err := uuid.Parse(clean); err == nil {
		if err := r.db.WithContext(ctx).Preload("Drivers").First(&ident, "id = ?", parsed).Error; err == nil {
			return &ident, nil
		}
	}

	stripped := strings.TrimPrefix(strings.TrimPrefix(strings.ToLower(clean), "ninja_"), "keeta_")
	err := r.db.WithContext(ctx).Preload("Drivers").
		Where("LOWER(TRIM(code)) = ? OR LOWER(TRIM(code)) = ? OR LOWER(TRIM(name)) = ? OR LOWER(TRIM(name)) = ?", strings.ToLower(clean), stripped, strings.ToLower(clean), stripped).
		First(&ident).Error
	if err == nil {
		return &ident, nil
	}

	appName := "NINJA"
	if strings.Contains(strings.ToLower(clean), "keeta") {
		appName = "KEETA"
	}
	newIdent := Identifier{
		ID:            uuid.New(),
		Name:          clean,
		AppName:       appName,
		Code:          stripped,
		MonthlyTarget: 460,
		DailyTarget:   18,
		IsActive:      true,
		AccountStatus: "ACTIVE",
	}
	if err := r.db.WithContext(ctx).Create(&newIdent).Error; err != nil {
		return nil, err
	}
	return &newIdent, nil
}

func (r *gormTargetRepository) LinkEmployeeToIdentifierCode(ctx context.Context, idOrCode string, employeeID *string) error {
	ident, err := r.FindIdentifierByIDOrCode(ctx, idOrCode)
	if err != nil || ident == nil {
		return err
	}

	if employeeID == nil || strings.TrimSpace(*employeeID) == "" {
		return nil
	}

	empUUID, err := uuid.Parse(strings.TrimSpace(*employeeID))
	if err != nil {
		return nil
	}

	codeToSet := ident.Code
	if codeToSet == "" {
		codeToSet = ident.Name
	}
	_ = r.db.WithContext(ctx).Table("employees").
		Where("id = ? AND deleted_at IS NULL", empUUID).
		Updates(map[string]interface{}{
			"application_id":   codeToSet,
			"application_type": ident.AppName,
		}).Error

	var emp struct {
		Name string `gorm:"column:name"`
	}
	if err := r.db.WithContext(ctx).Table("employees").Select("name").Where("id = ?", empUUID).Scan(&emp).Error; err == nil && emp.Name != "" {
		driver, _ := r.FindOrCreateDriver(ctx, emp.Name)
		if driver != nil {
			driver.EmployeeID = &empUUID
			_ = r.db.WithContext(ctx).Save(driver).Error
			_ = r.LinkDriverToIdentifier(ctx, ident.ID, driver.ID, time.Now().Format("2006-01-02"))
		}
	}

	return nil
}

func (r *gormTargetRepository) ListIdentifiers(ctx context.Context, search string, isActive *bool) ([]Identifier, error) {
	var list []Identifier
	q := r.db.WithContext(ctx).Model(&Identifier{}).Preload("Drivers")
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

func (r *gormTargetRepository) CreateIdentifier(ctx context.Context, ident *Identifier) error {
	return r.db.WithContext(ctx).Omit("Drivers", "DailyOrders").Create(ident).Error
}

func (r *gormTargetRepository) UpdateIdentifier(ctx context.Context, ident *Identifier) error {
	return r.db.WithContext(ctx).Model(&Identifier{}).
		Where("id = ?", ident.ID).
		Updates(map[string]interface{}{
			"name":           ident.Name,
			"app_name":       ident.AppName,
			"code":           ident.Code,
			"monthly_target": ident.MonthlyTarget,
			"daily_target":   ident.DailyTarget,
			"is_active":      ident.IsActive,
			"account_status": ident.AccountStatus,
			"updated_at":     time.Now(),
		}).Error
}

func (r *gormTargetRepository) DeleteIdentifier(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Identifier{}, "id = ?", id).Error
}

func (r *gormTargetRepository) DeleteAllIdentifiers(ctx context.Context) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		_ = tx.Exec("DELETE FROM target_alerts").Error
		_ = tx.Exec("DELETE FROM daily_orders").Error
		_ = tx.Exec("DELETE FROM identifier_drivers").Error
		_ = tx.Exec("DELETE FROM import_batches").Error
		_ = tx.Unscoped().Where("1 = 1").Delete(&Identifier{}).Error
		return tx.Unscoped().Where("1 = 1").Delete(&Driver{}).Error
	})
}

func (r *gormTargetRepository) FindDriverByID(ctx context.Context, id uuid.UUID) (*Driver, error) {
	var d Driver
	if err := r.db.WithContext(ctx).First(&d, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *gormTargetRepository) FindDriverByName(ctx context.Context, name string) (*Driver, error) {
	var d Driver
	if err := r.db.WithContext(ctx).Where("TRIM(LOWER(name)) = TRIM(LOWER(?))", name).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *gormTargetRepository) FindOrCreateDriver(ctx context.Context, name string) (*Driver, error) {
	d, err := r.FindDriverByName(ctx, name)
	if err == nil && d != nil {
		return d, nil
	}

	newDriver := Driver{
		Name:     name,
		IsActive: true,
	}
	if err := r.db.WithContext(ctx).Create(&newDriver).Error; err != nil {
		return r.FindDriverByName(ctx, name)
	}
	return &newDriver, nil
}

func (r *gormTargetRepository) ListDrivers(ctx context.Context, search string) ([]Driver, error) {
	var list []Driver
	q := r.db.WithContext(ctx).Model(&Driver{})
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
	var link IdentifierDriver
	err := r.db.WithContext(ctx).Where("identifier_id = ? AND driver_id = ?", identifierID, driverID).First(&link).Error
	if err == nil {
		if orderDate > link.LastSeenDate {
			link.LastSeenDate = orderDate
			link.IsActive = true
			return r.db.WithContext(ctx).Save(&link).Error
		}
		return nil
	}

	newLink := IdentifierDriver{
		IdentifierID:  identifierID,
		DriverID:      driverID,
		IsActive:      true,
		FirstSeenDate: orderDate,
		LastSeenDate:  orderDate,
	}
	return r.db.WithContext(ctx).Create(&newLink).Error
}

func (r *gormTargetRepository) GetDriversForIdentifier(ctx context.Context, identifierID uuid.UUID) ([]Driver, error) {
	var drivers []Driver
	err := r.db.WithContext(ctx).
		Table("drivers").
		Joins("JOIN identifier_drivers ON identifier_drivers.driver_id = drivers.id").
		Where("identifier_drivers.identifier_id = ?", identifierID).
		Find(&drivers).Error
	return drivers, err
}
