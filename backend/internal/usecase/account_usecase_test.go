package usecase

import (
	"context"
	"fmt"
	"testing"

	"sprout-backend/internal/domain"
)

// --- Mock Repository ---

type mockAccountRepo struct {
	accounts map[string]*domain.Account
}

func newMockRepo() *mockAccountRepo {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domain.Account),
	}
	// Seed a root system account
	root := &domain.Account{
		ID:        "root-asset",
		Code:      "100.000",
		Name:      "ASET",
		Type:      domain.AccountTypeAsset,
		ParentID:  nil,
		Level:     0,
		IsSystem:  true,
		IsControl: true,
		IsActive:  true,
	}
	repo.accounts[root.ID] = root
	return repo
}

func (m *mockAccountRepo) GetAll(_ context.Context) ([]domain.Account, error) {
	var result []domain.Account
	for _, a := range m.accounts {
		if a.IsActive {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *mockAccountRepo) GetAllWithBalances(_ context.Context) ([]domain.Account, error) {
	// In the mock, balances are just whatever is set on the account (default 0)
	var result []domain.Account
	for _, a := range m.accounts {
		if a.IsActive {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *mockAccountRepo) GetByID(_ context.Context, id string) (*domain.Account, error) {
	a, ok := m.accounts[id]
	if !ok {
		return nil, nil
	}
	return a, nil
}

func (m *mockAccountRepo) GetByCode(_ context.Context, code string) (*domain.Account, error) {
	for _, a := range m.accounts {
		if a.Code == code {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockAccountRepo) GetChildren(_ context.Context, parentID string) ([]domain.Account, error) {
	var result []domain.Account
	for _, a := range m.accounts {
		if a.ParentID != nil && *a.ParentID == parentID && a.IsActive {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *mockAccountRepo) HasChildren(_ context.Context, id string) (bool, error) {
	for _, a := range m.accounts {
		if a.ParentID != nil && *a.ParentID == id && a.IsActive {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockAccountRepo) IsReferencedInJournalLines(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *mockAccountRepo) DeleteRelatedJournalEntries(_ context.Context, _ string) error {
	return nil
}

func (m *mockAccountRepo) Create(_ context.Context, account *domain.Account) (*domain.Account, error) {
	account.ID = fmt.Sprintf("generated-%s", account.Code)
	m.accounts[account.ID] = account
	return account, nil
}

func (m *mockAccountRepo) Update(_ context.Context, account *domain.Account) (*domain.Account, error) {
	m.accounts[account.ID] = account
	return account, nil
}

func (m *mockAccountRepo) Delete(_ context.Context, id string) error {
	delete(m.accounts, id)
	return nil
}

func (m *mockAccountRepo) Search(ctx context.Context, _ domain.AccountFilter) ([]domain.Account, error) {
	return m.GetAll(ctx)
}

func (m *mockAccountRepo) CreateOpeningBalance(_ context.Context, accountID string, accountType domain.AccountType, amount float64, _ string) error {
	// Simulate balance effect: set the account balance directly
	acct, ok := m.accounts[accountID]
	if !ok {
		return fmt.Errorf("account not found")
	}
	acct.Balance = amount
	return nil
}

func (m *mockAccountRepo) AdjustBalance(_ context.Context, accountID string, accountType domain.AccountType, currentBalance, newBalance float64, _ string) error {
	acct, ok := m.accounts[accountID]
	if !ok {
		return fmt.Errorf("account not found")
	}
	acct.Balance = newBalance
	return nil
}

// --- Tests ---

func TestCreateAccount_Success(t *testing.T) {
	repo := newMockRepo()
	uc := NewAccountUseCase(repo)

	req := domain.CreateAccountRequest{
		Code:     "111.000",
		Name:     "Kas",
		ParentID: "root-asset",
	}

	account, err := uc.Create(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if account.Code != "111.000" {
		t.Errorf("expected code 111.000, got %s", account.Code)
	}
	if account.Type != domain.AccountTypeAsset {
		t.Errorf("expected type asset (inherited from parent), got %s", account.Type)
	}
	if account.Level != 1 {
		t.Errorf("expected level 1 (parent level + 1), got %d", account.Level)
	}
	if account.IsSystem {
		t.Errorf("expected is_system=false for user-created account")
	}
}

func TestCreateAccount_DuplicateCode(t *testing.T) {
	repo := newMockRepo()
	uc := NewAccountUseCase(repo)

	req := domain.CreateAccountRequest{
		Code:     "111.000",
		Name:     "Kas",
		ParentID: "root-asset",
	}

	_, err := uc.Create(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err = uc.Create(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for duplicate code, got nil")
	}
}

func TestCreateAccount_ParentNotFound(t *testing.T) {
	repo := newMockRepo()
	uc := NewAccountUseCase(repo)

	req := domain.CreateAccountRequest{
		Code:     "111.000",
		Name:     "Kas",
		ParentID: "non-existent",
	}

	_, err := uc.Create(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for non-existent parent, got nil")
	}
}

func TestDeleteAccount_SystemAccount(t *testing.T) {
	repo := newMockRepo()
	uc := NewAccountUseCase(repo)

	err := uc.Delete(context.Background(), "root-asset")
	if err == nil {
		t.Fatal("expected error deleting system account, got nil")
	}
}

func TestDeleteAccount_WithChildren(t *testing.T) {
	repo := newMockRepo()
	uc := NewAccountUseCase(repo)

	// Create a child first
	req := domain.CreateAccountRequest{
		Code:     "111.000",
		Name:     "Kas",
		ParentID: "root-asset",
	}
	child, err := uc.Create(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("create child failed: %v", err)
	}

	// Create a grandchild
	req2 := domain.CreateAccountRequest{
		Code:     "111.001",
		Name:     "Kas Kecil",
		ParentID: child.ID,
	}
	_, err = uc.Create(context.Background(), req2, "user-1")
	if err != nil {
		t.Fatalf("create grandchild failed: %v", err)
	}

	// Try to delete the child (has children)
	err = uc.Delete(context.Background(), child.ID)
	if err == nil {
		t.Fatal("expected error deleting account with children, got nil")
	}
}

func TestUpdateAccount_SystemAccount(t *testing.T) {
	repo := newMockRepo()
	uc := NewAccountUseCase(repo)

	name := "Changed Name"
	_, err := uc.Update(context.Background(), "root-asset", domain.UpdateAccountRequest{
		Name: &name,
	}, "user-1")
	if err == nil {
		t.Fatal("expected error editing system account, got nil")
	}
}

func TestGetTree(t *testing.T) {
	repo := newMockRepo()
	uc := NewAccountUseCase(repo)

	// Add a child
	req := domain.CreateAccountRequest{
		Code:     "111.000",
		Name:     "Kas",
		ParentID: "root-asset",
	}
	_, err := uc.Create(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("create child failed: %v", err)
	}

	tree, err := uc.GetTree(context.Background())
	if err != nil {
		t.Fatalf("get tree failed: %v", err)
	}
	if len(tree) == 0 {
		t.Fatal("expected at least one root node")
	}

	// Find the ASET root
	var assetRoot *domain.AccountTreeNode
	for _, node := range tree {
		if node.Account.Code == "100.000" {
			assetRoot = node
			break
		}
	}
	if assetRoot == nil {
		t.Fatal("expected ASET root in tree")
	}
	if len(assetRoot.Children) != 1 {
		t.Errorf("expected 1 child under ASET, got %d", len(assetRoot.Children))
	}
}

func TestGetTree_BalanceRollUp(t *testing.T) {
	repo := newMockRepo()
	uc := NewAccountUseCase(repo)

	// Create two children under ASET with preset balances
	parentID := "root-asset"
	child1 := &domain.Account{
		ID:       "child-kas",
		Code:     "111.000",
		Name:     "Kas",
		Type:     domain.AccountTypeAsset,
		ParentID: &parentID,
		Level:    1,
		IsActive: true,
		Balance:  5000000, // e.g. 5 million in cash
	}
	child2 := &domain.Account{
		ID:       "child-piutang",
		Code:     "112.000",
		Name:     "Piutang Usaha",
		Type:     domain.AccountTypeAsset,
		ParentID: &parentID,
		Level:    1,
		IsActive: true,
		Balance:  3000000, // e.g. 3 million receivable
	}
	repo.accounts[child1.ID] = child1
	repo.accounts[child2.ID] = child2

	tree, err := uc.GetTree(context.Background())
	if err != nil {
		t.Fatalf("get tree failed: %v", err)
	}

	// Find the ASET root
	var assetRoot *domain.AccountTreeNode
	for _, node := range tree {
		if node.Account.Code == "100.000" {
			assetRoot = node
			break
		}
	}
	if assetRoot == nil {
		t.Fatal("expected ASET root in tree")
	}

	// ASET root balance should be sum of children: 5M + 3M = 8M
	expectedBalance := 8000000.0
	if assetRoot.Account.Balance != expectedBalance {
		t.Errorf("expected ASET balance %.2f, got %.2f", expectedBalance, assetRoot.Account.Balance)
	}

	// Children should retain their own balances
	for _, child := range assetRoot.Children {
		if child.Account.Code == "111.000" && child.Account.Balance != 5000000 {
			t.Errorf("expected Kas balance 5000000, got %.2f", child.Account.Balance)
		}
		if child.Account.Code == "112.000" && child.Account.Balance != 3000000 {
			t.Errorf("expected Piutang balance 3000000, got %.2f", child.Account.Balance)
		}
	}
}

func TestCreateAccount_WithStartingBalance(t *testing.T) {
	repo := newMockRepo()
	uc := NewAccountUseCase(repo)

	req := domain.CreateAccountRequest{
		Code:            "111.000",
		Name:            "Kas",
		ParentID:        "root-asset",
		StartingBalance: 10000000, // 10 million
	}

	account, err := uc.Create(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify the account was created
	if account.Code != "111.000" {
		t.Errorf("expected code 111.000, got %s", account.Code)
	}

	// Verify opening balance was applied (mock sets Balance directly)
	stored := repo.accounts[account.ID]
	if stored.Balance != 10000000 {
		t.Errorf("expected balance 10000000 after opening balance, got %.2f", stored.Balance)
	}
}

func TestCreateAccount_ZeroStartingBalanceNoJournal(t *testing.T) {
	repo := newMockRepo()
	uc := NewAccountUseCase(repo)

	req := domain.CreateAccountRequest{
		Code:            "111.000",
		Name:            "Kas",
		ParentID:        "root-asset",
		StartingBalance: 0, // no starting balance
	}

	account, err := uc.Create(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Balance should remain 0 (no opening balance journal created)
	stored := repo.accounts[account.ID]
	if stored.Balance != 0 {
		t.Errorf("expected balance 0 when no starting balance requested, got %.2f", stored.Balance)
	}
}

func TestUpdateAccount_Balance(t *testing.T) {
	repo := newMockRepo()
	uc := NewAccountUseCase(repo)

	// Create a child account first
	req := domain.CreateAccountRequest{
		Code:            "111.000",
		Name:            "Kas",
		ParentID:        "root-asset",
		StartingBalance: 5000000,
	}
	account, err := uc.Create(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Update the balance to 8 million
	newBalance := 8000000.0
	updated, err := uc.Update(context.Background(), account.ID, domain.UpdateAccountRequest{
		Balance: &newBalance,
	}, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// The returned account should reflect the new balance
	if updated.Balance != 8000000 {
		t.Errorf("expected returned balance 8000000, got %.2f", updated.Balance)
	}

	// The stored account should also be updated
	stored := repo.accounts[account.ID]
	if stored.Balance != 8000000 {
		t.Errorf("expected stored balance 8000000, got %.2f", stored.Balance)
	}
}

func TestUpdateAccount_BalanceAndName(t *testing.T) {
	repo := newMockRepo()
	uc := NewAccountUseCase(repo)

	// Create a child account
	req := domain.CreateAccountRequest{
		Code:     "111.000",
		Name:     "Kas",
		ParentID: "root-asset",
	}
	account, err := uc.Create(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Update both name and balance at the same time
	newName := "Kas Bank"
	newBalance := 3000000.0
	updated, err := uc.Update(context.Background(), account.ID, domain.UpdateAccountRequest{
		Name:    &newName,
		Balance: &newBalance,
	}, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if updated.Name != "Kas Bank" {
		t.Errorf("expected name 'Kas Bank', got '%s'", updated.Name)
	}
	if updated.Balance != 3000000 {
		t.Errorf("expected balance 3000000, got %.2f", updated.Balance)
	}
}

func TestUpdateAccount_BalanceDecrease(t *testing.T) {
	repo := newMockRepo()
	uc := NewAccountUseCase(repo)

	// Create a child account with starting balance
	req := domain.CreateAccountRequest{
		Code:            "111.000",
		Name:            "Kas",
		ParentID:        "root-asset",
		StartingBalance: 10000000,
	}
	account, err := uc.Create(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Decrease the balance
	newBalance := 2000000.0
	updated, err := uc.Update(context.Background(), account.ID, domain.UpdateAccountRequest{
		Balance: &newBalance,
	}, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if updated.Balance != 2000000 {
		t.Errorf("expected balance 2000000, got %.2f", updated.Balance)
	}
}
