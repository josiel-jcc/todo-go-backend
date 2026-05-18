package services

import (
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/errors"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"
	"todo-go-backend/pkg/utils"

	"gorm.io/gorm"
)

type UserService interface {
	DeleteAccount(userID uint, password string) error
	ExportAccount(userID uint) (map[string]interface{}, error)
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) DeleteAccount(userID uint, password string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.NewUserNotFoundError()
	}
	if !utils.CheckPasswordHash(password, user.Password) {
		return errors.NewInvalidCredentialsError()
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.Notification{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&models.Tag{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&models.TaskSharedWith{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&models.Task{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Task{}).Where("assigned_by = ?", userID).Update("assigned_by", nil).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&models.User{}, userID).Error
	})
}

func (s *userService) ExportAccount(userID uint) (map[string]interface{}, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.NewUserNotFoundError()
	}

	var tasks []models.Task
	if err := database.DB.Where("user_id = ?", userID).Preload("Tags").Find(&tasks).Error; err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	var sharedTasks []models.Task
	subQuery := database.DB.Table("task_shared_with").Select("task_id").Where("user_id = ?", userID)
	if err := database.DB.Where("id IN (?)", subQuery).Preload("Tags").Find(&sharedTasks).Error; err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	var tags []models.Tag
	if err := database.DB.Where("user_id = ?", userID).Find(&tags).Error; err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	var comments []models.Comment
	if err := database.DB.Where("user_id = ?", userID).Find(&comments).Error; err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	profile := map[string]interface{}{
		"id":                    user.ID,
		"username":              user.Username,
		"email":                 user.Email,
		"notifications_enabled": user.NotificationsEnabled,
		"telegram_chat_id":      user.TelegramChatID,
		"terms_accepted_at":     user.TermsAcceptedAt,
		"created_at":            user.CreatedAt,
		"updated_at":            user.UpdatedAt,
	}

	return map[string]interface{}{
		"exported_at":    user.UpdatedAt,
		"profile":        profile,
		"tasks":          tasks,
		"shared_tasks":   sharedTasks,
		"tags":           tags,
		"comments":       comments,
	}, nil
}
