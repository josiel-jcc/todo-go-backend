package services

import (
	"time"
	"todo-go-backend/internal/errors"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"
	"todo-go-backend/pkg/utils"
)

type AuthService interface {
	Register(username, email, password string) (*models.User, string, error)
	Login(identifier, password string) (*models.User, string, error)
}

type authService struct {
	userRepo     repositories.UserRepository
	jwtSecret    string
	groupService GroupService
}

func NewAuthService(userRepo repositories.UserRepository, jwtSecret string, groupService GroupService) AuthService {
	return &authService{
		userRepo:     userRepo,
		jwtSecret:    jwtSecret,
		groupService: groupService,
	}
}

func (s *authService) Register(username, email, password string) (*models.User, string, error) {
	exists, err := s.userRepo.ExistsByUsernameOrEmail(username, email)
	if err != nil {
		return nil, "", errors.NewInternalServerError(err)
	}
	if exists {
		return nil, "", errors.NewUserAlreadyExistsError()
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", errors.NewInternalServerError(err)
	}

	now := time.Now()
	user := &models.User{
		Username:             username,
		Email:                email,
		Password:             hashedPassword,
		NotificationsEnabled: false,
		TermsAcceptedAt:      &now,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", errors.NewInternalServerError(err)
	}

	if s.groupService != nil {
		if err := s.groupService.AddUserToDefaultGroup(user.ID); err != nil {
			return nil, "", errors.NewInternalServerError(err)
		}
	}

	token, _, err := utils.GenerateToken(user.ID, user.Username, s.jwtSecret)
	if err != nil {
		return nil, "", errors.NewInternalServerError(err)
	}

	return user, token, nil
}

func (s *authService) Login(identifier, password string) (*models.User, string, error) {
	user, err := s.userRepo.FindByUsernameOrEmailValue(identifier)
	if err != nil {
		return nil, "", errors.NewInvalidCredentialsError()
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, "", errors.NewInvalidCredentialsError()
	}

	token, _, err := utils.GenerateToken(user.ID, user.Username, s.jwtSecret)
	if err != nil {
		return nil, "", errors.NewInternalServerError(err)
	}

	return user, token, nil
}
