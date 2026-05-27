package services

import (
	"todo-go-backend/internal/repositories"
)

// StatsService exposes task statistics for the authenticated user.
type StatsService interface {
	GetUserTaskStats(userID uint, hideStaleCompleted bool) (*repositories.UserTaskStats, error)
}

type statsService struct {
	statsRepo repositories.StatsRepository
}

// NewStatsService creates a StatsService.
func NewStatsService(statsRepo repositories.StatsRepository) StatsService {
	return &statsService{statsRepo: statsRepo}
}

func (s *statsService) GetUserTaskStats(userID uint, hideStaleCompleted bool) (*repositories.UserTaskStats, error) {
	return s.statsRepo.GetUserTaskStats(userID, hideStaleCompleted)
}
