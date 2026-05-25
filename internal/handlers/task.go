package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"todo-go-backend/internal/errors"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// TaskHandler manages task handlers
type TaskHandler struct {
	taskService services.TaskService
}

// NewTaskHandler creates a new instance of TaskHandler
func NewTaskHandler(taskService services.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// CreateTaskRequest represents a task creation request
type CreateTaskRequest struct {
	Title       string          `json:"title" binding:"required,min=1,max=200" example:"Clean the house"`
	Description string          `json:"description" example:"Clean all rooms"`
	Type        models.TaskType `json:"type" binding:"required,oneof=casa trabalho lazer saude" example:"casa"`
	Priority    *string         `json:"priority" binding:"omitempty,oneof=baixa media alta urgente" example:"alta"` // Optional: task priority
	DueDate               *string `json:"due_date" example:"2024-12-31T23:59:59Z"`                                    // ISO 8601 format
	ReminderMinutesBefore *int    `json:"reminder_minutes_before" example:"10"`                                       // Optional: minutes before due_date (5,10,15,30,60); omit to use user default
	UserID                *uint   `json:"user_id" example:"2"`                                                        // Optional: if provided, assign to another user
	TagIDs                []uint  `json:"tag_ids"`                                                                    // Optional: IDs of tags to associate
}

// ShareTaskRequest represents a request to share a task with users
type ShareTaskRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required,min=1" example:"2,3,4"`
}

// UpdateTaskRequest represents a task update request
type UpdateTaskRequest struct {
	Title       *string          `json:"title" example:"Updated title"`
	Description *string          `json:"description" example:"Updated description"`
	Type        *models.TaskType `json:"type" binding:"omitempty,oneof=casa trabalho lazer saude" example:"trabalho"`
	Priority    *string          `json:"priority" binding:"omitempty,oneof=baixa media alta urgente" example:"urgente"`
	DueDate               *string          `json:"due_date" example:"2024-12-31T23:59:59Z"`
	ReminderMinutesBefore *int             `json:"reminder_minutes_before" example:"15"` // Optional: 5,10,15,30,60; null clears override
	Completed             *bool            `json:"completed" example:"true"`
	TagIDs                *[]uint          `json:"tag_ids"` // Optional: nil = no change, [] = remove all, [1,2] = set tags
}

// CreateTask creates a new task
// @Summary      Create a new task
// @Description  Creates a new task for the authenticated user or assigns it to another user
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      CreateTaskRequest  true  "Task creation data"
// @Success      201      {object}  models.Task
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, errors.NewInvalidInputError(err.Error()))
		return
	}

	userID := c.GetUint("user_id")

	// Parse due date if provided
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsed, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			handleError(c, errors.NewInvalidInputError("Invalid date format. Use ISO 8601 (RFC3339)"))
			return
		}
		dueDate = &parsed
	}

	// Parse priority
	var priority *models.Priority
	if req.Priority != nil {
		p := models.Priority(*req.Priority)
		priority = &p
	}

	if req.ReminderMinutesBefore != nil && !services.IsValidReminderMinutes(*req.ReminderMinutesBefore) {
		handleError(c, errors.NewInvalidInputError("reminder_minutes_before must be one of: 5, 10, 15, 30, 60"))
		return
	}

	createReq := &services.CreateTaskRequest{
		Title:                 req.Title,
		Description:           req.Description,
		Type:                  req.Type,
		Priority:              priority,
		DueDate:               dueDate,
		ReminderMinutesBefore: req.ReminderMinutesBefore,
		UserID:                req.UserID,
		TagIDs:                req.TagIDs,
	}

	task, err := h.taskService.Create(userID, createReq)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GetTasks lists user tasks
