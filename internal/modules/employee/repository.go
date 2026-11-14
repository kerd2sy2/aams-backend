package employee

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeRepository interface {
	Create(ctx context.Context, emp *Employee) error
	Update(ctx context.Context, emp *Employee) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Employee, error)
	FindByNationalID(ctx context.Context, nationalID string) (*Employee, error)
	FindAll(ctx context.Context, filter EmployeeFilter) ([]Employee, int64, error)
	Search(ctx context.Context, query string, branchID *uuid.UUID) ([]Employee, error)
	GetWorkingEmployees(ctx context.Context, branchID *uuid.UUID) ([]Employee, error)
	GetLocations(ctx context.Context, branchID *uuid.UUID) ([]EmployeeLocationDTO, error)
	UpdateLocation(ctx context.Context, id uuid.UUID, lat, lng float64, isVPN, isMock, outOfZone bool) error
	UpdatePhone(ctx context.Context, id uuid.UUID, phone string) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	BatchSetOilChange(ctx context.Context, entries []OilSetupEntry) (int, error)
}

type gormEmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &gormEmployeeRepository{db: db}
}

func (r *gormEmployeeRepository) Create(ctx context.Context, emp *Employee) error {
	if emp.ID == uuid.Nil {
		emp.ID = uuid.New()
	}
	if emp.MotorcycleNumber != "" {
		cleanPlate := strings.TrimSpace(emp.MotorcycleNumber)
		var vehKey string
		if err := r.db.WithContext(ctx).Table("vehicles").
			Select("key_number").
			Where("TRIM(plate_number) = ? AND deleted_at IS NULL", cleanPlate).
			Scan(&vehKey).Error; err == nil && vehKey != "" {
			emp.KeyNumber = vehKey
		}
	}
	return r.db.WithContext(ctx).Create(emp).Error
}

func (r *gormEmployeeRepository) Update(ctx context.Context, emp *Employee) error {
	if emp.MotorcycleNumber != "" {
		cleanPlate := strings.TrimSpace(emp.MotorcycleNumber)
		var vehKey string
		if err := r.db.WithContext(ctx).Table("vehicles").
			Select("key_number").
			Where("TRIM(plate_number) = ? AND deleted_at IS NULL", cleanPlate).
			Scan(&vehKey).Error; err == nil && vehKey != "" {
			emp.KeyNumber = vehKey
		}
	}
	return r.db.WithContext(ctx).Save(emp).Error
}

func (r *gormEmployeeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Employee{}, "id = ?", id).Error
}

func (r *gormEmployeeRepository) FindByID(ctx context.Context, id uuid.UUID) (*Employee, error) {
	var emp Employee
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").First(&emp, "id = ?", id).Error
	if err != nil {
		// Fallback without preloads to protect against relationship errors
		if fallbackErr := r.db.WithContext(ctx).First(&emp, "id = ?", id).Error; fallbackErr == nil {
			return &emp, nil
		}
		return nil, err
	}
	return &emp, nil
}

func (r *gormEmployeeRepository) FindByNationalID(ctx context.Context, nationalID string) (*Employee, error) {
	var emp Employee
	err := r.db.WithContext(ctx).Preload("RoleObj").Preload("Branch").Where("national_id = ?", nationalID).First(&emp).Error
	if err != nil {
		// Fallback without preloads to protect against relationship errors
		if fallbackErr := r.db.WithContext(ctx).Where("national_id = ?", nationalID).First(&emp).Error; fallbackErr == nil {
			return &emp, nil
		}
		return nil, err
	}
	return &emp, nil
}

