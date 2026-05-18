package models

import "time"

// UserPublic is the minimal user info exposed for task assignment lists.
type UserPublic struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
