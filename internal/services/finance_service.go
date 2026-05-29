package services

import (
	"errors"
	"strings"
	"time"

	apperrors "todo-go-backend/internal/errors"
	"todo-go-backend/internal/models"
	"todo-go-backend/internal/repositories"

	"gorm.io/gorm"
)

// FinanceService handles finance module business logic.
type FinanceService interface {
	EnsureAccess(userID, groupID uint) (*models.FinanceMemberRole, error)
	ListAccounts(userID, groupID uint) ([]AccountWithBalance, error)
	GetAccount(userID, groupID, accountID uint) (*AccountWithBalance, error)
	CreateAccount(userID, groupID uint, req CreateFinanceAccountRequest) (*models.FinanceAccount, error)
	UpdateAccount(userID, groupID, accountID uint, req UpdateFinanceAccountRequest) (*models.FinanceAccount, error)
	DeleteAccount(userID, groupID, accountID uint) error

	ListCategories(userID, groupID uint, kind *string) ([]models.FinanceCategory, error)
	CreateCategory(userID, groupID uint, req CreateFinanceCategoryRequest) (*models.FinanceCategory, error)
	UpdateCategory(userID, groupID, categoryID uint, req UpdateFinanceCategoryRequest) (*models.FinanceCategory, error)
	DeleteCategory(userID, groupID, categoryID uint) error

	ListTransactions(userID, groupID uint, filter ListFinanceTransactionsFilter) ([]models.FinanceTransaction, error)
	GetTransaction(userID, groupID, transactionID uint) (*models.FinanceTransaction, error)
	CreateTransaction(userID, groupID uint, req CreateFinanceTransactionRequest) (*models.FinanceTransaction, error)
	UpdateTransaction(userID, groupID, transactionID uint, req UpdateFinanceTransactionRequest) (*models.FinanceTransaction, error)
	DeleteTransaction(userID, groupID, transactionID uint) error

	GetDashboard(userID, groupID uint, month string) (*FinanceDashboard, error)
	ListCategoryBudgets(userID, groupID uint, month string) ([]FinanceCategoryBudgetItem, error)
	SetCategoryBudgets(userID, groupID uint, month string, items []SetCategoryBudgetItem) ([]FinanceCategoryBudgetItem, error)

	ListGoals(userID, groupID uint, includeArchived bool) ([]FinanceGoalItem, error)
	GetGoal(userID, groupID, goalID uint) (*FinanceGoalItem, error)
	CreateGoal(userID, groupID uint, req CreateFinanceGoalRequest) (*FinanceGoalItem, error)
	UpdateGoal(userID, groupID, goalID uint, req UpdateFinanceGoalRequest) (*FinanceGoalItem, error)
	DeleteGoal(userID, groupID, goalID uint) error

	ListMemberRoles(userID, groupID uint) ([]models.FinanceMemberRole, error)
	UpdateMemberRole(userID, groupID, targetUserID uint, role models.FinanceMemberRoleName) (*models.FinanceMemberRole, error)
}

// AccountWithBalance includes computed balance.
type AccountWithBalance struct {
	models.FinanceAccount
	BalanceCents int64 `json:"balance_cents"`
}

// FinanceDashboard is the monthly summary response.
type FinanceDashboard struct {
	Month    string                    `json:"month"`
	Currency string                    `json:"currency"`
	Totals   FinanceDashboardTotals    `json:"totals"`
	ByCategory []FinanceCategoryBreakdown `json:"by_category"`
	Accounts []AccountWithBalance      `json:"accounts"`
}

// FinanceDashboardTotals holds income/expense/net in cents.
type FinanceDashboardTotals struct {
	IncomeCents  int64 `json:"income_cents"`
	ExpenseCents int64 `json:"expense_cents"`
	NetCents     int64 `json:"net_cents"`
}

// FinanceCategoryBreakdown is per-category totals.
type FinanceCategoryBreakdown struct {
	CategoryID  uint     `json:"category_id"`
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	TotalCents  int64    `json:"total_cents"`
	BudgetCents *int64   `json:"budget_cents,omitempty"`
	PercentUsed *float64 `json:"percent_used,omitempty"`
}

