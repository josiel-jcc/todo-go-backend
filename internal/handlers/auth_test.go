package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"
	"todo-go-backend/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")

	t.Run("Successful registration", func(t *testing.T) {
		reqBody := RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonValue, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "User created successfully", response["message"])
		cookies := w.Result().Cookies()
		var hasAuthCookie bool
		for _, cookie := range cookies {
			if cookie.Name == "auth_token" && cookie.HttpOnly {
				hasAuthCookie = true
			}
		}
		assert.True(t, hasAuthCookie)
	})

	t.Run("Duplicate username", func(t *testing.T) {
		reqBody := RegisterRequest{
			Username: "testuser",
			Email:    "test2@example.com",
			Password: "password123",
		}
		jsonValue, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid email", func(t *testing.T) {
		reqBody := RegisterRequest{
			Username: "newuser",
			Email:    "invalid-email",
			Password: "password123",
		}
		jsonValue, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLogin(t *testing.T) {
	setupTestDB()
	router := setupTestRouter("test-secret")

	// Create a test user first
	hashedPassword, _ := utils.HashPassword("password123")
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	database.DB.Create(&user)

	t.Run("Successful login", func(t *testing.T) {
		reqBody := LoginRequest{
			Username: "testuser",
			Password: "password123",
		}
		jsonValue, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		cookies := w.Result().Cookies()
		var hasAuthCookie bool
		for _, cookie := range cookies {
			if cookie.Name == "auth_token" {
				hasAuthCookie = true
			}
		}
		assert.True(t, hasAuthCookie)
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		reqBody := LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}
		jsonValue, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

