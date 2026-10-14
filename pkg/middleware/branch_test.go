package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/pkg/middleware"
)

func TestBranchScope_AllowsBranch(t *testing.T) {
	branchA := uuid.New()
	branchB := uuid.New()

	// 1. Global Scope
	globalScope := &middleware.BranchScope{
		IsGlobal:  true,
		Role:      "SUPER_ADMIN",
		HasBranch: false,
	}
	if !globalScope.AllowsBranch(&branchA) {
		t.Errorf("global scope should allow branch A")
	}
	if !globalScope.AllowsBranch(&branchB) {
		t.Errorf("global scope should allow branch B")
	}

	// 2. Scoped to Branch A
	scopedA := &middleware.BranchScope{
		IsGlobal:  false,
		BranchID:  &branchA,
		Role:      "SUPERVISOR",
		HasBranch: true,
	}
	if !scopedA.AllowsBranch(&branchA) {
		t.Errorf("scoped A should allow branch A")
	}
	if scopedA.AllowsBranch(&branchB) {
		t.Errorf("scoped A MUST DENY branch B")
	}
	if scopedA.AllowsBranch(nil) {
		t.Errorf("scoped A MUST DENY nil target branch")
	}
	nilUUID := uuid.Nil
	if scopedA.AllowsBranch(&nilUUID) {
		t.Errorf("scoped A MUST DENY uuid.Nil target branch")
	}

	// 3. Nil Scope or Missing Branch - Fail Closed
	var nilScope *middleware.BranchScope
	if nilScope.AllowsBranch(&branchA) {
		t.Errorf("nil scope MUST FAIL CLOSED")
	}

	missingBranchScope := &middleware.BranchScope{
		IsGlobal:  false,
		BranchID:  nil,
		Role:      "SUPERVISOR",
		HasBranch: false,
	}
	if missingBranchScope.AllowsBranch(&branchA) {
		t.Errorf("scope with missing branch MUST FAIL CLOSED")
	}
}

func TestResolveBranchScope_SecurityRules(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchA := uuid.New()
	branchB := uuid.New()

	// Case 1: Unauthenticated -> Fail closed
	t.Run("Unauthenticated fails closed", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		scope, err := middleware.ResolveBranchScope(c)
		if err == nil || scope != nil {
			t.Fatalf("expected error for unauthenticated context, got scope: %+v", scope)
		}
	})

	// Case 2: SUPER_ADMIN global access
	t.Run("SUPER_ADMIN gets global access by default", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("admin_role", "SUPER_ADMIN")
		scope, err := middleware.ResolveBranchScope(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !scope.IsGlobal {
			t.Fatalf("expected IsGlobal=true for SUPER_ADMIN")
		}
	})

	// Case 3: SUPER_ADMIN with explicit ?branch_id filter
	t.Run("SUPER_ADMIN can filter to specific branch", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest("GET", "/test?branch_id="+branchA.String(), nil)
		c.Request = req
		c.Set("admin_role", "SUPER_ADMIN")

		scope, err := middleware.ResolveBranchScope(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scope.IsGlobal {
			t.Fatalf("expected IsGlobal=false when filtered")
		}
		if scope.BranchID == nil || *scope.BranchID != branchA {
			t.Fatalf("expected BranchID=%s, got %v", branchA, scope.BranchID)
		}
	})

	// Case 4: SUPERVISOR strictly scoped to assigned branch
	t.Run("SUPERVISOR strictly scoped to assigned branch", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("admin_role", "SUPERVISOR")
		c.Set("branch_id", &branchA)

		scope, err := middleware.ResolveBranchScope(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scope.IsGlobal {
			t.Fatalf("SUPERVISOR must NEVER be global")
		}
		if scope.BranchID == nil || *scope.BranchID != branchA {
			t.Fatalf("expected BranchID=%s", branchA)
		}
	})

	// Case 5: SUPERVISOR with missing branch context -> Fail closed
	t.Run("SUPERVISOR without branch context fails closed", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("admin_role", "SUPERVISOR")
		c.Set("branch_id", nil)

		scope, err := middleware.ResolveBranchScope(c)
		if err == nil || scope != nil {
			t.Fatalf("SUPERVISOR without branch MUST FAIL CLOSED")
		}
	})

	// Case 6: SUPERVISOR attempting forged ?branch_id parameter -> Fail closed
	t.Run("SUPERVISOR with conflicting branch_id query fails closed", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest("GET", "/test?branch_id="+branchB.String(), nil)
		c.Request = req
		c.Set("admin_role", "SUPERVISOR")
		c.Set("branch_id", &branchA)

		scope, err := middleware.ResolveBranchScope(c)
		if err == nil || scope != nil {
			t.Fatalf("SUPERVISOR attempting cross-branch query MUST FAIL CLOSED with error")
		}
	})

	// Case 7: Employee / Driver strictly scoped to employee branch
	t.Run("Employee / Driver strictly scoped to branch", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("admin_role", "DRIVER")
		c.Set("is_employee", true)
		c.Set("branch_id", &branchA)

		scope, err := middleware.ResolveBranchScope(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scope.IsGlobal {
			t.Fatalf("DRIVER must NEVER be global")
		}
		if !scope.AllowsBranch(&branchA) {
			t.Fatalf("DRIVER must be allowed to access own branch")
		}
		if scope.AllowsBranch(&branchB) {
			t.Fatalf("DRIVER MUST BE DENIED access to another branch")
		}
	})
}
