package models

import (
	"time"

	"gorm.io/gorm"
)

// FinanceCategoryKind is income or expense.
type FinanceCategoryKind string

const (
	FinanceCategoryIncome  FinanceCategoryKind = "income"
	FinanceCategoryExpense FinanceCategoryKind = "expense"
)

// FinanceCategory groups transactions for reporting.
type FinanceCategory struct {
	ID        uint                `json:"id" gorm:"primaryKey"`
	GroupID   uint                `json:"group_id" gorm:"not null;index"`
	Name      string              `json:"name" gorm:"type:varchar(100);not null"`
	Kind      FinanceCategoryKind `json:"kind" gorm:"type:varchar(20);not null"`
	Color     string              `json:"color,omitempty" gorm:"type:varchar(7)"`
	IsSystem  bool                `json:"is_system" gorm:"not null;default:false"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	DeletedAt gorm.DeletedAt      `json:"-" gorm:"index"`
}

func (FinanceCategory) TableName() string {
	return "finance_categories"
}
