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

func TestFinanceCategoryBudget_DashboardShowsPercentUsed(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	user, token := createUserWithToken(t, "budget_user", "budget_user@test.com")
	group := createGroupWithMembers(t, user)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/finance/groups/%d/health", group.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	catReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/finance/groups/%d/categories?kind=expense", group.ID), nil)
	catReq.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, catReq)
	var cats []map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cats))
	require.NotEmpty(t, cats)
	categoryID := uint(cats[0]["id"].(float64))

	accBody, _ := json.Marshal(map[string]interface{}{"name": "Conta", "type": "checking"})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/finance/groups/%d/accounts", group.ID), bytes.NewBuffer(accBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var acc map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &acc))
	accountID := uint(acc["id"].(float64))

	budgetBody, _ := json.Marshal(map[string]interface{}{
		"items": []map[string]interface{}{{"category_id": categoryID, "limit_cents": 10000}},
	})
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/finance/groups/%d/budgets", group.ID), bytes.NewBuffer(budgetBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	txBody, _ := json.Marshal(map[string]interface{}{
		"type": "expense", "account_id": accountID, "category_id": categoryID,
		"amount_cents": 4000, "date": "2026-05-15", "visibility": "household",
	})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/finance/groups/%d/transactions", group.ID), bytes.NewBuffer(txBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/finance/groups/%d/dashboard?month=2026-05", group.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var dash map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &dash))
	byCat := dash["by_category"].([]interface{})
	require.NotEmpty(t, byCat)
	found := false
	for _, raw := range byCat {
		row := raw.(map[string]interface{})
		if uint(row["category_id"].(float64)) == categoryID {
			found = true
			assert.Equal(t, float64(10000), row["budget_cents"])
			assert.InDelta(t, 40.0, row["percent_used"], 0.01)
		}
	}
	assert.True(t, found)
}
