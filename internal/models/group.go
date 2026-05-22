package models

import (
	"time"

	"gorm.io/gorm"
)

// Group represents a user group for task collaboration.
// Table name is quoted because "groups" is a reserved word in MySQL.
func (Group) TableName() string {
	return "`groups`"
}

type Group struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"type:varchar(100);not null"`
	CreatedBy uint           `json:"created_by" gorm:"not null;index"`
	IsDefault bool           `json:"is_default" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Creator   User           `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	Members   []User         `json:"members,omitempty" gorm:"many2many:group_members;"`
}

// GroupMember is the join table for group membership.
type GroupMember struct {
	GroupID  uint      `gorm:"primaryKey"`
	UserID   uint      `gorm:"primaryKey"`
	JoinedAt time.Time `json:"joined_at"`
}

func (GroupMember) TableName() string {
	return "group_members"
}

// GroupInvitationStatus represents invitation state.
type GroupInvitationStatus string

const (
	GroupInvitationPending   GroupInvitationStatus = "pending"
	GroupInvitationAccepted  GroupInvitationStatus = "accepted"
	GroupInvitationDeclined  GroupInvitationStatus = "declined"
	GroupInvitationCancelled GroupInvitationStatus = "cancelled"
)

// GroupInvitation represents a pending or resolved group invite.
type GroupInvitation struct {
	ID              uint                  `json:"id" gorm:"primaryKey"`
	GroupID         uint                  `json:"group_id" gorm:"not null;index"`
	InvitedUserID   uint                  `json:"invited_user_id" gorm:"not null;index"`
	InvitedByUserID uint                  `json:"invited_by_user_id" gorm:"not null;index"`
	Status          GroupInvitationStatus `json:"status" gorm:"type:varchar(20);not null;default:pending"`
	RespondedAt     *time.Time            `json:"responded_at,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	Group           Group                 `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	InvitedUser     User                  `json:"invited_user,omitempty" gorm:"foreignKey:InvitedUserID"`
	InvitedByUser   User                  `json:"invited_by_user,omitempty" gorm:"foreignKey:InvitedByUserID"`
}
