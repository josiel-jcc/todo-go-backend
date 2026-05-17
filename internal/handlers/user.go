package handlers

import (
	"net/http"
	"strconv"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/errors"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/notifications"
	"todo-go-backend/internal/repositories"
	"todo-go-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// UserHandler manages user handlers
type UserHandler struct {
	notificationService *notifications.NotificationService
	userRepo            repositories.UserRepository
	userService         services.UserService
}

// NewUserHandler creates a new instance of UserHandler
func NewUserHandler(
	notificationService *notifications.NotificationService,
	userRepo repositories.UserRepository,
	userService services.UserService,
) *UserHandler {
	return &UserHandler{
		notificationService: notificationService,
		userRepo:            userRepo,
		userService:         userService,
	}
}

type DeleteAccountRequest struct {
	Password string `json:"password" binding:"required"`
}

// UpdateTelegramChatIDRequest represents a request to update Telegram chat ID
type UpdateTelegramChatIDRequest struct {
	TelegramChatID *string `json:"telegram_chat_id" example:"123456789"` // Telegram chat ID (must be numeric string, null to remove). User must send a message to the bot first.
}

// UpdateNotificationsEnabledRequest represents a request to update notifications enabled
type UpdateNotificationsEnabledRequest struct {
	NotificationsEnabled *bool `json:"notifications_enabled" example:"true"`
}

// UpdateTelegramChatID updates user's Telegram chat ID
// @Summary      Update Telegram chat ID
// @Description  Updates the Telegram chat ID for the authenticated user to receive notifications
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      UpdateTelegramChatIDRequest  true  "Telegram chat ID"
// @Success      200      {object}  SuccessResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /users/telegram-chat-id [put]
func (h *UserHandler) UpdateTelegramChatID(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req UpdateTelegramChatIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}

	// Validate Telegram Chat ID format if provided
	if req.TelegramChatID != nil && *req.TelegramChatID != "" {
		// Basic validation: should be numeric
		chatID := *req.TelegramChatID
		isNumeric := true
		for i, r := range chatID {
			if r < '0' || r > '9' {
				// Allow negative sign at the start (for group chats)
				if r == '-' && i == 0 && len(chatID) > 1 {
					continue
				}
				isNumeric = false
				break
			}
		}
		if !isNumeric {
			handleError(c, errors.NewInvalidInputError("telegram_chat_id must be a numeric string (e.g., '123456789'). For group chats, it can be negative (e.g., '-123456789')"))
			return
		}
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		handleError(c, errors.NewUserNotFoundError())
		return
	}

	user.TelegramChatID = req.TelegramChatID
	if err := database.DB.Save(&user).Error; err != nil {
		handleError(c, errors.NewInternalServerError(err))
		return
	}

	message := "Telegram chat ID updated successfully"
	if req.TelegramChatID == nil {
		message = "Telegram chat ID removed successfully"
	} else {
		message += ". Make sure you've sent a message to the bot first!"
	}

	handleSuccess(c, http.StatusOK, message, nil)
}

// UpdateNotificationsEnabled updates user's notifications enabled setting
// @Summary      Update notifications enabled
// @Description  Updates the notifications enabled setting for the authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      UpdateNotificationsEnabledRequest  true  "Notifications enabled"
// @Success      200      {object}  SuccessResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /users/notifications-enabled [put]
func (h *UserHandler) UpdateNotificationsEnabled(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req UpdateNotificationsEnabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}

	if req.NotificationsEnabled == nil {
		handleError(c, errors.NewInvalidInputError("notifications_enabled is required"))
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		handleError(c, errors.NewUserNotFoundError())
		return
	}

	user.NotificationsEnabled = *req.NotificationsEnabled
	if err := database.DB.Save(&user).Error; err != nil {
		handleError(c, errors.NewInternalServerError(err))
		return
	}

	message := "Notifications enabled"
	if !*req.NotificationsEnabled {
		message = "Notifications disabled"
	}

	handleSuccess(c, http.StatusOK, message, nil)
}

// TestNotifications manually triggers notification check (for testing)
// @Summary      Test notifications
// @Description  Manually triggers a notification check. Useful for testing without waiting for the scheduler. Check server logs for detailed information.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  SuccessResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /notifications/test [post]
func (h *UserHandler) TestNotifications(c *gin.Context) {
	if err := h.notificationService.CheckAndSendNotifications(); err != nil {
		handleError(c, errors.NewInternalServerError(err))
		return
	}

	handleSuccess(c, http.StatusOK, "Notification check completed. Check server logs for details and verify your email/Telegram.", nil)
}

// GetNotificationDebugInfo returns debug information about notification configuration
// @Summary      Get notification debug info
// @Description  Returns debug information about the current user's notification settings and recent tasks
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  map[string]interface{}
// @Failure      401      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /notifications/debug [get]
func (h *UserHandler) GetNotificationDebugInfo(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		handleError(c, errors.NewUserNotFoundError())
		return
	}

	// Get user's tasks with due dates
	var tasks []models.Task
	database.DB.Where("user_id = ? AND due_date IS NOT NULL AND completed = ?", userID, false).
		Order("due_date ASC").
		Limit(10).
		Find(&tasks)

	// Get recent notifications
	var notifications []models.Notification
	database.DB.Where("user_id = ?", userID).
		Order("sent_at DESC").
		Limit(10).
		Find(&notifications)

	debugInfo := map[string]interface{}{
		"user": map[string]interface{}{
			"id":                    user.ID,
			"username":              user.Username,
			"email":                 user.Email,
			"notifications_enabled": user.NotificationsEnabled,
			"telegram_chat_id":      user.TelegramChatID,
		},
		"tasks_count": len(tasks),
		"tasks":       tasks,
		"notifications_count": len(notifications),
		"recent_notifications": notifications,
	}

	handleSuccess(c, http.StatusOK, "Debug information retrieved", debugInfo)
}

// GetCurrentUser returns the authenticated user's profile
// @Summary      Get current user profile
// @Description  Returns the authenticated user's profile including notification settings
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.User
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /users/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		handleError(c, errors.NewUserNotFoundError())
		return
	}

	c.JSON(http.StatusOK, user)
}

// PaginatedUsersResponse represents a paginated response for users
type PaginatedUsersResponse struct {
	Users      []models.UserPublic `json:"users"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

// GetUsers lists all users in the system with pagination
// @Summary      List users
// @Description  Retrieves a paginated list of all users in the system. Returns only public information (id, username, email) for use in task assignment.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page   query     int     false  "Page number (default: 1)"
// @Param        limit  query     int     false  "Items per page (default: 10, max: 100)"
// @Success      200    {object}  PaginatedUsersResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      401    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			// Maximum limit is 100
			if limit > 100 {
				limit = 100
			}
		}
	}

	users, total, err := h.userRepo.FindAllPaginated(page, limit)
	if err != nil {
		handleError(c, errors.NewInternalServerError(err))
		return
	}

	// Calculate total pages
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	response := PaginatedUsersResponse{
		Users:      users,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteCurrentUser deletes the authenticated user's account and all associated data.
func (h *UserHandler) DeleteCurrentUser(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}

	if err := h.userService.DeleteAccount(userID, req.Password); err != nil {
		handleError(c, err)
		return
	}

	handleSuccess(c, http.StatusOK, "Account deleted successfully", nil)
}

// ExportCurrentUser exports all personal data for the authenticated user.
func (h *UserHandler) ExportCurrentUser(c *gin.Context) {
	userID := c.GetUint("user_id")

	data, err := h.userService.ExportAccount(userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, data)
}
