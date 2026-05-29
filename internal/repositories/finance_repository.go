package repositories

import (
	"time"

	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"

	"gorm.io/gorm"
)

// FinanceRepository persists finance domain data.
type FinanceRepository interface {
	CountCategoriesByGroup(groupID uint) (int64, error)
	CreateCategories(categories []models.FinanceCategory) error
	ListCategories(groupID uint, kind *models.FinanceCategoryKind) ([]models.FinanceCategory, error)
	FindCategoryByID(groupID, categoryID uint) (*models.FinanceCategory, error)
	CreateCategory(category *models.FinanceCategory) error
	UpdateCategory(category *models.FinanceCategory) error
	DeleteCategory(groupID, categoryID uint) error
	CountTransactionsByCategory(groupID, categoryID uint) (int64, error)

	CreateAccount(account *models.FinanceAccount) error
	ListAccounts(groupID uint, includeArchived bool) ([]models.FinanceAccount, error)
	FindAccountByID(groupID, accountID uint) (*models.FinanceAccount, error)
	UpdateAccount(account *models.FinanceAccount) error
	SoftDeleteAccount(groupID, accountID uint) error
	CountTransactionsByAccount(groupID, accountID uint) (int64, error)

	CreateTransaction(tx *models.FinanceTransaction) error
	UpdateTransaction(tx *models.FinanceTransaction) error
	SoftDeleteTransaction(groupID, transactionID uint) error
	FindTransactionByID(groupID, transactionID uint) (*models.FinanceTransaction, error)
	ListTransactions(groupID uint, filter FinanceTransactionFilter) ([]models.FinanceTransaction, error)

	SumAccountBalanceCents(groupID, accountID uint, viewerUserID uint) (int64, error)
	AggregateDashboard(groupID uint, viewerUserID uint, from, to time.Time) (*FinanceDashboardAgg, error)

	GetMemberRole(groupID, userID uint) (*models.FinanceMemberRole, error)
	UpsertMemberRole(role *models.FinanceMemberRole) error
	ListMemberRoles(groupID uint) ([]models.FinanceMemberRole, error)
	IsGroupMember(groupID, userID uint) (bool, error)
	GetGroupCreatedBy(groupID uint) (uint, error)
}

// FinanceTransactionFilter scopes transaction queries.
type FinanceTransactionFilter struct {
	From       *time.Time
	To         *time.Time
	AccountID  *uint
	CategoryID *uint
	Type       *models.FinanceTransactionType
	ViewerID   uint
}

// FinanceDashboardAgg holds aggregated dashboard metrics.
type FinanceDashboardAgg struct {
	IncomeCents  int64
	ExpenseCents int64
	ByCategory   []FinanceCategoryTotal
	AccountIDs   []uint
}

// FinanceCategoryTotal is expense/income per category.
type FinanceCategoryTotal struct {
	CategoryID   uint
	CategoryName string
	Kind         models.FinanceCategoryKind
	TotalCents   int64
}

type financeRepository struct{}

// NewFinanceRepository creates a FinanceRepository.
func NewFinanceRepository() FinanceRepository {
	return &financeRepository{}
}

func (r *financeRepository) CountCategoriesByGroup(groupID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.FinanceCategory{}).Where("group_id = ?", groupID).Count(&count).Error
	return count, err
}

func (r *financeRepository) CreateCategories(categories []models.FinanceCategory) error {
	if len(categories) == 0 {
		return nil
	}
	return database.DB.Create(&categories).Error
}