// @Summary      List user tasks
// @Description  Retrieves paginated tasks for the authenticated user with optional filters, search, and sorting
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page          query     int     false  "Page number (default: 1)"
// @Param        limit         query     int     false  "Items per page (default: 10, max: 100)"
// @Param        type          query     string  false  "Filter by task type (casa, trabalho, lazer, saude)"
// @Param        completed     query     bool    false  "Filter by completion status"
// @Param        search        query     string  false  "Search in title and description"
// @Param        due_date_from query     string  false  "Filter tasks with due date from (ISO 8601 format)"
// @Param        due_date_to   query     string  false  "Filter tasks with due date to (ISO 8601 format)"
// @Param        period        query     string  false  "Filter by period (overdue, today, this_week, this_month)"
// @Param        assigned_by   query     int     false  "Filter by user ID who assigned the task"
// @Param        priority      query     string  false  "Filter by priority (baixa, media, alta, urgente)"
// @Param        tag_ids       query     string  false  "Filter by tag IDs, comma-separated (e.g. 1,2,3)"
// @Param        sort_by       query     string  false  "Sort field (created_at, due_date, title, priority)"
// @Param        order         query     string  false  "Sort order (asc, desc)"
// @Success      200           {object}  services.PaginatedTasksResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /tasks [get]
func (h *TaskHandler) GetTasks(c *gin.Context) {
	userID := c.GetUint("user_id")

	filters := &services.TaskFilters{}

	// Parse pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	// Parse filters
	if taskType := c.Query("type"); taskType != "" {
		taskTypeEnum := models.TaskType(taskType)
		filters.Type = &taskTypeEnum
	}

	if completed := c.Query("completed"); completed != "" {
		completedBool := completed == "true"
		filters.Completed = &completedBool
	}

	if c.Query("hide_stale_completed") == "true" {
		filters.HideStaleCompleted = true
	}

	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	// Parse date filters and period filters
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	// Handle period filters (overdue, today, this_week, this_month)
	if period := c.Query("period"); period != "" {
		switch period {
		case "overdue":
			// Tasks with due_date in the past and not completed
			past := now
			filters.DueDateTo = &past
			notCompleted := false
			filters.Completed = &notCompleted
		case "today":
			filters.DueDateFrom = &todayStart
			filters.DueDateTo = &todayEnd
		case "this_week":
			filters.DueDateFrom = &weekStart
			filters.DueDateTo = &weekEnd
		case "this_month":
			filters.DueDateFrom = &monthStart
			filters.DueDateTo = &monthEnd
		}
	}

	// Parse explicit date filters (override period if both are provided)
	if dueDateFromStr := c.Query("due_date_from"); dueDateFromStr != "" {
		if dueDateFrom, err := time.Parse(time.RFC3339, dueDateFromStr); err == nil {
			filters.DueDateFrom = &dueDateFrom
		}
	}

	if dueDateToStr := c.Query("due_date_to"); dueDateToStr != "" {
		if dueDateTo, err := time.Parse(time.RFC3339, dueDateToStr); err == nil {
			filters.DueDateTo = &dueDateTo
		}
	}

	// Parse priority filter
	if priorityStr := c.Query("priority"); priorityStr != "" {
		priority := models.Priority(priorityStr)
		filters.Priority = &priority
	}

	// Parse assigned_by filter
	if assignedByStr := c.Query("assigned_by"); assignedByStr != "" {
		if assignedBy, err := strconv.ParseUint(assignedByStr, 10, 32); err == nil {
			assignedByUint := uint(assignedBy)
			filters.AssignedBy = &assignedByUint
		}
	}

	// Parse tag_ids filter (comma-separated)
	if tagIDsStr := c.Query("tag_ids"); tagIDsStr != "" {
		tagIDs := []uint{}
		// Support both comma-separated and array format
		if len(tagIDsStr) > 0 && tagIDsStr[0] == '[' {
			// Array format: [1,2,3]
			// Simple parsing for [1,2,3] format
			cleaned := tagIDsStr[1 : len(tagIDsStr)-1] // Remove [ and ]
			if cleaned != "" {
				for _, idStr := range strings.Split(cleaned, ",") {
					idStr = strings.TrimSpace(idStr)
					if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
						tagIDs = append(tagIDs, uint(id))
					}
				}
			}
		} else {
			// Comma-separated format: 1,2,3
			for _, idStr := range strings.Split(tagIDsStr, ",") {
				idStr = strings.TrimSpace(idStr)
				if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
					tagIDs = append(tagIDs, uint(id))
				}
			}
		}
		if len(tagIDs) > 0 {
			filters.TagIDs = tagIDs
		}
	}

	// Parse sorting
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filters.SortBy = sortBy
	}
	if order := c.Query("order"); order != "" {
		filters.Order = order
	}

	result, err := h.taskService.GetByUserID(userID, filters)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAssignedTasks lists tasks assigned by the authenticated user
