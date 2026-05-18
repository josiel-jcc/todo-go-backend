package utils

import (
	"time"
	"todo-go-backend/internal/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const TokenMaxAge = 24 * time.Hour

func GenerateToken(userID uint, username, jwtSecret string) (string, string, error) {
	jti := uuid.NewString()
	claims := &auth.Claims{
		UserID:   userID,
		Username: username,
		JTI:      jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenMaxAge)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(jwtSecret))
	return signed, jti, err
}
