package migrations

import (
	"time"
	"todo-go-backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const DefaultGroupName = "Os de casa"

// EnsureDefaultGroup creates the default group and adds all active users as members.
func EnsureDefaultGroup(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var group models.Group
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("is_default = ?", true).
			First(&group).Error
		if err == gorm.ErrRecordNotFound {
			var oldest models.User
			if err := tx.Order("id ASC").First(&oldest).Error; err != nil {
				return err
			}
			group = models.Group{
				Name:      DefaultGroupName,
				CreatedBy: oldest.ID,
				IsDefault: true,
			}
			if err := tx.Create(&group).Error; err != nil {
				return err
			}
			member := models.GroupMember{
				GroupID:  group.ID,
				UserID:   oldest.ID,
				JoinedAt: group.CreatedAt,
			}
			if err := tx.Where(models.GroupMember{GroupID: group.ID, UserID: oldest.ID}).
				FirstOrCreate(&member).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		var userIDs []uint
		if err := tx.Model(&models.User{}).Pluck("id", &userIDs).Error; err != nil {
			return err
		}

		now := time.Now()
		for _, uid := range userIDs {
			member := models.GroupMember{GroupID: group.ID, UserID: uid, JoinedAt: now}
			if err := tx.Where(models.GroupMember{GroupID: group.ID, UserID: uid}).
				FirstOrCreate(&member).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// AddUserToDefaultGroup adds a user to the default group, creating it if needed.
func AddUserToDefaultGroup(db *gorm.DB, userID uint) error {
	if err := EnsureDefaultGroup(db); err != nil {
		return err
	}
	var group models.Group
	if err := db.Where("is_default = ?", true).First(&group).Error; err != nil {
		return err
	}
	member := models.GroupMember{GroupID: group.ID, UserID: userID, JoinedAt: time.Now()}
	return db.Where(models.GroupMember{GroupID: group.ID, UserID: userID}).
		FirstOrCreate(&member).Error
}
