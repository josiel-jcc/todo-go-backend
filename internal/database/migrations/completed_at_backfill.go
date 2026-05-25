package migrations

import (
	"todo-go-backend/internal/models"

	"gorm.io/gorm"
)

// BackfillCompletedAt sets completed_at from updated_at for tasks already marked completed.
func BackfillCompletedAt(db *gorm.DB) error {
	return db.Model(&models.Task{}).
		Where("completed = ? AND completed_at IS NULL", true).
		UpdateColumn("completed_at", gorm.Expr("updated_at")).Error
}
