package dto

import (
	"time"

	"github.com/google/uuid"
)

type LoginRequest struct {
	Username string `json:"username"`
	Login    string `json:"login"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
	App      string `json:"app"` // "ohda", "eshraf", "edara", "manadeeb", "dashboard"
}

type LoginResponse struct {
	Token                 string          `json:"token"`
	AccessToken           string          `json:"access_token"`
	RefreshToken          string          `json:"refresh_token"`
	ExpiresAt             time.Time       `json:"expires_at"`
	Type                  string          `json:"type"` // "admin" or "employee"
	IsEmployee            bool            `json:"is_employee,omitempty"`
	Employee              any             `json:"employee,omitempty"`
	Admin                 *AdminInfo      `json:"admin,omitempty"`
	User                  *AdminInfo      `json:"user,omitempty"`
	MustChangePassword    bool            `json:"must_change_password"`
	PermissionsConfigured bool            `json:"permissions_configured"`
	AppPermissions        map[string]bool `json:"app_permissions,omitempty"`
}

type AdminInfo struct {
	ID                    uuid.UUID       `json:"id"`
	Email                 string          `json:"email"`
	Username              string          `json:"username"`
	Phone                 string          `json:"phone,omitempty"`
	EmployeeID            *uuid.UUID      `json:"employee_id,omitempty"`
	Name                  string          `json:"name"`
	Role                  string          `json:"role"`
	Permissions           []string        `json:"permissions"`
	PermissionsConfigured bool            `json:"permissions_configured"`
	MustChangePassword    bool            `json:"must_change_password"`
	AppPermissions        map[string]bool `json:"app_permissions,omitempty"`
	GoogleEmail           string     `json:"google_email,omitempty"`
	GoogleAvatar          string     `json:"google_avatar,omitempty"`
	IsGoogleLinked        bool       `json:"is_google_linked"`
	BranchID              *uuid.UUID `json:"branch_id"`
	BranchName            string     `json:"branch_name,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type GoogleLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	GoogleID string `json:"google_id"`
	Token    string `json:"token"`
}

type GoogleLinkRequest struct {
	Email    string `json:"email" binding:"required"`
	GoogleID string `json:"google_id"`
	Avatar   string `json:"avatar"`
}

type RequestOTPRequest struct {
	NationalID string `json:"national_id" binding:"required"`
	DeviceInfo string `json:"device_info"`
	DeviceUUID string `json:"device_uuid"`
}

type RequestOTPResponse struct {
	Success      bool      `json:"success"`
	Message      string    `json:"message"`
	NationalID   string    `json:"national_id"`
	EmployeeName string    `json:"employee_name"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type VerifyOTPRequest struct {
	NationalID string `json:"national_id" binding:"required"`
	OTPCode    string `json:"otp_code" binding:"required"`
	DeviceUUID string `json:"device_uuid"`
}

type OTPListQuery struct {
	Status string `form:"status"`
	Search string `form:"search"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	IsSystem    bool      `json:"is_system"`
	UsersCount  int64     `json:"users_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	DisplayName string   `json:"display_name" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type UpdateRoleRequest struct {
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type CreateAdminRequest struct {
	Name        string     `json:"name" binding:"required"`
	Email       string     `json:"email" binding:"required,email"`
	Username    string     `json:"username" binding:"required,min=3,max=50"`
	Phone       string     `json:"phone"`
	Password    string     `json:"password" binding:"required,min=8"`
	Role        string     `json:"role"`
	RoleID      *uuid.UUID `json:"role_id"`
	EmployeeID  *uuid.UUID `json:"employee_id,omitempty"`
	Permissions []string   `json:"permissions"`
	BranchID    *uuid.UUID `json:"branch_id"`
}

type UpdateAdminRequest struct {
	Name                  string     `json:"name"`
	Email                 string     `json:"email"`
	Username              string     `json:"username"`
	Phone                 string     `json:"phone"`
	Password              string     `json:"password"`
	Role                  string     `json:"role"`
	RoleID                *uuid.UUID `json:"role_id"`
	EmployeeID            *uuid.UUID `json:"employee_id,omitempty"`
	Permissions           []string   `json:"permissions"`
	PermissionsConfigured *bool      `json:"permissions_configured,omitempty"`
	BranchID              *uuid.UUID `json:"branch_id"`
}

type PromoteEmployeeRequest struct {
	EmployeeID  uuid.UUID  `json:"employee_id" binding:"required"`
	Role        string     `json:"role" binding:"required"`
	RoleID      *uuid.UUID `json:"role_id"`
	BranchID    *uuid.UUID `json:"branch_id"`
	Phone       string     `json:"phone"`
	Permissions []string   `json:"permissions"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type CreateBranchRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateBranchRequest struct {
	Name string `json:"name" binding:"required"`
}

type BranchResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	EmployeesCount int64     `json:"employees_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
