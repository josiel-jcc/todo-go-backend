package services

import (
	"encoding/json"
	"testing"

	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupActivityNotificationTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.Task{},
		&models.Comment{},
		&models.UserNotification{},
		&models.TaskSharedWith{},
	))
	database.DB = db
}

func TestActivityNotificationService_NotifyTaskComment(t *testing.T) {
	setupActivityNotificationTestDB(t)
	svc := NewActivityNotificationService(repositories.NewUserNotificationRepository())

	owner := models.User{Username: "owner", Email: "o@test.com", Password: "h"}
	assignee := models.User{Username: "assignee", Email: "a@test.com", Password: "h"}
	shared := models.User{Username: "shared", Email: "s@test.com", Password: "h"}
	commenter := models.User{Username: "commenter", Email: "c@test.com", Password: "h"}
	require.NoError(t, database.DB.Create(&owner).Error)
	require.NoError(t, database.DB.Create(&assignee).Error)
	require.NoError(t, database.DB.Create(&shared).Error)
	require.NoError(t, database.DB.Create(&commenter).Error)

	assignedBy := owner.ID
	task := models.Task{
		Title: "Tarefa", Type: models.TaskTypeCasa, UserID: assignee.ID,
		AssignedBy: &assignedBy, Priority: models.PriorityMedia,
	}
	require.NoError(t, database.DB.Create(&task).Error)
	require.NoError(t, database.DB.Create(&models.TaskSharedWith{TaskID: task.ID, UserID: shared.ID}).Error)
	require.NoError(t, database.DB.Preload("SharedWithUsers").First(&task, task.ID).Error)

	comment := &models.Comment{
		ID: 1, Content: "Olá, atualização!", TaskID: task.ID, UserID: commenter.ID,
		User: commenter,
	}

	err := svc.NotifyTaskComment(&task, comment, commenter.Username)
	require.NoError(t, err)

	var notifications []models.UserNotification
	require.NoError(t, database.DB.Where("type = ?", models.UserNotificationTypeTaskComment).Find(&notifications).Error)
	assert.Len(t, notifications, 3)

	userIDs := map[uint]bool{}
	for _, n := range notifications {
		userIDs[n.UserID] = true
		assert.Equal(t, models.UserNotificationTypeTaskComment, n.Type)
	}
	assert.True(t, userIDs[owner.ID])
	assert.True(t, userIDs[assignee.ID])
	assert.True(t, userIDs[shared.ID])
	assert.False(t, userIDs[commenter.ID])
}

func TestActivityNotificationService_NotifyDelegatedTaskCompleted(t *testing.T) {
	setupActivityNotificationTestDB(t)
	svc := NewActivityNotificationService(repositories.NewUserNotificationRepository())

	delegator := models.User{Username: "delegator", Email: "d@test.com", Password: "h"}
	assignee := models.User{Username: "assignee", Email: "a@test.com", Password: "h"}
	require.NoError(t, database.DB.Create(&delegator).Error)
	require.NoError(t, database.DB.Create(&assignee).Error)

	assignedBy := delegator.ID
	task := models.Task{
		Title: "Delegada", Type: models.TaskTypeTrabalho, UserID: assignee.ID,
		AssignedBy: &assignedBy, Completed: true, Priority: models.PriorityMedia,
	}
	require.NoError(t, database.DB.Create(&task).Error)

	err := svc.NotifyDelegatedTaskCompleted(&task, assignee.ID, assignee.Username)
	require.NoError(t, err)

	var n models.UserNotification
	require.NoError(t, database.DB.Where("user_id = ? AND type = ?", delegator.ID, models.UserNotificationTypeTaskCompleted).First(&n).Error)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(n.Payload), &payload))
	assert.Equal(t, float64(task.ID), payload["task_id"])
	assert.Equal(t, "Delegada", payload["task_title"])
	assert.Equal(t, "assignee", payload["completed_by_username"])
}

func TestActivityNotificationService_NotifyDelegatedTaskCompleted_skipsSelfComplete(t *testing.T) {
	setupActivityNotificationTestDB(t)
	svc := NewActivityNotificationService(repositories.NewUserNotificationRepository())

	user := models.User{Username: "solo", Email: "solo@test.com", Password: "h"}
	require.NoError(t, database.DB.Create(&user).Error)

	task := models.Task{
		Title: "Própria", Type: models.TaskTypeCasa, UserID: user.ID,
		Completed: true, Priority: models.PriorityMedia,
	}
	require.NoError(t, database.DB.Create(&task).Error)

	err := svc.NotifyDelegatedTaskCompleted(&task, user.ID, user.Username)
	require.NoError(t, err)

	var count int64
	require.NoError(t, database.DB.Model(&models.UserNotification{}).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestTruncateCommentPreview(t *testing.T) {
	longRunes := make([]rune, 200)
	for i := range longRunes {
		longRunes[i] = 'a'
	}
	out := truncateCommentPreview(string(longRunes))
	assert.True(t, len([]rune(out)) <= maxCommentPreviewRunes+3)
	assert.Contains(t, out, "…")
}
