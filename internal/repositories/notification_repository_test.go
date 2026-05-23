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

func setupNotificationRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.Task{},
		&models.Notification{},
	))
	database.DB = db
	return db
}

func TestExistsTaskReminder_dedupByDueDateFingerprint(t *testing.T) {
	setupNotificationRepoTestDB(t)
	repo := NewNotificationRepository()

	user := &models.User{Username: "u1", Email: "u1@test.com", Password: "hash"}
	require.NoError(t, database.DB.Create(user).Error)

	due1 := time.Date(2026, 5, 22, 15, 0, 0, 0, time.UTC)
	task := &models.Task{Title: "Task", UserID: user.ID, DueDate: &due1}
	require.NoError(t, database.DB.Create(task).Error)

	dueFingerprint := normalizeTaskDueDateUTC(due1)
	sentAt := time.Now().UTC()

	require.NoError(t, repo.Create(&models.Notification{
		UserID:      user.ID,
		TaskID:      task.ID,
		Type:        models.NotificationTypeTaskReminder,
		Channel:     models.NotificationChannelPush,
		TaskDueDate: &dueFingerprint,
		SentAt:      sentAt,
	}))

	exists, err := repo.ExistsTaskReminder(user.ID, task.ID, models.NotificationChannelPush, due1)
	require.NoError(t, err)
	assert.True(t, exists)

	existsOtherChannel, err := repo.ExistsTaskReminder(user.ID, task.ID, models.NotificationChannelEmail, due1)
	require.NoError(t, err)
	assert.False(t, existsOtherChannel)

	due2 := due1.Add(time.Hour)
	existsOtherDue, err := repo.ExistsTaskReminder(user.ID, task.ID, models.NotificationChannelPush, due2)
	require.NoError(t, err)
	assert.False(t, existsOtherDue)
}

func TestDeleteByTaskID_clearsReminderMarkers(t *testing.T) {
	setupNotificationRepoTestDB(t)
	repo := NewNotificationRepository()

	user := &models.User{Username: "u2", Email: "u2@test.com", Password: "hash"}
	require.NoError(t, database.DB.Create(user).Error)

	due := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	task := &models.Task{Title: "Task", UserID: user.ID, DueDate: &due}
	require.NoError(t, database.DB.Create(task).Error)

	dueFP := normalizeTaskDueDateUTC(due)
	for _, ch := range []models.NotificationChannel{
		models.NotificationChannelPush,
		models.NotificationChannelEmail,
	} {
		require.NoError(t, repo.Create(&models.Notification{
			UserID:      user.ID,
			TaskID:      task.ID,
			Type:        models.NotificationTypeTaskReminder,
			Channel:     ch,
			TaskDueDate: &dueFP,
			SentAt:      time.Now().UTC(),
		}))
	}

	require.NoError(t, repo.DeleteByTaskID(task.ID))

	exists, err := repo.ExistsTaskReminder(user.ID, task.ID, models.NotificationChannelPush, due)
	require.NoError(t, err)
	assert.False(t, exists)

	var count int64
	require.NoError(t, database.DB.Model(&models.Notification{}).Where("task_id = ?", task.ID).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}
