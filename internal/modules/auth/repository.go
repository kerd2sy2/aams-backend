package auth

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gormAdminRepository struct{ db *gorm.DB }
type gormRoleRepository struct{ db *gorm.DB }
type gormOTPRepository struct{ db *gorm.DB }
type gormBranchRepository struct{ db *gorm.DB }

type AdminRepository interface {
	Create(ctx context.Context, admin *Admin) error
	Update(ctx context.Context, admin *Admin) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Admin, error)
	FindByEmail(ctx context.Context, email string) (*Admin, error)
	FindByUsername(ctx context.Context, username string) (*Admin, error)
	FindByPhone(ctx context.Context, phone string) (*Admin, error)
	FindByLogin(ctx context.Context, userKey string, phoneVariations []string) (*Admin, error)
	FindByGoogleID(ctx context.Context, googleID string) (*Admin, error)
	FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*Admin, error)
	FindAll(ctx context.Context) ([]Admin, error)
	FindEmployeeByLogin(ctx context.Context, login string) (map[string]interface{}, error)
}

type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Role, error)
	FindByName(ctx context.Context, name string) (*Role, error)
	FindAll(ctx context.Context) ([]Role, error)
	CountUsersByRoleID(ctx context.Context, roleID uuid.UUID) (int64, error)
}

type OTPRepository interface {
	Create(ctx context.Context, otp *OTPRequest) error
	FindByID(ctx context.Context, id uuid.UUID) (*OTPRequest, error)
	FindByNationalID(ctx context.Context, nationalID string) (*OTPRequest, error)
	FindByCode(ctx context.Context, nationalID, code string) (*OTPRequest, error)
	Update(ctx context.Context, otp *OTPRequest) error
	FindAll(ctx context.Context, query OTPListQuery) ([]OTPRequest, int64, error)
	InvalidatePrevious(ctx context.Context, nationalID string) error
	FindEmployeeByNationalID(ctx context.Context, nationalID string) (map[string]interface{}, error)
	GetDevicesByNationalID(ctx context.Context, nationalID string) ([]map[string]interface{}, error)
	RevokeDevice(ctx context.Context, nationalID, deviceUUID string) error
}

type BranchRepository interface {
	Create(ctx context.Context, branch *Branch) error
	Update(ctx context.Context, branch *Branch) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Branch, error)
	FindByName(ctx context.Context, name string) (*Branch, error)
	FindAll(ctx context.Context) ([]Branch, error)
	CountEmployeesByBranchID(ctx context.Context, branchID uuid.UUID) (int64, error)
}




func NewAuthRepository(db *gorm.DB) (AdminRepository, RoleRepository, OTPRepository, BranchRepository) {
	return &gormAdminRepository{db: db}, &gormRoleRepository{db: db}, &gormOTPRepository{db: db}, &gormBranchRepository{db: db}
}

func NormalizeSaudiPhoneVariations(input string) []string {
	s := strings.TrimSpace(input)
	if s == "" {
		return []string{""}
	}
	eastern := []string{"٠", "١", "٢", "٣", "٤", "٥", "٦", "٧", "٨", "٩"}
	persian := []string{"۰", "۱", "۲", "۳", "۴", "۵", "۶", "۷", "۸", "۹"}
	for i := 0; i <= 9; i++ {
		s = strings.ReplaceAll(s, eastern[i], strconv.Itoa(i))
		s = strings.ReplaceAll(s, persian[i], strconv.Itoa(i))
	}

	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	digits := b.String()
	if digits == "" {
		return []string{input}
	}

	resMap := make(map[string]struct{})
	resMap[input] = struct{}{}
	resMap[digits] = struct{}{}

	var core string
	if strings.HasPrefix(digits, "009665") && len(digits) == 14 {
		core = digits[5:]
	} else if strings.HasPrefix(digits, "9665") && len(digits) == 12 {
		core = digits[3:]
	} else if strings.HasPrefix(digits, "05") && len(digits) == 10 {
		core = digits[1:]
	} else if strings.HasPrefix(digits, "5") && len(digits) == 9 {
		core = digits
	}

	if len(core) == 9 {
		resMap[core] = struct{}{}
		resMap["0"+core] = struct{}{}
		resMap["966"+core] = struct{}{}
		resMap["+966"+core] = struct{}{}
		resMap["00966"+core] = struct{}{}
	}

	result := make([]string, 0, len(resMap))
	for v := range resMap {
		result = append(result, v)
	}
	return result
}

