package repositories

import (
	"testing"
	"time"

	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupStatsTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.Task{},
		&models.TaskSharedWith{},
	))
	database.DB = db
}

func TestStatsRepository_GetUserTaskStats(t *testing.T) {
	setupStatsTestDB(t)
	repo := NewStatsRepository()

	owner := models.User{Username: "owner", Email: "o@test.com", Password: "hash"}
	require.NoError(t, database.DB.Create(&owner).Error)

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())
	yesterday := today.Add(-48 * time.Hour)

	tasks := []models.Task{
		{Title: "Done casa", Type: models.TaskTypeCasa, Priority: models.PriorityMedia, UserID: owner.ID, Completed: true, DueDate: &today},
		{Title: "Pending casa", Type: models.TaskTypeCasa, Priority: models.PriorityAlta, UserID: owner.ID, Completed: false, DueDate: &today},
		{Title: "Overdue", Type: models.TaskTypeTrabalho, Priority: models.PriorityUrgente, UserID: owner.ID, Completed: false, DueDate: &yesterday},
		{Title: "No due", Type: models.TaskTypeLazer, Priority: models.PriorityBaixa, UserID: owner.ID, Completed: false},
	}
	for i := range tasks {
		require.NoError(t, database.DB.Create(&tasks[i]).Error)
	}

	stats, err := repo.GetUserTaskStats(owner.ID, false)
	require.NoError(t, err)

	assert.Equal(t, int64(4), stats.Summary.Total)
	assert.Equal(t, int64(1), stats.Summary.Completed)
	assert.Equal(t, int64(3), stats.Summary.Pending)
	assert.Equal(t, int64(1), stats.Summary.Overdue)
	assert.Equal(t, int64(2), stats.Today.Total)
	assert.Equal(t, int64(1), stats.Today.Completed)
	assert.Equal(t, int64(3), stats.InProgress)
	assert.Len(t, stats.ByType, 3)
	assert.Len(t, stats.ByPriority, 4)
}

func TestStatsRepository_IncludesSharedTasks(t *testing.T) {
	setupStatsTestDB(t)
	repo := NewStatsRepository()

	owner := models.User{Username: "owner", Email: "o2@test.com", Password: "hash"}
	viewer := models.User{Username: "viewer", Email: "v@test.com", Password: "hash"}
	require.NoError(t, database.DB.Create(&owner).Error)
	require.NoError(t, database.DB.Create(&viewer).Error)

	task := models.Task{Title: "Shared", Type: models.TaskTypeCasa, UserID: owner.ID, Priority: models.PriorityMedia}
	require.NoError(t, database.DB.Create(&task).Error)
	require.NoError(t, database.DB.Create(&models.TaskSharedWith{TaskID: task.ID, UserID: viewer.ID}).Error)

	stats, err := repo.GetUserTaskStats(viewer.ID, false)
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats.Summary.Total)
}
