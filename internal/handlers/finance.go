package handlers

import (
	"net/http"
	"todo-go-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// FinanceHandler serves finance module endpoints.
type FinanceHandler struct {
	groupService services.GroupService
}

// NewFinanceHandler creates a FinanceHandler.
func NewFinanceHandler(groupService services.GroupService) *FinanceHandler {
	return &FinanceHandler{groupService: groupService}
}

// FinanceHealthResponse is returned by the finance module health check.
type FinanceHealthResponse struct {
	Status  string `json:"status"`
	GroupID uint   `json:"group_id"`
}

// Health confirms the finance module is available for a group the user belongs to.
func (h *FinanceHandler) Health(c *gin.Context) {
	userID := c.GetUint("user_id")
	groupID, err := parseUintParam(c, "groupId")
	if err != nil {
		handleError(c, err)
		return
	}

	if _, err := h.groupService.GetGroup(userID, groupID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, FinanceHealthResponse{
		Status:  "finance_module",
		GroupID: groupID,
	})
}
