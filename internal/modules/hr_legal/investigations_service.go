package hr_legal

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *hrLegalService) toInvestigationResponse(inv *Investigation) *InvestigationResponse {
	var questions, answers, items, images []string
	_ = json.Unmarshal([]byte(inv.Questions), &questions)
	_ = json.Unmarshal([]byte(inv.Answers), &answers)
	_ = json.Unmarshal([]byte(inv.Items), &items)
	_ = json.Unmarshal([]byte(inv.Images), &images)

	g := false
	if inv.IsGuilty != nil {
		g = *inv.IsGuilty
	}

	return &InvestigationResponse{
		ID:                 inv.ID,
		EmployeeID:         inv.EmployeeID,
		EmployeeName:       inv.EmployeeName,
		NationalID:         inv.NationalID,
		SupervisorID:       inv.SupervisorID,
		SupervisorName:     inv.SupervisorName,
		Type:               inv.Type,
		Questions:          questions,
		Answers:            answers,
		ReportText:         inv.ReportText,
		Images:             images,
		Amount:             inv.Amount,
		StartDate:          inv.StartDate,
		EndDate:            inv.EndDate,
		Items:              items,
		IsGuilty:           g,
		Notes:              inv.Notes,
		DeductionMonth:     inv.DeductionMonth,
		Status:             inv.Status,
		ApprovedBy:         inv.ApprovedBy,
		ApprovedByName:     inv.ApprovedByName,
		ApprovedByUsername: inv.ApprovedByUsername,
		RejectedByName:     inv.RejectedByName,
		RejectedByUsername: inv.RejectedByUsername,
		ApprovedAt:         inv.ApprovedAt,
		RejectedAt:         inv.RejectedAt,
		CreatedAt:          inv.CreatedAt,
		UpdatedAt:          inv.UpdatedAt,
	}
}

func (s *hrLegalService) CreateInvestigation(ctx context.Context, req CreateInvestigationRequest, supervisorID uuid.UUID) (*InvestigationResponse, error) {
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, errors.New("معرف الموظف غير صالح")
	}

	questionsJSON, _ := json.Marshal(req.Questions)
	answersJSON, _ := json.Marshal(req.Answers)
	itemsJSON, _ := json.Marshal(req.Items)
	imagesJSON, _ := json.Marshal(req.Images)

	invType := req.Type
	if invType == "" {
		invType = "investigation"
	}

	var startD, endD *time.Time
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startD = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endD = &t
		}
	}

	empName := strings.TrimSpace(req.EmployeeName)
	nationalID := strings.TrimSpace(req.NationalID)
	if empName == "" || nationalID == "" {
		if empInfo, err := s.repo.GetEmployeeInfo(ctx, empID); err == nil && empInfo != nil {
			if empName == "" {
				empName = empInfo.Name
			}
			if nationalID == "" {
				nationalID = empInfo.NationalID
			}
		}
	}

	inv := &Investigation{
		ID:             uuid.New(),
		EmployeeID:     empID,
		EmployeeName:   empName,
		SupervisorID:   supervisorID,
		SupervisorName: strings.TrimSpace(req.SupervisorName),
		NationalID:     nationalID,
		Type:           invType,
		Questions:      string(questionsJSON),
		Answers:        string(answersJSON),
		ReportText:     req.ReportText,
		Images:         string(imagesJSON),
		Amount:         req.Amount,
		StartDate:      startD,
		EndDate:        endD,
		Items:          string(itemsJSON),
		IsGuilty:       &req.IsGuilty,
		Notes:          req.Notes,
		DeductionMonth: req.DeductionMonth,
		Status:         "pending",
	}

	if err := s.repo.CreateInvestigation(ctx, inv); err != nil {
		return nil, err
	}
	if created, err := s.repo.FindInvestigationByID(ctx, inv.ID); err == nil && created != nil {
		return s.toInvestigationResponse(created), nil
	}
	return s.toInvestigationResponse(inv), nil
}

