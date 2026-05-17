package models

import "time"

// TokenDenylist stores revoked JWT IDs until expiration.
type TokenDenylist struct {
	ID        uint      `gorm:"primaryKey"`
	JTI       string    `gorm:"type:varchar(64);uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
}
