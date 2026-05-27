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
	userRepo             repositories.UserRepository
	groupRepo            repositories.GroupRepository
	groupInvitationRepo  repositories.GroupInvitationRepository
	userNotificationRepo repositories.UserNotificationRepository
}

func NewUserService(
	userRepo repositories.UserRepository,
	groupRepo repositories.GroupRepository,
	groupInvitationRepo repositories.GroupInvitationRepository,
	userNotificationRepo repositories.UserNotificationRepository,
) UserService {
	return &userService{
		userRepo:             userRepo,
		groupRepo:            groupRepo,
		groupInvitationRepo:  groupInvitationRepo,
		userNotificationRepo: userNotificationRepo,
	}
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
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserNotification{}).Error; err != nil {
			return err
		}
		if err := tx.Where("invited_user_id = ? OR invited_by_user_id = ?", userID, userID).
			Delete(&models.GroupInvitation{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&models.GroupMember{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&models.PushSubscription{}).Error; err != nil {
			return err
		}
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

	var groups []models.Group
	_ = database.DB.
		Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ?", userID).
		Find(&groups).Error

	var invitations []models.GroupInvitation
	_ = database.DB.Where("invited_user_id = ? OR invited_by_user_id = ?", userID, userID).Find(&invitations).Error

	var pushSubscriptions []models.PushSubscription
	if err := database.DB.Where("user_id = ?", userID).Find(&pushSubscriptions).Error; err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	var userNotifications []models.UserNotification
	_ = database.DB.Where("user_id = ?", userID).Find(&userNotifications).Error

	profile := map[string]interface{}{
		"id":                      user.ID,
		"username":                user.Username,
		"email":                   user.Email,
		"notifications_enabled":   user.NotificationsEnabled,
		"reminder_minutes_before": user.ReminderMinutesBefore,
		"telegram_chat_id":        user.TelegramChatID,
		"terms_accepted_at":       user.TermsAcceptedAt,
		"created_at":              user.CreatedAt,
		"updated_at":              user.UpdatedAt,
	}

	return map[string]interface{}{
		"exported_at":          user.UpdatedAt,
		"profile":              profile,
		"tasks":                tasks,
		"shared_tasks":         sharedTasks,
		"tags":                 tags,
		"comments":             comments,
		"groups":               groups,
		"group_invitations":    invitations,
		"push_subscriptions":   pushSubscriptions,
		"user_notifications":   userNotifications,
	}, nil
}
