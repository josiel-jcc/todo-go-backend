package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTaskStats(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	user, token := createTestUser(t)

	today := time.Now()
	due := time.Date(today.Year(), today.Month(), today.Day(), 15, 0, 0, 0, today.Location())
	database.DB.Create(&models.Task{
		Title: "Today task", Type: models.TaskTypeCasa, UserID: user.ID,
		Completed: false, DueDate: &due, Priority: models.PriorityMedia,
	})
	database.DB.Create(&models.Task{
		Title: "Done", Type: models.TaskTypeTrabalho, UserID: user.ID,
		Completed: true, Priority: models.PriorityBaixa,
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats repositories.UserTaskStats
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
	assert.Equal(t, int64(2), stats.Summary.Total)
	assert.Equal(t, int64(1), stats.Summary.Completed)
	assert.Equal(t, int64(1), stats.Today.Total)
}