// FinanceCategoryBudgetItem is a category budget for a month.
type FinanceCategoryBudgetItem struct {
	CategoryID   uint   `json:"category_id"`
	CategoryName string `json:"category_name"`
	LimitCents   int64  `json:"limit_cents"`
}

// SetCategoryBudgetItem sets or updates a budget line.
type SetCategoryBudgetItem struct {
	CategoryID uint
	LimitCents int64
}

// FinanceGoalItem is a savings goal with computed progress.
type FinanceGoalItem struct {
	models.FinanceGoal
	PercentComplete float64 `json:"percent_complete"`
	IsCompleted     bool    `json:"is_completed"`
}

type CreateFinanceGoalRequest struct {
	Name         string
	TargetCents  int64
	CurrentCents int64
	TargetDate   *time.Time
}

type UpdateFinanceGoalRequest struct {
	Name         *string
	TargetCents  *int64
	CurrentCents *int64
	TargetDate   *time.Time
	IsArchived   *bool
}

type CreateFinanceAccountRequest struct {
	Name                string
	Type                models.FinanceAccountType
	Currency            string
	InitialBalanceCents int64
}

type UpdateFinanceAccountRequest struct {
	Name       *string
	Type       *models.FinanceAccountType
	IsArchived *bool
}

type CreateFinanceCategoryRequest struct {
	Name  string
	Kind  models.FinanceCategoryKind
	Color string
}

type UpdateFinanceCategoryRequest struct {
	Name  *string
	Color *string
}

type CreateFinanceTransactionRequest struct {
	Type              models.FinanceTransactionType
	AccountID         uint
	TransferAccountID *uint
	CategoryID        *uint
	AmountCents       int64
	Description       string
	Date              time.Time
	Visibility        models.FinanceTransactionVisibility
}

type UpdateFinanceTransactionRequest struct {
	AccountID         *uint
	TransferAccountID *uint
	CategoryID        *uint
	AmountCents       *int64
	Description       *string
	Date              *time.Time
	Visibility        *models.FinanceTransactionVisibility
}

type ListFinanceTransactionsFilter struct {
	From       *time.Time
	To         *time.Time
	AccountID  *uint
	CategoryID *uint
	Type       *models.FinanceTransactionType
}

type financeService struct {
	financeRepo repositories.FinanceRepository
}

// NewFinanceService creates a FinanceService.
func NewFinanceService(financeRepo repositories.FinanceRepository) FinanceService {
	return &financeService{financeRepo: financeRepo}
}

func roleLevel(role models.FinanceMemberRoleName) int {
	switch role {
	case models.FinanceRoleAdmin:
		return 3
	case models.FinanceRoleEditor:
		return 2
	case models.FinanceRoleViewer:
		return 1
	default:
		return 0
	}
}

func (s *financeService) EnsureAccess(userID, groupID uint) (*models.FinanceMemberRole, error) {
	ok, err := s.financeRepo.IsGroupMember(groupID, userID)
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	if !ok {
		return nil, apperrors.NewForbiddenError()
	}

	role, err := s.financeRepo.GetMemberRole(groupID, userID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewInternalServerError(err)
		}
		if err := s.seedDefaults(groupID, userID); err != nil {
			return nil, err
		}
		role, err = s.financeRepo.GetMemberRole(groupID, userID)
		if err != nil {
			return nil, apperrors.NewInternalServerError(err)
		}
	}
	return role, nil
}

func (s *financeService) requireRole(userID, groupID uint, min models.FinanceMemberRoleName) (*models.FinanceMemberRole, error) {
	role, err := s.EnsureAccess(userID, groupID)
	if err != nil {
		return nil, err
	}
	if roleLevel(role.Role) < roleLevel(min) {
		return nil, apperrors.NewForbiddenError()
	}
	return role, nil
}

