package handlers

import (
	"net/http"
	"strconv"
	"todo-go-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	groupService services.GroupService
}

func NewGroupHandler(groupService services.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

type CreateGroupRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100" example:"Equipe"`
}

type UpdateGroupRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100" example:"Equipe atualizada"`
}

type InviteUserRequest struct {
	UserID uint `json:"user_id" binding:"required" example:"2"`
}

// ListGroups returns groups the authenticated user belongs to.
// @Summary      List my groups
// @Description  Returns all groups where the authenticated user is a member.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   GroupListItemResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /groups [get]
func (h *GroupHandler) ListGroups(c *gin.Context) {
	userID := c.GetUint("user_id")
	groups, err := h.groupService.ListGroups(userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, groups)
}

// CreateGroup creates a new group; the creator becomes a member immediately.
// @Summary      Create a group
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      CreateGroupRequest  true  "Group name"
// @Success      201      {object}  GroupResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /groups [post]
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}
	userID := c.GetUint("user_id")
	group, err := h.groupService.CreateGroup(userID, req.Name)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, group)
}

// GetGroup returns group details including members and pending invitations.
// @Summary      Get group details
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Group ID"
// @Success      200  {object}  GroupDetailResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /groups/{id} [get]
func (h *GroupHandler) GetGroup(c *gin.Context) {
	userID := c.GetUint("user_id")
	groupID, err := parseUintParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}
	group, err := h.groupService.GetGroup(userID, groupID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, group)
}

// UpdateGroup renames a group (members only).
// @Summary      Update group name
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                 true  "Group ID"
// @Param        request  body      UpdateGroupRequest  true  "New name"
// @Success      200      {object}  GroupResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /groups/{id} [put]
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}
	userID := c.GetUint("user_id")
	groupID, err := parseUintParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}
	group, err := h.groupService.UpdateGroup(userID, groupID, req.Name)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, group)
}

// DeleteGroup deletes a group (not allowed for default group "Os de casa").
// @Summary      Delete a group
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Group ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /groups/{id} [delete]
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	userID := c.GetUint("user_id")
	groupID, err := parseUintParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}
	if err := h.groupService.DeleteGroup(userID, groupID); err != nil {
		handleError(c, err)
		return
	}
	handleSuccess(c, http.StatusOK, "Group deleted", nil)
}

// InviteUser sends a group invitation and creates an in-app notification.
// @Summary      Invite user to group
// @Description  Creates a pending invitation; the user must accept to join. Cannot invite existing members.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                true  "Group ID"
// @Param        request  body      InviteUserRequest  true  "User to invite"
// @Success      201      {object}  GroupInvitationListItem
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /groups/{id}/invitations [post]
func (h *GroupHandler) InviteUser(c *gin.Context) {
	var req InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}
	userID := c.GetUint("user_id")
	groupID, err := parseUintParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}
	inv, err := h.groupService.InviteUser(userID, groupID, req.UserID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, inv)
}

// CancelInvitation cancels a pending group invitation.
// @Summary      Cancel group invitation
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id              path      int  true  "Group ID"
// @Param        invitation_id   path      int  true  "Invitation ID"
// @Success      200             {object}  SuccessResponse
// @Failure      400             {object}  ErrorResponse
// @Failure      401             {object}  ErrorResponse
// @Failure      403             {object}  ErrorResponse
// @Failure      500             {object}  ErrorResponse
// @Router       /groups/{id}/invitations/{invitation_id} [delete]
func (h *GroupHandler) CancelInvitation(c *gin.Context) {
	userID := c.GetUint("user_id")
	groupID, err := parseUintParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}
	invitationID, err := parseUintParam(c, "invitation_id")
	if err != nil {
		handleError(c, err)
		return
	}
	if err := h.groupService.CancelInvitation(userID, groupID, invitationID); err != nil {
		handleError(c, err)
		return
	}
	handleSuccess(c, http.StatusOK, "Invitation cancelled", nil)
}

// RemoveMember removes a member from the group (or leave the group).
// @Summary      Remove group member
// @Description  Any member can remove another member or leave by passing their own user ID.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id        path      int  true  "Group ID"
// @Param        user_id   path      int  true  "User ID to remove"
// @Success      200       {object}  SuccessResponse
// @Failure      401       {object}  ErrorResponse
// @Failure      403       {object}  ErrorResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /groups/{id}/members/{user_id} [delete]
func (h *GroupHandler) RemoveMember(c *gin.Context) {
	userID := c.GetUint("user_id")
	groupID, err := parseUintParam(c, "id")
	if err != nil {
		handleError(c, err)
		return
	}
	memberID, err := parseUintParam(c, "user_id")
	if err != nil {
		handleError(c, err)
		return
	}
	if err := h.groupService.RemoveMember(userID, groupID, memberID); err != nil {
		handleError(c, err)
		return
	}
	handleSuccess(c, http.StatusOK, "Member removed", nil)
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	val, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(val), nil
}
