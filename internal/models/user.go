package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID                   uint           `json:"id" gorm:"primaryKey"`
	Username             string         `json:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
	Email                string         `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Password             string         `json:"-" gorm:"type:varchar(255);not null"`       // Hashed password, not exposed in JSON
	TelegramChatID       *string        `json:"telegram_chat_id" gorm:"type:varchar(50)"`  // Telegram chat ID for notifications
	NotificationsEnabled  bool           `json:"notifications_enabled" gorm:"default:false"` // Opt-in notifications
	ReminderMinutesBefore int            `json:"reminder_minutes_before" gorm:"not null;default:10"`
	TermsAcceptedAt       *time.Time     `json:"terms_accepted_at,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`
}
