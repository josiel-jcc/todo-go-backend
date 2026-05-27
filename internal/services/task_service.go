package services

import (
	"time"
	"todo-go-backend/internal/errors"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"
)

// TaskService defines the interface for task operations
type TaskService interface {
	Create(userID uint, req *CreateTaskRequest) (*models.Task, error)
	GetByID(userID, taskID uint) (*models.Task, error)
	GetByUserID(userID uint, filters *TaskFilters) (*PaginatedTasksResponse, error)
	GetAssignedByUser(assignedByID uint, filters *TaskFilters) (*PaginatedTasksResponse, error)
	Update(userID, taskID uint, req *UpdateTaskRequest) (*models.Task, error)
	Delete(userID, taskID uint) error
	ShareTask(ownerID, taskID uint, userIDs []uint) error
	UnshareTask(ownerID, taskID uint, sharedUserID uint) error
}

// CreateTaskRequest represents a task creation request
type CreateTaskRequest struct {
	Title       string
	Description string
	Type        models.TaskType
	Priority    *models.Priority // Optional: task priority
	DueDate               *time.Time
	ReminderMinutesBefore *int
	UserID                *uint  // Optional: ID of the user to whom the task will be assigned
	TagIDs                []uint // Optional: IDs of tags to associate with the task
}

// UpdateTaskRequest represents a task update request
type UpdateTaskRequest struct {
	Title       *string
	Description *string
	Type        *models.TaskType
	Priority    *models.Priority
	DueDate                 *time.Time
	ReminderMinutesBefore   *int
	Completed               *bool
	TagIDs                  *[]uint // Optional: IDs of tags to associate with the task (nil = no change, empty = remove all)
}

// TaskFilters defines filters for task search
type TaskFilters struct {
	Type               *models.TaskType
	Completed          *bool
	HideStaleCompleted bool
	Priority           *models.Priority
	Search      *string
	DueDateFrom *time.Time
	DueDateTo   *time.Time
	AssignedBy  *uint
	TagIDs      []uint // Filter by tag IDs
	Page        int
	Limit       int
	SortBy      string // created_at, due_date, title, priority
	Order       string // asc, desc
}