func (s *hrLegalService) UpdateInvestigation(ctx context.Context, id uuid.UUID, req UpdateInvestigationRequest) (*InvestigationResponse, error) {
	inv, err := s.repo.FindInvestigationByID(ctx, id)
	if err != nil {
		return nil, errors.New("سجل التحقيق غير موجود")
	}

	if req.EmployeeID != "" {
		if empUUID, err := uuid.Parse(req.EmployeeID); err == nil {
			inv.EmployeeID = empUUID
		}
	}
	if req.EmployeeName != "" {
		inv.EmployeeName = strings.TrimSpace(req.EmployeeName)
	}
	if req.NationalID != "" {
		inv.NationalID = strings.TrimSpace(req.NationalID)
	}
	if req.SupervisorName != "" {
		inv.SupervisorName = strings.TrimSpace(req.SupervisorName)
	}
	if req.Type != "" {
		inv.Type = req.Type
	}
	if req.Questions != nil {
		b, _ := json.Marshal(req.Questions)
		inv.Questions = string(b)
	}
	if req.Answers != nil {
		b, _ := json.Marshal(req.Answers)
		inv.Answers = string(b)
	}
	if req.ReportText != "" {
		inv.ReportText = req.ReportText
	}
	if req.Images != nil {
		b, _ := json.Marshal(req.Images)
		inv.Images = string(b)
	}
	if req.Amount != nil {
		inv.Amount = req.Amount
	}
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			inv.StartDate = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			inv.EndDate = &t
		}
	}
	if req.Items != nil {
		b, _ := json.Marshal(req.Items)
		inv.Items = string(b)
	}
	inv.IsGuilty = &req.IsGuilty
	if req.Notes != "" {
		inv.Notes = req.Notes
	}
	if req.DeductionMonth != "" {
		inv.DeductionMonth = req.DeductionMonth
	}

	if err := s.repo.UpdateInvestigation(ctx, inv); err != nil {
		return nil, err
	}
	if updated, err := s.repo.FindInvestigationByID(ctx, id); err == nil && updated != nil {
		return s.toInvestigationResponse(updated), nil
	}
	return s.toInvestigationResponse(inv), nil
}

func (s *hrLegalService) DeleteInvestigation(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteInvestigation(ctx, id)
}

func (s *hrLegalService) GetAllInvestigations(ctx context.Context, branchID *uuid.UUID) ([]InvestigationResponse, error) {
	list, err := s.repo.FindInvestigations(ctx, branchID)
	if err != nil {
		return nil, err
	}
	result := make([]InvestigationResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toInvestigationResponse(&list[i]))
	}
	return result, nil
}

func (s *hrLegalService) GetInvestigationByID(ctx context.Context, id uuid.UUID) (*InvestigationResponse, error) {
	inv, err := s.repo.FindInvestigationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toInvestigationResponse(inv), nil
}

func (s *hrLegalService) ApproveInvestigation(ctx context.Context, id uuid.UUID, adminID uuid.UUID, adminName, adminUsername string) (*InvestigationResponse, error) {
	inv, err := s.repo.FindInvestigationByID(ctx, id)
	if err != nil {
		return nil, ErrInvestigationNotFound
	}

	now := time.Now()
	inv.Status = "approved"
	inv.ApprovedBy = &adminID
	inv.ApprovedByName = adminName
	inv.ApprovedByUsername = adminUsername
	inv.ApprovedAt = &now

	if err := s.repo.UpdateInvestigation(ctx, inv); err != nil {
		return nil, err
	}
	return s.toInvestigationResponse(inv), nil
}

func (s *hrLegalService) RejectInvestigation(ctx context.Context, id uuid.UUID, adminID uuid.UUID, adminName, adminUsername string) (*InvestigationResponse, error) {
	inv, err := s.repo.FindInvestigationByID(ctx, id)
	if err != nil {
		return nil, ErrInvestigationNotFound
	}

	now := time.Now()
	inv.Status = "rejected"
	inv.RejectedBy = &adminID
	inv.RejectedByName = adminName
	inv.RejectedByUsername = adminUsername
	inv.RejectedAt = &now

	if err := s.repo.UpdateInvestigation(ctx, inv); err != nil {
		return nil, err
	}
	return s.toInvestigationResponse(inv), nil
}

func (s *hrLegalService) GetPendingInvestigationCount(ctx context.Context) (int64, error) {
	return s.repo.CountPendingInvestigations(ctx)
}
