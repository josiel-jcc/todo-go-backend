package repositories

import (
	"time"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"
)

// NotificationRepository defines the interface for notification operations
type NotificationRepository interface {
	Create(notification *models.Notification) error
	Exists(userID, taskID uint, notificationType models.NotificationType, channel models.NotificationChannel, date time.Time) (bool, error)
	ExistsTaskReminder(userID, taskID uint, channel models.NotificationChannel, taskDueDate time.Time) (bool, error)
	DeleteByTaskID(taskID uint) error
	FindByUserID(userID uint) ([]models.Notification, error)
}

type notificationRepository struct{}

// NewNotificationRepository creates a new instance of NotificationRepository
func NewNotificationRepository() NotificationRepository {
	return &notificationRepository{}
}

func (r *notificationRepository) Create(notification *models.Notification) error {
	return database.DB.Create(notification).Error
}

// Exists checks if a notification was already sent for a task on a specific date
func (r *notificationRepository) Exists(userID, taskID uint, notificationType models.NotificationType, channel models.NotificationChannel, date time.Time) (bool, error) {
	var count int64
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND task_id = ? AND type = ? AND channel = ? AND sent_at BETWEEN ? AND ?",
			userID, taskID, notificationType, channel, startOfDay, endOfDay).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func normalizeTaskDueDateUTC(t time.Time) time.Time {
	return t.UTC().Truncate(time.Second)
}

// ExistsTaskReminder checks if a task_reminder was already sent for the given due_date fingerprint.
func (r *notificationRepository) ExistsTaskReminder(userID, taskID uint, channel models.NotificationChannel, taskDueDate time.Time) (bool, error) {
	var count int64
	dueUTC := normalizeTaskDueDateUTC(taskDueDate)

	err := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND task_id = ? AND type = ? AND channel = ? AND task_due_date = ?",
			userID, taskID, models.NotificationTypeTaskReminder, channel, dueUTC).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *notificationRepository) DeleteByTaskID(taskID uint) error {
	return database.DB.Where("task_id = ?", taskID).Delete(&models.Notification{}).Error
}

func (r *notificationRepository) FindByUserID(userID uint) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := database.DB.
		Where("user_id = ?", userID).
		Preload("Task").
		Order("sent_at DESC").
		Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

