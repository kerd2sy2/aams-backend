package contracts

import (
	"context"

	"github.com/google/uuid"
)

type AdminDTO struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	Username     string     `json:"username"`
	Name         string     `json:"name"`
	Role         string     `json:"role"`
	BranchID     *uuid.UUID `json:"branch_id"`
	BranchName   string     `json:"branch_name"`
	Permissions  []string   `json:"permissions"`
	GoogleEmail  string     `json:"google_email,omitempty"`
	GoogleAvatar string     `json:"google_avatar,omitempty"`
}

type IAuthContract interface {
	GetAdminByID(ctx context.Context, id uuid.UUID) (*AdminDTO, error)
	GetAdminByEmail(ctx context.Context, email string) (*AdminDTO, error)
	HasPermission(ctx context.Context, adminID uuid.UUID, permission string) (bool, error)
}
