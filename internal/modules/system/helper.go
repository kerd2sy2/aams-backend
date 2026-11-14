package system

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"delivery-backend/pkg/middleware"
)

func getBranchID(c *gin.Context) *uuid.UUID {
	if scope, err := middleware.ResolveBranchScope(c); err == nil && scope != nil {
		if scope.BranchID != nil {
			return scope.BranchID
		}
	}
	if val, exists := c.Get("branch_id"); exists && val != nil {
		if bIDPtr, ok := val.(*uuid.UUID); ok && bIDPtr != nil && *bIDPtr != uuid.Nil {
			return bIDPtr
		}
		if bID, ok := val.(uuid.UUID); ok && bID != uuid.Nil {
			return &bID
		}
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
