package repositories

import (
	"time"

	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"

	"gorm.io/gorm"
)

// TaskStatsBucket holds aggregate counts for a dimension (summary, type, priority).
type TaskStatsBucket struct {
	Total          int64   `json:"total"`
	Completed      int64   `json:"completed"`
	Pending        int64   `json:"pending"`
	Overdue        int64   `json:"overdue"`
	CompletionRate float64 `json:"completion_rate"`
}

// TaskStatsToday holds counts for tasks due today.
type TaskStatsToday struct {
	Total          int64   `json:"total"`
	Completed      int64   `json:"completed"`
	Pending        int64   `json:"pending"`
	CompletionRate float64 `json:"completion_rate"`
}

// TaskStatsByType is per task type.
type TaskStatsByType struct {
	Type           models.TaskType `json:"type"`
	Total          int64           `json:"total"`
	Completed      int64           `json:"completed"`
	Pending        int64           `json:"pending"`
	CompletionRate float64         `json:"completion_rate"`
}

// TaskStatsByPriority is per priority level.
type TaskStatsByPriority struct {
	Priority       models.Priority `json:"priority"`
	Total          int64           `json:"total"`
	Completed      int64           `json:"completed"`
	Pending        int64           `json:"pending"`
	CompletionRate float64         `json:"completion_rate"`
}

// UserTaskStats is the full stats payload for a user.
type UserTaskStats struct {
	Summary    TaskStatsBucket       `json:"summary"`
	Today      TaskStatsToday        `json:"today"`
	ByType     []TaskStatsByType     `json:"by_type"`
	ByPriority []TaskStatsByPriority `json:"by_priority"`
	InProgress int64                 `json:"in_progress"`
}

// StatsRepository loads aggregated task statistics.
type StatsRepository interface {
	GetUserTaskStats(userID uint, hideStaleCompleted bool) (*UserTaskStats, error)
}

type statsRepository struct{}

// NewStatsRepository creates a StatsRepository.
func NewStatsRepository() StatsRepository {
	return &statsRepository{}
}

func (r *statsRepository) visibleTasksQuery(userID uint, hideStaleCompleted bool) *gorm.DB {
	subQuery := database.DB.Table("task_shared_with").Select("task_id").Where("user_id = ?", userID)
	q := database.DB.Model(&models.Task{}).Where("user_id = ? OR id IN (?)", userID, subQuery)
	if hideStaleCompleted {
		cutoff := time.Now().Add(-staleCompletedAge)
		q = q.Where("completed = ? OR completed_at IS NULL OR completed_at >= ?", false, cutoff)
	}
	return q
}

func countQuery(q *gorm.DB) (int64, error) {
	var n int64
	err := q.Count(&n).Error
	return n, err
}

func completionRate(completed, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(completed) / float64(total) * 100
}

func (r *statsRepository) GetUserTaskStats(userID uint, hideStaleCompleted bool) (*UserTaskStats, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)

	total, err := countQuery(r.visibleTasksQuery(userID, hideStaleCompleted))
	if err != nil {
		return nil, err
	}
	completed, err := countQuery(r.visibleTasksQuery(userID, hideStaleCompleted).Where("completed = ?", true))
	if err != nil {
		return nil, err
	}
	overdue, err := countQuery(r.visibleTasksQuery(userID, hideStaleCompleted).
		Where("completed = ? AND due_date IS NOT NULL AND due_date < ?", false, now))
	if err != nil {
		return nil, err
	}
	summary := TaskStatsBucket{
		Total:          total,
		Completed:      completed,
		Pending:        total - completed,
		Overdue:        overdue,
		CompletionRate: completionRate(completed, total),
	}

	todayTotal, err := countQuery(r.visibleTasksQuery(userID, hideStaleCompleted).
		Where("due_date >= ? AND due_date < ?", todayStart, todayEnd))
	if err != nil {
		return nil, err
	}
	todayCompleted, err := countQuery(r.visibleTasksQuery(userID, hideStaleCompleted).
		Where("due_date >= ? AND due_date < ? AND completed = ?", todayStart, todayEnd, true))
	if err != nil {
		return nil, err
	}
	today := TaskStatsToday{
		Total:          todayTotal,
		Completed:      todayCompleted,
		Pending:        todayTotal - todayCompleted,
		CompletionRate: completionRate(todayCompleted, todayTotal),
	}

	inProgress, err := countQuery(r.visibleTasksQuery(userID, hideStaleCompleted).Where("completed = ?", false))
	if err != nil {
		return nil, err
	}

	byType, err := r.aggregateByType(userID, hideStaleCompleted)
	if err != nil {
		return nil, err
	}
	byPriority, err := r.aggregateByPriority(userID, hideStaleCompleted)
	if err != nil {
		return nil, err
	}

	return &UserTaskStats{
		Summary:    summary,
		Today:      today,
		ByType:     byType,
		ByPriority: byPriority,
		InProgress: inProgress,
	}, nil
}

type typeAggRow struct {
	Type      models.TaskType
	Total     int64
	Completed int64
}

func (r *statsRepository) aggregateByType(userID uint, hideStaleCompleted bool) ([]TaskStatsByType, error) {
	var rows []typeAggRow
	err := r.visibleTasksQuery(userID, hideStaleCompleted).
		Select("type, COUNT(*) as total, SUM(CASE WHEN completed = 1 THEN 1 ELSE 0 END) as completed").
		Group("type").
		Order("type ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]TaskStatsByType, 0, len(rows))
	for _, row := range rows {
		out = append(out, TaskStatsByType{
			Type:           row.Type,
			Total:          row.Total,
			Completed:      row.Completed,
			Pending:        row.Total - row.Completed,
			CompletionRate: completionRate(row.Completed, row.Total),
		})
	}
	return out, nil
}

type priorityAggRow struct {
	Priority  models.Priority
	Total     int64
	Completed int64
}

func (r *statsRepository) aggregateByPriority(userID uint, hideStaleCompleted bool) ([]TaskStatsByPriority, error) {
	var rows []priorityAggRow
	err := r.visibleTasksQuery(userID, hideStaleCompleted).
		Select("priority, COUNT(*) as total, SUM(CASE WHEN completed = 1 THEN 1 ELSE 0 END) as completed").
		Group("priority").
		Order("priority ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]TaskStatsByPriority, 0, len(rows))
	for _, row := range rows {
		out = append(out, TaskStatsByPriority{
			Priority:       row.Priority,
			Total:          row.Total,
			Completed:      row.Completed,
			Pending:        row.Total - row.Completed,
			CompletionRate: completionRate(row.Completed, row.Total),
		})
	}
	return out, nil
}
