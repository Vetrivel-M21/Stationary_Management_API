package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID             uint   `json:"userId"`
	Email              string `json:"email"`
	Role               string `json:"role"`
	BranchID           *uint  `json:"branchId"`
	ApproverAccessType string `json:"approverAccessType"`
	FirstLogin         bool   `json:"firstLogin"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, email, role string, branchID *uint, approverAccessType string, firstLogin bool, secret string, expirationHours int) (string, error) {
	claims := JWTClaims{
		UserID:             userID,
		Email:              email,
		Role:               role,
		BranchID:           branchID,
		ApproverAccessType: approverAccessType,
		FirstLogin:         firstLogin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
