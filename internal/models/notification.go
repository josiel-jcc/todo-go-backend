package models

import (
	"time"

	"gorm.io/gorm"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// NotificationTypeDueSoon represents notification for tasks due soon (1 day before)
	NotificationTypeDueSoon NotificationType = "due_soon"
	// NotificationTypeDueToday represents notification for tasks due today
	NotificationTypeDueToday NotificationType = "due_today"
	// NotificationTypeOverdue represents notification for overdue tasks
	NotificationTypeOverdue NotificationType = "overdue"
	// NotificationTypeTaskReminder represents a timed reminder before task due_date
	NotificationTypeTaskReminder NotificationType = "task_reminder"
)

// NotificationChannel represents the channel used to send notification
type NotificationChannel string

const (
	// NotificationChannelEmail represents email channel
	NotificationChannelEmail NotificationChannel = "email"
	// NotificationChannelTelegram represents Telegram channel
	NotificationChannelTelegram NotificationChannel = "telegram"
	// NotificationChannelPush represents Web Push channel
	NotificationChannelPush NotificationChannel = "push"
)

// Notification represents a sent notification
type Notification struct {
	ID        uint                `json:"id" gorm:"primaryKey"`
	UserID    uint                 `json:"user_id" gorm:"not null;index"`
	TaskID    uint                 `json:"task_id" gorm:"not null;index"`
	Type      NotificationType     `json:"type" gorm:"type:varchar(20);not null"`
	Channel     NotificationChannel  `json:"channel" gorm:"type:varchar(20);not null"`
	TaskDueDate *time.Time           `json:"task_due_date,omitempty"` // UTC due_date at send time (dedup fingerprint)
	SentAt      time.Time            `json:"sent_at"`
	User      User                 `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Task      Task                 `json:"task,omitempty" gorm:"foreignKey:TaskID"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
	DeletedAt gorm.DeletedAt       `json:"-" gorm:"index"`
}

