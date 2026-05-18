package database

import (
	"fmt"
	"todo-go-backend/internal/config"
	"todo-go-backend/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect connects to the database using the provided configuration
func Connect(cfg *config.Config) error {
	var err error
	var dialector gorm.Dialector

	// Use MySQL if configured, otherwise use SQLite
	if cfg.UseMySQL() {
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DatabaseUser,
			cfg.DatabasePassword,
			cfg.DatabaseHost,
			cfg.DatabasePort,
			cfg.DatabaseName,
		)
		dialector = mysql.Open(dsn)
	} else {
		dialector = sqlite.Open(cfg.DatabasePath)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return err
	}

	// Auto migrate
	err = DB.AutoMigrate(
		&models.User{},
		&models.Task{},
		&models.TaskSharedWith{},
		&models.Tag{},
		&models.Comment{},
		&models.Notification{},
		&models.TokenDenylist{},
	)

	return err
}
