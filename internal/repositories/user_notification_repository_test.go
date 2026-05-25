package repositories

import (
	"testing"
	"time"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupUserNotificationRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.UserNotification{}))
	database.DB = db
	return db
}

func TestUserNotificationRepository_activeOnlyAndMarkRead(t *testing.T) {
	setupUserNotificationRepoTestDB(t)
	repo := NewUserNotificationRepository()

	user := &models.User{Username: "u1", Email: "u1@test.com", Password: "hash"}
	require.NoError(t, database.DB.Create(user).Error)

	now := time.Now()
	staleReadAt := now.Add(-48 * time.Hour)
	recentReadAt := now.Add(-1 * time.Hour)

	notifications := []models.UserNotification{
		{UserID: user.ID, Type: models.UserNotificationTypeTaskReminder, Payload: `{"task_id":1}`, Read: false},
		{UserID: user.ID, Type: models.UserNotificationTypeTaskReminder, Payload: `{"task_id":2}`, Read: true, ReadAt: &recentReadAt},
		{UserID: user.ID, Type: models.UserNotificationTypeTaskReminder, Payload: `{"task_id":3}`, Read: true, ReadAt: &staleReadAt},
	}
	for i := range notifications {
		require.NoError(t, database.DB.Create(&notifications[i]).Error)
	}

	active, total, err := repo.ListByUserID(user.ID, false, true, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, active, 2)

	require.NoError(t, repo.MarkRead(notifications[0].ID, user.ID))

	var updated models.UserNotification
	require.NoError(t, database.DB.First(&updated, notifications[0].ID).Error)
	assert.True(t, updated.Read)
	require.NotNil(t, updated.ReadAt)

	cutoff := now.Add(-25 * time.Hour)
	require.NoError(t, repo.DeleteStaleRead(cutoff))

	var remaining int64
	require.NoError(t, database.DB.Model(&models.UserNotification{}).Where("user_id = ?", user.ID).Count(&remaining).Error)
	assert.Equal(t, int64(2), remaining)
}
