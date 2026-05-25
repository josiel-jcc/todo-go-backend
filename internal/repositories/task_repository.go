package repositories

import (
	"time"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"
)

// TaskRepository defines the interface for task operations
type TaskRepository interface {
	Create(task *models.Task) error
	FindByID(id uint) (*models.Task, error)
	FindByUserID(userID uint, filters *TaskFilters) ([]models.Task, int64, error)
	FindByAssignedBy(assignedByID uint, filters *TaskFilters) ([]models.Task, int64, error)
	Update(task *models.Task) error
	Delete(id uint) error
	Exists(id uint) (bool, error)
	AddSharedWith(taskID, userID uint) error
	RemoveSharedWith(taskID, userID uint) error
	UserCanAccessTask(taskID, userID uint) (bool, error)
	FindReminderCandidates(now time.Time) ([]models.Task, error)
}

const staleCompletedAge = 24 * time.Hour

// TaskFilters defines filters for task search
type TaskFilters struct {
	Type                *models.TaskType
	Completed           *bool
	HideStaleCompleted  bool
	Priority            *models.Priority
	Search       *string // Search in title and description
	DueDateFrom  *time.Time
	DueDateTo    *time.Time
	AssignedBy   *uint
	TagIDs       []uint  // Filter by tag IDs
	Page         int
	Limit        int
	SortBy       string // created_at, due_date, title, priority
	Order        string // asc, desc
}

type taskRepository struct{}

// NewTaskRepository creates a new instance of TaskRepository
func NewTaskRepository() TaskRepository {
	return &taskRepository{}
}

func (r *taskRepository) Create(task *models.Task) error {
	return database.DB.Create(task).Error
}

func (r *taskRepository) FindByID(id uint) (*models.Task, error) {
	var task models.Task
	if err := database.DB.
		Preload("User").
		Preload("AssignedByUser").
		Preload("SharedWithUsers").
		Preload("Tags").
		First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) FindByUserID(userID uint, filters *TaskFilters) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	// Base query: tasks owned by user OR shared with user
	subQuery := database.DB.Table("task_shared_with").Select("task_id").Where("user_id = ?", userID)
	query := database.DB.Model(&models.Task{}).Where("user_id = ? OR id IN (?)", userID, subQuery)

	// Apply filters
	if filters != nil {
		if filters.Type != nil {
			query = query.Where("type = ?", *filters.Type)
		}
		if filters.Completed != nil {
			query = query.Where("completed = ?", *filters.Completed)
		}
		if filters.HideStaleCompleted {
			cutoff := time.Now().Add(-staleCompletedAge)
			query = query.Where("completed = ? OR completed_at IS NULL OR completed_at >= ?", false, cutoff)
		}
		if filters.Priority != nil {
			query = query.Where("priority = ?", *filters.Priority)
		}
		if filters.Search != nil && *filters.Search != "" {
			searchPattern := "%" + *filters.Search + "%"
			query = query.Where("(title LIKE ? OR description LIKE ?)", searchPattern, searchPattern)
		}
		if filters.DueDateFrom != nil {
			query = query.Where("due_date >= ?", *filters.DueDateFrom)
		}
		if filters.DueDateTo != nil {
			query = query.Where("due_date <= ?", *filters.DueDateTo)
		}
		if filters.AssignedBy != nil {
			query = query.Where("assigned_by = ?", *filters.AssignedBy)
		}
		// Filter by tags (tasks that have ALL specified tags)
		if len(filters.TagIDs) > 0 {
			query = query.Joins("JOIN task_tags ON tasks.id = task_tags.task_id").
				Where("task_tags.tag_id IN ?", filters.TagIDs).
				Group("tasks.id").
				Having("COUNT(DISTINCT task_tags.tag_id) = ?", len(filters.TagIDs))
		}
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "created_at"
	order := "DESC"
	if filters != nil {
		if filters.SortBy != "" {
			validSortFields := map[string]bool{
				"created_at": true,
				"due_date":   true,
				"title":      true,
				"priority":   true,
			}
			if validSortFields[filters.SortBy] {
				sortBy = filters.SortBy
			}
		}
		if filters.Order != "" {
			if filters.Order == "asc" || filters.Order == "desc" {
				order = filters.Order
			}
		}
	}
	query = query.Order(sortBy + " " + order)

	// Apply pagination
	if filters != nil && filters.Limit > 0 {
		query = query.Limit(filters.Limit)
		if filters.Page > 0 {
			offset := (filters.Page - 1) * filters.Limit
			query = query.Offset(offset)
		}
	}

	// Execute query with preloads
	if err := query.Preload("User").Preload("AssignedByUser").Preload("SharedWithUsers").Preload("Tags").Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *taskRepository) FindByAssignedBy(assignedByID uint, filters *TaskFilters) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	// Tasks delegated to others (exclude self-assigned, including legacy rows)
	query := database.DB.Model(&models.Task{}).
		Where("assigned_by = ?", assignedByID).
		Where("user_id != ?", assignedByID)

	// Apply filters
	if filters != nil {
		if filters.Type != nil {
			query = query.Where("type = ?", *filters.Type)
		}
		if filters.Completed != nil {
			query = query.Where("completed = ?", *filters.Completed)
		}
		if filters.HideStaleCompleted {
			cutoff := time.Now().Add(-staleCompletedAge)
			query = query.Where("completed = ? OR completed_at IS NULL OR completed_at >= ?", false, cutoff)
		}
		if filters.Priority != nil {
			query = query.Where("priority = ?", *filters.Priority)
		}
		if filters.Search != nil && *filters.Search != "" {
			searchPattern := "%" + *filters.Search + "%"
			query = query.Where("(title LIKE ? OR description LIKE ?)", searchPattern, searchPattern)
		}
		if filters.DueDateFrom != nil {
			query = query.Where("due_date >= ?", *filters.DueDateFrom)
		}
		if filters.DueDateTo != nil {
			query = query.Where("due_date <= ?", *filters.DueDateTo)
		}
		// Filter by tags (tasks that have ALL specified tags)
		if len(filters.TagIDs) > 0 {
			query = query.Joins("JOIN task_tags ON tasks.id = task_tags.task_id").
				Where("task_tags.tag_id IN ?", filters.TagIDs).
				Group("tasks.id").
				Having("COUNT(DISTINCT task_tags.tag_id) = ?", len(filters.TagIDs))
		}
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "created_at"
	order := "DESC"
	if filters != nil {
		if filters.SortBy != "" {
			validSortFields := map[string]bool{
				"created_at": true,
				"due_date":   true,
				"title":      true,
				"priority":   true,
			}
			if validSortFields[filters.SortBy] {
				sortBy = filters.SortBy
			}
		}
		if filters.Order != "" {
			if filters.Order == "asc" || filters.Order == "desc" {
				order = filters.Order
			}
		}
	}
	query = query.Order(sortBy + " " + order)

	// Apply pagination
	if filters != nil && filters.Limit > 0 {
		query = query.Limit(filters.Limit)
		if filters.Page > 0 {
			offset := (filters.Page - 1) * filters.Limit
			query = query.Offset(offset)
		}
	}

	// Execute query with preloads
	if err := query.Preload("User").Preload("AssignedByUser").Preload("SharedWithUsers").Preload("Tags").Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *taskRepository) AddSharedWith(taskID, userID uint) error {
	// FirstOrCreate avoids duplicate (DB-agnostic)
	return database.DB.Where(models.TaskSharedWith{TaskID: taskID, UserID: userID}).
		FirstOrCreate(&models.TaskSharedWith{TaskID: taskID, UserID: userID}).Error
}

