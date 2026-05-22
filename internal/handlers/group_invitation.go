package handlers

import (
	"net/http"
	"todo-go-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type GroupInvitationHandler struct {
	groupService services.GroupService
}

func NewGroupInvitationHandler(groupService services.GroupService) *GroupInvitationHandler {
	return &GroupInvitationHandler{groupService: groupService}
}

// ListReceived returns pending group invitations for the authenticated user.
// @Summary      List received invitations
// @Description  Returns pending group invitations where the authenticated user is the invitee.
// @Tags         group-invitations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   GroupInvitationListItem
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /group-invitations [get]
func (h *GroupInvitationHandler) ListReceived(c *gin.Context) {
	userID := c.GetUint("user_id")
	list, err := h.groupService.ListReceivedInvitations(userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// Accept accepts a group invitation and adds the user as a member.
// @Summary      Accept group invitation
// @Tags         group-invitations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Invitation ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /group-invitations/{id}/accept [post]
func (h *GroupInvitationHandler) Accept(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, err := parseUintParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}
	if err := h.groupService.AcceptInvitation(userID, id); err != nil {
		handleError(c, err)
		return
	}
	handleSuccess(c, http.StatusOK, "Invitation accepted", nil)
}

// Decline declines a group invitation.
// @Summary      Decline group invitation
// @Tags         group-invitations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Invitation ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /group-invitations/{id}/decline [post]
func (h *GroupInvitationHandler) Decline(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, err := parseUintParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}
	if err := h.groupService.DeclineInvitation(userID, id); err != nil {
		handleError(c, err)
		return
	}
	handleSuccess(c, http.StatusOK, "Invitation declined", nil)
}
