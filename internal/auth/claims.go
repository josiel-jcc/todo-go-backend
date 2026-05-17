package auth

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	JTI      string `json:"jti"`
	jwt.RegisteredClaims
}
