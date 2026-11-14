package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (r *gormOTPRepository) Create(ctx context.Context, otp *OTPRequest) error {
	if otp.ID == uuid.Nil {
		otp.ID = uuid.New()
	}

	// Auto-lookup employee details by NationalID if not already set
	if otp.EmployeeID == nil || *otp.EmployeeID == uuid.Nil {
		var emp struct {
			ID         uuid.UUID
			Name       string
			BranchID   *uuid.UUID
			BranchName string
		}
		if err := r.db.WithContext(ctx).Table("employees").
			Select("employees.id, employees.name, employees.branch_id, branches.name as branch_name").
			Joins("LEFT JOIN branches ON branches.id = employees.branch_id").
			Where("employees.national_id = ?", otp.NationalID).
			First(&emp).Error; err == nil && emp.ID != uuid.Nil {
			otp.EmployeeID = &emp.ID
			if otp.EmployeeName == "" {
				otp.EmployeeName = emp.Name
			}
			if otp.BranchID == nil {
				otp.BranchID = emp.BranchID
			}
			if otp.BranchName == "" {
				otp.BranchName = emp.BranchName
			}
		}
	}

	return r.db.WithContext(ctx).Create(otp).Error
}

func (r *gormOTPRepository) FindByID(ctx context.Context, id uuid.UUID) (*OTPRequest, error) {
	var otp OTPRequest
	err := r.db.WithContext(ctx).First(&otp, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *gormOTPRepository) FindByNationalID(ctx context.Context, nationalID string) (*OTPRequest, error) {
	var otp OTPRequest
	err := r.db.WithContext(ctx).Where("national_id = ?", nationalID).Order("created_at DESC").First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *gormOTPRepository) FindByCode(ctx context.Context, nationalID, code string) (*OTPRequest, error) {
	var otp OTPRequest
	err := r.db.WithContext(ctx).Where("national_id = ? AND otp_code = ? AND LOWER(status) = ? AND expires_at > ?", nationalID, code, "pending", time.Now()).Order("created_at DESC").First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *gormOTPRepository) Update(ctx context.Context, otp *OTPRequest) error {
	return r.db.WithContext(ctx).Save(otp).Error
}

func (r *gormOTPRepository) FindAll(ctx context.Context, query OTPListQuery) ([]OTPRequest, int64, error) {
	var list []OTPRequest
	var total int64
	q := r.db.WithContext(ctx).Model(&OTPRequest{})
	if query.Status != "" {
		q = q.Where("LOWER(status) = LOWER(?)", query.Status)
	}
	if query.Search != "" {
		s := "%" + query.Search + "%"
		q = q.Where("national_id LIKE ? OR employee_name LIKE ?", s, s)
	}
	_ = q.Count(&total)
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	err := q.Order("created_at DESC").Offset(query.Offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *gormOTPRepository) InvalidatePrevious(ctx context.Context, nationalID string) error {
	return r.db.WithContext(ctx).Model(&OTPRequest{}).Where("national_id = ? AND LOWER(status) = 'pending'", nationalID).Update("status", "EXPIRED").Error
}

func (r *gormOTPRepository) FindEmployeeByNationalID(ctx context.Context, nationalID string) (map[string]interface{}, error) {
	var emp struct {
		ID               uuid.UUID  `gorm:"column:id"`
		Name             string     `gorm:"column:name"`
		NationalID       string     `gorm:"column:national_id"`
		EmployeeNumber   string     `gorm:"column:employee_number"`
		JobRole          string     `gorm:"column:job_role"`
		Phone            string     `gorm:"column:phone"`
		PersonalImage    string     `gorm:"column:personal_image"`
		MotorcycleNumber string     `gorm:"column:motorcycle_number"`
		KeyNumber        string     `gorm:"column:key_number"`
		VehicleType      string     `gorm:"column:vehicle_type"`
		Shift            string     `gorm:"column:shift"`
		BranchID         *uuid.UUID `gorm:"column:branch_id"`
		BranchName       string     `gorm:"column:branch_name"`
	}

	err := r.db.WithContext(ctx).Table("employees").
		Select("employees.id, employees.name, employees.national_id, employees.employee_number, employees.job_role, employees.phone, employees.personal_image, employees.motorcycle_number, employees.key_number, employees.vehicle_type, employees.shift, employees.branch_id, branches.name as branch_name").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id").
		Where("employees.national_id = ? AND employees.deleted_at IS NULL", nationalID).
		First(&emp).Error
	if err != nil {
		return nil, err
	}

	res := map[string]interface{}{
		"id":                emp.ID.String(),
		"name":              emp.Name,
		"national_id":       emp.NationalID,
		"employee_number":   emp.EmployeeNumber,
		"job_role":          emp.JobRole,
		"phone":             emp.Phone,
		"personal_image":    emp.PersonalImage,
		"motorcycle_number": emp.MotorcycleNumber,
		"key_number":        emp.KeyNumber,
		"vehicle_type":      emp.VehicleType,
		"shift":             emp.Shift,
		"branch_name":       emp.BranchName,
		"is_employee":       true,
		"role":              emp.JobRole,
	}
	if emp.BranchID != nil {
		res["branch_id"] = emp.BranchID.String()
	}
	return res, nil
}

func (r *gormOTPRepository) GetDevicesByNationalID(ctx context.Context, nationalID string) ([]map[string]interface{}, error) {
	var rows []struct {
		DeviceUUID string    `gorm:"column:device_uuid"`
		DeviceInfo string    `gorm:"column:device_info"`
		CreatedAt  time.Time `gorm:"column:created_at"`
	}

	err := r.db.WithContext(ctx).Table("otp_requests").
		Select("device_uuid, device_info, MAX(created_at) as created_at").
		Where("national_id = ? AND LOWER(status) = 'verified' AND device_uuid != ''", nationalID).
		Group("device_uuid, device_info").
		Order("created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	var res []map[string]interface{}
	for _, row := range rows {
		res = append(res, map[string]interface{}{
			"uuid":       row.DeviceUUID,
			"name":       row.DeviceInfo,
			"trusted_at": row.CreatedAt.Format(time.RFC3339),
		})
	}
	return res, nil
}

func (r *gormOTPRepository) RevokeDevice(ctx context.Context, nationalID, deviceUUID string) error {
	return r.db.WithContext(ctx).Table("otp_requests").
		Where("national_id = ? AND device_uuid = ?", nationalID, deviceUUID).
		Update("status", "REVOKED").Error
}

// ------------------------------------------------------------------
