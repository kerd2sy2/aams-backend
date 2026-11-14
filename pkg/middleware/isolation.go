package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ModuleStatus represents the runtime bulkhead health of an isolated backend module.
type ModuleStatus struct {
	Name        string     `json:"name"`
	TitleAr     string     `json:"title_ar"`
	Status      string     `json:"status"` // "healthy" | "degraded"
	LastFaultAt *time.Time `json:"last_fault_at,omitempty"`
	LastError   string     `json:"last_error,omitempty"`
	FaultCount  int64      `json:"fault_count"`
}

var (
	moduleRegistryMu sync.RWMutex
	moduleOrder      = []string{
		"auth",
		"employee",
		"fleet",
		"work",
		"hr_legal",
		"inventory",
		"custody",
		"target",
		"operation_video",
		"ticket",
		"system",
	}
	moduleStatuses = map[string]*ModuleStatus{
		"auth":            {Name: "auth", TitleAr: "المصادقة والصلاحيات", Status: "healthy"},
		"employee":        {Name: "employee", TitleAr: "شؤون الموظفين", Status: "healthy"},
		"fleet":           {Name: "fleet", TitleAr: "إدارة الأسطول والدبابات", Status: "healthy"},
		"work":            {Name: "work", TitleAr: "العمل والمناوبات", Status: "healthy"},
		"hr_legal":        {Name: "hr_legal", TitleAr: "الموارد البشرية والشؤون القانونية", Status: "healthy"},
		"inventory":       {Name: "inventory", TitleAr: "المستودع والمخزون", Status: "healthy"},
		"custody":         {Name: "custody", TitleAr: "العهدة والمصروفات", Status: "healthy"},
		"target":          {Name: "target", TitleAr: "التارجت والطلبات", Status: "healthy"},
		"operation_video": {Name: "operation_video", TitleAr: "فيديوهات التشغيل", Status: "healthy"},
		"ticket":          {Name: "ticket", TitleAr: "تذاكر الدعم الفني", Status: "healthy"},
		"system":          {Name: "system", TitleAr: "إعدادات النظام والتقارير", Status: "healthy"},
	}
)

func recordModuleFault(moduleName, errMsg string) {
	if len(errMsg) > 500 {
		errMsg = errMsg[:500]
	}
	now := time.Now()

	moduleRegistryMu.Lock()
	defer moduleRegistryMu.Unlock()

	st, exists := moduleStatuses[moduleName]
	if !exists {
		st = &ModuleStatus{
			Name:    moduleName,
			TitleAr: moduleName,
		}
		moduleStatuses[moduleName] = st
		moduleOrder = append(moduleOrder, moduleName)
	}
	st.Status = "degraded"
	st.LastFaultAt = &now
	st.LastError = errMsg
	st.FaultCount++
}

// GetModulesHealth returns a snapshot of every isolated module's health status.
func GetModulesHealth() []ModuleStatus {
	moduleRegistryMu.RLock()
	defer moduleRegistryMu.RUnlock()

	out := make([]ModuleStatus, 0, len(moduleOrder))
	for _, name := range moduleOrder {
		if st, ok := moduleStatuses[name]; ok {
			copySt := *st
			if st.LastFaultAt != nil {
				t := *st.LastFaultAt
				copySt.LastFaultAt = &t
			}
			out = append(out, copySt)
		}
	}
	return out
}

// ModuleIsolationMiddleware implements the Bulkhead Fault-Isolation pattern per bounded context.
// If any handler or service inside the module panics, ONLY that module's request returns a
// localized 503 MODULE_UNAVAILABLE payload while the Go process and all sibling modules continue
// serving traffic at 100% capacity.
func ModuleIsolationMiddleware(moduleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("active_module", moduleName)

		defer func() {
			if rec := recover(); rec != nil {
				stack := string(debug.Stack())
				log.Printf("[BULKHEAD_ISOLATION][module=%s][path=%s] Panic recovered: %v\n%s",
					moduleName, c.Request.URL.Path, rec, stack)
				recordModuleFault(moduleName, fmt.Sprintf("%v", rec))

				if !c.Writer.Written() {
					c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
						"error":     "MODULE_UNAVAILABLE",
						"module":    moduleName,
						"message":   fmt.Sprintf("حدث خطأ مؤقت في خدمة (%s)، باقي خدمات النظام تعمل بشكل طبيعي. يرجى المحاولة مرة أخرى.", moduleName),
						"retryable": true,
					})
				} else {
					c.Abort()
				}
			}
		}()

		c.Next()
	}
}

// ModuleTimeoutMiddleware attaches a bounded context.WithTimeout to every request entering a module
// so that a slow database query or external call in one module never exhausts server goroutines.
func ModuleTimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
