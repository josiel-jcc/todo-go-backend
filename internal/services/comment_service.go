package services

import (
	"todo-go-backend/internal/errors"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"
)

// CommentService defines the interface for comment operations
type CommentService interface {
	Create(userID uint, req *CreateCommentRequest) (*models.Comment, error)
	GetByID(userID, commentID uint) (*models.Comment, error)
	GetByTaskID(userID, taskID uint) ([]models.Comment, error)
	Update(userID, commentID uint, req *UpdateCommentRequest) (*models.Comment, error)
	Delete(userID, commentID uint) error
}

// CreateCommentRequest represents a comment creation request
type CreateCommentRequest struct {
	Content string
	TaskID  uint
}

// UpdateCommentRequest represents a comment update request
type UpdateCommentRequest struct {
	Content *string
}

type commentService struct {
	commentRepo          repositories.CommentRepository
	taskRepo             repositories.TaskRepository
	activityNotifications ActivityNotificationService
}

// NewCommentService creates a new instance of CommentService
func NewCommentService(
	commentRepo repositories.CommentRepository,
	taskRepo repositories.TaskRepository,
	activityNotifications ActivityNotificationService,
) CommentService {
	return &commentService{
		commentRepo:          commentRepo,
		taskRepo:             taskRepo,
		activityNotifications: activityNotifications,
	}
}

func (s *commentService) Create(userID uint, req *CreateCommentRequest) (*models.Comment, error) {
	// Validate content
	if req.Content == "" || len(req.Content) > 5000 {
		return nil, errors.NewInvalidInputError("Comment content must be between 1 and 5000 characters")
	}

	if _, err := s.taskRepo.FindByID(req.TaskID); err != nil {
		return nil, errors.NewTaskNotFoundError()
	}
	canAccess, err := s.taskRepo.UserCanAccessTask(req.TaskID, userID)
	if err != nil || !canAccess {
		return nil, errors.NewForbiddenError()
	}

	comment := &models.Comment{
		Content: req.Content,
		TaskID:  req.TaskID,
		UserID:  userID,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	// Reload with relationships
	comment, err = s.commentRepo.FindByID(comment.ID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	task, err := s.taskRepo.FindByID(req.TaskID)
	if err == nil && s.activityNotifications != nil {
		authorName := comment.User.Username
		if authorName == "" {
			authorName = "Usuário"
		}
		_ = s.activityNotifications.NotifyTaskComment(task, comment, authorName)
	}

	return comment, nil
}

func (s *commentService) GetByID(userID, commentID uint) (*models.Comment, error) {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return nil, errors.NewTaskNotFoundError() // Reuse error type
	}

	canAccess, err := s.taskRepo.UserCanAccessTask(comment.TaskID, userID)
	if err != nil || !canAccess {
		return nil, errors.NewForbiddenError()
	}

	return comment, nil
}

func (s *commentService) GetByTaskID(userID, taskID uint) ([]models.Comment, error) {
	if _, err := s.taskRepo.FindByID(taskID); err != nil {
		return nil, errors.NewTaskNotFoundError()
	}
	canAccess, err := s.taskRepo.UserCanAccessTask(taskID, userID)
	if err != nil || !canAccess {
		return nil, errors.NewForbiddenError()
	}

	comments, err := s.commentRepo.FindByTaskID(taskID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	return comments, nil
}

func (s *commentService) Update(userID, commentID uint, req *UpdateCommentRequest) (*models.Comment, error) {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return nil, errors.NewTaskNotFoundError()
	}

	// Only the comment author can update their comment
	if comment.UserID != userID {
		return nil, errors.NewForbiddenError()
	}

	// Validate content if provided
	if req.Content != nil {
		if *req.Content == "" || len(*req.Content) > 5000 {
			return nil, errors.NewInvalidInputError("Comment content must be between 1 and 5000 characters")
		}
		comment.Content = *req.Content
	}

	if err := s.commentRepo.Update(comment); err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	// Reload with relationships
	comment, err = s.commentRepo.FindByID(comment.ID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	return comment, nil
}

func (s *commentService) Delete(userID, commentID uint) error {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return errors.NewTaskNotFoundError()
	}

	// Only the comment author can delete their comment
	if comment.UserID != userID {
		return errors.NewForbiddenError()
	}

	if err := s.commentRepo.Delete(commentID); err != nil {
		return errors.NewInternalServerError(err)
	}

	return nil
}