func (r *financeRepository) ListCategories(groupID uint, kind *models.FinanceCategoryKind) ([]models.FinanceCategory, error) {
	var categories []models.FinanceCategory
	q := database.DB.Where("group_id = ?", groupID).Order("is_system DESC, name ASC")
	if kind != nil {
		q = q.Where("kind = ?", *kind)
	}
	if err := q.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *financeRepository) FindCategoryByID(groupID, categoryID uint) (*models.FinanceCategory, error) {
	var cat models.FinanceCategory
	if err := database.DB.Where("group_id = ? AND id = ?", groupID, categoryID).First(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *financeRepository) CreateCategory(category *models.FinanceCategory) error {
	return database.DB.Create(category).Error
}

func (r *financeRepository) UpdateCategory(category *models.FinanceCategory) error {
	return database.DB.Save(category).Error
}

func (r *financeRepository) DeleteCategory(groupID, categoryID uint) error {
	return database.DB.Where("group_id = ? AND id = ?", groupID, categoryID).Delete(&models.FinanceCategory{}).Error
}

func (r *financeRepository) CountTransactionsByCategory(groupID, categoryID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.FinanceTransaction{}).
		Where("group_id = ? AND category_id = ?", groupID, categoryID).
		Count(&count).Error
	return count, err
}

func (r *financeRepository) CreateAccount(account *models.FinanceAccount) error {
	return database.DB.Create(account).Error
}

func (r *financeRepository) ListAccounts(groupID uint, includeArchived bool) ([]models.FinanceAccount, error) {
	var accounts []models.FinanceAccount
	q := database.DB.Where("group_id = ?", groupID).Order("name ASC")
	if !includeArchived {
		q = q.Where("is_archived = ?", false)
	}
	if err := q.Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *financeRepository) FindAccountByID(groupID, accountID uint) (*models.FinanceAccount, error) {
	var account models.FinanceAccount
	if err := database.DB.Where("group_id = ? AND id = ?", groupID, accountID).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *financeRepository) UpdateAccount(account *models.FinanceAccount) error {
	return database.DB.Save(account).Error
}

func (r *financeRepository) SoftDeleteAccount(groupID, accountID uint) error {
	return database.DB.Where("group_id = ? AND id = ?", groupID, accountID).Delete(&models.FinanceAccount{}).Error
}

func (r *financeRepository) CountTransactionsByAccount(groupID, accountID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.FinanceTransaction{}).
		Where("group_id = ? AND (account_id = ? OR transfer_account_id = ?)", groupID, accountID, accountID).
		Count(&count).Error
	return count, err
}

func (r *financeRepository) CreateTransaction(tx *models.FinanceTransaction) error {
	return database.DB.Create(tx).Error
}

func (r *financeRepository) UpdateTransaction(tx *models.FinanceTransaction) error {
	return database.DB.Save(tx).Error
}

func (r *financeRepository) SoftDeleteTransaction(groupID, transactionID uint) error {
	return database.DB.Where("group_id = ? AND id = ?", groupID, transactionID).Delete(&models.FinanceTransaction{}).Error
}

func (r *financeRepository) FindTransactionByID(groupID, transactionID uint) (*models.FinanceTransaction, error) {
	var tx models.FinanceTransaction
	if err := database.DB.Where("group_id = ? AND id = ?", groupID, transactionID).First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func visibilityScope(viewerID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"(visibility = ? OR (visibility = ? AND created_by = ?))",
			models.FinanceVisibilityHousehold,
			models.FinanceVisibilityPrivate,
			viewerID,
		)
	}
}

func (r *financeRepository) ListTransactions(groupID uint, filter FinanceTransactionFilter) ([]models.FinanceTransaction, error) {
	var txs []models.FinanceTransaction
	q := database.DB.Where("group_id = ?", groupID).Scopes(visibilityScope(filter.ViewerID))
	if filter.From != nil {
		q = q.Where("date >= ?", filter.From.Format("2006-01-02"))
	}
	if filter.To != nil {
		q = q.Where("date <= ?", filter.To.Format("2006-01-02"))
	}
	if filter.AccountID != nil {
		q = q.Where("account_id = ? OR transfer_account_id = ?", *filter.AccountID, *filter.AccountID)
	}
	if filter.CategoryID != nil {
		q = q.Where("category_id = ?", *filter.CategoryID)
	}
	if filter.Type != nil {
		q = q.Where("type = ?", *filter.Type)
	}
	if err := q.Order("date DESC, id DESC").Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}