func (r *taskRepository) RemoveSharedWith(taskID, userID uint) error {
	return database.DB.Delete(&models.TaskSharedWith{}, "task_id = ? AND user_id = ?", taskID, userID).Error
}

func (r *taskRepository) UserCanAccessTask(taskID, userID uint) (bool, error) {
	var task models.Task
	if err := database.DB.Select("id", "user_id", "assigned_by").First(&task, taskID).Error; err != nil {
		return false, err
	}
	if task.UserID == userID {
		return true, nil
	}
	if task.AssignedBy != nil && *task.AssignedBy == userID {
		return true, nil
	}
	var count int64
	if err := database.DB.Table("task_shared_with").Where("task_id = ? AND user_id = ?", taskID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *taskRepository) Update(task *models.Task) error {
	return database.DB.Save(task).Error
}

func (r *taskRepository) Delete(id uint) error {
	return database.DB.Delete(&models.Task{}, id).Error
}

func (r *taskRepository) Exists(id uint) (bool, error) {
	var count int64
	if err := database.DB.Model(&models.Task{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindReminderCandidates returns incomplete tasks with due_date in the scheduler window
// for users who have notifications enabled.
func (r *taskRepository) FindReminderCandidates(now time.Time) ([]models.Task, error) {
	nowUTC := now.UTC()
	windowStart := nowUTC.Add(4 * time.Minute)
	windowEnd := nowUTC.Add(61 * time.Minute)

	var tasks []models.Task
	err := database.DB.
		Joins("JOIN users ON users.id = tasks.user_id").
		Where("tasks.completed = ?", false).
		Where("tasks.due_date IS NOT NULL").
		Where("users.notifications_enabled = ?", true).
		Where("tasks.due_date >= ? AND tasks.due_date <= ?", windowStart, windowEnd).
		Preload("User").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

