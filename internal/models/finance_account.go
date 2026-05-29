package models

import (
	"time"

	"gorm.io/gorm"
)

// FinanceAccountType classifies a financial account.
type FinanceAccountType string

const (
	FinanceAccountChecking FinanceAccountType = "checking"
	FinanceAccountSavings  FinanceAccountType = "savings"
	FinanceAccountCash     FinanceAccountType = "cash"
	FinanceAccountOther    FinanceAccountType = "other"
)

// FinanceAccount is a household account (checking, wallet, etc.).
type FinanceAccount struct {
	ID                  uint               `json:"id" gorm:"primaryKey"`
	GroupID             uint               `json:"group_id" gorm:"not null;index"`
	Name                string             `json:"name" gorm:"type:varchar(100);not null"`
	Type                FinanceAccountType `json:"type" gorm:"type:varchar(20);not null"`
	Currency            string             `json:"currency" gorm:"type:varchar(3);not null;default:BRL"`
	InitialBalanceCents int64              `json:"initial_balance_cents" gorm:"not null;default:0"`
	IsArchived          bool               `json:"is_archived" gorm:"not null;default:false"`
	CreatedBy           uint               `json:"created_by" gorm:"not null;index"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
	DeletedAt           gorm.DeletedAt     `json:"-" gorm:"index"`
}

func (FinanceAccount) TableName() string {
	return "finance_accounts"
}
