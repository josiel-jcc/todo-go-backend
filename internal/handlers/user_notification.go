package handlers

import (
	"net/http"
	"strconv"
	"todo-go-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type UserNotificationHandler struct {
	notificationService services.UserNotificationService
}

func NewUserNotificationHandler(notificationService services.UserNotificationService) *UserNotificationHandler {
	return &UserNotificationHandler{notificationService: notificationService}
}

// List returns paginated in-app notifications for the authenticated user.
// @Summary      List in-app notifications
// @Description  Returns in-app notifications (e.g. group invites). Use unread_only=true for unread only.
// @Tags         notifications-in-app
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page         query     int   false  "Page number (default: 1)"
// @Param        limit        query     int   false  "Items per page (default: 20, max: 100)"
// @Param        unread_only  query     bool  false  "Only unread notifications"
// @Success      200          {object}  PaginatedInAppNotificationsResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /notifications/in-app [get]
func (h *UserNotificationHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, limit := parsePagination(c)
	unreadOnly := c.Query("unread_only") == "true"

	list, total, err := h.notificationService.List(userID, unreadOnly, page, limit)
	if err != nil {
		handleError(c, err)
		return
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}
	c.JSON(http.StatusOK, PaginatedInAppNotificationsResponse{
		Notifications: list,
		Total:         total,
		Page:          page,
		Limit:         limit,
		TotalPages:    totalPages,
	})
}

// UnreadCount returns the count of unread in-app notifications.
// @Summary      Unread in-app notification count
// @Description  Returns how many in-app notifications are unread (for badge display).
// @Tags         notifications-in-app
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  UnreadCountResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /notifications/in-app/unread-count [get]
func (h *UserNotificationHandler) UnreadCount(c *gin.Context) {
	userID := c.GetUint("user_id")
	count, err := h.notificationService.UnreadCount(userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, UnreadCountResponse{Count: count})
}

// MarkRead marks a single in-app notification as read.
// @Summary      Mark notification as read
// @Tags         notifications-in-app
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Notification ID"
// @Success      200  {object}  SuccessResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /notifications/in-app/{id}/read [patch]
func (h *UserNotificationHandler) MarkRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, err := parseUintParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}
	if err := h.notificationService.MarkRead(userID, id); err != nil {
		handleError(c, err)
		return
	}
	handleSuccess(c, http.StatusOK, "Notification marked as read", nil)
}

// MarkAllRead marks all in-app notifications as read.
// @Summary      Mark all notifications as read
// @Tags         notifications-in-app
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  SuccessResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /notifications/in-app/read-all [post]
func (h *UserNotificationHandler) MarkAllRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	if err := h.notificationService.MarkAllRead(userID); err != nil {
		handleError(c, err)
		return
	}
	handleSuccess(c, http.StatusOK, "All notifications marked as read", nil)
}

func parsePagination(c *gin.Context) (int, int) {
	page := 1
	limit := 20
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
		if limit > 100 {
			limit = 100
		}
	}
	return page, limit
}
