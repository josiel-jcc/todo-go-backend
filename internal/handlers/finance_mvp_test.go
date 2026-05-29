package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFinanceMVP_PrivateTransactionHiddenFromOtherMember(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	owner, ownerToken := createUserWithToken(t, "fin_owner", "fin_owner@test.com")
	member, memberToken := createUserWithToken(t, "fin_member", "fin_member@test.com")
	group := createGroupWithMembers(t, owner, member.ID)

	// Bootstrap finance for owner
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/finance/groups/%d/health", group.ID), nil)
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	accBody, _ := json.Marshal(map[string]interface{}{
		"name": "Carteira", "type": "cash", "initial_balance_cents": 0,
	})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/finance/groups/%d/accounts", group.ID), bytes.NewBuffer(accBody))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var acc map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &acc))
	accountID := uint(acc["id"].(float64))

	catReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/finance/groups/%d/categories?kind=expense", group.ID), nil)
	catReq.Header.Set("Authorization", "Bearer "+ownerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, catReq)
	require.Equal(t, http.StatusOK, w.Code)
	var cats []map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cats))
	require.NotEmpty(t, cats)
	categoryID := uint(cats[0]["id"].(float64))

	today := time.Now().Format("2006-01-02")
	txBody, _ := json.Marshal(map[string]interface{}{
		"type": "expense", "account_id": accountID, "category_id": categoryID,
		"amount_cents": 5000, "date": today, "visibility": "private", "description": "segredo",
	})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/finance/groups/%d/transactions", group.ID), bytes.NewBuffer(txBody))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Member lists transactions — should not see private expense
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/finance/groups/%d/transactions", group.ID), nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var txs []map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &txs))
	assert.Empty(t, txs)

	// Owner sees it
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &txs))
	assert.Len(t, txs, 1)
}

func TestFinanceMVP_ViewerCannotCreateAccount(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	owner, ownerToken := createUserWithToken(t, "adm", "adm@test.com")
	viewer, viewerToken := createUserWithToken(t, "view", "view@test.com")
	group := createGroupWithMembers(t, owner, viewer.ID)

	// Bootstrap owner (admin)
	h := fmt.Sprintf("/api/v1/finance/groups/%d/health", group.ID)
	req, _ := http.NewRequest(http.MethodGet, h, nil)
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Set viewer role
	roleBody, _ := json.Marshal(map[string]string{"role": "viewer"})
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/finance/groups/%d/members/%d/role", group.ID, viewer.ID), bytes.NewBuffer(roleBody))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Viewer bootstrap as editor by health - need viewer to call health first which gives editor default
	// Actually viewer already has role from admin update
	accBody, _ := json.Marshal(map[string]interface{}{"name": "X", "type": "cash"})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/finance/groups/%d/accounts", group.ID), bytes.NewBuffer(accBody))
	req.Header.Set("Authorization", "Bearer "+viewerToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestFinanceMVP_TransferUpdatesBalances(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	owner, token := createUserWithToken(t, "xfer", "xfer@test.com")
	group := createGroupWithMembers(t, owner)

	bootstrap := fmt.Sprintf("/api/v1/finance/groups/%d/health", group.ID)
	req, _ := http.NewRequest(http.MethodGet, bootstrap, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	createAcc := func(name string) uint {
		body, _ := json.Marshal(map[string]interface{}{
			"name": name, "type": "checking", "initial_balance_cents": 10000,
		})
		r, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/finance/groups/%d/accounts", group.ID), bytes.NewBuffer(body))
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, r)
		require.Equal(t, http.StatusCreated, rec.Code)
		var acc map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &acc))
		return uint(acc["id"].(float64))
	}

	fromID := createAcc("Origem")
	toID := createAcc("Destino")
	today := time.Now().Format("2006-01-02")
	txBody, _ := json.Marshal(map[string]interface{}{
		"type": "transfer", "account_id": fromID, "transfer_account_id": toID,
		"amount_cents": 2500, "date": today, "visibility": "household",
	})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/finance/groups/%d/transactions", group.ID), bytes.NewBuffer(txBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/finance/groups/%d/accounts", group.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var accounts []map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &accounts))
	balances := map[uint]int64{}
	for _, a := range accounts {
		balances[uint(a["id"].(float64))] = int64(a["balance_cents"].(float64))
	}
	assert.Equal(t, int64(7500), balances[fromID])
	assert.Equal(t, int64(12500), balances[toID])
}