// PaginatedTasksResponse represents a paginated response
type PaginatedTasksResponse struct {
	Tasks      []models.Task `json:"tasks"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

type taskService struct {
	taskRepo              repositories.TaskRepository
	userRepo              repositories.UserRepository
	tagRepo               repositories.TagRepository
	groupService          GroupService
	notificationRepo      repositories.NotificationRepository
	activityNotifications ActivityNotificationService
}

// NewTaskService creates a new instance of TaskService
func NewTaskService(
	taskRepo repositories.TaskRepository,
	userRepo repositories.UserRepository,
	tagRepo repositories.TagRepository,
	groupService GroupService,
	notificationRepo repositories.NotificationRepository,
	activityNotifications ActivityNotificationService,
) TaskService {
	return &taskService{
		taskRepo:              taskRepo,
		userRepo:              userRepo,
		tagRepo:               tagRepo,
		groupService:          groupService,
		notificationRepo:      notificationRepo,
		activityNotifications: activityNotifications,
	}
}

func (s *taskService) Create(userID uint, req *CreateTaskRequest) (*models.Task, error) {
	// Validate task type
	if !isValidTaskType(req.Type) {
		return nil, errors.NewInvalidInputError("Invalid task type. Must be one of: casa, trabalho, lazer, saude")
	}

	// Validate priority if provided
	priority := models.PriorityMedia // Default priority
	if req.Priority != nil {
		if !isValidPriority(*req.Priority) {
			return nil, errors.NewInvalidInputError("Invalid priority. Must be one of: baixa, media, alta, urgente")
		}
		priority = *req.Priority
	}

	if req.ReminderMinutesBefore != nil && !IsValidReminderMinutes(*req.ReminderMinutesBefore) {
		return nil, errors.NewInvalidInputError("reminder_minutes_before must be one of: 5, 10, 15, 30, 60")
	}

	// Determine target user
	targetUserID := userID
	if req.UserID != nil {
		// Check if target user exists
		_, err := s.userRepo.FindByID(*req.UserID)
		if err != nil {
			return nil, errors.NewUserNotFoundError()
		}
		targetUserID = *req.UserID
		if targetUserID != userID {
			if err := s.groupService.AssertCanCollaborateWith(userID, targetUserID); err != nil {
				return nil, err
			}
		}
	}

	// Validate tags if provided
	var tags []models.Tag
	if len(req.TagIDs) > 0 {
		foundTags, err := s.tagRepo.FindByIDs(req.TagIDs, targetUserID)
		if err != nil {
			return nil, errors.NewInvalidInputError("One or more tags not found or don't belong to the user")
		}
		if len(foundTags) != len(req.TagIDs) {
			return nil, errors.NewInvalidInputError("One or more tags not found or don't belong to the user")
		}
		tags = foundTags
	}

	// AssignedBy is set only when delegating to another user (nil for self-created tasks)
	var assignedBy *uint
	if req.UserID != nil && *req.UserID != userID {
		assignedBy = &userID
	}
	task := &models.Task{
		Title:                 req.Title,
		Description:           req.Description,
		Type:                  req.Type,
		Priority:              priority,
		DueDate:               req.DueDate,
		ReminderMinutesBefore: req.ReminderMinutesBefore,
		UserID:                targetUserID,
		AssignedBy:            assignedBy,
		Completed:             false,
		Tags:                  tags,
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	// When a user creates a task for another, share it with the creator so both have access
	if req.UserID != nil && *req.UserID != userID {
		if err := s.taskRepo.AddSharedWith(task.ID, userID); err != nil {
			return nil, errors.NewInternalServerError(err)
		}
	}

	// Reload with relationships
	task, err := s.taskRepo.FindByID(task.ID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	return task, nil
}

func (s *taskService) GetByID(userID, taskID uint) (*models.Task, error) {
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, errors.NewTaskNotFoundError()
	}

	canAccess, err := s.taskRepo.UserCanAccessTask(taskID, userID)
	if err != nil || !canAccess {
		return nil, errors.NewForbiddenError()
	}

	return task, nil
}

func (s *taskService) GetByUserID(userID uint, filters *TaskFilters) (*PaginatedTasksResponse, error) {
	repoFilters := &repositories.TaskFilters{}

	// Set default pagination
	page := 1
	limit := 10
	if filters != nil {
		if filters.Page > 0 {
			page = filters.Page
		}
		if filters.Limit > 0 {
			limit = filters.Limit
			// Maximum limit is 100
			if limit > 100 {
				limit = 100
			}
		}
		repoFilters.Page = page
		repoFilters.Limit = limit

		// Apply filters
		if filters.Type != nil {
			if !isValidTaskType(*filters.Type) {
				return nil, errors.NewInvalidInputError("Invalid task type filter")
			}
			repoFilters.Type = filters.Type
		}
		if filters.Priority != nil {
			if !isValidPriority(*filters.Priority) {
				return nil, errors.NewInvalidInputError("Invalid priority filter")
			}
			repoFilters.Priority = filters.Priority
		}
		repoFilters.Completed = filters.Completed
		repoFilters.HideStaleCompleted = filters.HideStaleCompleted
		repoFilters.Search = filters.Search
		repoFilters.DueDateFrom = filters.DueDateFrom
		repoFilters.DueDateTo = filters.DueDateTo
		repoFilters.AssignedBy = filters.AssignedBy
		repoFilters.TagIDs = filters.TagIDs
		repoFilters.SortBy = filters.SortBy
		repoFilters.Order = filters.Order
	} else {
		repoFilters.Page = page
		repoFilters.Limit = limit
	}

	tasks, total, err := s.taskRepo.FindByUserID(userID, repoFilters)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	// Calculate total pages
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	return &PaginatedTasksResponse{
		Tasks:      tasks,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *taskService) GetAssignedByUser(assignedByID uint, filters *TaskFilters) (*PaginatedTasksResponse, error) {
	repoFilters := &repositories.TaskFilters{}

	// Set default pagination
	page := 1
	limit := 10
	if filters != nil {
		if filters.Page > 0 {
			page = filters.Page
		}
		if filters.Limit > 0 {
			limit = filters.Limit
			// Maximum limit is 100
			if limit > 100 {
				limit = 100
			}
		}
		repoFilters.Page = page
		repoFilters.Limit = limit

		// Apply filters
		if filters.Type != nil {
			if !isValidTaskType(*filters.Type) {
				return nil, errors.NewInvalidInputError("Invalid task type filter")
			}
			repoFilters.Type = filters.Type
		}
		if filters.Priority != nil {
			if !isValidPriority(*filters.Priority) {
				return nil, errors.NewInvalidInputError("Invalid priority filter")
			}
			repoFilters.Priority = filters.Priority
		}
		repoFilters.Completed = filters.Completed
		repoFilters.HideStaleCompleted = filters.HideStaleCompleted
		repoFilters.Search = filters.Search
		repoFilters.DueDateFrom = filters.DueDateFrom
		repoFilters.DueDateTo = filters.DueDateTo
		repoFilters.TagIDs = filters.TagIDs
		repoFilters.SortBy = filters.SortBy
		repoFilters.Order = filters.Order
	} else {
		repoFilters.Page = page
		repoFilters.Limit = limit
	}

	tasks, total, err := s.taskRepo.FindByAssignedBy(assignedByID, repoFilters)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	// Calculate total pages
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	return &PaginatedTasksResponse{
		Tasks:      tasks,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *taskService) Update(userID, taskID uint, req *UpdateTaskRequest) (*models.Task, error) {
	// Find task
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, errors.NewTaskNotFoundError()
	}

	canAccess, err := s.taskRepo.UserCanAccessTask(taskID, userID)
	if err != nil || !canAccess {
		return nil, errors.NewForbiddenError()
	}

	oldDueDate := copyTimePtr(task.DueDate)
	oldReminderMinutes := copyIntPtr(task.ReminderMinutesBefore)

	// Update fields
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Type != nil {
		if !isValidTaskType(*req.Type) {
			return nil, errors.NewInvalidInputError("Invalid task type. Must be one of: casa, trabalho, lazer, saude")
		}
		task.Type = *req.Type
	}
	if req.Priority != nil {
		if !isValidPriority(*req.Priority) {
			return nil, errors.NewInvalidInputError("Invalid priority. Must be one of: baixa, media, alta, urgente")
		}
		task.Priority = *req.Priority
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}
	if req.ReminderMinutesBefore != nil {
		if !IsValidReminderMinutes(*req.ReminderMinutesBefore) {
			return nil, errors.NewInvalidInputError("reminder_minutes_before must be one of: 5, 10, 15, 30, 60")
		}
		task.ReminderMinutesBefore = req.ReminderMinutesBefore
	}
	wasCompleted := task.Completed
	if req.Completed != nil {
		if *req.Completed && !task.Completed {
			now := time.Now()
			task.CompletedAt = &now
		} else if !*req.Completed {
			task.CompletedAt = nil
		}
		task.Completed = *req.Completed
	}

	// Update tags if provided
	if req.TagIDs != nil {
		if len(*req.TagIDs) == 0 {
			// Remove all tags
			task.Tags = []models.Tag{}
		} else {
			// Validate and set new tags (use task owner for tag ownership)
			foundTags, err := s.tagRepo.FindByIDs(*req.TagIDs, task.UserID)
			if err != nil {
				return nil, errors.NewInvalidInputError("One or more tags not found or don't belong to the user")
			}
			if len(foundTags) != len(*req.TagIDs) {
				return nil, errors.NewInvalidInputError("One or more tags not found or don't belong to the user")
			}
			task.Tags = foundTags
		}
	}

	if dueDateChanged(oldDueDate, task.DueDate) || reminderMinutesChanged(oldReminderMinutes, task.ReminderMinutesBefore) {
		if err := s.notificationRepo.DeleteByTaskID(taskID); err != nil {
			return nil, errors.NewInternalServerError(err)
		}
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	// Reload with relationships
	task, err = s.taskRepo.FindByID(task.ID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	if req.Completed != nil && *req.Completed && !wasCompleted && s.activityNotifications != nil {
		completerName := ""
		if completer, err := s.userRepo.FindByID(userID); err == nil {
			completerName = completer.Username
		}
		if completerName == "" {
			completerName = "Usuário"
		}
		_ = s.activityNotifications.NotifyDelegatedTaskCompleted(task, userID, completerName)
	}

	return task, nil
}

func (s *taskService) Delete(userID, taskID uint) error {
	// Find task
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return errors.NewTaskNotFoundError()
	}

	// Only the task owner can delete the task
	if task.UserID != userID {
		return errors.NewForbiddenError()
	}

	if err := s.taskRepo.Delete(taskID); err != nil {
		return errors.NewInternalServerError(err)
	}

	return nil
}

// ShareTask adds users to the task's shared list. Only the task owner can share.
func (s *taskService) ShareTask(ownerID, taskID uint, userIDs []uint) error {
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return errors.NewTaskNotFoundError()
	}
	if task.UserID != ownerID {
		return errors.NewForbiddenError()
	}
	for _, uid := range userIDs {
		if uid == ownerID {
			continue // owner already has access
		}
		if _, err := s.userRepo.FindByID(uid); err != nil {
			return errors.NewInvalidInputError("One or more user IDs are invalid")
		}
		if err := s.groupService.AssertCanCollaborateWith(ownerID, uid); err != nil {
			return err
		}
		if err := s.taskRepo.AddSharedWith(taskID, uid); err != nil {
			return errors.NewInternalServerError(err)
		}
	}
	return nil
}

// UnshareTask removes a user from the task's shared list. Only the task owner can unshare.
func (s *taskService) UnshareTask(ownerID, taskID uint, sharedUserID uint) error {
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return errors.NewTaskNotFoundError()
	}
	if task.UserID != ownerID {
		return errors.NewForbiddenError()
	}
	if err := s.taskRepo.RemoveSharedWith(taskID, sharedUserID); err != nil {
		return errors.NewInternalServerError(err)
	}
	return nil
}

func copyTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := *t
	return &v
}

func copyIntPtr(v *int) *int {
	if v == nil {
		return nil
	}
	n := *v
	return &n
}

func dueDateChanged(before, after *time.Time) bool {
	return !timePtrEqual(before, after)
}

func reminderMinutesChanged(before, after *int) bool {
	return !intPtrEqual(before, after)
}

func timePtrEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.UTC().Truncate(time.Second).Equal(b.UTC().Truncate(time.Second))
}

func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// isValidTaskType checks if the task type is valid
func isValidTaskType(taskType models.TaskType) bool {
	switch taskType {
	case models.TaskTypeCasa, models.TaskTypeTrabalho, models.TaskTypeLazer, models.TaskTypeSaude:
		return true
	default:
		return false
	}
}

// isValidPriority checks if the priority is valid
func isValidPriority(priority models.Priority) bool {
	switch priority {
	case models.PriorityBaixa, models.PriorityMedia, models.PriorityAlta, models.PriorityUrgente:
		return true
	default:
		return false
	}
}

