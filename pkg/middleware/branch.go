package middleware

import (
	"errors"
	"strings"

	"delivery-backend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	ErrUnauthenticatedScope   = errors.New("غير مصرح: لا توجد جلسة مستخدم صالحة")
	ErrBranchContextMissing   = errors.New("غير مصرح: حساب المستخدم غير مرتبط بفرع مصرح به")
	ErrUnauthorizedCrossBranch = errors.New("غير مصرح: لا تملك صلاحية الوصول إلى فروع أخرى")
)

// BranchScope encapsulates the evaluated branch authorization for the request.
type BranchScope struct {
	// IsGlobal indicates explicit authorization to access data across all branches
	// (strictly reserved for SUPER_ADMIN or roles with explicit global permissions).
	IsGlobal bool

	// BranchID is the single authorized branch UUID (or the selected branch if global).
	BranchID *uuid.UUID

	// Role is the evaluated role of the caller.
	Role string

	// HasBranch indicates whether a valid non-nil BranchID is present.
	HasBranch bool
}

// AllowsBranch checks whether the scope allows reading or modifying a record belonging to targetBranchID.
// Fails closed if scope is nil, or if targetBranchID does not match when IsGlobal is false.
func (s *BranchScope) AllowsBranch(targetBranchID *uuid.UUID) bool {
	if s == nil {
		return false
	}
	if s.IsGlobal {
		return true
	}
	if !s.HasBranch || s.BranchID == nil {
		return false
	}
	if targetBranchID == nil || *targetBranchID == uuid.Nil {
		return false
	}
	return *s.BranchID == *targetBranchID
}

// extractRequestedBranchID extracts an explicit branch UUID requested by the client via query parameter or header.
func extractRequestedBranchID(c *gin.Context) *uuid.UUID {
	if c == nil || c.Request == nil {
		return nil
	}
	if s := c.Query("branch_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil && id != uuid.Nil {
			return &id
		}
	}
	if s := c.GetHeader("X-Branch-ID"); s != "" {
		if id, err := uuid.Parse(s); err == nil && id != uuid.Nil {
			return &id
		}
	}
	return nil
}

// hasCrossBranchPermission checks if the user has permission to access or switch between multiple branches.
func hasCrossBranchPermission(admin *domain.Admin, role string) bool {
	if isExecutiveRole(role) {
		return true
	}
	if admin != nil && admin.RoleObj != nil && isExecutiveRole(admin.RoleObj.Name) {
		return true
	}
	if admin != nil {
		multiBranchPerms := []string{
			"*",
			"all_branches",
			"branches.manage",
			"branches:all",
			"view_all_branches",
			"manage_branches",
			"custody:all_branches",
			"custody.all_branches",
			"custody",
			"custody.view",
			"custody.manage",
			"apps.ohda",
			"app.ohda",
		}
		for _, p := range multiBranchPerms {
			if checkAdminPermission(admin, p) {
				return true
			}
		}
	}
	if strings.EqualFold(role, "ADMIN") || strings.EqualFold(role, "ACCOUNTANT") || strings.EqualFold(role, "OPERATIONS") {
		return true
	}
	return false
}

// ResolveBranchScope derives and verifies the branch authorization scope from trusted server-side authentication context.
// Fails closed if caller is not authorized or if a branch-scoped user lacks an assigned branch.
func ResolveBranchScope(c *gin.Context) (*BranchScope, error) {
	roleVal, exists := c.Get("admin_role")
	role, _ := roleVal.(string)

	adminVal, _ := c.Get("admin")
	admin, _ := adminVal.(*domain.Admin)
	if (role == "" || !exists) && admin != nil {
		role = admin.Role
		if role == "" && admin.RoleObj != nil {
			role = admin.RoleObj.Name
		}
	}

	if role == "" && !exists {
		return nil, ErrUnauthenticatedScope
	}

	isEmployee := false
	if isEmpVal, ok := c.Get("is_employee"); ok {
		if b, okBool := isEmpVal.(bool); okBool {
			isEmployee = b
		}
	}

	// 1. Employee / Driver: strictly scoped to employee's assigned branch unless granted cross-branch
	if isEmployee || strings.EqualFold(role, "DRIVER") || strings.EqualFold(role, "EMPLOYEE") {
		if !hasCrossBranchPermission(admin, role) {
			bid := extractContextBranchID(c)
			if bid == nil || *bid == uuid.Nil {
				return nil, ErrBranchContextMissing
			}
			return &BranchScope{
				IsGlobal:  false,
				BranchID:  bid,
				Role:      role,
				HasBranch: true,
			}, nil
		}
	}

	// 2. Roles with cross-branch global or multi-branch permission:
	// SUPER_ADMIN, OWNER, GENERAL_MANAGER, ADMIN, ACCOUNTANT, OPERATIONS, or users with explicit multi-branch/custody permissions.
	if hasCrossBranchPermission(admin, role) {
		if reqBranch := extractRequestedBranchID(c); reqBranch != nil {
			return &BranchScope{
				IsGlobal:  false,
				BranchID:  reqBranch,
				Role:      role,
				HasBranch: true,
			}, nil
		}

		assignedBranch := extractContextBranchID(c)
		return &BranchScope{
			IsGlobal:  true,
			BranchID:  assignedBranch,
			Role:      role,
			HasBranch: assignedBranch != nil && *assignedBranch != uuid.Nil,
		}, nil
	}

	// 3. Single-branch scoped roles (e.g. SUPERVISOR without cross-branch permissions):
	// Strictly scoped to server-side assigned branch. Forbids cross-branch access.
	bid := extractContextBranchID(c)
	if bid == nil || *bid == uuid.Nil {
		return nil, ErrBranchContextMissing
	}

	if reqBranch := extractRequestedBranchID(c); reqBranch != nil {
		if *reqBranch != *bid {
			return nil, ErrUnauthorizedCrossBranch
		}
	}

	return &BranchScope{
		IsGlobal:  false,
		BranchID:  bid,
		Role:      role,
		HasBranch: true,
	}, nil
}

func extractContextBranchID(c *gin.Context) *uuid.UUID {
	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if val, ok := bid.(*uuid.UUID); ok && val != nil && *val != uuid.Nil {
			return val
		}
		if val, ok := bid.(uuid.UUID); ok && val != uuid.Nil {
			return &val
		}
	}
	return nil
}
