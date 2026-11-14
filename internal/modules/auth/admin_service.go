package auth

import (
	"context"

	"encoding/json"
	"errors"
	"fmt"

	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	rootDomain "delivery-backend/internal/domain"
	sharedRepo "delivery-backend/internal/repository"

)

type adminService struct {
	repo    AdminRepository
	empRepo sharedRepo.EmployeeRepository
}

func NewAdminService(repo AdminRepository, empRepo ...sharedRepo.EmployeeRepository) AdminService {
	svc := &adminService{repo: repo}
	if len(empRepo) > 0 {
		svc.empRepo = empRepo[0]
	}
	return svc
}

func (s *adminService) Create(ctx context.Context, req CreateAdminRequest) (*Admin, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	permsJSON, _ := json.Marshal(req.Permissions)
	admin := &Admin{
		Name:                  req.Name,
		Email:                 req.Email,
		Username:              req.Username,
		Phone:                 req.Phone,
		Password:              string(hashedPwd),
		Role:                  req.Role,
		RoleID:                req.RoleID,
		EmployeeID:            req.EmployeeID,
		Permissions:           string(permsJSON),
		PermissionsConfigured: req.Permissions != nil,
		MustChangePassword:    false,
		BranchID:              req.BranchID,
	}
	if admin.Role == "" {
		admin.Role = "ADMIN"
	}
	if err := s.repo.Create(ctx, admin); err != nil {
		return nil, err
	}
	return admin, nil
}

func (s *adminService) Update(ctx context.Context, id uuid.UUID, req UpdateAdminRequest) (*Admin, error) {
	admin, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("المستخدم غير موجود")
	}
	if req.Name != "" {
		admin.Name = req.Name
	}
	if req.Email != "" {
		admin.Email = req.Email
	}
	if req.Username != "" {
		admin.Username = req.Username
	}
	if req.Phone != "" {
		admin.Phone = req.Phone
	}
	if req.Role != "" {
		admin.Role = req.Role
	}
	if req.RoleID != nil {
		admin.RoleID = req.RoleID
	}
	if req.EmployeeID != nil {
		admin.EmployeeID = req.EmployeeID
	}
	if req.Permissions != nil {
		cleanPerms := sanitizePermissions(req.Permissions)
		permsJSON, _ := json.Marshal(cleanPerms)
		admin.Permissions = string(permsJSON)
		admin.PermissionsConfigured = true
	}
	if req.PermissionsConfigured != nil {
		admin.PermissionsConfigured = *req.PermissionsConfigured
	}
	if req.ClearBranch {
		admin.BranchID = nil
	} else if req.BranchID != nil {
		admin.BranchID = req.BranchID
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err == nil {
			admin.Password = string(hashed)
		}
	}
	if req.Avatar != "" {
		admin.Avatar = req.Avatar
		if admin.GoogleAvatar == "" {
			admin.GoogleAvatar = req.Avatar
		}
	}
	admin.Branch = nil
	admin.RoleObj = nil
	if err := s.repo.Update(ctx, admin); err != nil {
		return nil, err
	}

	// Bidirectional sync: If this Admin is linked to an Employee, sync permissions & role to Employee
	if s.empRepo != nil {
		var targetEmp *rootDomain.Employee
		if admin.EmployeeID != nil && *admin.EmployeeID != uuid.Nil {
			targetEmp, _ = s.empRepo.FindByID(ctx, *admin.EmployeeID)
		}
		if targetEmp == nil && admin.Phone != "" {
			variations := NormalizeSaudiPhoneVariations(admin.Phone)
			for _, v := range variations {
				if emps, err := s.empRepo.FindBySearchTerm(ctx, v, nil); err == nil && len(emps) > 0 {
					targetEmp = &emps[0]
					break
				}
			}
		}
		if targetEmp != nil {
			if admin.EmployeeID == nil || *admin.EmployeeID == uuid.Nil {
				admin.EmployeeID = &targetEmp.ID
				_ = s.repo.Update(ctx, admin)
			}
			if req.Role != "" {
				targetEmp.JobRole = req.Role
			}
			if req.RoleID != nil {
				targetEmp.RoleID = req.RoleID
			}
			if req.Permissions != nil {
				targetEmp.Permissions = admin.Permissions
			}
			if req.BranchID != nil || req.ClearBranch {
				targetEmp.BranchID = admin.BranchID
			}
			_ = s.empRepo.Update(ctx, targetEmp)
		}
	}

	return s.repo.FindByID(ctx, admin.ID)
}

