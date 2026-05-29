package models

import "time"

// FinanceMemberRoleName is the financial permission level within a group.
type FinanceMemberRoleName string

const (
	FinanceRoleAdmin  FinanceMemberRoleName = "admin"
	FinanceRoleEditor FinanceMemberRoleName = "editor"
	FinanceRoleViewer FinanceMemberRoleName = "viewer"
)

// FinanceMemberRole assigns finance permissions per user per group.
type FinanceMemberRole struct {
	GroupID    uint                  `json:"group_id" gorm:"primaryKey"`
	UserID     uint                  `json:"user_id" gorm:"primaryKey"`
	Role       FinanceMemberRoleName `json:"role" gorm:"type:varchar(20);not null"`
	AssignedAt time.Time             `json:"assigned_at"`
}

func (FinanceMemberRole) TableName() string {
	return "finance_member_roles"
}
