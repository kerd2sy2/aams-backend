package hr_legal

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

func (s *hrLegalService) CreateBankAccount(ctx context.Context, req CreateEmployeeBankAccountRequest) (*EmployeeBankAccount, error) {
	if req.IsDefault {
		_ = s.repo.ClearDefaultBankAccount(ctx, req.EmployeeID)
	}

	acc := &EmployeeBankAccount{
		ID:               uuid.New(),
		EmployeeID:       req.EmployeeID,
		BankName:         req.BankName,
		IBAN:             req.IBAN,
		AccountOwnerName: req.AccountOwnerName,
		IsDefault:        req.IsDefault,
	}

	if err := s.repo.CreateBankAccount(ctx, acc); err != nil {
		return nil, err
	}
	return s.repo.FindBankAccountByID(ctx, acc.ID)
}

func (s *hrLegalService) UpdateBankAccount(ctx context.Context, id uuid.UUID, req UpdateEmployeeBankAccountRequest) (*EmployeeBankAccount, error) {
	acc, err := s.repo.FindBankAccountByID(ctx, id)
	if err != nil {
		return nil, errors.New("الحساب البنكي غير موجود")
	}

	if req.BankName != nil {
		acc.BankName = *req.BankName
	}
	if req.IBAN != nil {
		acc.IBAN = *req.IBAN
	}
	if req.AccountOwnerName != nil {
		acc.AccountOwnerName = *req.AccountOwnerName
	}
	if req.IsDefault != nil && *req.IsDefault {
		_ = s.repo.ClearDefaultBankAccount(ctx, acc.EmployeeID)
		acc.IsDefault = true
	} else if req.IsDefault != nil {
		acc.IsDefault = *req.IsDefault
	}

	if err := s.repo.UpdateBankAccount(ctx, acc); err != nil {
		return nil, err
	}
	return s.repo.FindBankAccountByID(ctx, id)
}

func (s *hrLegalService) DeleteBankAccount(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteBankAccount(ctx, id)
}

func (s *hrLegalService) GetBankAccountByID(ctx context.Context, id uuid.UUID) (*EmployeeBankAccount, error) {
	return s.repo.FindBankAccountByID(ctx, id)
}

func (s *hrLegalService) GetAllBankAccounts(ctx context.Context, filter EmployeeBankAccountFilter) ([]EmployeeBankAccount, int64, error) {
	return s.repo.FindBankAccounts(ctx, filter)
}
