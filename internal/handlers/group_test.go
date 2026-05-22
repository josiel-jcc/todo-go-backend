package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"
	"todo-go-backend/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createUserWithToken(t *testing.T, username, email string) (models.User, string) {
	t.Helper()
	hashedPassword, _ := utils.HashPassword("password123")
	user := models.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}
	require.NoError(t, database.DB.Create(&user).Error)
	token, _, _ := utils.GenerateToken(user.ID, user.Username, "test-secret")
	return user, token
}

func createGroupWithMembers(t *testing.T, owner models.User, memberIDs ...uint) models.Group {
	t.Helper()
	groupRepo := repositories.NewGroupRepository()
	g := &models.Group{Name: "Team", CreatedBy: owner.ID}
	require.NoError(t, groupRepo.Create(g))
	require.NoError(t, groupRepo.AddMember(g.ID, owner.ID))
	for _, id := range memberIDs {
		require.NoError(t, groupRepo.AddMember(g.ID, id))
	}
	return *g
}

func TestCreateGroup(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	_, token := createTestUser(t)

	body, _ := json.Marshal(CreateGroupRequest{Name: "Equipe"})
	req, _ := http.NewRequest("POST", "/api/v1/groups", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var group models.Group
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &group))
	assert.Equal(t, "Equipe", group.Name)
}

func TestListGroups(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	user, token := createTestUser(t)
	createGroupWithMembers(t, user)

	req, _ := http.NewRequest("GET", "/api/v1/groups", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var groups []map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &groups))
	assert.NotEmpty(t, groups)
}

func TestGetGroup_ForbiddenForNonMember(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	owner, _ := createTestUser(t)
	_, outsiderToken := createUserWithToken(t, "outsider", "outsider@test.com")
	group := createGroupWithMembers(t, owner)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/groups/%d", group.ID), nil)
	req.Header.Set("Authorization", "Bearer "+outsiderToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestInviteUser(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	owner, ownerToken := createTestUser(t)
	invitee, _ := createUserWithToken(t, "invitee", "invitee@test.com")
	group := createGroupWithMembers(t, owner)

	body, _ := json.Marshal(InviteUserRequest{UserID: invitee.ID})
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/groups/%d/invitations", group.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var notifCount int64
	database.DB.Model(&models.UserNotification{}).Where("user_id = ?", invitee.ID).Count(&notifCount)
	assert.Equal(t, int64(1), notifCount)
}

func TestDeleteDefaultGroup_Returns400(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	user, token := createTestUser(t)

	defaultGroup := &models.Group{Name: "Os de casa", CreatedBy: user.ID, IsDefault: true}
	require.NoError(t, database.DB.Create(defaultGroup).Error)
	require.NoError(t, database.DB.Create(&models.GroupMember{GroupID: defaultGroup.ID, UserID: user.ID}).Error)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/groups/%d", defaultGroup.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUsers_InviteScopeForbidden(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	owner, _ := createTestUser(t)
	_, outsiderToken := createUserWithToken(t, "outsider2", "outsider2@test.com")
	group := createGroupWithMembers(t, owner)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users?scope=invite&group_id=%d", group.ID), nil)
	req.Header.Set("Authorization", "Bearer "+outsiderToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
