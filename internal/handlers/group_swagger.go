package handlers

import (
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/services"
)

// GroupListItemResponse is used for OpenAPI documentation.
type GroupListItemResponse = services.GroupListItem

// GroupDetailResponse is used for OpenAPI documentation.
type GroupDetailResponse = services.GroupDetail

// PaginatedInAppNotificationsResponse is the paginated in-app notifications list.
type PaginatedInAppNotificationsResponse struct {
	Notifications []models.UserNotification `json:"notifications"`
	Total         int64                     `json:"total"`
	Page          int                       `json:"page"`
	Limit         int                       `json:"limit"`
	TotalPages    int                       `json:"total_pages"`
}

// UnreadCountResponse is the unread notification count.
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

// GroupResponse is used for OpenAPI documentation.
type GroupResponse = models.Group

// GroupInvitationListItem is used for OpenAPI documentation.
type GroupInvitationListItem = models.GroupInvitation
