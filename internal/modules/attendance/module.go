package attendance

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"delivery-backend/pkg/middleware"
)

type gormEmployeeFinder struct {
	db *gorm.DB
}

func (f *gormEmployeeFinder) FindAllEmployees(ctx context.Context, branchID *uuid.UUID) ([]EmployeeInfo, error) {
	var rows []struct {
		ID          uuid.UUID  `gorm:"column:id"`
		Name        string     `gorm:"column:name"`
		NationalID  string     `gorm:"column:national_id"`
		BranchID    *uuid.UUID `gorm:"column:branch_id"`
		VehicleType string     `gorm:"column:vehicle_type"`
		BranchName  string     `gorm:"column:branch_name"`
	}

	q := f.db.WithContext(ctx).Table("employees").
		Select("employees.id, employees.name, employees.national_id, employees.branch_id, employees.vehicle_type, branches.name as branch_name").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id").
		Where("employees.deleted_at IS NULL")

	if branchID != nil && *branchID != uuid.Nil {
		q = q.Where("employees.branch_id = ?", *branchID)
	}

	if err := q.Order("employees.name ASC").Limit(500).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]EmployeeInfo, len(rows))
	for i, r := range rows {
		result[i] = EmployeeInfo{
			ID:          r.ID,
			Name:        r.Name,
			NationalID:  r.NationalID,
			BranchID:    r.BranchID,
			BranchName:  r.BranchName,
			VehicleType: r.VehicleType,
		}
	}
	return result, nil
}

func (f *gormEmployeeFinder) FindEmployeeByID(ctx context.Context, id uuid.UUID) (*EmployeeInfo, error) {
	var r struct {
		ID          uuid.UUID  `gorm:"column:id"`
		Name        string     `gorm:"column:name"`
		NationalID  string     `gorm:"column:national_id"`
		BranchID    *uuid.UUID `gorm:"column:branch_id"`
		VehicleType string     `gorm:"column:vehicle_type"`
		BranchName  string     `gorm:"column:branch_name"`
	}

	err := f.db.WithContext(ctx).Table("employees").
		Select("employees.id, employees.name, employees.national_id, employees.branch_id, employees.vehicle_type, branches.name as branch_name").
		Joins("LEFT JOIN branches ON branches.id = employees.branch_id").
		Where("employees.id = ? AND employees.deleted_at IS NULL", id).
		First(&r).Error
	if err != nil {
		return nil, err
	}

	return &EmployeeInfo{
		ID:          r.ID,
		Name:        r.Name,
		NationalID:  r.NationalID,
		BranchID:    r.BranchID,
		BranchName:  r.BranchName,
		VehicleType: r.VehicleType,
	}, nil
}

type Module struct {
	Service AttendanceService
	Handler *AttendanceHandler
}

func NewModule(db *gorm.DB) *Module {
	repo := NewAttendanceRepository(db)
	finder := &gormEmployeeFinder{db: db}
	svc := NewAttendanceService(repo, finder)
	handler := NewAttendanceHandler(svc)
	return &Module{
		Service: svc,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/attendance", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.GetAttendance)
	rg.POST("/attendance/:employee_id", middleware.RequireRoles("SUPER_ADMIN", "ADMIN", "SUPERVISOR"), m.Handler.ToggleAttendance)
}
