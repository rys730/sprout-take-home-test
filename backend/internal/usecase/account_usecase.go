package usecase

import (
	"context"
	"fmt"

	"sprout-backend/internal/domain"
)

type AccountUseCase struct {
	repo domain.AccountRepository
}

func NewAccountUseCase(repo domain.AccountRepository) *AccountUseCase {
	return &AccountUseCase{repo: repo}
}

// List returns accounts filtered/searched by the given criteria, with balances.
func (uc *AccountUseCase) List(ctx context.Context, filter domain.AccountFilter) ([]domain.Account, error) {
	if filter.Search != "" || filter.Type != "" || filter.ParentID != "" {
		return uc.repo.Search(ctx, filter)
	}
	return uc.repo.GetAllWithBalances(ctx)
}

// GetTree returns the full chart of accounts as a hierarchical tree with balances.
// Leaf account balances come from posted journal entries.
// Parent account balances are the sum of their children's balances.
func (uc *AccountUseCase) GetTree(ctx context.Context) ([]*domain.AccountTreeNode, error) {
	accounts, err := uc.repo.GetAllWithBalances(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all accounts with balances: %w", err)
	}
	tree := buildTree(accounts)
	// Roll up balances from children to parents
	for _, root := range tree {
		rollUpBalances(root)
	}
	return tree, nil
}

// GetByID retrieves a single account by its ID.
func (uc *AccountUseCase) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	account, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account not found")
	}
	return account, nil
}

// Create validates business rules and creates a new child account.
func (uc *AccountUseCase) Create(ctx context.Context, req domain.CreateAccountRequest, createdBy string) (*domain.Account, error) {
	// 1. Validate parent exists
	parent, err := uc.repo.GetByID(ctx, req.ParentID)
	if err != nil {
		return nil, fmt.Errorf("get parent account: %w", err)
	}
	if parent == nil {
		return nil, fmt.Errorf("parent account not found")
	}

	// 2. Check code uniqueness
	existing, err := uc.repo.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("check code uniqueness: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("account code '%s' already exists", req.Code)
	}

	// 3. Inherit type from parent, calculate level
	account := &domain.Account{
		Code:      req.Code,
		Name:      req.Name,
		Type:      parent.Type,
		ParentID:  &req.ParentID,
		Level:     parent.Level + 1,
		IsSystem:  false,
		IsControl: false,
		IsActive:  true,
		CreatedBy: &createdBy,
	}

	created, err := uc.repo.Create(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	// 4. If a starting balance is provided, create an opening-balance journal entry
	if req.StartingBalance > 0 {
		if err := uc.repo.CreateOpeningBalance(ctx, created.ID, created.Type, req.StartingBalance, createdBy); err != nil {
			return nil, fmt.Errorf("create opening balance: %w", err)
		}
	}

	return created, nil
}

// Update modifies an account's code, name, and/or balance.
// System/control accounts cannot be updated.
// When balance is provided, an adjustment journal entry is created.
func (uc *AccountUseCase) Update(ctx context.Context, id string, req domain.UpdateAccountRequest, updatedBy string) (*domain.Account, error) {
	// 1. Get existing account
	account, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account not found")
	}

	// 2. Prevent editing system/control accounts
	if account.IsSystem || account.IsControl {
		return nil, fmt.Errorf("cannot edit system or control accounts")
	}

	// 3. Apply code/name updates
	if req.Name != nil {
		account.Name = *req.Name
	}
	if req.Code != nil {
		// Check code uniqueness (only if changing)
		if *req.Code != account.Code {
			existing, err := uc.repo.GetByCode(ctx, *req.Code)
			if err != nil {
				return nil, fmt.Errorf("check code uniqueness: %w", err)
			}
			if existing != nil {
				return nil, fmt.Errorf("account code '%s' already exists", *req.Code)
			}
			account.Code = *req.Code
		}
	}

	updated, err := uc.repo.Update(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("update account: %w", err)
	}

	// 4. If a new balance is provided, create an adjustment journal entry
	if req.Balance != nil {
		// Get the current computed balance for this account
		allAccounts, err := uc.repo.GetAllWithBalances(ctx)
		if err != nil {
			return nil, fmt.Errorf("get current balance: %w", err)
		}
		var currentBalance float64
		for _, a := range allAccounts {
			if a.ID == id {
				currentBalance = a.Balance
				break
			}
		}

		if *req.Balance != currentBalance {
			if err := uc.repo.AdjustBalance(ctx, id, account.Type, currentBalance, *req.Balance, updatedBy); err != nil {
				return nil, fmt.Errorf("adjust balance: %w", err)
			}
		}
		updated.Balance = *req.Balance
	}

	return updated, nil
}

// Delete removes an account after validating business rules.
// Cannot delete system/control accounts, accounts with children, or accounts referenced in journals.
func (uc *AccountUseCase) Delete(ctx context.Context, id string) error {
	// 1. Get existing account
	account, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("account not found")
	}

	// 2. Prevent deleting system/control accounts
	if account.IsSystem || account.IsControl {
		return fmt.Errorf("cannot delete system or control accounts")
	}

	// 3. Prevent deleting accounts with children
	hasChildren, err := uc.repo.HasChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("check children: %w", err)
	}
	if hasChildren {
		return fmt.Errorf("cannot delete account with child accounts")
	}

	// 4. Prevent deleting accounts referenced in journal entries
	isReferenced, err := uc.repo.IsReferencedInJournalLines(ctx, id)
	if err != nil {
		return fmt.Errorf("check journal references: %w", err)
	}
	if isReferenced {
		return fmt.Errorf("cannot delete account referenced in journal entries")
	}

	return uc.repo.Delete(ctx, id)
}

// buildTree constructs a tree from a flat list of accounts.
func buildTree(accounts []domain.Account) []*domain.AccountTreeNode {
	nodeMap := make(map[string]*domain.AccountTreeNode)
	var roots []*domain.AccountTreeNode

	// Create nodes for all accounts
	for i := range accounts {
		nodeMap[accounts[i].ID] = &domain.AccountTreeNode{
			Account:  accounts[i],
			Children: []*domain.AccountTreeNode{},
		}
	}

	// Build parent-child relationships
	for i := range accounts {
		node := nodeMap[accounts[i].ID]
		if accounts[i].ParentID == nil {
			roots = append(roots, node)
		} else {
			parentNode, ok := nodeMap[*accounts[i].ParentID]
			if ok {
				parentNode.Children = append(parentNode.Children, node)
			} else {
				// Orphan node — treat as root
				roots = append(roots, node)
			}
		}
	}

	return roots
}

// rollUpBalances recursively sums children's balances into the parent.
// A parent's balance = its own direct balance + sum of all children's balances.
func rollUpBalances(node *domain.AccountTreeNode) float64 {
	childSum := 0.0
	for _, child := range node.Children {
		childSum += rollUpBalances(child)
	}
	node.Account.Balance += childSum
	return node.Account.Balance
}
