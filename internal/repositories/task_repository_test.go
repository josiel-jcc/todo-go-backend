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

func setupTaskRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Task{}))
	database.DB = db
	return db
}

func TestFindReminderCandidates_filtersWindowAndNotifications(t *testing.T) {
	setupTaskRepoTestDB(t)
	repo := NewTaskRepository()

	now := time.Date(2026, 5, 22, 14, 0, 0, 0, time.UTC)

	enabledUser := &models.User{
		Username:              "enabled",
		Email:                 "enabled@test.com",
		Password:              "hash",
		NotificationsEnabled:  true,
		ReminderMinutesBefore: 10,
	}
	disabledUser := &models.User{
		Username:             "disabled",
		Email:                "disabled@test.com",
		Password:             "hash",
		NotificationsEnabled: false,
	}
	require.NoError(t, database.DB.Create(enabledUser).Error)
	require.NoError(t, database.DB.Create(disabledUser).Error)

	inWindow := now.Add(30 * time.Minute)
	outsideWindow := now.Add(2 * time.Hour)
	tooSoon := now.Add(2 * time.Minute)

	tasks := []models.Task{
		{Title: "in window", UserID: enabledUser.ID, DueDate: &inWindow},
		{Title: "outside window", UserID: enabledUser.ID, DueDate: &outsideWindow},
		{Title: "too soon", UserID: enabledUser.ID, DueDate: &tooSoon},
		{Title: "disabled user", UserID: disabledUser.ID, DueDate: &inWindow},
		{Title: "completed", UserID: enabledUser.ID, DueDate: &inWindow, Completed: true},
	}
	for i := range tasks {
		require.NoError(t, database.DB.Create(&tasks[i]).Error)
	}

	candidates, err := repo.FindReminderCandidates(now)
	require.NoError(t, err)
	require.Len(t, candidates, 1)
	assert.Equal(t, "in window", candidates[0].Title)
	assert.True(t, candidates[0].User.NotificationsEnabled)
}
