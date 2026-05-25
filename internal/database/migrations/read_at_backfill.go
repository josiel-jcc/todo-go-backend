package migrations

import (
	"todo-go-backend/internal/models"

	"gorm.io/gorm"
)

// BackfillReadAt sets read_at from updated_at for notifications already marked read.
func BackfillReadAt(db *gorm.DB) error {
	return db.Model(&models.UserNotification{}).
		Where("`read` = ? AND read_at IS NULL", true).
		UpdateColumn("read_at", gorm.Expr("updated_at")).Error
}
