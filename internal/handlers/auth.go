package handlers

import (
	"net/http"
	"time"
	"todo-go-backend/internal/config"
	"todo-go-backend/internal/auth"
	"todo-go-backend/internal/middleware"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/services"
	"todo-go-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	authService services.AuthService
	cfg         *config.Config
}

func NewAuthHandler(authService services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{authService: authService, cfg: cfg}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"johndoe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"johndoe"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type AuthResponse struct {
	Message string      `json:"message" example:"Login successful"`
	Token   string      `json:"token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User    interface{} `json:"user"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}

	user, token, err := h.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		handleError(c, err)
		return
	}

	middleware.SetAuthCookie(c, h.cfg, token)

	c.JSON(http.StatusCreated, AuthResponse{
		Message: "User created successfully",
		Token:   token,
		User:    userProfileResponse(*user),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}

	user, token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		handleError(c, err)
		return
	}

	middleware.SetAuthCookie(c, h.cfg, token)

	c.JSON(http.StatusOK, AuthResponse{
		Message: "Login successful",
		Token:   token,
		User:    userProfileResponse(*user),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if jti, ok := c.Get("token_jti"); ok {
		if jtiStr, ok := jti.(string); ok && jtiStr != "" {
			_ = middleware.RevokeTokenJTI(jtiStr, time.Now().Add(utils.TokenMaxAge))
		}
	} else if cookie, err := c.Cookie("auth_token"); err == nil && cookie != "" {
		revokeFromTokenString("Bearer "+cookie, h.cfg.JWTSecret)
	} else if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		revokeFromTokenString(authHeader, h.cfg.JWTSecret)
	}

	middleware.ClearAuthCookie(c, h.cfg)
	handleSuccess(c, http.StatusOK, "Logout successful", nil)
}

func revokeFromTokenString(authHeader, jwtSecret string) {
	const prefix = "Bearer "
	if len(authHeader) <= len(prefix) {
		return
	}
	tokenString := authHeader[len(prefix):]
	claims := &auth.Claims{}
	_, _ = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if claims.JTI != "" && claims.ExpiresAt != nil {
		_ = middleware.RevokeTokenJTI(claims.JTI, claims.ExpiresAt.Time)
	}
}

func userProfileResponse(user models.User) gin.H {
	profile := gin.H{
		"id":                    user.ID,
		"username":              user.Username,
		"email":                 user.Email,
		"notifications_enabled": user.NotificationsEnabled,
		"created_at":            user.CreatedAt,
		"updated_at":            user.UpdatedAt,
	}
	if user.TermsAcceptedAt != nil {
		profile["terms_accepted_at"] = user.TermsAcceptedAt
	}
	if user.TelegramChatID != nil {
		profile["telegram_chat_id"] = *user.TelegramChatID
	} else {
		profile["telegram_chat_id"] = ""
	}
	return profile
}
