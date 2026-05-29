package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFinanceHealth_Unauthorized(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	owner, _ := createTestUser(t)
	group := createGroupWithMembers(t, owner)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/finance/groups/%d/health", group.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFinanceHealth_ForbiddenForNonMember(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	owner, _ := createTestUser(t)
	_, outsiderToken := createUserWithToken(t, "outsider", "outsider-finance@test.com")
	group := createGroupWithMembers(t, owner)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/finance/groups/%d/health", group.ID), nil)
	req.Header.Set("Authorization", "Bearer "+outsiderToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestFinanceHealth_OKForMember(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	owner, token := createUserWithToken(t, "member", "member-finance@test.com")
	group := createGroupWithMembers(t, owner)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/finance/groups/%d/health", group.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "finance_module")
}
