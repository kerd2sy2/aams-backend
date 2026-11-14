package hr_legal

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

func (s *hrLegalService) CreateDocument(ctx context.Context, req CreateEmployeeDocumentRequest) (*EmployeeDocument, error) {
	status := "VALID"
	if req.Status != "" {
		status = req.Status
	}

	var issueDate, expiryDate *time.Time
	if req.IssueDate != nil && *req.IssueDate != "" {
		if t, err := time.Parse("2006-01-02", *req.IssueDate); err == nil {
			issueDate = &t
		}
	}
	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		if t, err := time.Parse("2006-01-02", *req.ExpiryDate); err == nil {
			expiryDate = &t
			if t.Before(time.Now()) {
				status = "EXPIRED"
			}
		}
	}

	doc := &EmployeeDocument{
		ID:         uuid.New(),
		EmployeeID: req.EmployeeID,
		DocType:    req.DocType,
		Title:      req.Title,
		DocNumber:  req.DocNumber,
		FileURL:    req.FileURL,
		IssueDate:  issueDate,
		ExpiryDate: expiryDate,
		Status:     status,
		Notes:      req.Notes,
	}

	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, err
	}
	return s.repo.FindDocumentByID(ctx, doc.ID)
}

func (s *hrLegalService) UpdateDocument(ctx context.Context, id uuid.UUID, req UpdateEmployeeDocumentRequest) (*EmployeeDocument, error) {
	doc, err := s.repo.FindDocumentByID(ctx, id)
	if err != nil {
		return nil, errors.New("المستند غير موجود")
	}

	if req.DocType != nil {
		doc.DocType = *req.DocType
	}
	if req.Title != nil {
		doc.Title = *req.Title
	}
	if req.DocNumber != nil {
		doc.DocNumber = *req.DocNumber
	}
	if req.FileURL != nil {
		doc.FileURL = *req.FileURL
	}
	if req.IssueDate != nil {
		if *req.IssueDate == "" {
			doc.IssueDate = nil
		} else if t, err := time.Parse("2006-01-02", *req.IssueDate); err == nil {
			doc.IssueDate = &t
		}
	}
	if req.ExpiryDate != nil {
		if *req.ExpiryDate == "" {
			doc.ExpiryDate = nil
		} else if t, err := time.Parse("2006-01-02", *req.ExpiryDate); err == nil {
			doc.ExpiryDate = &t
		}
	}
	if req.Status != nil {
		doc.Status = *req.Status
	}
	if req.Notes != nil {
		doc.Notes = *req.Notes
	}

	if err := s.repo.UpdateDocument(ctx, doc); err != nil {
		return nil, err
	}
	return s.repo.FindDocumentByID(ctx, id)
}

func (s *hrLegalService) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteDocument(ctx, id)
}

func (s *hrLegalService) GetDocumentByID(ctx context.Context, id uuid.UUID) (*EmployeeDocument, error) {
	return s.repo.FindDocumentByID(ctx, id)
}

func (s *hrLegalService) GetAllDocuments(ctx context.Context, filter EmployeeDocumentFilter) ([]EmployeeDocument, int64, error) {
	return s.repo.FindDocuments(ctx, filter)
}

func (s *hrLegalService) GetExpiringDocuments(ctx context.Context, days int) ([]EmployeeDocument, error) {
	if days <= 0 {
		days = 30
	}
	return s.repo.FindExpiringDocuments(ctx, days)
}
