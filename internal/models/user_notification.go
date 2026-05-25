package models

import (
	"time"

	"gorm.io/gorm"
)

// UserNotificationType represents in-app notification types.
type UserNotificationType string

const (
	UserNotificationTypeGroupInvite  UserNotificationType = "group_invite"
	UserNotificationTypeTaskReminder UserNotificationType = "task_reminder"
)

// UserNotification represents an in-app notification for a user.
type UserNotification struct {
	ID        uint                 `json:"id" gorm:"primaryKey"`
	UserID    uint                 `json:"user_id" gorm:"not null;index"`
	Type      UserNotificationType `json:"type" gorm:"type:varchar(30);not null"`
	Payload   string               `json:"payload" gorm:"type:text;not null"`
	Read      bool                 `json:"read" gorm:"column:read;default:false"`
	ReadAt    *time.Time           `json:"read_at,omitempty"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
	DeletedAt gorm.DeletedAt       `json:"-" gorm:"index"`
}
