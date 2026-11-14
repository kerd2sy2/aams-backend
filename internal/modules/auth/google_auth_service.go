package auth

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"delivery-backend/pkg/jwt"

	"github.com/google/uuid"
)

func (s *authService) GoogleLogin(ctx context.Context, req GoogleLoginRequest) (*LoginResponse, error) {
	admin, err := s.adminRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("حساب Google هذا غير مرتبط بأي مسؤول مسجل في النظام")
	}

	if admin.GoogleID == "" && req.GoogleID != "" {
		admin.GoogleID = req.GoogleID
		admin.GoogleEmail = req.Email
		_ = s.adminRepo.Update(ctx, admin)
	}

	token, exp, err := jwt.GenerateToken(admin.ID, admin.Email, admin.Role, s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, _, _ := jwt.GenerateToken(admin.ID, admin.Email, admin.Role, s.cfg.JWTSecret)

	var perms []string
	if admin.Permissions != "" {
		_ = json.Unmarshal([]byte(admin.Permissions), &perms)
	}
	branchName := ""
	if admin.Branch != nil {
		branchName = admin.Branch.Name
	}

	effectiveAvatar := admin.Avatar
	if effectiveAvatar == "" {
		effectiveAvatar = admin.GoogleAvatar
	}

	adminInfo := &AdminInfo{
		ID:             admin.ID,
		Email:          admin.Email,
		Username:       admin.Username,
		Name:           admin.Name,
		Role:           admin.Role,
		Permissions:    perms,
		GoogleEmail:    admin.GoogleEmail,
		GoogleAvatar:   admin.GoogleAvatar,
		Avatar:         effectiveAvatar,
		IsGoogleLinked: admin.GoogleID != "",
		BranchID:       admin.BranchID,
		BranchName:     branchName,
	}

	return &LoginResponse{
		Token:        token,
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresAt:    exp,
		Type:         "admin",
		Admin:        adminInfo,
		User:         adminInfo,
	}, nil
}

func (s *authService) LinkGoogle(ctx context.Context, adminID uuid.UUID, req GoogleLinkRequest) error {
	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return errors.New("المستخدم غير موجود")
	}
	admin.GoogleID = req.GoogleID
	admin.GoogleEmail = req.Email
	admin.GoogleAvatar = req.Avatar
	if admin.Avatar == "" && req.Avatar != "" {
		admin.Avatar = req.Avatar
	}
	return s.adminRepo.Update(ctx, admin)
}

func (s *authService) UnlinkGoogle(ctx context.Context, adminID uuid.UUID) error {
	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return errors.New("المستخدم غير موجود")
	}
	admin.GoogleID = ""
	admin.GoogleEmail = ""
	admin.GoogleAvatar = ""
	return s.adminRepo.Update(ctx, admin)
}

func (s *authService) UpdateAvatar(ctx context.Context, adminID uuid.UUID, avatar string) error {
	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return errors.New("المستخدم غير موجود")
	}
	cleanAvatar := strings.TrimSpace(avatar)
	admin.Avatar = cleanAvatar
	if admin.GoogleAvatar == "" {
		admin.GoogleAvatar = cleanAvatar
	}
	return s.adminRepo.Update(ctx, admin)
}
