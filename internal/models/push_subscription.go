package models

import (
	"time"

	"gorm.io/gorm"
)

// PushSubscription stores a Web Push subscription for a user device.
type PushSubscription struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null;uniqueIndex:idx_push_subscriptions_user_endpoint"`
	Endpoint  string         `json:"endpoint" gorm:"type:text;not null;uniqueIndex:idx_push_subscriptions_user_endpoint"`
	P256dh    string         `json:"p256dh" gorm:"type:varchar(255);not null"`
	Auth      string         `json:"auth" gorm:"type:varchar(255);not null"`
	UserAgent string         `json:"user_agent,omitempty" gorm:"type:varchar(512)"`
	User      User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
