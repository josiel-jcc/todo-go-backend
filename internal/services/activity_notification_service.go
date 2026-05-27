package services

import (
	"encoding/json"
	"strings"

	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"
)

const maxCommentPreviewRunes = 120

// ActivityNotificationService sends in-app notifications for task activity.
type ActivityNotificationService interface {
	NotifyTaskComment(task *models.Task, comment *models.Comment, commenterUsername string) error
	NotifyDelegatedTaskCompleted(task *models.Task, completedByUserID uint, completedByUsername string) error
}

type activityNotificationService struct {
	userNotificationRepo repositories.UserNotificationRepository
}

// NewActivityNotificationService creates an ActivityNotificationService.
func NewActivityNotificationService(userNotificationRepo repositories.UserNotificationRepository) ActivityNotificationService {
	return &activityNotificationService{userNotificationRepo: userNotificationRepo}
}

func (s *activityNotificationService) NotifyTaskComment(
	task *models.Task,
	comment *models.Comment,
	commenterUsername string,
) error {
	if task == nil || comment == nil {
		return nil
	}

	recipients := taskActivityRecipientIDs(task, comment.UserID)
	if len(recipients) == 0 {
		return nil
	}

	payload, err := json.Marshal(map[string]interface{}{
		"task_id":           task.ID,
		"task_title":        task.Title,
		"comment_id":        comment.ID,
		"comment_preview":   truncateCommentPreview(comment.Content),
		"author_username":   commenterUsername,
	})
	if err != nil {
		return err
	}

	return s.createForRecipients(recipients, models.UserNotificationTypeTaskComment, string(payload))
}

func (s *activityNotificationService) NotifyDelegatedTaskCompleted(
	task *models.Task,
	completedByUserID uint,
	completedByUsername string,
) error {
	if task == nil || task.AssignedBy == nil {
		return nil
	}
	delegatorID := *task.AssignedBy
	if delegatorID == completedByUserID || task.UserID != completedByUserID {
		return nil
	}

	payload, err := json.Marshal(map[string]interface{}{
		"task_id":                 task.ID,
		"task_title":              task.Title,
		"completed_by_username":   completedByUsername,
		"completed_by_user_id":    completedByUserID,
	})
	if err != nil {
		return err
	}

	return s.userNotificationRepo.Create(&models.UserNotification{
		UserID:  delegatorID,
		Type:    models.UserNotificationTypeTaskCompleted,
		Payload: string(payload),
		Read:    false,
	})
}

func taskActivityRecipientIDs(task *models.Task, excludeUserID uint) []uint {
	seen := make(map[uint]struct{})
	add := func(id uint) {
		if id == 0 || id == excludeUserID {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
	}

	add(task.UserID)
	if task.AssignedBy != nil {
		add(*task.AssignedBy)
	}
	for _, u := range task.SharedWithUsers {
		add(u.ID)
	}

	out := make([]uint, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out
}

func (s *activityNotificationService) createForRecipients(
	recipientIDs []uint,
	typ models.UserNotificationType,
	payload string,
) error {
	for _, userID := range recipientIDs {
		if err := s.userNotificationRepo.Create(&models.UserNotification{
			UserID:  userID,
			Type:    typ,
			Payload: payload,
			Read:    false,
		}); err != nil {
			return err
		}
	}
	return nil
}

func truncateCommentPreview(content string) string {
	content = strings.TrimSpace(content)
	runes := []rune(content)
	if len(runes) <= maxCommentPreviewRunes {
		return content
	}
	return string(runes[:maxCommentPreviewRunes]) + "…"
}
