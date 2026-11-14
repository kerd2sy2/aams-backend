package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestModuleIsolationMiddleware_BulkheadRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Module A: simulates a fatal panic inside target module
	targetGroup := r.Group("/api/v1")
	targetGroup.Use(ModuleIsolationMiddleware("target"), ModuleTimeoutMiddleware(5*time.Second))
	targetGroup.GET("/targets/crash", func(c *gin.Context) {
		panic("simulated nil pointer dereference in target analytics")
	})

	// Module B: simulates normal work shift operation
	workGroup := r.Group("/api/v1")
	workGroup.Use(ModuleIsolationMiddleware("work"), ModuleTimeoutMiddleware(5*time.Second))
	workGroup.GET("/work/active", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ACTIVE"})
	})

	// 1. Trigger panic in Module A
	reqA := httptest.NewRequest(http.MethodGet, "/api/v1/targets/crash", nil)
	recA := httptest.NewRecorder()
	r.ServeHTTP(recA, reqA)

	if recA.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 for crashed module, got %d", recA.Code)
	}

	var bodyA map[string]interface{}
	if err := json.Unmarshal(recA.Body.Bytes(), &bodyA); err != nil {
		t.Fatalf("failed to decode 503 response: %v", err)
	}
	if bodyA["error"] != "MODULE_UNAVAILABLE" || bodyA["module"] != "target" {
		t.Errorf("unexpected bulkhead payload: %+v", bodyA)
	}

	// 2. Verify Module B continues operating at 100% capacity immediately after Module A panicked
	reqB := httptest.NewRequest(http.MethodGet, "/api/v1/work/active", nil)
	recB := httptest.NewRecorder()
	r.ServeHTTP(recB, reqB)

	if recB.Code != http.StatusOK {
		t.Fatalf("expected status 200 for healthy sibling module, got %d", recB.Code)
	}

	// 3. Verify health registry reflects "target" as degraded and "work" as healthy
	statuses := GetModulesHealth()
	var targetFound, workFound bool
	for _, st := range statuses {
		if st.Name == "target" {
			targetFound = true
			if st.Status != "degraded" || st.FaultCount < 1 || st.LastFaultAt == nil {
				t.Errorf("expected target module to be degraded with fault_count>=1, got %+v", st)
			}
		}
		if st.Name == "work" {
			workFound = true
			if st.Status != "healthy" || st.FaultCount != 0 {
				t.Errorf("expected work module to stay healthy, got %+v", st)
			}
		}
	}
	if !targetFound || !workFound {
		t.Errorf("missing expected modules in registry: target=%v work=%v", targetFound, workFound)
	}
}
