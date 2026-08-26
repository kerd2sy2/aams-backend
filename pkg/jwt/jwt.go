package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	AdminID  uuid.UUID  `json:"admin_id"`
	Email    string     `json:"email"`
	Name     string     `json:"name"`
	Role     string     `json:"role"`
	BranchID *uuid.UUID `json:"branch_id"`
	jwt.RegisteredClaims
}

func GenerateTokens(adminID uuid.UUID, email, name, role string, branchID *uuid.UUID, secret, refreshSecret string) (accessToken string, refreshToken string, err error) {
	// Access token (valid for 24 hours)
	accessClaims := &Claims{
		AdminID:  adminID,
		Email:    email,
		Name:     name,
		Role:     role,
		BranchID: branchID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   adminID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = token.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	// Refresh token (valid for 7 days)
	refreshClaims := &Claims{
		AdminID:  adminID,
		Email:    email,
		Name:     name,
		Role:     role,
		BranchID: branchID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   adminID.String(),
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(refreshSecret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func ValidateToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