// @Summary      List tasks assigned by user
// @Description  Retrieves paginated tasks that were created/assigned by the authenticated user to other users. This allows users to follow tasks they created for others.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page          query     int     false  "Page number (default: 1)"
// @Param        limit         query     int     false  "Items per page (default: 10, max: 100)"
// @Param        type          query     string  false  "Filter by task type (casa, trabalho, lazer, saude)"
// @Param        completed     query     bool    false  "Filter by completion status"
// @Param        search        query     string  false  "Search in title and description"
// @Param        due_date_from query     string  false  "Filter tasks with due date from (ISO 8601 format)"
// @Param        due_date_to   query     string  false  "Filter tasks with due date to (ISO 8601 format)"
// @Param        period        query     string  false  "Filter by period (overdue, today, this_week, this_month)"
// @Param        priority      query     string  false  "Filter by priority (baixa, media, alta, urgente)"
// @Param        tag_ids       query     string  false  "Filter by tag IDs, comma-separated (e.g. 1,2,3)"
// @Param        sort_by       query     string  false  "Sort field (created_at, due_date, title, priority)"
// @Param        order         query     string  false  "Sort order (asc, desc)"
// @Success      200           {object}  services.PaginatedTasksResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /tasks/assigned [get]
func (h *TaskHandler) GetAssignedTasks(c *gin.Context) {
	userID := c.GetUint("user_id")

	filters := &services.TaskFilters{}

	// Parse pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	// Parse filters
	if typeStr := c.Query("type"); typeStr != "" {
		taskType := models.TaskType(typeStr)
		filters.Type = &taskType
	}

	if completedStr := c.Query("completed"); completedStr != "" {
		if completed, err := strconv.ParseBool(completedStr); err == nil {
			filters.Completed = &completed
		}
	}

	if c.Query("hide_stale_completed") == "true" {
		filters.HideStaleCompleted = true
	}

	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	// Parse period filter
	if period := c.Query("period"); period != "" {
		now := time.Now()

		switch period {
		case "overdue":
			dueDateTo := now
			filters.DueDateTo = &dueDateTo
			notCompleted := false
			filters.Completed = &notCompleted
		case "today":
			startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			endOfDay := startOfDay.Add(24 * time.Hour).Add(-1 * time.Second)
			filters.DueDateFrom = &startOfDay
			filters.DueDateTo = &endOfDay
		case "this_week":
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startOfWeek := now.AddDate(0, 0, -weekday+1)
			startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
			endOfWeek := startOfWeek.AddDate(0, 0, 7).Add(-1 * time.Second)
			filters.DueDateFrom = &startOfWeek
			filters.DueDateTo = &endOfWeek
		case "this_month":
			startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-1 * time.Second)
			filters.DueDateFrom = &startOfMonth
			filters.DueDateTo = &endOfMonth
		}
	}

	if dueDateFromStr := c.Query("due_date_from"); dueDateFromStr != "" {
		if dueDateFrom, err := time.Parse(time.RFC3339, dueDateFromStr); err == nil {
			filters.DueDateFrom = &dueDateFrom
		}
	}

	if dueDateToStr := c.Query("due_date_to"); dueDateToStr != "" {
		if dueDateTo, err := time.Parse(time.RFC3339, dueDateToStr); err == nil {
			filters.DueDateTo = &dueDateTo
		}
	}

	// Parse priority filter
	if priorityStr := c.Query("priority"); priorityStr != "" {
		priority := models.Priority(priorityStr)
		filters.Priority = &priority
	}

	// Parse tag_ids filter (comma-separated)
	if tagIDsStr := c.Query("tag_ids"); tagIDsStr != "" {
		tagIDs := []uint{}
		// Support both comma-separated and array format
		if len(tagIDsStr) > 0 && tagIDsStr[0] == '[' {
			// Array format: [1,2,3]
			// Simple parsing for [1,2,3] format
			cleaned := tagIDsStr[1 : len(tagIDsStr)-1] // Remove [ and ]
			if cleaned != "" {
				for _, idStr := range strings.Split(cleaned, ",") {
					idStr = strings.TrimSpace(idStr)
					if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
						tagIDs = append(tagIDs, uint(id))
					}
				}
			}
		} else {
			// Comma-separated format: 1,2,3
			for _, idStr := range strings.Split(tagIDsStr, ",") {
				idStr = strings.TrimSpace(idStr)
				if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
					tagIDs = append(tagIDs, uint(id))
				}
			}
		}
		if len(tagIDs) > 0 {
			filters.TagIDs = tagIDs
		}
	}

	// Parse sorting
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filters.SortBy = sortBy
	}
	if order := c.Query("order"); order != "" {
		filters.Order = order
	}

	result, err := h.taskService.GetAssignedByUser(userID, filters)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetTask retrieves a specific task
