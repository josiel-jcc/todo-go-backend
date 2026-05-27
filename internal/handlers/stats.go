package handlers

import (
	"net/http"

	"todo-go-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// StatsHandler serves task statistics endpoints.
type StatsHandler struct {
	statsService services.StatsService
}

// NewStatsHandler creates a StatsHandler.
func NewStatsHandler(statsService services.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

// GetTaskStats returns aggregated task statistics for the authenticated user.
// @Summary      Task statistics
// @Description  Aggregated counts: summary, today, by type, by priority, and in-progress total
// @Tags         stats
// @Produce      json
// @Security     BearerAuth
// @Param        hide_stale_completed  query  bool  false  "Exclude completed tasks older than 24h (default true)"
// @Success      200  {object}  repositories.UserTaskStats
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /stats [get]
func (h *StatsHandler) GetTaskStats(c *gin.Context) {
	userID := c.GetUint("user_id")

	hideStale := true
	if v := c.Query("hide_stale_completed"); v == "false" {
		hideStale = false
	}

	stats, err := h.statsService.GetUserTaskStats(userID, hideStale)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}
