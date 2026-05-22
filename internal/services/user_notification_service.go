package services

import (
	"todo-go-backend/internal/errors"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"
)

type UserNotificationService interface {
	List(userID uint, unreadOnly bool, page, limit int) ([]models.UserNotification, int64, error)
	UnreadCount(userID uint) (int64, error)
	MarkRead(userID, notificationID uint) error
	MarkAllRead(userID uint) error
}

type userNotificationService struct {
	repo repositories.UserNotificationRepository
}

func NewUserNotificationService(repo repositories.UserNotificationRepository) UserNotificationService {
	return &userNotificationService{repo: repo}
}

func (s *userNotificationService) List(userID uint, unreadOnly bool, page, limit int) ([]models.UserNotification, int64, error) {
	list, total, err := s.repo.ListByUserID(userID, unreadOnly, page, limit)
	if err != nil {
		return nil, 0, errors.NewInternalServerError(err)
	}
	return list, total, nil
}

func (s *userNotificationService) UnreadCount(userID uint) (int64, error) {
	count, err := s.repo.CountUnread(userID)
	if err != nil {
		return 0, errors.NewInternalServerError(err)
	}
	return count, nil
}

func (s *userNotificationService) MarkRead(userID, notificationID uint) error {
	n, err := s.repo.FindByID(notificationID)
	if err != nil || n.UserID != userID {
		return errors.NewForbiddenError()
	}
	return s.repo.MarkRead(notificationID, userID)
}

func (s *userNotificationService) MarkAllRead(userID uint) error {
	return s.repo.MarkAllRead(userID)
}
