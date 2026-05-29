package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFinanceGoal_CRUDAndProgress(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	user, token := createUserWithToken(t, "goal_user", "goal_user@test.com")
	group := createGroupWithMembers(t, user)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/finance/groups/%d/health", group.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	body, _ := json.Marshal(map[string]interface{}{
		"name": "Reserva de emergência", "target_cents": 50000, "current_cents": 12500,
		"target_date": "2026-12-31",
	})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/finance/groups/%d/goals", group.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var created map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	goalID := uint(created["id"].(float64))
	assert.Equal(t, float64(25), created["percent_complete"])
	assert.Equal(t, false, created["is_completed"])

	patch, _ := json.Marshal(map[string]interface{}{"current_cents": 50000})
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/finance/groups/%d/goals/%d", group.ID, goalID), bytes.NewBuffer(patch))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var updated map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, float64(100), updated["percent_complete"])
	assert.Equal(t, true, updated["is_completed"])

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/finance/groups/%d/goals", group.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var list []map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Len(t, list, 1)

	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/finance/groups/%d/goals/%d", group.ID, goalID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)
}
