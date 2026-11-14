package auth

import (
	"context"

	"errors"

	"strings"

	"github.com/google/uuid"

)

type branchService struct {
	repo BranchRepository
}

func NewBranchService(repo BranchRepository) BranchService {
	return &branchService{repo: repo}
}

func (s *branchService) Create(ctx context.Context, req CreateBranchRequest) (*Branch, error) {
	b := &Branch{Name: req.Name}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *branchService) Update(ctx context.Context, id uuid.UUID, req UpdateBranchRequest) (*Branch, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("الفرع غير موجود")
	}
	b.Name = req.Name
	if err := s.repo.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *branchService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *branchService) FindByID(ctx context.Context, id uuid.UUID) (*Branch, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *branchService) FindAll(ctx context.Context) ([]BranchResponse, error) {
	branches, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]BranchResponse, len(branches))
	for i, b := range branches {
		count, _ := s.repo.CountEmployeesByBranchID(ctx, b.ID)
		res[i] = BranchResponse{
			ID:             b.ID,
			Name:           b.Name,
			EmployeesCount: count,
			CreatedAt:      b.CreatedAt,
			UpdatedAt:      b.UpdatedAt,
		}
	}
	return res, nil
}

func sanitizePermissions(perms []string) []string {
	if len(perms) <= 1 {
		return perms
	}
	var res []string
	for _, p := range perms {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" && trimmed != "*" {
			res = append(res, trimmed)
		}
	}
	if len(res) == 0 {
		return []string{"*"}
	}
	return res
}

