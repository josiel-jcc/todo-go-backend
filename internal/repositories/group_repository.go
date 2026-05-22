package repositories

import (
	"time"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"

	"gorm.io/gorm"
)

type GroupRepository interface {
	Create(group *models.Group) error
	FindByID(id uint) (*models.Group, error)
	FindByName(name string) (*models.Group, error)
	FindDefaultGroup() (*models.Group, error)
	FindByUserID(userID uint) ([]models.Group, error)
	Update(group *models.Group) error
	Delete(id uint) error
	IsMember(groupID, userID uint) (bool, error)
	AddMember(groupID, userID uint) error
	RemoveMember(groupID, userID uint) error
	ListMemberIDs(groupID uint) ([]uint, error)
	ListMembers(groupID uint) ([]models.User, error)
	CountMembers(groupID uint) (int64, error)
	UsersShareAtLeastOneGroup(userID, otherUserID uint) (bool, error)
	FindCoGroupUserIDs(userID uint) ([]uint, error)
	FindAllUserIDs() ([]uint, error)
	FindOldestUserID() (*uint, error)
}

type groupRepository struct{}

func NewGroupRepository() GroupRepository {
	return &groupRepository{}
}

func (r *groupRepository) Create(group *models.Group) error {
	return database.DB.Create(group).Error
}

func (r *groupRepository) FindByID(id uint) (*models.Group, error) {
	var group models.Group
	if err := database.DB.First(&group, id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) FindByName(name string) (*models.Group, error) {
	var group models.Group
	if err := database.DB.Where("name = ?", name).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) FindDefaultGroup() (*models.Group, error) {
	var group models.Group
	if err := database.DB.Where("is_default = ?", true).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) FindByUserID(userID uint) ([]models.Group, error) {
	var groups []models.Group
	err := database.DB.
		Joins("JOIN group_members ON group_members.group_id = `groups`.id").
		Where("group_members.user_id = ?", userID).
		Find(&groups).Error
	return groups, err
}

func (r *groupRepository) Update(group *models.Group) error {
	return database.DB.Save(group).Error
}

func (r *groupRepository) Delete(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", id).Delete(&models.GroupMember{}).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", id).Delete(&models.GroupInvitation{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Group{}, id).Error
	})
}

func (r *groupRepository) IsMember(groupID, userID uint) (bool, error) {
	var count int64
	err := database.DB.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *groupRepository) AddMember(groupID, userID uint) error {
	member := models.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}
	return database.DB.Where(models.GroupMember{GroupID: groupID, UserID: userID}).
		FirstOrCreate(&member).Error
}

func (r *groupRepository) RemoveMember(groupID, userID uint) error {
	return database.DB.Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&models.GroupMember{}).Error
}

func (r *groupRepository) ListMemberIDs(groupID uint) ([]uint, error) {
	var ids []uint
	err := database.DB.Model(&models.GroupMember{}).
		Where("group_id = ?", groupID).
		Pluck("user_id", &ids).Error
	return ids, err
}

func (r *groupRepository) ListMembers(groupID uint) ([]models.User, error) {
	var users []models.User
	err := database.DB.
		Joins("JOIN group_members ON group_members.user_id = users.id").
		Where("group_members.group_id = ?", groupID).
		Find(&users).Error
	return users, err
}

func (r *groupRepository) CountMembers(groupID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.GroupMember{}).
		Where("group_id = ?", groupID).
		Count(&count).Error
	return count, err
}

func (r *groupRepository) UsersShareAtLeastOneGroup(userID, otherUserID uint) (bool, error) {
	if userID == otherUserID {
		return true, nil
	}
	var count int64
	err := database.DB.Table("group_members AS a").
		Joins("JOIN group_members AS b ON a.group_id = b.group_id").
		Where("a.user_id = ? AND b.user_id = ?", userID, otherUserID).
		Count(&count).Error
	return count > 0, err
}

func (r *groupRepository) FindCoGroupUserIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := database.DB.Table("group_members AS a").
		Joins("JOIN group_members AS b ON a.group_id = b.group_id").
		Where("a.user_id = ? AND b.user_id != ?", userID, userID).
		Distinct("b.user_id").
		Pluck("b.user_id", &ids).Error
	return ids, err
}

func (r *groupRepository) FindAllUserIDs() ([]uint, error) {
	var ids []uint
	err := database.DB.Model(&models.User{}).Pluck("id", &ids).Error
	return ids, err
}

func (r *groupRepository) FindOldestUserID() (*uint, error) {
	var user models.User
	err := database.DB.Order("id ASC").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user.ID, nil
}
