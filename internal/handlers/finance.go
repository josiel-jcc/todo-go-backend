package handlers

import (
	"net/http"
	"strconv"
	"time"

	"todo-go-backend/internal/models"
	"todo-go-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// FinanceHandler serves finance module endpoints.
type FinanceHandler struct {
	financeService services.FinanceService
	groupService   services.GroupService
}

// NewFinanceHandler creates a FinanceHandler.
func NewFinanceHandler(financeService services.FinanceService, groupService services.GroupService) *FinanceHandler {
	return &FinanceHandler{financeService: financeService, groupService: groupService}
}

// FinanceHealthResponse is returned by the finance module health check.
type FinanceHealthResponse struct {
	Status  string `json:"status"`
	GroupID uint   `json:"group_id"`
}

// Health confirms the finance module is available for a group the user belongs to.
func (h *FinanceHandler) Health(c *gin.Context) {
	userID := c.GetUint("user_id")
	groupID, err := parseUintParam(c, "groupId")
	if err != nil {
		handleError(c, err)
		return
	}
	if _, err := h.groupService.GetGroup(userID, groupID); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, FinanceHealthResponse{Status: "finance_module", GroupID: groupID})
}

func (h *FinanceHandler) groupIDParam(c *gin.Context) (uint, uint, bool) {
	userID := c.GetUint("user_id")
	groupID, err := parseUintParam(c, "groupId")
	if err != nil {
		handleError(c, err)
		return 0, 0, false
	}
	return userID, groupID, true
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

type createAccountBody struct {
	Name                string                     `json:"name" binding:"required,min=1,max=100"`
	Type                models.FinanceAccountType  `json:"type" binding:"required"`
	Currency            string                     `json:"currency"`
	InitialBalanceCents int64                      `json:"initial_balance_cents"`
}

type updateAccountBody struct {
	Name       *string                    `json:"name"`
	Type       *models.FinanceAccountType `json:"type"`
	IsArchived *bool                      `json:"is_archived"`
}

func (h *FinanceHandler) ListAccounts(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	accounts, err := h.financeService.ListAccounts(userID, groupID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func (h *FinanceHandler) GetAccount(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	accountID, err := parseUintParam(c, "accountId")
	if err != nil {
		handleError(c, err)
		return
	}
	account, err := h.financeService.GetAccount(userID, groupID, accountID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, account)
}

func (h *FinanceHandler) CreateAccount(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	var body createAccountBody
	if err := c.ShouldBindJSON(&body); err != nil {
		handleValidationError(c, err)
		return
	}
	account, err := h.financeService.CreateAccount(userID, groupID, services.CreateFinanceAccountRequest{
		Name: body.Name, Type: body.Type, Currency: body.Currency,
		InitialBalanceCents: body.InitialBalanceCents,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, account)
}

func (h *FinanceHandler) UpdateAccount(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	accountID, err := parseUintParam(c, "accountId")
	if err != nil {
		handleError(c, err)
		return
	}
	var body updateAccountBody
	if err := c.ShouldBindJSON(&body); err != nil {
		handleValidationError(c, err)
		return
	}
	account, err := h.financeService.UpdateAccount(userID, groupID, accountID, services.UpdateFinanceAccountRequest{
		Name: body.Name, Type: body.Type, IsArchived: body.IsArchived,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, account)
}

func (h *FinanceHandler) DeleteAccount(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	accountID, err := parseUintParam(c, "accountId")
	if err != nil {
		handleError(c, err)
		return
	}
	if err := h.financeService.DeleteAccount(userID, groupID, accountID); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type createCategoryBody struct {
	Name  string                    `json:"name" binding:"required,min=1,max=100"`
	Kind  models.FinanceCategoryKind `json:"kind" binding:"required"`
	Color string                    `json:"color"`
}

type updateCategoryBody struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

func (h *FinanceHandler) ListCategories(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	kind := c.Query("kind")
	var kindPtr *string
	if kind != "" {
		kindPtr = &kind
	}
	categories, err := h.financeService.ListCategories(userID, groupID, kindPtr)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (h *FinanceHandler) CreateCategory(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	var body createCategoryBody
	if err := c.ShouldBindJSON(&body); err != nil {
		handleValidationError(c, err)
		return
	}
	cat, err := h.financeService.CreateCategory(userID, groupID, services.CreateFinanceCategoryRequest{
		Name: body.Name, Kind: body.Kind, Color: body.Color,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, cat)
}

func (h *FinanceHandler) UpdateCategory(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	categoryID, err := parseUintParam(c, "categoryId")
	if err != nil {
		handleError(c, err)
		return
	}
	var body updateCategoryBody
	if err := c.ShouldBindJSON(&body); err != nil {
		handleValidationError(c, err)
		return
	}
	cat, err := h.financeService.UpdateCategory(userID, groupID, categoryID, services.UpdateFinanceCategoryRequest{
		Name: body.Name, Color: body.Color,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, cat)
}

func (h *FinanceHandler) DeleteCategory(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	categoryID, err := parseUintParam(c, "categoryId")
	if err != nil {
		handleError(c, err)
		return
	}
	if err := h.financeService.DeleteCategory(userID, groupID, categoryID); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type createTransactionBody struct {
	Type              models.FinanceTransactionType       `json:"type" binding:"required"`
	AccountID         uint                                `json:"account_id" binding:"required"`
	TransferAccountID *uint                               `json:"transfer_account_id"`
	CategoryID        *uint                               `json:"category_id"`
	AmountCents       int64                               `json:"amount_cents" binding:"required"`
	Description       string                              `json:"description"`
	Date              string                              `json:"date" binding:"required"`
	Visibility        models.FinanceTransactionVisibility `json:"visibility"`
}

type updateTransactionBody struct {
	AccountID         *uint                                `json:"account_id"`
	TransferAccountID *uint                                `json:"transfer_account_id"`
	CategoryID        *uint                                `json:"category_id"`
	AmountCents       *int64                               `json:"amount_cents"`
	Description       *string                              `json:"description"`
	Date              *string                              `json:"date"`
	Visibility        *models.FinanceTransactionVisibility `json:"visibility"`
}

func (h *FinanceHandler) ListTransactions(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	filter := services.ListFinanceTransactionsFilter{}
	if from := c.Query("from"); from != "" {
		t, err := parseDate(from)
		if err != nil {
			handleValidationError(c, err)
			return
		}
		filter.From = &t
	}
	if to := c.Query("to"); to != "" {
		t, err := parseDate(to)
		if err != nil {
			handleValidationError(c, err)
			return
		}
		filter.To = &t
	}
	if aid := c.Query("account_id"); aid != "" {
		id, err := parseUintQuery(aid)
		if err != nil {
			handleValidationError(c, err)
			return
		}
		filter.AccountID = &id
	}
	if cid := c.Query("category_id"); cid != "" {
		id, err := parseUintQuery(cid)
		if err != nil {
			handleValidationError(c, err)
			return
		}
		filter.CategoryID = &id
	}
	if typ := c.Query("type"); typ != "" {
		t := models.FinanceTransactionType(typ)
		filter.Type = &t
	}
	txs, err := h.financeService.ListTransactions(userID, groupID, filter)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, txs)
}

func parseUintQuery(s string) (uint, error) {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(val), nil
}

func (h *FinanceHandler) GetTransaction(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	transactionID, err := parseUintParam(c, "transactionId")
	if err != nil {
		handleError(c, err)
		return
	}
	tx, err := h.financeService.GetTransaction(userID, groupID, transactionID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, tx)
}

func (h *FinanceHandler) CreateTransaction(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	var body createTransactionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		handleValidationError(c, err)
		return
	}
	date, err := parseDate(body.Date)
	if err != nil {
		handleValidationError(c, err)
		return
	}
	tx, err := h.financeService.CreateTransaction(userID, groupID, services.CreateFinanceTransactionRequest{
		Type: body.Type, AccountID: body.AccountID, TransferAccountID: body.TransferAccountID,
		CategoryID: body.CategoryID, AmountCents: body.AmountCents, Description: body.Description,
		Date: date, Visibility: body.Visibility,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, tx)
}

func (h *FinanceHandler) UpdateTransaction(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	transactionID, err := parseUintParam(c, "transactionId")
	if err != nil {
		handleError(c, err)
		return
	}
	var body updateTransactionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		handleValidationError(c, err)
		return
	}
	req := services.UpdateFinanceTransactionRequest{
		AccountID: body.AccountID, TransferAccountID: body.TransferAccountID,
		CategoryID: body.CategoryID, AmountCents: body.AmountCents,
		Description: body.Description, Visibility: body.Visibility,
	}
	if body.Date != nil {
		d, err := parseDate(*body.Date)
		if err != nil {
			handleValidationError(c, err)
			return
		}
		req.Date = &d
	}
	tx, err := h.financeService.UpdateTransaction(userID, groupID, transactionID, req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, tx)
}

func (h *FinanceHandler) DeleteTransaction(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	transactionID, err := parseUintParam(c, "transactionId")
	if err != nil {
		handleError(c, err)
		return
	}
	if err := h.financeService.DeleteTransaction(userID, groupID, transactionID); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *FinanceHandler) GetDashboard(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	dash, err := h.financeService.GetDashboard(userID, groupID, c.Query("month"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dash)
}

type updateMemberRoleBody struct {
	Role models.FinanceMemberRoleName `json:"role" binding:"required"`
}

func (h *FinanceHandler) ListMemberRoles(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	roles, err := h.financeService.ListMemberRoles(userID, groupID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *FinanceHandler) UpdateMemberRole(c *gin.Context) {
	userID, groupID, ok := h.groupIDParam(c)
	if !ok {
		return
	}
	targetUserID, err := parseUintParam(c, "userId")
	if err != nil {
		handleError(c, err)
		return
	}
	var body updateMemberRoleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		handleValidationError(c, err)
		return
	}
	role, err := h.financeService.UpdateMemberRole(userID, groupID, targetUserID, body.Role)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, role)
}
