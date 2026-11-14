package auth

import (
	"context"

	"encoding/json"
	"errors"

	"github.com/google/uuid"

)

type roleService struct {
	repo RoleRepository
}

func NewRoleService(repo RoleRepository) RoleService {
	return &roleService{repo: repo}
}

func (s *roleService) Create(ctx context.Context, req CreateRoleRequest) (*Role, error) {
	permsJSON, _ := json.Marshal(req.Permissions)
	role := &Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Permissions: string(permsJSON),
	}
	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) Update(ctx context.Context, id uuid.UUID, req UpdateRoleRequest) (*Role, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("الدور غير موجود")
	}
	if req.DisplayName != "" {
		role.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.Permissions != nil {
		permsJSON, _ := json.Marshal(req.Permissions)
		role.Permissions = string(permsJSON)
	}
	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *roleService) FindByID(ctx context.Context, id uuid.UUID) (*Role, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *roleService) FindAll(ctx context.Context) ([]RoleResponse, error) {
	roles, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]RoleResponse, len(roles))
	for i, r := range roles {
		var perms []string
		_ = json.Unmarshal([]byte(r.Permissions), &perms)
		count, _ := s.repo.CountUsersByRoleID(ctx, r.ID)
		res[i] = RoleResponse{
			ID:          r.ID,
			Name:        r.Name,
			DisplayName: r.DisplayName,
			Description: r.Description,
			Permissions: perms,
			IsSystem:    r.IsSystem,
			UsersCount:  count,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		}
	}
	return res, nil
}

// AdminService