func (s *financeService) seedDefaults(groupID, userID uint) error {
	createdBy, err := s.financeRepo.GetGroupCreatedBy(groupID)
	if err != nil {
		return apperrors.NewInternalServerError(err)
	}
	defaultRole := models.FinanceRoleEditor
	if createdBy == userID {
		defaultRole = models.FinanceRoleAdmin
	}
	if err := s.financeRepo.UpsertMemberRole(&models.FinanceMemberRole{
		GroupID:    groupID,
		UserID:     userID,
		Role:       defaultRole,
		AssignedAt: time.Now(),
	}); err != nil {
		return apperrors.NewInternalServerError(err)
	}

	count, err := s.financeRepo.CountCategoriesByGroup(groupID)
	if err != nil {
		return apperrors.NewInternalServerError(err)
	}
	if count > 0 {
		return nil
	}

	now := time.Now()
	seeds := []models.FinanceCategory{
		{GroupID: groupID, Name: "Salário", Kind: models.FinanceCategoryIncome, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{GroupID: groupID, Name: "Outras receitas", Kind: models.FinanceCategoryIncome, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{GroupID: groupID, Name: "Alimentação", Kind: models.FinanceCategoryExpense, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{GroupID: groupID, Name: "Moradia", Kind: models.FinanceCategoryExpense, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{GroupID: groupID, Name: "Transporte", Kind: models.FinanceCategoryExpense, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{GroupID: groupID, Name: "Outras despesas", Kind: models.FinanceCategoryExpense, IsSystem: true, CreatedAt: now, UpdatedAt: now},
	}
	if err := s.financeRepo.CreateCategories(seeds); err != nil {
		return apperrors.NewInternalServerError(err)
	}
	return nil
}

func (s *financeService) ListAccounts(userID, groupID uint) ([]AccountWithBalance, error) {
	if _, err := s.EnsureAccess(userID, groupID); err != nil {
		return nil, err
	}
	accounts, err := s.financeRepo.ListAccounts(groupID, false)
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	out := make([]AccountWithBalance, 0, len(accounts))
	for _, a := range accounts {
		bal, err := s.financeRepo.SumAccountBalanceCents(groupID, a.ID, userID)
		if err != nil {
			return nil, apperrors.NewInternalServerError(err)
		}
		out = append(out, AccountWithBalance{FinanceAccount: a, BalanceCents: bal})
	}
	return out, nil
}

func (s *financeService) GetAccount(userID, groupID, accountID uint) (*AccountWithBalance, error) {
	if _, err := s.EnsureAccess(userID, groupID); err != nil {
		return nil, err
	}
	account, err := s.financeRepo.FindAccountByID(groupID, accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewTaskNotFoundError()
		}
		return nil, apperrors.NewInternalServerError(err)
	}
	bal, err := s.financeRepo.SumAccountBalanceCents(groupID, accountID, userID)
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	return &AccountWithBalance{FinanceAccount: *account, BalanceCents: bal}, nil
}

func (s *financeService) CreateAccount(userID, groupID uint, req CreateFinanceAccountRequest) (*models.FinanceAccount, error) {
	if _, err := s.requireRole(userID, groupID, models.FinanceRoleEditor); err != nil {
		return nil, err
	}
	currency := req.Currency
	if currency == "" {
		currency = "BRL"
	}
	account := &models.FinanceAccount{
		GroupID:             groupID,
		Name:                strings.TrimSpace(req.Name),
		Type:                req.Type,
		Currency:            currency,
		InitialBalanceCents: req.InitialBalanceCents,
		CreatedBy:           userID,
	}
	if account.Name == "" {
		return nil, apperrors.NewInvalidInputError("name is required")
	}
	if err := s.financeRepo.CreateAccount(account); err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	return account, nil
}

func (s *financeService) UpdateAccount(userID, groupID, accountID uint, req UpdateFinanceAccountRequest) (*models.FinanceAccount, error) {
	if _, err := s.requireRole(userID, groupID, models.FinanceRoleEditor); err != nil {
		return nil, err
	}
	account, err := s.financeRepo.FindAccountByID(groupID, accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewTaskNotFoundError()
		}
		return nil, apperrors.NewInternalServerError(err)
	}
	if req.Name != nil {
		account.Name = strings.TrimSpace(*req.Name)
	}
	if req.Type != nil {
		account.Type = *req.Type
	}
	if req.IsArchived != nil {
		account.IsArchived = *req.IsArchived
	}
	if err := s.financeRepo.UpdateAccount(account); err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	return account, nil
}

func (s *financeService) DeleteAccount(userID, groupID, accountID uint) error {
	role, err := s.requireRole(userID, groupID, models.FinanceRoleEditor)
	if err != nil {
		return err
	}
	if _, err := s.financeRepo.FindAccountByID(groupID, accountID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewTaskNotFoundError()
		}
		return apperrors.NewInternalServerError(err)
	}
	count, err := s.financeRepo.CountTransactionsByAccount(groupID, accountID)
	if err != nil {
		return apperrors.NewInternalServerError(err)
	}
	if count > 0 && role.Role != models.FinanceRoleAdmin {
		return apperrors.NewInvalidInputError("account has transactions; only admin can archive")
	}
	if err := s.financeRepo.SoftDeleteAccount(groupID, accountID); err != nil {
		return apperrors.NewInternalServerError(err)
	}
	return nil
}

func (s *financeService) ListCategories(userID, groupID uint, kind *string) ([]models.FinanceCategory, error) {
	if _, err := s.EnsureAccess(userID, groupID); err != nil {
		return nil, err
	}
	var k *models.FinanceCategoryKind
	if kind != nil && *kind != "" {
		kk := models.FinanceCategoryKind(*kind)
		k = &kk
	}
	cats, err := s.financeRepo.ListCategories(groupID, k)
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	return cats, nil
}

func (s *financeService) CreateCategory(userID, groupID uint, req CreateFinanceCategoryRequest) (*models.FinanceCategory, error) {
	if _, err := s.requireRole(userID, groupID, models.FinanceRoleEditor); err != nil {
		return nil, err
	}
	cat := &models.FinanceCategory{
		GroupID: groupID,
		Name:    strings.TrimSpace(req.Name),
		Kind:    req.Kind,
		Color:   req.Color,
	}
	if cat.Name == "" {
		return nil, apperrors.NewInvalidInputError("name is required")
	}
	if err := s.financeRepo.CreateCategory(cat); err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	return cat, nil
}

func (s *financeService) UpdateCategory(userID, groupID, categoryID uint, req UpdateFinanceCategoryRequest) (*models.FinanceCategory, error) {
	if _, err := s.requireRole(userID, groupID, models.FinanceRoleEditor); err != nil {
		return nil, err
	}
	cat, err := s.financeRepo.FindCategoryByID(groupID, categoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewTaskNotFoundError()
		}
		return nil, apperrors.NewInternalServerError(err)
	}
	if cat.IsSystem {
		return nil, apperrors.NewInvalidInputError("system categories cannot be edited")
	}
	if req.Name != nil {
		cat.Name = strings.TrimSpace(*req.Name)
	}
	if req.Color != nil {
		cat.Color = *req.Color
	}
	if err := s.financeRepo.UpdateCategory(cat); err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	return cat, nil
}

func (s *financeService) DeleteCategory(userID, groupID, categoryID uint) error {
	if _, err := s.requireRole(userID, groupID, models.FinanceRoleEditor); err != nil {
		return err
	}
	cat, err := s.financeRepo.FindCategoryByID(groupID, categoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewTaskNotFoundError()
		}
		return apperrors.NewInternalServerError(err)
	}
	if cat.IsSystem {
		return apperrors.NewInvalidInputError("system categories cannot be deleted")
	}
	count, err := s.financeRepo.CountTransactionsByCategory(groupID, categoryID)
	if err != nil {
		return apperrors.NewInternalServerError(err)
	}
	if count > 0 {
		return apperrors.NewInvalidInputError("category is in use")
	}
	if err := s.financeRepo.DeleteCategory(groupID, categoryID); err != nil {
		return apperrors.NewInternalServerError(err)
	}
	return nil
}

