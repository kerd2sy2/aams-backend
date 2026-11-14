package hr_legal

import (

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/pkg/middleware"
)

type HRLegalHandler struct {
	svc HRLegalService
}

func NewHRLegalHandler(svc HRLegalService) *HRLegalHandler {
	return &HRLegalHandler{svc: svc}
}

func getBranchID(c *gin.Context) *uuid.UUID {
	scope, err := middleware.ResolveBranchScope(c)
	if err == nil && scope != nil && !scope.IsGlobal {
		return scope.BranchID
	}
	if bid, exists := c.Get("branch_id"); exists && bid != nil {
		if val, ok := bid.(*uuid.UUID); ok && val != nil {
			return val
		}
		if val, ok := bid.(uuid.UUID); ok && val != uuid.Nil {
			return &val
		}
	}
	return nil
}

// ------------------------------------------------------------------
// 1. Investigation Endpoints
// ------------------------------------------------------------------