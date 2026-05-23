package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestUpdateReminderSettings(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	_, token := createTestUser(t)

	t.Run("valid value 15", func(t *testing.T) {
		body, _ := json.Marshal(UpdateReminderSettingsRequest{ReminderMinutesBefore: 15})
		req, _ := http.NewRequest("PUT", "/api/v1/users/reminder-settings", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var user models.User
		json.Unmarshal(w.Body.Bytes(), &user)
		assert.Equal(t, 15, user.ReminderMinutesBefore)
	})

	t.Run("invalid value 7", func(t *testing.T) {
		body, _ := json.Marshal(UpdateReminderSettingsRequest{ReminderMinutesBefore: 7})
		req, _ := http.NewRequest("PUT", "/api/v1/users/reminder-settings", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPushSubscribe(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	user, token := createTestUser(t)

	t.Run("subscribe and unsubscribe", func(t *testing.T) {
		subBody, _ := json.Marshal(PushSubscribeRequest{
			Endpoint:  "https://push.example.com/sub/abc",
			UserAgent: "TestBrowser/1.0",
			Keys: PushSubscribeKeys{
				P256dh: "test-p256dh-key",
				Auth:   "test-auth-secret",
			},
		})
		req, _ := http.NewRequest("POST", "/api/v1/notifications/push/subscribe", bytes.NewBuffer(subBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var count int64
		database.DB.Model(&models.PushSubscription{}).Where("user_id = ?", user.ID).Count(&count)
		assert.Equal(t, int64(1), count)

		unsubBody, _ := json.Marshal(PushUnsubscribeRequest{Endpoint: "https://push.example.com/sub/abc"})
		delReq, _ := http.NewRequest("DELETE", "/api/v1/notifications/push/subscribe", bytes.NewBuffer(unsubBody))
		delReq.Header.Set("Content-Type", "application/json")
		delReq.Header.Set("Authorization", "Bearer "+token)
		delW := httptest.NewRecorder()
		router.ServeHTTP(delW, delReq)
		assert.Equal(t, http.StatusOK, delW.Code)

		database.DB.Model(&models.PushSubscription{}).Where("user_id = ?", user.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("vapid public key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/notifications/push/vapid-public-key", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp VAPIDPublicKeyResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "test-vapid-public-key", resp.PublicKey)
	})
}
