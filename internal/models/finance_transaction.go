package models

import (
	"time"

	"gorm.io/gorm"
)

// FinanceTransactionType is income, expense, or transfer.
type FinanceTransactionType string

const (
	FinanceTransactionIncome   FinanceTransactionType = "income"
	FinanceTransactionExpense  FinanceTransactionType = "expense"
	FinanceTransactionTransfer FinanceTransactionType = "transfer"
)

// FinanceTransactionVisibility controls household vs private scope.
type FinanceTransactionVisibility string

const (
	FinanceVisibilityPrivate   FinanceTransactionVisibility = "private"
	FinanceVisibilityHousehold FinanceTransactionVisibility = "household"
)

// FinanceTransaction is a ledger entry.
type FinanceTransaction struct {
	ID             uint                         `json:"id" gorm:"primaryKey"`
	GroupID        uint                         `json:"group_id" gorm:"not null;index:idx_fin_tx_group_date"`
	AccountID          uint                         `json:"account_id" gorm:"not null;index"`
	TransferAccountID  *uint                        `json:"transfer_account_id,omitempty" gorm:"index"`
	CategoryID         *uint                        `json:"category_id,omitempty" gorm:"index"`
	Type               FinanceTransactionType       `json:"type" gorm:"type:varchar(20);not null"`
	AmountCents        int64                        `json:"amount_cents" gorm:"not null"`
	Description        string                       `json:"description" gorm:"type:varchar(255)"`
	Date               time.Time                    `json:"date" gorm:"type:date;not null;index:idx_fin_tx_group_date"`
	Visibility         FinanceTransactionVisibility `json:"visibility" gorm:"type:varchar(20);not null;default:household"`
	CreatedBy      uint                         `json:"created_by" gorm:"not null;index"`
	CreatedAt      time.Time                    `json:"created_at"`
	UpdatedAt      time.Time                    `json:"updated_at"`
	DeletedAt      gorm.DeletedAt               `json:"-" gorm:"index"`
}

func (FinanceTransaction) TableName() string {
	return "finance_transactions"
}
