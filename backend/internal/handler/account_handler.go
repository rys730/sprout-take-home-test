package handler

import (
	"net/http"

	"sprout-backend/internal/domain"
	"sprout-backend/internal/infrastructure/logger"

	"github.com/labstack/echo/v4"
)

type AccountHandler struct {
	useCase domain.AccountUseCase
}

func NewAccountHandler(useCase domain.AccountUseCase) *AccountHandler {
	return &AccountHandler{useCase: useCase}
}

// ListAccounts godoc
// @Summary     List accounts
// @Description Retrieve a list of accounts with optional filtering
// @Tags        accounts
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       search    query    string false "Search by code or name"
// @Param       type      query    string false "Filter by account type (asset, liability, equity, revenue, expense)"
// @Param       parent_id query    string false "Filter by parent account ID"
// @Success     200       {object} map[string]interface{} "data: []Account, count: int"
// @Failure     401       {object} map[string]string       "Unauthorized"
// @Failure     500       {object} map[string]string       "Failed to retrieve accounts"
// @Router      /accounts [get]
func (h *AccountHandler) ListAccounts(c echo.Context) error {
	filter := domain.AccountFilter{
		Search:   c.QueryParam("search"),
		Type:     c.QueryParam("type"),
		ParentID: c.QueryParam("parent_id"),
	}

	accounts, err := h.useCase.List(c.Request().Context(), filter)
	if err != nil {
		logger.Errorf("Failed to list accounts: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve accounts",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  accounts,
		"count": len(accounts),
	})
}

// GetAccountTree godoc
// @Summary     Get account tree
// @Description Returns the full chart of accounts as a hierarchical tree with rolled-up balances
// @Tags        accounts
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} map[string]interface{} "data: []AccountTreeNode"
// @Failure     401 {object} map[string]string       "Unauthorized"
// @Failure     500 {object} map[string]string       "Failed to retrieve account tree"
// @Router      /accounts/tree [get]
func (h *AccountHandler) GetAccountTree(c echo.Context) error {
	tree, err := h.useCase.GetTree(c.Request().Context())
	if err != nil {
		logger.Errorf("Failed to get account tree: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve account tree",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": tree,
	})
}

// GetAccount godoc
// @Summary     Get account by ID
// @Description Retrieve a single account by its UUID
// @Tags        accounts
// @Produce     json
// @Security    BearerAuth
// @Param       id  path     string true "Account UUID"
// @Success     200 {object} map[string]interface{} "data: Account"
// @Failure     401 {object} map[string]string       "Unauthorized"
// @Failure     404 {object} map[string]string       "Account not found"
// @Router      /accounts/{id} [get]
func (h *AccountHandler) GetAccount(c echo.Context) error {
	id := c.Param("id")

	account, err := h.useCase.GetByID(c.Request().Context(), id)
	if err != nil {
		logger.Errorf("Failed to get account %s: %v", id, err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": account,
	})
}

// CreateAccount godoc
// @Summary     Create a new account
// @Description Create a new account in the chart of accounts. Optionally set a starting balance.
// @Tags        accounts
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body     domain.CreateAccountRequest true "Account data"
// @Success     201  {object} map[string]interface{}       "data: Account, message: string"
// @Failure     400  {object} map[string]string             "Validation error or duplicate code"
// @Failure     401  {object} map[string]string             "Unauthorized"
// @Router      /accounts [post]
func (h *AccountHandler) CreateAccount(c echo.Context) error {
	var req domain.CreateAccountRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.Code == "" || req.Name == "" || req.ParentID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "code, name, and parent_id are required",
		})
	}

	// Extract user ID from JWT claims
	createdBy := getUserIDFromContext(c)

	account, err := h.useCase.Create(c.Request().Context(), req, createdBy)
	if err != nil {
		logger.Errorf("Failed to create account: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data":    account,
		"message": "Account created successfully",
	})
}

// UpdateAccount godoc
// @Summary     Update an account
// @Description Update an existing account's code, name, or balance. At least one field must be provided.
// @Tags        accounts
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path     string                     true "Account UUID"
// @Param       body body     domain.UpdateAccountRequest true "Fields to update"
// @Success     200  {object} map[string]interface{}       "data: Account, message: string"
// @Failure     400  {object} map[string]string             "Validation error"
// @Failure     401  {object} map[string]string             "Unauthorized"
// @Router      /accounts/{id} [put]
func (h *AccountHandler) UpdateAccount(c echo.Context) error {
	id := c.Param("id")

	var req domain.UpdateAccountRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.Code == nil && req.Name == nil && req.Balance == nil && req.ParentID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "At least one of code, name, parent_id, or balance must be provided",
		})
	}

	updatedBy := getUserIDFromContext(c)

	account, err := h.useCase.Update(c.Request().Context(), id, req, updatedBy)
	if err != nil {
		logger.Errorf("Failed to update account %s: %v", id, err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":    account,
		"message": "Account updated successfully",
	})
}

// DeleteAccount godoc
// @Summary     Delete an account
// @Description Delete an account by ID. Cannot delete system accounts, control accounts, or accounts with children/journal references.
// @Tags        accounts
// @Produce     json
// @Security    BearerAuth
// @Param       id  path     string true "Account UUID"
// @Success     200 {object} map[string]string "Account deleted successfully"
// @Failure     400 {object} map[string]string "Cannot delete (system/children/referenced)"
// @Failure     401 {object} map[string]string "Unauthorized"
// @Router      /accounts/{id} [delete]
func (h *AccountHandler) DeleteAccount(c echo.Context) error {
	id := c.Param("id")

	if err := h.useCase.Delete(c.Request().Context(), id); err != nil {
		logger.Errorf("Failed to delete account %s: %v", id, err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Account deleted successfully",
	})
}

// getUserIDFromContext extracts the user ID from JWT claims in the echo context.
func getUserIDFromContext(c echo.Context) string {
	user := c.Get("user")
	if user == nil {
		return ""
	}

	// The JWT middleware stores the token in context.
	// Extract the user_id claim from it.
	token, ok := user.(interface{ Claims() interface{} })
	if !ok {
		return ""
	}

	_ = token
	// Fallback: for now return empty — will be properly wired when auth is complete.
	return ""
}
