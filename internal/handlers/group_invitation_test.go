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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListReceivedInvitations(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	owner, ownerToken := createTestUser(t)
	invitee, inviteeToken := createUserWithToken(t, "invitee3", "invitee3@test.com")
	group := createGroupWithMembers(t, owner)

	body, _ := json.Marshal(InviteUserRequest{UserID: invitee.ID})
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/groups/%d/invitations", group.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	listReq, _ := http.NewRequest("GET", "/api/v1/group-invitations", nil)
	listReq.Header.Set("Authorization", "Bearer "+inviteeToken)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	assert.Equal(t, http.StatusOK, listW.Code)
	var invitations []map[string]interface{}
	require.NoError(t, json.Unmarshal(listW.Body.Bytes(), &invitations))
	assert.Len(t, invitations, 1)
}

func TestAcceptInvitation(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	owner, ownerToken := createTestUser(t)
	invitee, inviteeToken := createUserWithToken(t, "invitee4", "invitee4@test.com")
	group := createGroupWithMembers(t, owner)

	body, _ := json.Marshal(InviteUserRequest{UserID: invitee.ID})
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/groups/%d/invitations", group.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var inv models.GroupInvitation
	require.NoError(t, database.DB.Where("invited_user_id = ?", invitee.ID).First(&inv).Error)

	acceptReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/group-invitations/%d/accept", inv.ID), nil)
	acceptReq.Header.Set("Authorization", "Bearer "+inviteeToken)
	acceptW := httptest.NewRecorder()
	router.ServeHTTP(acceptW, acceptReq)

	assert.Equal(t, http.StatusOK, acceptW.Code)

	groupRepo := repositories.NewGroupRepository()
	ok, err := groupRepo.IsMember(group.ID, invitee.ID)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestDeclineInvitation(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	owner, ownerToken := createTestUser(t)
	invitee, inviteeToken := createUserWithToken(t, "invitee5", "invitee5@test.com")
	group := createGroupWithMembers(t, owner)

	body, _ := json.Marshal(InviteUserRequest{UserID: invitee.ID})
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/groups/%d/invitations", group.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var inv models.GroupInvitation
	require.NoError(t, database.DB.Where("invited_user_id = ?", invitee.ID).First(&inv).Error)

	declineReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/group-invitations/%d/decline", inv.ID), nil)
	declineReq.Header.Set("Authorization", "Bearer "+inviteeToken)
	declineW := httptest.NewRecorder()
	router.ServeHTTP(declineW, declineReq)

	assert.Equal(t, http.StatusOK, declineW.Code)

	require.NoError(t, database.DB.First(&inv, inv.ID).Error)
	assert.Equal(t, models.GroupInvitationDeclined, inv.Status)

	groupRepo := repositories.NewGroupRepository()
	ok, err := groupRepo.IsMember(group.ID, invitee.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}