func (s *financeService) ListTransactions(userID, groupID uint, filter ListFinanceTransactionsFilter) ([]models.FinanceTransaction, error) {
	if _, err := s.EnsureAccess(userID, groupID); err != nil {
		return nil, err
	}
	txs, err := s.financeRepo.ListTransactions(groupID, repositories.FinanceTransactionFilter{
		From: filter.From, To: filter.To,
		AccountID: filter.AccountID, CategoryID: filter.CategoryID,
		Type: filter.Type, ViewerID: userID,
	})
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	return txs, nil
}

func (s *financeService) GetTransaction(userID, groupID, transactionID uint) (*models.FinanceTransaction, error) {
	if _, err := s.EnsureAccess(userID, groupID); err != nil {
		return nil, err
	}
	tx, err := s.financeRepo.FindTransactionByID(groupID, transactionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewTaskNotFoundError()
		}
		return nil, apperrors.NewInternalServerError(err)
	}
	if !s.canViewTransaction(userID, tx) {
		return nil, apperrors.NewForbiddenError()
	}
	return tx, nil
}

func (s *financeService) canViewTransaction(userID uint, tx *models.FinanceTransaction) bool {
	if tx.Visibility == models.FinanceVisibilityHousehold {
		return true
	}
	return tx.CreatedBy == userID
}

