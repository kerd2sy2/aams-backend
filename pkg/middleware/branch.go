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

	// 1. Employee / Driver: strictly scoped to employee's assigned branch
	if isEmployee || strings.EqualFold(role, "DRIVER") || strings.EqualFold(role, "EMPLOYEE") {
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

	// 2. SUPER_ADMIN, OWNER, GENERAL_MANAGER, or unassigned ADMIN: explicitly authorized for cross-branch global access.
	// Can optionally filter by branch if query param is explicitly provided.
	isSuperRole := isExecutiveRole(role)
	if admin != nil && admin.RoleObj != nil && isExecutiveRole(admin.RoleObj.Name) {
		isSuperRole = true
	}
	if !isSuperRole && admin != nil && checkAdminPermission(admin, "*") {
		isSuperRole = true
	}

	if isSuperRole || (strings.EqualFold(role, "ADMIN") && (extractContextBranchID(c) == nil || *extractContextBranchID(c) == uuid.Nil)) {
		scope := &BranchScope{
			IsGlobal:  true,
			Role:      role,
			HasBranch: false,
		}
		if s := c.Query("branch_id"); s != "" {
			if id, err := uuid.Parse(s); err == nil && id != uuid.Nil {
				scope.BranchID = &id
				scope.HasBranch = true
				scope.IsGlobal = false // explicitly filtered to specific branch
			}
		}
		return scope, nil
	}

	// 3. SUPERVISOR: strictly scoped to the server-side assigned branch
	if strings.EqualFold(role, "SUPERVISOR") {
		bid := extractContextBranchID(c)
		if bid == nil || *bid == uuid.Nil {
			return nil, ErrBranchContextMissing
		}

		// Security check: if client sent a conflicting branch_id query param, fail closed
		if clientBranchStr := c.Query("branch_id"); clientBranchStr != "" {
			if clientID, err := uuid.Parse(clientBranchStr); err == nil {
				if clientID != *bid {
					return nil, ErrUnauthorizedCrossBranch
				}
			}
		}

		return &BranchScope{
			IsGlobal:  false,
			BranchID:  bid,
			Role:      role,
			HasBranch: true,
		}, nil
	}

	// 4. Other Roles (e.g. ADMIN, OPERATIONS, ACCOUNTANT):
	// Must have an assigned branch or fail closed. Do NOT grant global access implicitly.
	bid := extractContextBranchID(c)
	if bid != nil && *bid != uuid.Nil {
		if clientBranchStr := c.Query("branch_id"); clientBranchStr != "" {
			if clientID, err := uuid.Parse(clientBranchStr); err == nil {
				if clientID != *bid {
					return nil, ErrUnauthorizedCrossBranch
				}
			}
		}
		return &BranchScope{
			IsGlobal:  false,
			BranchID:  bid,
			Role:      role,
			HasBranch: true,
		}, nil
	}

	// Missing branch context for non-SUPER_ADMIN: fail closed
	return nil, ErrBranchContextMissing
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
