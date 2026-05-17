package repositories

import (
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"
)

// UserRepository defines the interface for user operations
type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByUsernameOrEmail(username, email string) (*models.User, error)
	FindByUsernameOrEmailValue(identifier string) (*models.User, error) // Find by username or email using a single value
	ExistsByUsernameOrEmail(username, email string) (bool, error)
	FindAll() ([]models.User, error) // Find all users
	FindAllPaginated(page, limit int) ([]models.UserPublic, int64, error) // Find all users with pagination (public fields only)
}

type userRepository struct{}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) Create(user *models.User) error {
	return database.DB.Create(user).Error
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUsernameOrEmail(username, email string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("username = ? OR email = ?", username, email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUsernameOrEmailValue(identifier string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("username = ? OR email = ?", identifier, identifier).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) ExistsByUsernameOrEmail(username, email string) (bool, error) {
	var count int64
	if err := database.DB.Model(&models.User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) FindAll() ([]models.User, error) {
	var users []models.User
	if err := database.DB.Select("id", "username", "email", "created_at", "updated_at").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) FindAllPaginated(page, limit int) ([]models.UserPublic, int64, error) {
	var users []models.UserPublic
	var total int64

	if err := database.DB.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	if err := database.DB.Model(&models.User{}).
		Select("id", "username", "created_at", "updated_at").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

