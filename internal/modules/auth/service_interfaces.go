package auth

import (
	"context"

	"github.com/google/uuid"

)

type AuthService interface {
	IAuthContract
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	GoogleLogin(ctx context.Context, req GoogleLoginRequest) (*LoginResponse, error)
	LinkGoogle(ctx context.Context, adminID uuid.UUID, req GoogleLinkRequest) error
	UnlinkGoogle(ctx context.Context, adminID uuid.UUID) error
	UpdateAvatar(ctx context.Context, adminID uuid.UUID, avatar string) error
}

type OTPService interface {
	RequestOTP(ctx context.Context, req RequestOTPRequest) (*RequestOTPResponse, error)
	VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*LoginResponse, error)
	GetOTPList(ctx context.Context, query OTPListQuery) ([]OTPRequest, int64, error)
	CancelOTP(ctx context.Context, id uuid.UUID) error
	GetDevices(ctx context.Context, nationalID string) ([]map[string]interface{}, error)
	RevokeDevice(ctx context.Context, nationalID, deviceUUID string) error
}

type RoleService interface {
	Create(ctx context.Context, req CreateRoleRequest) (*Role, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRoleRequest) (*Role, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Role, error)
	FindAll(ctx context.Context) ([]RoleResponse, error)
}

type AdminService interface {
	Create(ctx context.Context, req CreateAdminRequest) (*Admin, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateAdminRequest) (*Admin, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Admin, error)
	FindAll(ctx context.Context) ([]Admin, error)
	ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error
	PromoteEmployee(ctx context.Context, req PromoteEmployeeRequest) (*Admin, error)
}

type BranchService interface {
	Create(ctx context.Context, req CreateBranchRequest) (*Branch, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateBranchRequest) (*Branch, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Branch, error)
	FindAll(ctx context.Context) ([]BranchResponse, error)
}
