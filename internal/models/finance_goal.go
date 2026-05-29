package models

import (
	"time"

	"gorm.io/gorm"
)

// FinanceGoal is a household savings target.
type FinanceGoal struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	GroupID      uint           `json:"group_id" gorm:"not null;index"`
	Name         string         `json:"name" gorm:"type:varchar(120);not null"`
	TargetCents  int64          `json:"target_cents" gorm:"not null"`
	CurrentCents int64          `json:"current_cents" gorm:"not null;default:0"`
	TargetDate   *time.Time     `json:"target_date,omitempty" gorm:"type:date"`
	IsArchived   bool           `json:"is_archived" gorm:"not null;default:false"`
	CreatedBy    uint           `json:"created_by" gorm:"not null;index"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (FinanceGoal) TableName() string {
	return "finance_goals"
}
