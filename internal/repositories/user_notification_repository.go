package repositories

import (
	"fmt"
	"time"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"
)

type UserNotificationRepository interface {
	Create(n *models.UserNotification) error
	ExistsTaskReminder(userID, taskID uint, dueDate time.Time) (bool, error)
	FindByID(id uint) (*models.UserNotification, error)
	ListByUserID(userID uint, unreadOnly bool, page, limit int) ([]models.UserNotification, int64, error)
	CountUnread(userID uint) (int64, error)
	MarkRead(id, userID uint) error
	MarkAllRead(userID uint) error
	MarkReadByInvitationID(userID uint, invitationID uint) error
	DeleteByUserID(userID uint) error
}

type userNotificationRepository struct{}

func NewUserNotificationRepository() UserNotificationRepository {
	return &userNotificationRepository{}
}

func (r *userNotificationRepository) Create(n *models.UserNotification) error {
	return database.DB.Create(n).Error
}

// ExistsTaskReminder reports whether an in-app task_reminder already exists for the due_date snapshot.
func (r *userNotificationRepository) ExistsTaskReminder(userID, taskID uint, dueDate time.Time) (bool, error) {
	dueStr := dueDate.UTC().Format(time.RFC3339)
	var count int64
	err := database.DB.Model(&models.UserNotification{}).
		Where("user_id = ? AND type = ?", userID, models.UserNotificationTypeTaskReminder).
		Where("payload LIKE ? AND payload LIKE ?",
			fmt.Sprintf(`%%"task_id":%d%%`, taskID),
			fmt.Sprintf(`%%"due_date":"%s"%%`, dueStr)).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userNotificationRepository) FindByID(id uint) (*models.UserNotification, error) {
	var n models.UserNotification
	if err := database.DB.First(&n, id).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *userNotificationRepository) ListByUserID(userID uint, unreadOnly bool, page, limit int) ([]models.UserNotification, int64, error) {
	query := database.DB.Model(&models.UserNotification{}).Where("user_id = ?", userID)
	if unreadOnly {
		query = query.Where("`read` = ?", false)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	var list []models.UserNotification
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *userNotificationRepository) CountUnread(userID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.UserNotification{}).
		Where("user_id = ? AND `read` = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *userNotificationRepository) MarkRead(id, userID uint) error {
	return database.DB.Model(&models.UserNotification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("`read`", true).Error
}

func (r *userNotificationRepository) MarkAllRead(userID uint) error {
	return database.DB.Model(&models.UserNotification{}).
		Where("user_id = ?", userID).
		Update("`read`", true).Error
}

func (r *userNotificationRepository) MarkReadByInvitationID(userID uint, invitationID uint) error {
	pattern := fmt.Sprintf("%%\"invitation_id\":%d%%", invitationID)
	return database.DB.Model(&models.UserNotification{}).
		Where("user_id = ? AND type = ? AND `read` = ? AND payload LIKE ?",
			userID, models.UserNotificationTypeGroupInvite, false, pattern).
		Update("`read`", true).Error
}

func (r *userNotificationRepository) DeleteByUserID(userID uint) error {
	return database.DB.Where("user_id = ?", userID).Delete(&models.UserNotification{}).Error
}