// @Summary      Get a task by ID
// @Description  Retrieves a specific task by its ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Task ID"
// @Success      200  {object}  models.Task
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		handleError(c, errors.NewInvalidInputError("Invalid task ID"))
		return
	}

	task, err := h.taskService.GetByID(userID, uint(taskID))
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTask updates a task
// @Summary      Update a task
// @Description  Updates an existing task
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                true  "Task ID"
// @Param        request  body      UpdateTaskRequest  true  "Task update data"
// @Success      200      {object}  models.Task
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		handleError(c, errors.NewInvalidInputError("Invalid task ID"))
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, errors.NewInvalidInputError(err.Error()))
		return
	}

	// Parse due date if provided
	var dueDate *time.Time
	if req.DueDate != nil {
		if *req.DueDate == "" {
			dueDate = nil
		} else {
			parsed, err := time.Parse(time.RFC3339, *req.DueDate)
			if err != nil {
				handleError(c, errors.NewInvalidInputError("Invalid date format. Use ISO 8601 (RFC3339)"))
				return
			}
			dueDate = &parsed
		}
	}

	// Parse priority
	var priority *models.Priority
	if req.Priority != nil {
		p := models.Priority(*req.Priority)
		priority = &p
	}

	if req.ReminderMinutesBefore != nil && !services.IsValidReminderMinutes(*req.ReminderMinutesBefore) {
		handleError(c, errors.NewInvalidInputError("reminder_minutes_before must be one of: 5, 10, 15, 30, 60"))
		return
	}

	updateReq := &services.UpdateTaskRequest{
		Title:                 req.Title,
		Description:           req.Description,
		Type:                  req.Type,
		Priority:              priority,
		DueDate:               dueDate,
		ReminderMinutesBefore: req.ReminderMinutesBefore,
		Completed:             req.Completed,
		TagIDs:                req.TagIDs,
	}

	task, err := h.taskService.Update(userID, uint(taskID), updateReq)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask deletes a task
// @Summary      Delete a task
// @Description  Deletes a task by its ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Task ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		handleError(c, errors.NewInvalidInputError("Invalid task ID"))
		return
	}

	if err := h.taskService.Delete(userID, uint(taskID)); err != nil {
		handleError(c, err)
		return
	}

	handleSuccess(c, http.StatusOK, "Task deleted successfully", nil)
}

// ShareTask shares a task with other users (owner only). No limit on how many users.
// @Summary      Share a task with users
// @Description  Adds users to the task's shared list. Only the owner can share. Target users must be members of at least one group in common with the owner.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int               true  "Task ID"
// @Param        request  body      ShareTaskRequest  true  "User IDs to share with"
// @Success      200      {object}  SuccessResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /tasks/{id}/share [post]
func (h *TaskHandler) ShareTask(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		handleError(c, errors.NewInvalidInputError("Invalid task ID"))
		return
	}

	var req ShareTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, errors.NewInvalidInputError(err.Error()))
		return
	}

	if err := h.taskService.ShareTask(userID, uint(taskID), req.UserIDs); err != nil {
		handleError(c, err)
		return
	}

	handleSuccess(c, http.StatusOK, "Task shared successfully", nil)
}

// UnshareTask removes a user from the task's shared list (owner only).
// @Summary      Unshare a task with a user
// @Description  Removes the given user from the task's shared list. Only the task owner can unshare.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int  true  "Task ID"
// @Param        user_id  path      int  true  "User ID to remove from shared list"
// @Success      200      {object}  SuccessResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /tasks/{id}/share/{user_id} [delete]
func (h *TaskHandler) UnshareTask(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		handleError(c, errors.NewInvalidInputError("Invalid task ID"))
		return
	}
	sharedUserID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		handleError(c, errors.NewInvalidInputError("Invalid user ID"))
		return
	}

	if err := h.taskService.UnshareTask(userID, uint(taskID), uint(sharedUserID)); err != nil {
		handleError(c, err)
		return
	}

	handleSuccess(c, http.StatusOK, "User removed from shared list", nil)
}
