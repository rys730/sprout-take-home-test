package domain

import (
	"context"
	"time"
)

// AccountType represents the category of a financial account.
type AccountType string

const (
	AccountTypeAsset     AccountType = "asset"
	AccountTypeLiability AccountType = "liability"
	AccountTypeEquity    AccountType = "equity"
	AccountTypeRevenue   AccountType = "revenue"
	AccountTypeExpense   AccountType = "expense"
)

// Account represents a single entry in the Chart of Accounts.
type Account struct {
	ID        string      `json:"id"`
	Code      string      `json:"code"`
	Name      string      `json:"name"`
	Type      AccountType `json:"type"`
	ParentID  *string     `json:"parent_id"`
	Level     int         `json:"level"`
	IsSystem  bool        `json:"is_system"`
	IsControl bool        `json:"is_control"`
	IsActive  bool        `json:"is_active"`
	Balance   float64     `json:"balance"` // computed: net balance from posted journal entries
	CreatedBy *string     `json:"created_by,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// AccountTreeNode wraps an Account with its children for tree display.
type AccountTreeNode struct {
	Account  Account            `json:"account"`
	Children []*AccountTreeNode `json:"children,omitempty"`
}

// CreateAccountRequest holds the data needed to create a new account.
type CreateAccountRequest struct {
	Code            string  `json:"code" validate:"required"`
	Name            string  `json:"name" validate:"required"`
	ParentID        string  `json:"parent_id" validate:"required"`
	StartingBalance float64 `json:"starting_balance"` // optional: creates an opening balance journal entry
}

// UpdateAccountRequest holds the data allowed to update on an account.
type UpdateAccountRequest struct {
	Code     *string  `json:"code,omitempty"`
	Name     *string  `json:"name,omitempty"`
	ParentID *string  `json:"parent_id,omitempty"`
	Balance  *float64 `json:"balance,omitempty"` // optional: adjusts balance via journal entry
}

// AccountFilter holds optional filter/search parameters.
type AccountFilter struct {
	Search   string // search by code or name
	Type     string // filter by account type
	ParentID string // filter by parent
	IsActive *bool  // filter by active status
}

// AccountRepository defines the persistence interface for accounts.
type AccountRepository interface {
	GetAll(ctx context.Context) ([]Account, error)
	GetAllWithBalances(ctx context.Context) ([]Account, error)
	GetByID(ctx context.Context, id string) (*Account, error)
	GetByCode(ctx context.Context, code string) (*Account, error)
	GetChildren(ctx context.Context, parentID string) ([]Account, error)
	HasChildren(ctx context.Context, id string) (bool, error)
	IsReferencedInJournalLines(ctx context.Context, id string) (bool, error)
	DeleteRelatedJournalEntries(ctx context.Context, accountID string) error
	Create(ctx context.Context, account *Account) (*Account, error)
	Update(ctx context.Context, account *Account) (*Account, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, filter AccountFilter) ([]Account, error)
	CreateOpeningBalance(ctx context.Context, accountID string, accountType AccountType, amount float64, createdBy string) error
	AdjustBalance(ctx context.Context, accountID string, accountType AccountType, currentBalance, newBalance float64, createdBy string) error
}

// AccountUseCase defines the business logic interface for accounts.
type AccountUseCase interface {
	List(ctx context.Context, filter AccountFilter) ([]Account, error)
	GetTree(ctx context.Context) ([]*AccountTreeNode, error)
	GetByID(ctx context.Context, id string) (*Account, error)
	Create(ctx context.Context, req CreateAccountRequest, createdBy string) (*Account, error)
	Update(ctx context.Context, id string, req UpdateAccountRequest, updatedBy string) (*Account, error)
	Delete(ctx context.Context, id string) error
}
