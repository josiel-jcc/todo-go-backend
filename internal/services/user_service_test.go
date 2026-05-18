package services

import (
	"testing"
	"todo-go-backend/internal/errors"
	"todo-go-backend/internal/models"
	"todo-go-backend/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestUserService_DeleteAccount_InvalidPassword(t *testing.T) {
	mockRepo := NewMockUserRepository()
	hashed, _ := utils.HashPassword("correct-password")
	_ = mockRepo.Create(&models.User{
		Username: "user1",
		Email:    "user1@example.com",
		Password: hashed,
	})

	service := NewUserService(mockRepo)
	err := service.DeleteAccount(1, "wrong-password")

	assert.Error(t, err)
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrInvalidCredentials, appErr.Err)
}
