package handlers

import (
	"net/http"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthHandler manages authentication handlers
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"johndoe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// LoginRequest represents a user login request
// The username field accepts either username or email address
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"johndoe"` // Username or email address (e.g., "johndoe" or "john@example.com")
	Password string `json:"password" binding:"required" example:"password123"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Message string      `json:"message" example:"Login successful"`
	Token   string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User    interface{} `json:"user"`
}

// Register registers a new user
// @Summary      Register a new user
// @Description  Creates a new user account and returns a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequest  true  "User registration data"
// @Success      201      {object}  AuthResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /auth/register [post]
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

	c.JSON(http.StatusCreated, AuthResponse{
		Message: "User created successfully",
		Token:   token,
		User:    userProfileResponse(*user),
	})
}

// Login authenticates a user
// @Summary      Login user
// @Description  Authenticates a user by username or email and returns a JWT token. The username field accepts either username or email address.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "User login credentials. The 'username' field accepts either username (e.g., 'johndoe') or email address (e.g., 'john@example.com')."
// @Success      200      {object}  AuthResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /auth/login [post]
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

	c.JSON(http.StatusOK, AuthResponse{
		Message: "Login successful",
		Token:   token,
		User:    userProfileResponse(*user),
	})
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
	if user.TelegramChatID != nil {
		profile["telegram_chat_id"] = *user.TelegramChatID
	} else {
		profile["telegram_chat_id"] = ""
	}
	return profile
}
