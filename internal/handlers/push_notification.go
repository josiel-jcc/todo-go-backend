package handlers

import (
	"net/http"
	"todo-go-backend/internal/errors"
	"todo-go-backend/internal/notifications"
	"todo-go-backend/internal/repositories"

	"github.com/gin-gonic/gin"
)

// PushNotificationHandler manages Web Push subscription endpoints.
type PushNotificationHandler struct {
	pushService *notifications.PushService
	pushRepo    repositories.PushSubscriptionRepository
}

// NewPushNotificationHandler creates a new PushNotificationHandler.
func NewPushNotificationHandler(
	pushService *notifications.PushService,
	pushRepo repositories.PushSubscriptionRepository,
) *PushNotificationHandler {
	return &PushNotificationHandler{
		pushService: pushService,
		pushRepo:    pushRepo,
	}
}

// VAPIDPublicKeyResponse is returned by GET /notifications/push/vapid-public-key.
type VAPIDPublicKeyResponse struct {
	PublicKey string `json:"public_key" example:"BEl62iUYgUivxIkv69yViEuiBIa-Ib9-SkvMeAtA3LFgDzkrxZJjSgSnfckjBJuBkr3qBUYIHBQaIQhZf9htdY"`
}

// PushSubscribeKeys holds Web Push encryption keys.
type PushSubscribeKeys struct {
	P256dh string `json:"p256dh" binding:"required"`
	Auth   string `json:"auth" binding:"required"`
}

// PushSubscribeRequest is the body for POST /notifications/push/subscribe.
type PushSubscribeRequest struct {
	Endpoint  string            `json:"endpoint" binding:"required"`
	Keys      PushSubscribeKeys `json:"keys" binding:"required"`
	UserAgent string            `json:"user_agent,omitempty"`
}

// PushUnsubscribeRequest is the body for DELETE /notifications/push/subscribe.
type PushUnsubscribeRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
}

// GetVAPIDPublicKey returns the VAPID public key for Web Push subscription.
// @Summary      Get VAPID public key
// @Description  Returns the VAPID public key used by the browser to subscribe to Web Push
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  VAPIDPublicKeyResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /notifications/push/vapid-public-key [get]
func (h *PushNotificationHandler) GetVAPIDPublicKey(c *gin.Context) {
	c.JSON(http.StatusOK, VAPIDPublicKeyResponse{
		PublicKey: h.pushService.PublicKey(),
	})
}

// Subscribe registers or updates a Web Push subscription for the authenticated user.
// @Summary      Subscribe to Web Push
// @Description  Stores or updates a Web Push subscription for the current device
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      PushSubscribeRequest  true  "Push subscription"
// @Success      200      {object}  SuccessResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /notifications/push/subscribe [post]
func (h *PushNotificationHandler) Subscribe(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req PushSubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}

	if err := h.pushRepo.Upsert(
		userID,
		req.Endpoint,
		req.Keys.P256dh,
		req.Keys.Auth,
		req.UserAgent,
	); err != nil {
		handleError(c, errors.NewInternalServerError(err))
		return
	}

	handleSuccess(c, http.StatusOK, "Push subscription saved", nil)
}

// Unsubscribe removes a Web Push subscription for the authenticated user.
// @Summary      Unsubscribe from Web Push
// @Description  Removes a Web Push subscription by endpoint for the current user
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      PushUnsubscribeRequest  true  "Endpoint to remove"
// @Success      200      {object}  SuccessResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /notifications/push/subscribe [delete]
func (h *PushNotificationHandler) Unsubscribe(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req PushUnsubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}

	if err := h.pushRepo.DeleteByEndpoint(userID, req.Endpoint); err != nil {
		handleError(c, errors.NewInternalServerError(err))
		return
	}

	handleSuccess(c, http.StatusOK, "Push subscription removed", nil)
}