func (s *financeService) canEditTransaction(userID uint, role *models.FinanceMemberRole, tx *models.FinanceTransaction) bool {
	if tx.Visibility == models.FinanceVisibilityPrivate {
		return tx.CreatedBy == userID
	}
	if role.Role == models.FinanceRoleAdmin {
		return true
	}
	if role.Role == models.FinanceRoleEditor {
		return tx.CreatedBy == userID
	}
	return false
}

func (s *financeService) validateTransaction(groupID uint, req CreateFinanceTransactionRequest) error {
	if req.AmountCents <= 0 {
		return apperrors.NewInvalidInputError("amount must be positive")
	}
	account, err := s.financeRepo.FindAccountByID(groupID, req.AccountID)
	if err != nil {
		return apperrors.NewInvalidInputError("invalid account")
	}
	if account.IsArchived {
		return apperrors.NewInvalidInputError("account is archived")
	}

	switch req.Type {
	case models.FinanceTransactionIncome, models.FinanceTransactionExpense:
		if req.CategoryID == nil {
			return apperrors.NewInvalidInputError("category is required")
		}
		cat, err := s.financeRepo.FindCategoryByID(groupID, *req.CategoryID)
		if err != nil {
			return apperrors.NewInvalidInputError("invalid category")
		}
		if req.Type == models.FinanceTransactionIncome && cat.Kind != models.FinanceCategoryIncome {
			return apperrors.NewInvalidInputError("category must be income type")
		}
		if req.Type == models.FinanceTransactionExpense && cat.Kind != models.FinanceCategoryExpense {
			return apperrors.NewInvalidInputError("category must be expense type")
		}
	case models.FinanceTransactionTransfer:
		if req.TransferAccountID == nil {
			return apperrors.NewInvalidInputError("transfer_account_id is required for transfers")
		}
		if *req.TransferAccountID == req.AccountID {
			return apperrors.NewInvalidInputError("transfer accounts must differ")
		}
		dest, err := s.financeRepo.FindAccountByID(groupID, *req.TransferAccountID)
		if err != nil || dest.IsArchived {
			return apperrors.NewInvalidInputError("invalid transfer destination account")
		}
	default:
		return apperrors.NewInvalidInputError("invalid transaction type")
	}
	return nil
}

func (s *financeService) CreateTransaction(userID, groupID uint, req CreateFinanceTransactionRequest) (*models.FinanceTransaction, error) {
	role, err := s.requireRole(userID, groupID, models.FinanceRoleEditor)
	if err != nil {
		return nil, err
	}
	_ = role
	if err := s.validateTransaction(groupID, req); err != nil {
		return nil, err
	}
	vis := req.Visibility
	if vis == "" {
		vis = models.FinanceVisibilityHousehold
	}
	tx := &models.FinanceTransaction{
		GroupID:           groupID,
		Type:              req.Type,
		AccountID:         req.AccountID,
		TransferAccountID: req.TransferAccountID,
		CategoryID:        req.CategoryID,
		AmountCents:       req.AmountCents,
		Description:       strings.TrimSpace(req.Description),
		Date:              req.Date,
		Visibility:        vis,
		CreatedBy:         userID,
	}
	if err := s.financeRepo.CreateTransaction(tx); err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	return tx, nil
}

