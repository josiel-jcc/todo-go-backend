package notifications

import (
	"log"
	"time"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"
)

// NotificationService handles notification logic
type NotificationService struct {
	emailService     *EmailService
	telegramService  *TelegramService
	notificationRepo repositories.NotificationRepository
	taskRepo         repositories.TaskRepository
	userRepo         repositories.UserRepository
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	emailService *EmailService,
	telegramService *TelegramService,
	notificationRepo repositories.NotificationRepository,
	taskRepo repositories.TaskRepository,
	userRepo repositories.UserRepository,
) *NotificationService {
	return &NotificationService{
		emailService:     emailService,
		telegramService:  telegramService,
		notificationRepo: notificationRepo,
		taskRepo:         taskRepo,
		userRepo:         userRepo,
	}
}

// CheckAndSendNotifications checks for tasks that need notifications and sends them
func (s *NotificationService) CheckAndSendNotifications() error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)

	log.Printf("Starting notification check at %s", now.Format("2006-01-02 15:04:05"))
	log.Printf("Today: %s, Tomorrow: %s", today.Format("2006-01-02"), tomorrow.Format("2006-01-02"))

	// Get all active tasks (not completed)
	var tasks []models.Task
	if err := database.DB.
		Where("completed = ? AND due_date IS NOT NULL", false).
		Preload("User").
		Find(&tasks).Error; err != nil {
		log.Printf("Error fetching tasks: %v", err)
		return err
	}

	log.Printf("Found %d tasks with due dates", len(tasks))

	processedCount := 0
	skippedCount := 0
	notificationCount := 0

	for _, task := range tasks {
		if task.DueDate == nil {
			log.Printf("Task %d: skipping (no due date)", task.ID)
			skippedCount++
			continue
		}

		dueDate := time.Date(task.DueDate.Year(), task.DueDate.Month(), task.DueDate.Day(), 0, 0, 0, 0, task.DueDate.Location())

		// Check if user has notifications enabled
		if !task.User.NotificationsEnabled {
			log.Printf("Task %d: skipping (user notifications disabled)", task.ID)
			skippedCount++
			continue
		}

		log.Printf("Task %d: due_date=%s, user_id=%d, notifications_enabled=%v",
			task.ID, dueDate.Format("2006-01-02"), task.UserID, task.User.NotificationsEnabled)

		// Check for overdue tasks
		if dueDate.Before(today) {
			log.Printf("Task %d: OVERDUE (due %s)", task.ID, dueDate.Format("2006-01-02"))
			s.sendNotification(&task, models.NotificationTypeOverdue, today)
			notificationCount++
		} else if dueDate.Equal(today) {
			log.Printf("Task %d: DUE TODAY", task.ID)
			s.sendNotification(&task, models.NotificationTypeDueToday, today)
			notificationCount++
		} else if dueDate.Equal(tomorrow) {
			log.Printf("Task %d: DUE SOON (due tomorrow)", task.ID)
			s.sendNotification(&task, models.NotificationTypeDueSoon, today)
			notificationCount++
		} else {
			log.Printf("Task %d: not due yet (due %s)", task.ID, dueDate.Format("2006-01-02"))
		}
		processedCount++
	}

	log.Printf("Notification check completed: %d processed, %d skipped, %d notifications sent", processedCount, skippedCount, notificationCount)
	return nil
}

// sendNotification sends notification via configured channels
func (s *NotificationService) sendNotification(task *models.Task, notificationType models.NotificationType, date time.Time) {
	user := task.User

	// Send email notification
	if user.Email != "" {
		log.Printf("Checking if email notification already sent for task %d, type %s", task.ID, notificationType)
		exists, err := s.notificationRepo.Exists(
			task.UserID,
			task.ID,
			notificationType,
			models.NotificationChannelEmail,
			date,
		)
		if err != nil {
			log.Printf("Error checking email notification existence: %v", err)
		} else if exists {
			log.Printf("Email notification already sent today for task %d, skipping", task.ID)
		} else {
			log.Printf("Sending email notification for task %d to user_id=%d", task.ID, user.ID)
			if err := s.emailService.SendNotification(&user, task, notificationType); err != nil {
				log.Printf("Failed to send email notification: %v", err)
			} else {
				log.Printf("Email notification sent successfully for task %d", task.ID)
				// Record notification
				notification := &models.Notification{
					UserID:  task.UserID,
					TaskID:  task.ID,
					Type:    notificationType,
					Channel: models.NotificationChannelEmail,
					SentAt:  time.Now(),
				}
				if err := s.notificationRepo.Create(notification); err != nil {
					log.Printf("Failed to record email notification: %v", err)
				}
			}
		}
	} else {
		log.Printf("Task %d: user has no email address, skipping email notification", task.ID)
	}

	// Send Telegram notification
	if user.TelegramChatID != nil && *user.TelegramChatID != "" {
		log.Printf("Checking if telegram notification already sent for task %d, type %s", task.ID, notificationType)
		exists, err := s.notificationRepo.Exists(
			task.UserID,
			task.ID,
			notificationType,
			models.NotificationChannelTelegram,
			date,
		)
		if err != nil {
			log.Printf("Error checking telegram notification existence: %v", err)
		} else if exists {
			log.Printf("Telegram notification already sent today for task %d, skipping", task.ID)
		} else {
			log.Printf("Sending telegram notification for task %d to user_id=%d", task.ID, user.ID)
			if err := s.telegramService.SendNotification(*user.TelegramChatID, task, notificationType); err != nil {
				log.Printf("Failed to send telegram notification: %v", err)
			} else {
				log.Printf("Telegram notification sent successfully for task %d", task.ID)
				// Record notification
				notification := &models.Notification{
					UserID:  task.UserID,
					TaskID:  task.ID,
					Type:    notificationType,
					Channel: models.NotificationChannelTelegram,
					SentAt:  time.Now(),
				}
				if err := s.notificationRepo.Create(notification); err != nil {
					log.Printf("Failed to record telegram notification: %v", err)
				}
			}
		}
	} else {
		log.Printf("Task %d: user has no telegram chat ID, skipping telegram notification", task.ID)
	}
}