func (r *financeRepository) SumAccountBalanceCents(groupID, accountID uint, viewerUserID uint) (int64, error) {
	account, err := r.FindAccountByID(groupID, accountID)
	if err != nil {
		return 0, err
	}
	balance := account.InitialBalanceCents

	var txs []models.FinanceTransaction
	err = database.DB.Where("group_id = ?", groupID).
		Where("account_id = ? OR transfer_account_id = ?", accountID, accountID).
		Scopes(visibilityScope(viewerUserID)).
		Find(&txs).Error
	if err != nil {
		return 0, err
	}

	for _, tx := range txs {
		switch tx.Type {
		case models.FinanceTransactionIncome:
			if tx.AccountID == accountID {
				balance += tx.AmountCents
			}
		case models.FinanceTransactionExpense:
			if tx.AccountID == accountID {
				balance -= tx.AmountCents
			}
		case models.FinanceTransactionTransfer:
			if tx.AccountID == accountID {
				balance -= tx.AmountCents
			}
			if tx.TransferAccountID != nil && *tx.TransferAccountID == accountID {
				balance += tx.AmountCents
			}
		}
	}
	return balance, nil
}

func (r *financeRepository) AggregateDashboard(groupID uint, viewerUserID uint, from, to time.Time) (*FinanceDashboardAgg, error) {
	var txs []models.FinanceTransaction
	err := database.DB.Where("group_id = ?", groupID).
		Where("date >= ? AND date <= ?", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Scopes(visibilityScope(viewerUserID)).
		Find(&txs).Error
	if err != nil {
		return nil, err
	}

	agg := &FinanceDashboardAgg{}
	catTotals := make(map[uint]*FinanceCategoryTotal)
	accountSet := map[uint]struct{}{}

	for _, tx := range txs {
		accountSet[tx.AccountID] = struct{}{}
		if tx.TransferAccountID != nil {
			accountSet[*tx.TransferAccountID] = struct{}{}
		}
		switch tx.Type {
		case models.FinanceTransactionIncome:
			agg.IncomeCents += tx.AmountCents
			if tx.CategoryID != nil {
				r.addCategoryTotal(catTotals, groupID, *tx.CategoryID, tx.AmountCents)
			}
		case models.FinanceTransactionExpense:
			agg.ExpenseCents += tx.AmountCents
			if tx.CategoryID != nil {
				r.addCategoryTotal(catTotals, groupID, *tx.CategoryID, tx.AmountCents)
			}
		}
	}

	for _, v := range catTotals {
		agg.ByCategory = append(agg.ByCategory, *v)
	}
	for id := range accountSet {
		agg.AccountIDs = append(agg.AccountIDs, id)
	}
	return agg, nil
}

func (r *financeRepository) addCategoryTotal(m map[uint]*FinanceCategoryTotal, groupID, categoryID uint, cents int64) {
	if _, ok := m[categoryID]; ok {
		m[categoryID].TotalCents += cents
		return
	}
	cat, err := r.FindCategoryByID(groupID, categoryID)
	if err != nil {
		return
	}
	m[categoryID] = &FinanceCategoryTotal{
		CategoryID:   categoryID,
		CategoryName: cat.Name,
		Kind:         cat.Kind,
		TotalCents:   cents,
	}
}

func (r *financeRepository) GetMemberRole(groupID, userID uint) (*models.FinanceMemberRole, error) {
	var role models.FinanceMemberRole
	if err := database.DB.Where("group_id = ? AND user_id = ?", groupID, userID).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *financeRepository) UpsertMemberRole(role *models.FinanceMemberRole) error {
	return database.DB.Save(role).Error
}

func (r *financeRepository) ListMemberRoles(groupID uint) ([]models.FinanceMemberRole, error) {
	var roles []models.FinanceMemberRole
	if err := database.DB.Where("group_id = ?", groupID).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *financeRepository) IsGroupMember(groupID, userID uint) (bool, error) {
	var count int64
	err := database.DB.Table("group_members").
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *financeRepository) GetGroupCreatedBy(groupID uint) (uint, error) {
	var g models.Group
	if err := database.DB.Select("created_by").First(&g, groupID).Error; err != nil {
		return 0, err
	}
	return g.CreatedBy, nil
}