func (s *financeService) UpdateTransaction(userID, groupID, transactionID uint, req UpdateFinanceTransactionRequest) (*models.FinanceTransaction, error) {
	role, err := s.requireRole(userID, groupID, models.FinanceRoleEditor)
	if err != nil {
		return nil, err
	}
	tx, err := s.financeRepo.FindTransactionByID(groupID, transactionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewTaskNotFoundError()
		}
		return nil, apperrors.NewInternalServerError(err)
	}
	if !s.canEditTransaction(userID, role, tx) {
		return nil, apperrors.NewForbiddenError()
	}
	if req.AccountID != nil {
		tx.AccountID = *req.AccountID
	}
	if req.TransferAccountID != nil {
		tx.TransferAccountID = req.TransferAccountID
	}
	if req.CategoryID != nil {
		tx.CategoryID = req.CategoryID
	}
	if req.AmountCents != nil {
		tx.AmountCents = *req.AmountCents
	}
	if req.Description != nil {
		tx.Description = strings.TrimSpace(*req.Description)
	}
	if req.Date != nil {
		tx.Date = *req.Date
	}
	if req.Visibility != nil {
		tx.Visibility = *req.Visibility
	}
	validateReq := CreateFinanceTransactionRequest{
		Type:              tx.Type,
		AccountID:         tx.AccountID,
		TransferAccountID: tx.TransferAccountID,
		CategoryID:        tx.CategoryID,
		AmountCents:       tx.AmountCents,
		Date:              tx.Date,
		Visibility:        tx.Visibility,
	}
	if err := s.validateTransaction(groupID, validateReq); err != nil {
		return nil, err
	}
	if err := s.financeRepo.UpdateTransaction(tx); err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	return tx, nil
}

func (s *financeService) DeleteTransaction(userID, groupID, transactionID uint) error {
	role, err := s.requireRole(userID, groupID, models.FinanceRoleEditor)
	if err != nil {
		return err
	}
	tx, err := s.financeRepo.FindTransactionByID(groupID, transactionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewTaskNotFoundError()
		}
		return apperrors.NewInternalServerError(err)
	}
	if !s.canEditTransaction(userID, role, tx) {
		return apperrors.NewForbiddenError()
	}
	if err := s.financeRepo.SoftDeleteTransaction(groupID, transactionID); err != nil {
		return apperrors.NewInternalServerError(err)
	}
	return nil
}

func parseMonth(month string) (time.Time, time.Time, string, error) {
	if month == "" {
		now := time.Now()
		month = now.Format("2006-01")
	}
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return time.Time{}, time.Time{}, "", apperrors.NewInvalidInputError("month must be YYYY-MM")
	}
	from := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, -1)
	return from, to, month, nil
}

func (s *financeService) GetDashboard(userID, groupID uint, month string) (*FinanceDashboard, error) {
	if _, err := s.EnsureAccess(userID, groupID); err != nil {
		return nil, err
	}
	from, to, monthStr, err := parseMonth(month)
	if err != nil {
		return nil, err
	}
	agg, err := s.financeRepo.AggregateDashboard(groupID, userID, from, to)
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	dash := &FinanceDashboard{
		Month:    monthStr,
		Currency: "BRL",
		Totals: FinanceDashboardTotals{
			IncomeCents:  agg.IncomeCents,
			ExpenseCents: agg.ExpenseCents,
			NetCents:     agg.IncomeCents - agg.ExpenseCents,
		},
	}
	budgets, err := s.financeRepo.ListCategoryBudgets(groupID, monthStr)
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	budgetByCategory := map[uint]int64{}
	for _, b := range budgets {
		budgetByCategory[b.CategoryID] = b.LimitCents
	}
	for _, c := range agg.ByCategory {
		item := FinanceCategoryBreakdown{
			CategoryID: c.CategoryID, Name: c.CategoryName, Kind: string(c.Kind), TotalCents: c.TotalCents,
		}
		if limit, ok := budgetByCategory[c.CategoryID]; ok && c.Kind == models.FinanceCategoryExpense {
			item.BudgetCents = &limit
			if limit > 0 {
				pct := float64(c.TotalCents) / float64(limit) * 100
				item.PercentUsed = &pct
			}
		}
		dash.ByCategory = append(dash.ByCategory, item)
	}
	accounts, err := s.ListAccounts(userID, groupID)
	if err != nil {
		return nil, err
	}
	accountMap := map[uint]AccountWithBalance{}
	for _, a := range accounts {
		accountMap[a.ID] = a
	}
	for _, id := range agg.AccountIDs {
		if a, ok := accountMap[id]; ok {
			dash.Accounts = append(dash.Accounts, a)
		}
	}
	if len(dash.Accounts) == 0 {
		dash.Accounts = accounts
	}
	return dash, nil
}

