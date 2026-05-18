package services

import (
	"testing"
	"todo-go-backend/internal/errors"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"

	"github.com/stretchr/testify/assert"
)

// MockUserRepository é um mock do UserRepository para testes
type MockUserRepository struct {
	users        map[uint]*models.User
	usersByUser  map[string]*models.User
	usersByEmail map[string]*models.User
	nextID       uint
}

func NewMockUserRepository() repositories.UserRepository {
	return &MockUserRepository{
		users:        make(map[uint]*models.User),
		usersByUser:  make(map[string]*models.User),
		usersByEmail: make(map[string]*models.User),
		nextID:       1,
	}
}

func (m *MockUserRepository) Create(user *models.User) error {
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	m.usersByUser[user.Username] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepository) FindByUsername(username string) (*models.User, error) {
	user, ok := m.usersByUser[username]
	if !ok {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	user, ok := m.usersByEmail[email]
	if !ok {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepository) FindByUsernameOrEmail(username, email string) (*models.User, error) {
	if user, ok := m.usersByUser[username]; ok {
		return user, nil
	}
	if user, ok := m.usersByEmail[email]; ok {
		return user, nil
	}
	return nil, errors.ErrUserNotFound
}

func (m *MockUserRepository) FindByUsernameOrEmailValue(identifier string) (*models.User, error) {
	if user, ok := m.usersByUser[identifier]; ok {
		return user, nil
	}
	if user, ok := m.usersByEmail[identifier]; ok {
		return user, nil
	}
	return nil, errors.ErrUserNotFound
}

func (m *MockUserRepository) ExistsByUsernameOrEmail(username, email string) (bool, error) {
	_, userExists := m.usersByUser[username]
	_, emailExists := m.usersByEmail[email]
	return userExists || emailExists, nil
}

func (m *MockUserRepository) FindAll() ([]models.User, error) {
	users := make([]models.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, nil
}

func (m *MockUserRepository) FindAllPaginated(page, limit int) ([]models.UserPublic, int64, error) {
	allUsers := make([]models.UserPublic, 0, len(m.users))
	for _, user := range m.users {
		allUsers = append(allUsers, models.UserPublic{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	total := int64(len(allUsers))
	offset := (page - 1) * limit
	start := offset
	end := offset + limit
	if start > int(total) {
		start = int(total)
	}
	if end > int(total) {
		end = int(total)
	}
	if start < 0 {
		start = 0
	}

	var paginatedUsers []models.UserPublic
	if start < int(total) {
		paginatedUsers = allUsers[start:end]
	}

	return paginatedUsers, total, nil
}

func TestAuthService_Register(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewAuthService(mockRepo, "test-secret")

	t.Run("Register new user successfully", func(t *testing.T) {
		user, token, err := service.Register("testuser", "test@example.com", "password123")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, token)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("Register duplicate username", func(t *testing.T) {
		_, _, err := service.Register("testuser", "test2@example.com", "password123")

		assert.Error(t, err)
		assert.IsType(t, &errors.AppError{}, err)
		appErr := err.(*errors.AppError)
		assert.Equal(t, errors.ErrUserAlreadyExists, appErr.Err)
	})
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewAuthService(mockRepo, "test-secret")

	// Create a user first
	_, _, _ = service.Register("testuser", "test@example.com", "password123")

	t.Run("Login with valid credentials", func(t *testing.T) {
		user, token, err := service.Login("testuser", "password123")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, token)
	})

	t.Run("Login with invalid password", func(t *testing.T) {
		_, _, err := service.Login("testuser", "wrongpassword")

		assert.Error(t, err)
		assert.IsType(t, &errors.AppError{}, err)
		appErr := err.(*errors.AppError)
		assert.Equal(t, errors.ErrInvalidCredentials, appErr.Err)
	})

	t.Run("Login with non-existent user", func(t *testing.T) {
		_, _, err := service.Login("nonexistent", "password123")

		assert.Error(t, err)
		assert.IsType(t, &errors.AppError{}, err)
		appErr := err.(*errors.AppError)
		assert.Equal(t, errors.ErrInvalidCredentials, appErr.Err)
	})
}
