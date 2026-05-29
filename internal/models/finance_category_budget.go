package models

import "time"

// FinanceCategoryBudget is a monthly spending limit for an expense category.
type FinanceCategoryBudget struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	GroupID     uint      `json:"group_id" gorm:"not null;uniqueIndex:uk_fin_budget_group_cat_month"`
	CategoryID  uint      `json:"category_id" gorm:"not null;uniqueIndex:uk_fin_budget_group_cat_month"`
	Month       string    `json:"month" gorm:"type:varchar(7);not null;uniqueIndex:uk_fin_budget_group_cat_month"`
	LimitCents  int64     `json:"limit_cents" gorm:"not null"`
	CreatedBy   uint      `json:"created_by" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (FinanceCategoryBudget) TableName() string {
	return "finance_category_budgets"
}