func (s *financeService) ListCategoryBudgets(userID, groupID uint, month string) ([]FinanceCategoryBudgetItem, error) {
	if _, err := s.EnsureAccess(userID, groupID); err != nil {
		return nil, err
	}
	_, _, monthStr, err := parseMonth(month)
	if err != nil {
		return nil, err
	}
	return s.buildBudgetItems(groupID, monthStr)
}

func (s *financeService) SetCategoryBudgets(userID, groupID uint, month string, items []SetCategoryBudgetItem) ([]FinanceCategoryBudgetItem, error) {
	if _, err := s.requireRole(userID, groupID, models.FinanceRoleEditor); err != nil {
		return nil, err
	}
	_, _, monthStr, err := parseMonth(month)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.LimitCents < 0 {
			return nil, apperrors.NewInvalidInputError("limit must be non-negative")
		}
		cat, err := s.financeRepo.FindCategoryByID(groupID, item.CategoryID)
		if err != nil {
			return nil, apperrors.NewInvalidInputError("invalid category")
		}
		if cat.Kind != models.FinanceCategoryExpense {
			return nil, apperrors.NewInvalidInputError("budget applies to expense categories only")
		}
		if err := s.financeRepo.UpsertCategoryBudget(&models.FinanceCategoryBudget{
			GroupID: groupID, CategoryID: item.CategoryID, Month: monthStr,
			LimitCents: item.LimitCents, CreatedBy: userID,
		}); err != nil {
			return nil, apperrors.NewInternalServerError(err)
		}
	}
	return s.buildBudgetItems(groupID, monthStr)
}

func (s *financeService) buildBudgetItems(groupID uint, month string) ([]FinanceCategoryBudgetItem, error) {
	budgets, err := s.financeRepo.ListCategoryBudgets(groupID, month)
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	out := make([]FinanceCategoryBudgetItem, 0, len(budgets))
	for _, b := range budgets {
		cat, err := s.financeRepo.FindCategoryByID(groupID, b.CategoryID)
		if err != nil {
			continue
		}
		out = append(out, FinanceCategoryBudgetItem{
			CategoryID: b.CategoryID, CategoryName: cat.Name, LimitCents: b.LimitCents,
		})
	}
	return out, nil
}

func goalProgress(goal *models.FinanceGoal) FinanceGoalItem {
	item := FinanceGoalItem{FinanceGoal: *goal}
	if goal.TargetCents > 0 {
		pct := float64(goal.CurrentCents) / float64(goal.TargetCents) * 100
		if pct > 100 {
			pct = 100
		}
		item.PercentComplete = pct
	}
	item.IsCompleted = goal.CurrentCents >= goal.TargetCents && goal.TargetCents > 0
	return item
}

func validateGoalAmounts(targetCents, currentCents int64) error {
	if targetCents <= 0 {
		return apperrors.NewInvalidInputError("target must be positive")
	}
	if currentCents < 0 {
		return apperrors.NewInvalidInputError("current amount must be non-negative")
	}
	return nil
}

func (s *financeService) ListGoals(userID, groupID uint, includeArchived bool) ([]FinanceGoalItem, error) {
	if _, err := s.EnsureAccess(userID, groupID); err != nil {
		return nil, err
	}
	goals, err := s.financeRepo.ListGoals(groupID, includeArchived)
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	out := make([]FinanceGoalItem, 0, len(goals))
	for i := range goals {
		out = append(out, goalProgress(&goals[i]))
	}
	return out, nil
}