func (r *gormEmployeeRepository) FindAll(ctx context.Context, filter EmployeeFilter) ([]Employee, int64, error) {
	var emps []Employee
	var total int64

	q := r.db.WithContext(ctx).Model(&Employee{})
	if filter.Search != "" {
		s := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(name) LIKE ? OR national_id LIKE ? OR phone LIKE ? OR employee_number LIKE ? OR application_id LIKE ?", s, s, s, s, s)
	}
	if filter.ApplicationID != "" {
		q = q.Where("application_id = ?", filter.ApplicationID)
	}
	if filter.ApplicationType != "" {
		q = q.Where("application_type = ?", filter.ApplicationType)
	}
	if filter.BranchID != nil && *filter.BranchID != uuid.Nil {
		q = q.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.JobRole != "" {
		q = q.Where("job_role = ? OR role_name = ?", filter.JobRole, filter.JobRole)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	isPostgres := r.db != nil && r.db.Dialector != nil && r.db.Dialector.Name() == "postgres"
	var orderClause string
	if isPostgres {
		orderClause = "NULLIF(regexp_replace(key_number, '\\D', '', 'g'), '')::numeric ASC NULLS LAST, name ASC"
	} else {
		orderClause = "key_number ASC, name ASC"
	}

	if sortBy := strings.TrimSpace(filter.SortBy); sortBy != "" {
		orderDirection := "ASC"
		if strings.EqualFold(strings.TrimSpace(filter.Order), "DESC") {
			orderDirection = "DESC"
		}

		allowedSortColumns := map[string]string{
			"name":              "name",
			"created_at":        "created_at",
			"updated_at":        "updated_at",
			"employee_number":   "employee_number",
			"national_id":       "national_id",
			"phone":             "phone",
			"job_role":          "job_role",
			"shift":             "shift",
			"is_working":        "is_working",
			"motorcycle_number": "motorcycle_number",
			"application_id":    "application_id",
			"application_type":  "application_type",
			"vehicle_type":      "vehicle_type",
			"role_name":         "role_name",
			"total_distance":    "total_distance",
		}

		if sortBy == "key_number" {
			if isPostgres {
				orderClause = "NULLIF(regexp_replace(key_number, '\\D', '', 'g'), '')::numeric " + orderDirection + " NULLS LAST, key_number " + orderDirection + ", name ASC"
			} else {
				orderClause = "key_number " + orderDirection + ", name ASC"
			}
		} else if safeCol, ok := allowedSortColumns[sortBy]; ok {
			orderClause = safeCol + " " + orderDirection
		}
	}

	err := q.Omit("barcode", "qr_code").Preload("Branch").Preload("RoleObj").Order(orderClause).Offset(offset).Limit(filter.Limit).Find(&emps).Error
	if err == nil {
		for i := range emps {
			emps[i].Barcode = ""
			emps[i].QRCode = ""
			if strings.HasPrefix(emps[i].PersonalImage, "data:") && len(emps[i].PersonalImage) > 4096 {
				emps[i].PersonalImage = ""
			}
			if strings.HasPrefix(emps[i].NationalIDImage, "data:") && len(emps[i].NationalIDImage) > 4096 {
				emps[i].NationalIDImage = ""
			}
			if strings.HasPrefix(emps[i].DrivingLicenseImage, "data:") && len(emps[i].DrivingLicenseImage) > 4096 {
				emps[i].DrivingLicenseImage = ""
			}
			if strings.HasPrefix(emps[i].PassportImage, "data:") && len(emps[i].PassportImage) > 4096 {
				emps[i].PassportImage = ""
			}
			if strings.HasPrefix(emps[i].VehicleRegistrationImage, "data:") && len(emps[i].VehicleRegistrationImage) > 4096 {
				emps[i].VehicleRegistrationImage = ""
			}
		}
	}
	return emps, total, err
}

func (r *gormEmployeeRepository) Search(ctx context.Context, query string, branchID *uuid.UUID) ([]Employee, error) {
	var emps []Employee
	s := "%" + strings.ToLower(query) + "%"
	q := r.db.WithContext(ctx).Where("LOWER(name) LIKE ? OR national_id LIKE ? OR phone LIKE ? OR motorcycle_number LIKE ? OR employee_number LIKE ? OR application_id LIKE ?", s, s, s, s, s, s)
	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("branch_id = ?", *branchID)
	}
	err := q.Preload("Branch").Preload("RoleObj").Limit(50).Find(&emps).Error
	return emps, err
}

func (r *gormEmployeeRepository) GetWorkingEmployees(ctx context.Context, branchID *uuid.UUID) ([]Employee, error) {
	var emps []Employee
	q := r.db.WithContext(ctx).Where("is_working = ? AND (job_role = 'DRIVER' OR role_name = 'DRIVER' OR (job_role NOT IN ('SUPERVISOR', 'ADMIN', 'SUPER_ADMIN') AND role_name NOT IN ('SUPERVISOR', 'ADMIN', 'SUPER_ADMIN')))", true)
	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("branch_id = ?", *branchID)
	}
	err := q.Find(&emps).Error
	if err != nil {
		return nil, err
	}
	if len(emps) == 0 {
		return []Employee{}, nil
	}

	empIDs := make([]uuid.UUID, len(emps))
	for i, emp := range emps {
		empIDs[i] = emp.ID
	}

	type activeSessionInfo struct {
		ID               uuid.UUID `gorm:"column:id"`
		EmployeeID       uuid.UUID `gorm:"column:employee_id"`
		StartTime        time.Time `gorm:"column:start_time"`
		StartKM          float64   `gorm:"column:start_km"`
		MotorcycleNumber string    `gorm:"column:motorcycle_number"`
	}
	var sessions []activeSessionInfo
	_ = r.db.WithContext(ctx).Table("work_sessions").
		Select("id, employee_id, start_time, start_km, motorcycle_number").
		Where("employee_id IN ? AND status = 'ACTIVE' AND deleted_at IS NULL", empIDs).
		Order("start_time DESC").
		Scan(&sessions)

	sessMap := make(map[uuid.UUID]activeSessionInfo)
	for _, s := range sessions {
		if _, exists := sessMap[s.EmployeeID]; !exists {
			sessMap[s.EmployeeID] = s
		}
	}

	activeEmps := make([]Employee, 0, len(emps))
	var staleEmpIDs []uuid.UUID

	for i := range emps {
		if s, ok := sessMap[emps[i].ID]; ok {
			startTimeCopy := s.StartTime
			emps[i].ShiftStartTime = &startTimeCopy
			emps[i].StartKM = s.StartKM
			sessIDCopy := s.ID
			emps[i].ActiveSessionID = &sessIDCopy
			if strings.TrimSpace(s.MotorcycleNumber) != "" {
				emps[i].ActiveMotorcycle = strings.TrimSpace(s.MotorcycleNumber)
			} else {
				emps[i].ActiveMotorcycle = emps[i].MotorcycleNumber
			}
			activeEmps = append(activeEmps, emps[i])
		} else {
			staleEmpIDs = append(staleEmpIDs, emps[i].ID)
		}
	}

	// Auto-heal database: clear stuck is_working flags for employees with no active session
	if len(staleEmpIDs) > 0 {
		_ = r.db.WithContext(ctx).Table("employees").Where("id IN ?", staleEmpIDs).Update("is_working", false).Error
	}

	return activeEmps, nil
}

