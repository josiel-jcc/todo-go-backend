package repositories

import (
	"time"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"
)

type GroupInvitationRepository interface {
	Create(inv *models.GroupInvitation) error
	FindByID(id uint) (*models.GroupInvitation, error)
	FindPendingByGroupAndUser(groupID, invitedUserID uint) (*models.GroupInvitation, error)
	ListPendingByGroup(groupID uint) ([]models.GroupInvitation, error)
	ListPendingReceived(userID uint) ([]models.GroupInvitation, error)
	UpdateStatus(id uint, status models.GroupInvitationStatus) error
}

type groupInvitationRepository struct{}

func NewGroupInvitationRepository() GroupInvitationRepository {
	return &groupInvitationRepository{}
}

func (r *groupInvitationRepository) Create(inv *models.GroupInvitation) error {
	return database.DB.Create(inv).Error
}

func (r *groupInvitationRepository) FindByID(id uint) (*models.GroupInvitation, error) {
	var inv models.GroupInvitation
	err := database.DB.Preload("Group").Preload("InvitedByUser").First(&inv, id).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *groupInvitationRepository) FindPendingByGroupAndUser(groupID, invitedUserID uint) (*models.GroupInvitation, error) {
	var inv models.GroupInvitation
	err := database.DB.
		Where("group_id = ? AND invited_user_id = ? AND status = ?", groupID, invitedUserID, models.GroupInvitationPending).
		First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *groupInvitationRepository) ListPendingByGroup(groupID uint) ([]models.GroupInvitation, error) {
	var list []models.GroupInvitation
	err := database.DB.
		Preload("InvitedUser").
		Where("group_id = ? AND status = ?", groupID, models.GroupInvitationPending).
		Find(&list).Error
	return list, err
}

func (r *groupInvitationRepository) ListPendingReceived(userID uint) ([]models.GroupInvitation, error) {
	var list []models.GroupInvitation
	err := database.DB.
		Preload("Group").
		Preload("InvitedByUser").
		Where("invited_user_id = ? AND status = ?", userID, models.GroupInvitationPending).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *groupInvitationRepository) UpdateStatus(id uint, status models.GroupInvitationStatus) error {
	now := time.Now()
	return database.DB.Model(&models.GroupInvitation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       status,
			"responded_at": now,
		}).Error
}
