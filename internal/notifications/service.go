package notifications

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"
)

const maxConcurrentReminderSends = 10

// NotificationService handles notification logic
type NotificationService struct {
	emailService         *EmailService
	telegramService      *TelegramService
	pushService          *PushService
	notificationRepo     repositories.NotificationRepository
	userNotificationRepo repositories.UserNotificationRepository
	taskRepo             repositories.TaskRepository
	userRepo             repositories.UserRepository
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	emailService *EmailService,
	telegramService *TelegramService,
	pushService *PushService,
	notificationRepo repositories.NotificationRepository,
	userNotificationRepo repositories.UserNotificationRepository,
	taskRepo repositories.TaskRepository,
	userRepo repositories.UserRepository,
) *NotificationService {
	return &NotificationService{
		emailService:         emailService,
		telegramService:      telegramService,
		pushService:          pushService,
		notificationRepo:     notificationRepo,
		userNotificationRepo: userNotificationRepo,
		taskRepo:             taskRepo,
		userRepo:             userRepo,
	}
}

// effectiveReminderMinutes returns task override or user default reminder offset.
func effectiveReminderMinutes(task models.Task, user models.User) int {
	if task.ReminderMinutesBefore != nil {
		return *task.ReminderMinutesBefore
	}
	return user.ReminderMinutesBefore
}

// shouldSendInWindow is true when reminderAt falls in [now-1min, now).
func shouldSendInWindow(reminderAt, now time.Time) bool {
	nowUTC := now.UTC()
	reminderUTC := reminderAt.UTC()
	windowStart := nowUTC.Add(-1 * time.Minute)
	return !reminderUTC.Before(windowStart) && reminderUTC.Before(nowUTC)
}

func taskDueDateFingerprint(dueDate time.Time) time.Time {
	return dueDate.UTC().Truncate(time.Second)
}

// CheckAndSendNotifications runs the timed task_reminder scheduler tick.
func (s *NotificationService) CheckAndSendNotifications() error {
	start := time.Now()
	now := start.UTC()

	candidates, err := s.taskRepo.FindReminderCandidates(now)
	if err != nil {
		return err
	}

	sem := make(chan struct{}, maxConcurrentReminderSends)
	var wg sync.WaitGroup
	var mu sync.Mutex
	sent := 0

	for i := range candidates {
		task := &candidates[i]
		if task.DueDate == nil {
			continue
		}

		offset := effectiveReminderMinutes(*task, task.User)
		reminderAt := task.DueDate.UTC().Add(-time.Duration(offset) * time.Minute)
		if !shouldSendInWindow(reminderAt, now) {
			continue
		}

		wg.Add(1)
		go func(t *models.Task, minutesBefore int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			n := s.sendTaskReminder(t, minutesBefore)
			mu.Lock()
			sent += n
			mu.Unlock()
		}(task, offset)
	}

	wg.Wait()

	elapsed := time.Since(start).Milliseconds()
	log.Printf("notification tick: candidates=%d sent=%d elapsed_ms=%d", len(candidates), sent, elapsed)
	return nil
}

// sendTaskReminder sends task_reminder on configured channels (email, telegram, push, in-app).
func (s *NotificationService) sendTaskReminder(task *models.Task, minutesBefore int) int {
	sent := 0
	user := task.User
	dueFP := taskDueDateFingerprint(*task.DueDate)
	notificationType := models.NotificationTypeTaskReminder

	if user.Email != "" {
		exists, err := s.notificationRepo.ExistsTaskReminder(
			task.UserID, task.ID, models.NotificationChannelEmail, *task.DueDate,
		)
		if err != nil {
			log.Printf("Error checking email task_reminder for task %d: %v", task.ID, err)
		} else if !exists {
			if err := s.emailService.SendNotification(&user, task, notificationType, minutesBefore); err != nil {
				log.Printf("Failed to send email task_reminder for task %d: %v", task.ID, err)
			} else if s.recordTaskReminder(task, models.NotificationChannelEmail, dueFP) {
				sent++
			}
		}
	}

	if user.TelegramChatID != nil && *user.TelegramChatID != "" {
		exists, err := s.notificationRepo.ExistsTaskReminder(
			task.UserID, task.ID, models.NotificationChannelTelegram, *task.DueDate,
		)
		if err != nil {
			log.Printf("Error checking telegram task_reminder for task %d: %v", task.ID, err)
		} else if !exists {
			if err := s.telegramService.SendNotification(*user.TelegramChatID, task, notificationType, minutesBefore); err != nil {
				log.Printf("Failed to send telegram task_reminder for task %d: %v", task.ID, err)
			} else if s.recordTaskReminder(task, models.NotificationChannelTelegram, dueFP) {
				sent++
			}
		}
	}

	exists, err := s.notificationRepo.ExistsTaskReminder(
		task.UserID, task.ID, models.NotificationChannelPush, *task.DueDate,
	)
	if err != nil {
		log.Printf("Error checking push task_reminder for task %d: %v", task.ID, err)
	} else if !exists {
		payload := PushPayload{
			Title: "Lembrete de tarefa",
			Body:  fmt.Sprintf("%s vence em %d minutos", task.Title, minutesBefore),
			URL:   fmt.Sprintf("/tasks/%d", task.ID),
		}
		if err := s.pushService.SendToUser(task.UserID, payload); err != nil {
			log.Printf("Failed to send push task_reminder for task %d: %v", task.ID, err)
		} else if s.recordTaskReminder(task, models.NotificationChannelPush, dueFP) {
			sent++
		}
	}

	if created, err := s.sendInAppTaskReminder(task, minutesBefore); err != nil {
		log.Printf("Failed to create in-app task_reminder for task %d: %v", task.ID, err)
	} else if created {
		sent++
	}

	return sent
}

func (s *NotificationService) sendInAppTaskReminder(task *models.Task, minutesBefore int) (bool, error) {
	exists, err := s.userNotificationRepo.ExistsTaskReminder(task.UserID, task.ID, *task.DueDate)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	payload, err := json.Marshal(map[string]interface{}{
		"task_id":        task.ID,
		"title":          task.Title,
		"due_date":       task.DueDate.UTC().Format(time.RFC3339),
		"minutes_before": minutesBefore,
	})
	if err != nil {
		return false, fmt.Errorf("marshal in-app payload: %w", err)
	}

	err = s.userNotificationRepo.Create(&models.UserNotification{
		UserID:  task.UserID,
		Type:    models.UserNotificationTypeTaskReminder,
		Payload: string(payload),
		Read:    false,
	})
	return err == nil, err
}

func (s *NotificationService) recordTaskReminder(task *models.Task, channel models.NotificationChannel, dueFP time.Time) bool {
	notification := &models.Notification{
		UserID:      task.UserID,
		TaskID:      task.ID,
		Type:        models.NotificationTypeTaskReminder,
		Channel:     channel,
		TaskDueDate: &dueFP,
		SentAt:      time.Now().UTC(),
	}
	if err := s.notificationRepo.Create(notification); err != nil {
		log.Printf("Failed to record %s task_reminder for task %d: %v", channel, task.ID, err)
		return false
	}
	return true
}