func (r *gormEmployeeRepository) GetLocations(ctx context.Context, branchID *uuid.UUID) ([]EmployeeLocationDTO, error) {
	results := make([]EmployeeLocationDTO, 0)
	q := r.db.WithContext(ctx).Table("employees").
		Select("employees.id, employees.name, employees.job_role, employees.employee_number, employees.phone, employees.personal_image, employees.national_id, employees.key_number, employees.motorcycle_number, employees.application_type, employees.shift, employees.branch_id, branches.name as branch_name, employees.latitude, employees.longitude, employees.last_location_at, employees.is_working as is_shift_active, employees.is_vpn, employees.is_mock_location, employees.out_of_zone").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id").
		Where("employees.deleted_at IS NULL AND employees.latitude IS NOT NULL AND employees.longitude IS NOT NULL")

	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("employees.branch_id = ?", *branchID)
	}

	err := q.Scan(&results).Error
	return results, err
}

func (r *gormEmployeeRepository) UpdateLocation(ctx context.Context, id uuid.UUID, lat, lng float64, isVPN, isMock, outOfZone bool) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&Employee{}).Where("id = ?", id).Updates(map[string]interface{}{
		"latitude":         lat,
		"longitude":        lng,
		"last_location_at": &now,
		"is_vpn":           isVPN,
		"is_mock_location": isMock,
		"out_of_zone":      outOfZone,
	}).Error
}

func (r *gormEmployeeRepository) UpdatePhone(ctx context.Context, id uuid.UUID, phone string) error {
	return r.db.WithContext(ctx).Model(&Employee{}).Where("id = ?", id).Update("phone", phone).Error
}

func (r *gormEmployeeRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.db.WithContext(ctx).Model(&Employee{}).Where("id = ?", id).Updates(map[string]interface{}{
		"password_hash":        passwordHash,
		"must_change_password": false,
	}).Error
}

func (r *gormEmployeeRepository) BatchSetOilChange(ctx context.Context, entries []OilSetupEntry) (int, error) {
	updated := 0
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, e := range entries {
			id, err := uuid.Parse(e.EmployeeID)
			if err != nil {
				continue
			}
			if err := tx.Model(&Employee{}).Where("id = ?", id).Update("last_oil_change_distance", e.LastOilChangeDistance).Error; err == nil {
				updated++
			}
		}
		return nil
	})
	return updated, err
}