func (s *adminService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *adminService) FindByID(ctx context.Context, id uuid.UUID) (*Admin, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *adminService) FindAll(ctx context.Context) ([]Admin, error) {
	return s.repo.FindAll(ctx)
}

func (s *adminService) ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	admin, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("المستخدم غير موجود")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(oldPassword)); err != nil {
		matched := false
		if admin.Phone != "" {
			phoneDigits := extractDigits(admin.Phone)
			if len(phoneDigits) >= 4 && oldPassword == phoneDigits[len(phoneDigits)-4:] {
				matched = true
			}
		}
		if !matched {
			return errors.New("كلمة المرور الحالية غير صحيحة")
		}
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	admin.Password = string(hashed)
	admin.MustChangePassword = false
	return s.repo.Update(ctx, admin)
}

func (s *adminService) PromoteEmployee(ctx context.Context, req PromoteEmployeeRequest) (*Admin, error) {
	if req.EmployeeID == uuid.Nil {
		return nil, errors.New("معرف الموظف مطلوب")
	}
	if req.Role == "" {
		return nil, errors.New("الدور الوظيفي مطلوب")
	}

	empMap, err := s.repo.FindEmployeeByLogin(ctx, req.EmployeeID.String())
	if err != nil || empMap == nil {
		return nil, errors.New("الموظف غير موجود")
	}

	empName, _ := empMap["name"].(string)
	empPhone, _ := empMap["phone"].(string)
	empNumber, _ := empMap["employee_number"].(string)

	targetPhone := strings.TrimSpace(req.Phone)
	if targetPhone == "" {
		targetPhone = empPhone
	}

	admin, err := s.repo.FindByEmployeeID(ctx, req.EmployeeID)
	if err != nil || admin == nil {
		if targetPhone != "" {
			variations := NormalizeSaudiPhoneVariations(targetPhone)
			for _, v := range variations {
				if a, pErr := s.repo.FindByPhone(ctx, v); pErr == nil && a != nil {
					admin = a
					break
				}
			}
		}
	}

	var permsJSON []byte
	hasCustomPerms := false
	if req.Permissions != nil {
		cleanPerms := sanitizePermissions(req.Permissions)
		permsJSON, _ = json.Marshal(cleanPerms)
		hasCustomPerms = true
	}

	if admin != nil {
		admin.Role = req.Role
		if req.RoleID != nil {
			admin.RoleID = req.RoleID
		}
		if req.BranchID != nil {
			admin.BranchID = req.BranchID
		}
		if targetPhone != "" {
			admin.Phone = targetPhone
		}
		admin.EmployeeID = &req.EmployeeID
		if hasCustomPerms {
			admin.Permissions = string(permsJSON)
			admin.PermissionsConfigured = true
		}
		if err := s.repo.Update(ctx, admin); err != nil {
			return nil, err
		}
		return admin, nil
	}

	phoneDigits := extractDigits(targetPhone)
	tempPass := "1234"
	if len(phoneDigits) >= 4 {
		tempPass = phoneDigits[len(phoneDigits)-4:]
	}
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(tempPass), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	email := fmt.Sprintf("emp_%s@aams-logistic.com", req.EmployeeID.String()[:8])
	username := targetPhone
	if username == "" {
		username = empNumber
	}
	if username == "" {
		username = fmt.Sprintf("emp_%s", req.EmployeeID.String()[:8])
	}

	newAdmin := &Admin{
		Name:                  empName,
		Email:                 email,
		Username:              username,
		Phone:                 targetPhone,
		Password:              string(hashedPwd),
		Role:                  req.Role,
		RoleID:                req.RoleID,
		EmployeeID:            &req.EmployeeID,
		Permissions:           string(permsJSON),
		PermissionsConfigured: hasCustomPerms,
		MustChangePassword:    true,
		BranchID:              req.BranchID,
	}

	if err := s.repo.Create(ctx, newAdmin); err != nil {
		return nil, err
	}
	return newAdmin, nil
}

// BranchService