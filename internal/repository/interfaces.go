package repository

import (
	"context"

	"delivery-backend/internal/domain"

	"github.com/google/uuid"
)

type RoleRepository interface {
	FindAll(ctx context.Context) ([]domain.Role, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	FindByName(ctx context.Context, name string) (*domain.Role, error)
	Create(ctx context.Context, role *domain.Role) error
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountUsersByRoleID(ctx context.Context, roleID uuid.UUID) (int64, error)
}

type AdminRepository interface {
	FindByEmail(ctx context.Context, email string) (*domain.Admin, error)
	FindByUsername(ctx context.Context, username string) (*domain.Admin, error)
	FindByPhone(ctx context.Context, phone string) (*domain.Admin, error)
	FindByLogin(ctx context.Context, login string) (*domain.Admin, error)
	FindByGoogleEmail(ctx context.Context, email string) (*domain.Admin, error)
	FindByGoogleID(ctx context.Context, googleID string) (*domain.Admin, error)
	UpdateGoogleLink(ctx context.Context, adminID uuid.UUID, googleEmail, googleID, googleAvatar string) error
	UnlinkGoogle(ctx context.Context, adminID uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error)
	FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*domain.Admin, error)
	Create(ctx context.Context, admin *domain.Admin) error
	Update(ctx context.Context, admin *domain.Admin) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]domain.Admin, error)
}

type EmployeeRepository interface {
	Create(ctx context.Context, emp *domain.Employee) error
	Update(ctx context.Context, emp *domain.Employee) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error)
	FindByNationalID(ctx context.Context, nationalID string) (*domain.Employee, error)
	FindBySearchTerm(ctx context.Context, term string, branchID *uuid.UUID) ([]domain.Employee, error)
	CountAll(ctx context.Context, branchID *uuid.UUID) (int64, error)
	UpdateLocation(ctx context.Context, id uuid.UUID, lat, lng float64, isVPN, isMock, outOfZone bool) error
	GetLocations(ctx context.Context, branchID *uuid.UUID) ([]domain.Employee, error)
}

type AuditRepository interface {
	CreateLog(ctx context.Context, log *domain.AuditLog) error
	GetLatestLogs(ctx context.Context, limit int) ([]domain.AuditLog, error)
	GetLatestLogsByBranch(ctx context.Context, branchID uuid.UUID, limit int) ([]domain.AuditLog, error)
	FindAll(ctx context.Context, branchID *uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	DeleteBulk(ctx context.Context, ids []uuid.UUID) error
	ClearAll(ctx context.Context, branchID *uuid.UUID) error
}
