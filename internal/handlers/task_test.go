package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"
	"todo-go-backend/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func createTestUser(t *testing.T) (models.User, string) {
	hashedPassword, _ := utils.HashPassword("password123")
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	database.DB.Create(&user)

	token, _, _ := utils.GenerateToken(user.ID, user.Username, "test-secret")
	return user, token
}

func TestCreateTask(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	user, token := createTestUser(t)

	t.Run("Create task successfully", func(t *testing.T) {
		dueDate := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
		reqBody := CreateTaskRequest{
			Title:       "Test Task",
			Description: "Test Description",
			Type:        models.TaskTypeCasa,
			DueDate:     &dueDate,
		}
		jsonValue, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var task models.Task
		json.Unmarshal(w.Body.Bytes(), &task)
		assert.Equal(t, "Test Task", task.Title)
		assert.Equal(t, user.ID, task.UserID)
		assert.Nil(t, task.AssignedBy)
	})

	t.Run("Create task for another user in same group", func(t *testing.T) {
		otherUser := models.User{
			Username: "otheruser",
			Email:    "other@example.com",
			Password: "hashed",
		}
		database.DB.Create(&otherUser)
		createGroupWithMembers(t, user, otherUser.ID)

		reqBody := CreateTaskRequest{
			Title:  "Task for other user",
			Type:   models.TaskTypeTrabalho,
			UserID: &otherUser.ID,
		}
		jsonValue, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var task models.Task
		json.Unmarshal(w.Body.Bytes(), &task)
		assert.Equal(t, otherUser.ID, task.UserID)
		assert.Equal(t, user.ID, *task.AssignedBy)
	})

	t.Run("Create task for user not in shared group", func(t *testing.T) {
		stranger := models.User{
			Username: "stranger",
			Email:    "stranger@example.com",
			Password: "hashed",
		}
		database.DB.Create(&stranger)

		reqBody := CreateTaskRequest{
			Title:  "Task for stranger",
			Type:   models.TaskTypeCasa,
			UserID: &stranger.ID,
		}
		jsonValue, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestShareTask(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	owner, token := createTestUser(t)

	t.Run("Share with user in same group", func(t *testing.T) {
		colleague := models.User{
			Username: "colleague",
			Email:    "colleague@example.com",
			Password: "hashed",
		}
		database.DB.Create(&colleague)
		createGroupWithMembers(t, owner, colleague.ID)

		task := models.Task{
			Title:  "Shareable",
			Type:   models.TaskTypeCasa,
			UserID: owner.ID,
		}
		database.DB.Create(&task)

		body, _ := json.Marshal(ShareTaskRequest{UserIDs: []uint{colleague.ID}})
		req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/tasks/%d/share", task.ID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Share rejected without shared group", func(t *testing.T) {
		stranger := models.User{
			Username: "stranger2",
			Email:    "stranger2@example.com",
			Password: "hashed",
		}
		database.DB.Create(&stranger)

		task := models.Task{
			Title:  "Private",
			Type:   models.TaskTypeCasa,
			UserID: owner.ID,
		}
		database.DB.Create(&task)

		body, _ := json.Marshal(ShareTaskRequest{UserIDs: []uint{stranger.ID}})
		req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/tasks/%d/share", task.ID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetAssignedTasks(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	user, token := createTestUser(t)

	otherUser := models.User{
		Username: "delegatee",
		Email:    "delegatee@example.com",
		Password: "hashed",
	}
	database.DB.Create(&otherUser)

	// Legacy-style self task (assigned_by = creator, user_id = creator) — must not appear
	selfAssigned := user.ID
	selfTask := models.Task{
		Title:      "My own task",
		Type:       models.TaskTypeCasa,
		UserID:     user.ID,
		AssignedBy: &selfAssigned,
	}
	database.DB.Create(&selfTask)

	delegatedTask := models.Task{
		Title:      "Delegated task",
		Type:       models.TaskTypeTrabalho,
		UserID:     otherUser.ID,
		AssignedBy: &user.ID,
	}
	database.DB.Create(&delegatedTask)

	req, _ := http.NewRequest("GET", "/api/v1/tasks/assigned", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	tasks := response["tasks"].([]interface{})
	assert.Len(t, tasks, 1)
	task := tasks[0].(map[string]interface{})
	assert.Equal(t, "Delegated task", task["title"])
}

func TestGetTasks(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	user, token := createTestUser(t)

	// Create test tasks
	task1 := models.Task{
		Title:  "Task 1",
		Type:   models.TaskTypeCasa,
		UserID: user.ID,
	}
	task2 := models.Task{
		Title:     "Task 2",
		Type:      models.TaskTypeTrabalho,
		UserID:    user.ID,
		Completed: true,
	}
	database.DB.Create(&task1)
	database.DB.Create(&task2)

	t.Run("Get all tasks", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["tasks"])
		assert.NotNil(t, response["total"])
		assert.NotNil(t, response["page"])
		assert.NotNil(t, response["limit"])
		tasks := response["tasks"].([]interface{})
		assert.GreaterOrEqual(t, len(tasks), 2)
	})

	t.Run("Filter by type", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/tasks?type=casa", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		tasks := response["tasks"].([]interface{})
		for _, taskInterface := range tasks {
			task := taskInterface.(map[string]interface{})
			assert.Equal(t, "casa", task["type"])
		}
	})

	t.Run("Search tasks", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/tasks?search=Task", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["tasks"])
	})

	t.Run("Pagination", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/tasks?page=1&limit=1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(1), response["page"])
		assert.Equal(t, float64(1), response["limit"])
		tasks := response["tasks"].([]interface{})
		assert.LessOrEqual(t, len(tasks), 1)
	})
}

func TestUpdateTask(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	user, token := createTestUser(t)

	task := models.Task{
		Title:  "Original Title",
		Type:   models.TaskTypeCasa,
		UserID: user.ID,
	}
	database.DB.Create(&task)

	t.Run("Update task successfully", func(t *testing.T) {
		newTitle := "Updated Title"
		completed := true
		reqBody := UpdateTaskRequest{
			Title:     &newTitle,
			Completed: &completed,
		}
		jsonValue, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/api/v1/tasks/"+fmt.Sprintf("%d", task.ID), bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var updatedTask models.Task
		json.Unmarshal(w.Body.Bytes(), &updatedTask)
		assert.Equal(t, "Updated Title", updatedTask.Title)
		assert.True(t, updatedTask.Completed)
	})
}

func TestCompleteRecurringTask(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	user, token := createTestUser(t)

	dueDate := time.Now().Add(24 * time.Hour)
	rule := models.RecurrenceWeekly
	nextDue := dueDate
	task := models.Task{
		Title:             "Weekly chore",
		Type:              models.TaskTypeCasa,
		UserID:            user.ID,
		DueDate:           &dueDate,
		RecurrenceRule:    &rule,
		RecurrenceNextDue: &nextDue,
	}
	database.DB.Create(&task)

	completed := true
	reqBody := UpdateTaskRequest{
		Title:     &task.Title,
		Type:      &task.Type,
		Completed: &completed,
	}
	jsonValue, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/tasks/%d", task.ID), bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var updated models.Task
	json.Unmarshal(w.Body.Bytes(), &updated)
	assert.False(t, updated.Completed)
	assert.NotNil(t, updated.RecurrenceRule)
	assert.Equal(t, models.RecurrenceWeekly, *updated.RecurrenceRule)
	assert.NotNil(t, updated.DueDate)
	assert.True(t, updated.DueDate.After(dueDate))
}

func TestDeleteTask(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")
	user, token := createTestUser(t)

	task := models.Task{
		Title:  "Task to delete",
		Type:   models.TaskTypeCasa,
		UserID: user.ID,
	}
	database.DB.Create(&task)

	req, _ := http.NewRequest("DELETE", "/api/v1/tasks/"+fmt.Sprintf("%d", task.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify task is deleted
	var deletedTask models.Task
	result := database.DB.First(&deletedTask, task.ID)
	assert.Error(t, result.Error)
}