func (s *financeService) GetGoal(userID, groupID, goalID uint) (*FinanceGoalItem, error) {
	if _, err := s.EnsureAccess(userID, groupID); err != nil {
		return nil, err
	}
	goal, err := s.financeRepo.FindGoalByID(groupID, goalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewTaskNotFoundError()
		}
		return nil, apperrors.NewInternalServerError(err)
	}
	item := goalProgress(goal)
	return &item, nil
}

func (s *financeService) CreateGoal(userID, groupID uint, req CreateFinanceGoalRequest) (*FinanceGoalItem, error) {
	if _, err := s.requireRole(userID, groupID, models.FinanceRoleEditor); err != nil {
		return nil, err
	}
	if err := validateGoalAmounts(req.TargetCents, req.CurrentCents); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.NewInvalidInputError("name is required")
	}
	goal := &models.FinanceGoal{
		GroupID: groupID, Name: name, TargetCents: req.TargetCents,
		CurrentCents: req.CurrentCents, TargetDate: req.TargetDate, CreatedBy: userID,
	}
	if err := s.financeRepo.CreateGoal(goal); err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	item := goalProgress(goal)
	return &item, nil
}

func (s *financeService) UpdateGoal(userID, groupID, goalID uint, req UpdateFinanceGoalRequest) (*FinanceGoalItem, error) {
	if _, err := s.requireRole(userID, groupID, models.FinanceRoleEditor); err != nil {
		return nil, err
	}
	goal, err := s.financeRepo.FindGoalByID(groupID, goalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewTaskNotFoundError()
		}
		return nil, apperrors.NewInternalServerError(err)
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, apperrors.NewInvalidInputError("name is required")
		}
		goal.Name = name
	}
	if req.TargetCents != nil {
		goal.TargetCents = *req.TargetCents
	}
	if req.CurrentCents != nil {
		goal.CurrentCents = *req.CurrentCents
	}
	if req.TargetDate != nil {
		goal.TargetDate = req.TargetDate
	}
	if req.IsArchived != nil {
		goal.IsArchived = *req.IsArchived
	}
	if err := validateGoalAmounts(goal.TargetCents, goal.CurrentCents); err != nil {
		return nil, err
	}
	if err := s.financeRepo.UpdateGoal(goal); err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	item := goalProgress(goal)
	return &item, nil
}

func (s *financeService) DeleteGoal(userID, groupID, goalID uint) error {
	if _, err := s.requireRole(userID, groupID, models.FinanceRoleEditor); err != nil {
		return err
	}
	if _, err := s.financeRepo.FindGoalByID(groupID, goalID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewTaskNotFoundError()
		}
		return apperrors.NewInternalServerError(err)
	}
	if err := s.financeRepo.SoftDeleteGoal(groupID, goalID); err != nil {
		return apperrors.NewInternalServerError(err)
	}
	return nil
}

func (s *financeService) ListMemberRoles(userID, groupID uint) ([]models.FinanceMemberRole, error) {
	if _, err := s.EnsureAccess(userID, groupID); err != nil {
		return nil, err
	}
	roles, err := s.financeRepo.ListMemberRoles(groupID)
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	return roles, nil
}

func (s *financeService) UpdateMemberRole(userID, groupID, targetUserID uint, role models.FinanceMemberRoleName) (*models.FinanceMemberRole, error) {
	if _, err := s.requireRole(userID, groupID, models.FinanceRoleAdmin); err != nil {
		return nil, err
	}
	ok, err := s.financeRepo.IsGroupMember(groupID, targetUserID)
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	if !ok {
		return nil, apperrors.NewInvalidInputError("user is not a group member")
	}
	switch role {
	case models.FinanceRoleAdmin, models.FinanceRoleEditor, models.FinanceRoleViewer:
	default:
		return nil, apperrors.NewInvalidInputError("invalid role")
	}
	mr := &models.FinanceMemberRole{
		GroupID: groupID, UserID: targetUserID, Role: role, AssignedAt: time.Now(),
	}
	if err := s.financeRepo.UpsertMemberRole(mr); err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}
	return mr, nil
}
