package repository

import (
	"context"
	"errors"
	"fmt"

	"sprout-backend/db/queries"
	"sprout-backend/internal/domain"
	"sprout-backend/internal/utils"

	"github.com/jackc/pgx/v5" // utils

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository struct {
	pool *pgxpool.Pool
	q    *queries.Queries
}

func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{
		pool: pool,
		q:    queries.New(pool),
	}
}

// accountToDomain converts a sqlc Account model to domain.Account.
func accountToDomain(a queries.Account) domain.Account {
	return domain.Account{
		ID:        utils.UUIDToString(a.ID),
		Code:      a.Code,
		Name:      a.Name,
		Type:      domain.AccountType(a.Type),
		ParentID:  utils.UUIDToStringPtr(a.ParentID),
		Level:     int(a.Level),
		IsSystem:  a.IsSystem,
		IsControl: a.IsControl,
		IsActive:  a.IsActive,
		CreatedBy: utils.UUIDToStringPtr(a.CreatedBy),
		CreatedAt: utils.TimestamptzToTime(a.CreatedAt),
		UpdatedAt: utils.TimestamptzToTime(a.UpdatedAt),
	}
}

// accountWithBalanceToDomain converts a sqlc GetAllAccountsWithBalancesRow to domain.Account.
func accountWithBalanceToDomain(a queries.GetAllAccountsWithBalancesRow) domain.Account {
	return domain.Account{
		ID:        utils.UUIDToString(a.ID),
		Code:      a.Code,
		Name:      a.Name,
		Type:      domain.AccountType(a.Type),
		ParentID:  utils.UUIDToStringPtr(a.ParentID),
		Level:     int(a.Level),
		IsSystem:  a.IsSystem,
		IsControl: a.IsControl,
		IsActive:  a.IsActive,
		Balance:   utils.NumericToFloat64(a.Balance),
		CreatedBy: utils.UUIDToStringPtr(a.CreatedBy),
		CreatedAt: utils.TimestamptzToTime(a.CreatedAt),
		UpdatedAt: utils.TimestamptzToTime(a.UpdatedAt),
	}
}

// ---------------------------------------------------------------------------
// AccountRepository methods
// ---------------------------------------------------------------------------

func (r *AccountRepository) GetAll(ctx context.Context) ([]domain.Account, error) {
	rows, err := r.q.GetAllAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("query accounts: %w", err)
	}
	accounts := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, accountToDomain(row))
	}
	return accounts, nil
}

func (r *AccountRepository) GetAllWithBalances(ctx context.Context) ([]domain.Account, error) {
	rows, err := r.q.GetAllAccountsWithBalances(ctx)
	if err != nil {
		return nil, fmt.Errorf("query accounts with balances: %w", err)
	}
	accounts := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, accountWithBalanceToDomain(row))
	}
	return accounts, nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	row, err := r.q.GetAccountByID(ctx, utils.ParseUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query account by id: %w", err)
	}
	a := accountToDomain(row)
	return &a, nil
}

func (r *AccountRepository) GetByCode(ctx context.Context, code string) (*domain.Account, error) {
	row, err := r.q.GetAccountByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query account by code: %w", err)
	}
	a := accountToDomain(row)
	return &a, nil
}

func (r *AccountRepository) GetChildren(ctx context.Context, parentID string) ([]domain.Account, error) {
	rows, err := r.q.GetAccountChildren(ctx, utils.ParseUUID(parentID))
	if err != nil {
		return nil, fmt.Errorf("query children: %w", err)
	}
	accounts := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, accountToDomain(row))
	}
	return accounts, nil
}

func (r *AccountRepository) HasChildren(ctx context.Context, id string) (bool, error) {
	has, err := r.q.HasAccountChildren(ctx, utils.ParseUUID(id))
	if err != nil {
		return false, fmt.Errorf("check children: %w", err)
	}
	return has, nil
}

func (r *AccountRepository) IsReferencedInJournalLines(ctx context.Context, id string) (bool, error) {
	ref, err := r.q.IsAccountReferencedInJournalLines(ctx, utils.ParseUUID(id))
	if err != nil {
		return false, fmt.Errorf("check journal references: %w", err)
	}
	return ref, nil
}

// DeleteRelatedJournalEntries removes all journal entry lines referencing the
// given account and then deletes any journal entries that become empty as a result.
func (r *AccountRepository) DeleteRelatedJournalEntries(ctx context.Context, accountID string) error {
	uid := utils.ParseUUID(accountID)

	// 1. Find all journal entries that reference this account
	entryIDs, err := r.q.GetJournalEntryIDsByAccountID(ctx, uid)
	if err != nil {
		return fmt.Errorf("get journal entry ids for account: %w", err)
	}

	// 2. Delete all journal entry lines for this account
	if err := r.q.DeleteJournalEntryLinesByAccountID(ctx, uid); err != nil {
		return fmt.Errorf("delete journal entry lines by account: %w", err)
	}

	// 3. Delete the now-orphaned journal entries
	//    (force delete regardless of status — these are system-generated opening/adjustment entries)
	for _, entryID := range entryIDs {
		// Delete any remaining lines for this entry first
		if err := r.q.DeleteJournalEntryLinesByEntryID(ctx, entryID); err != nil {
			return fmt.Errorf("delete remaining journal entry lines: %w", err)
		}
		if err := r.q.ForceDeleteJournalEntry(ctx, entryID); err != nil {
			return fmt.Errorf("force delete journal entry: %w", err)
		}
	}

	return nil
}

func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) (*domain.Account, error) {
	row, err := r.q.CreateAccount(ctx, queries.CreateAccountParams{
		Code:      account.Code,
		Name:      account.Name,
		Type:      queries.AccountType(account.Type),
		ParentID:  utils.StringPtrToUUID(account.ParentID),
		Level:     int32(account.Level),
		IsSystem:  account.IsSystem,
		IsControl: account.IsControl,
		IsActive:  account.IsActive,
		CreatedBy: utils.StringPtrToUUID(account.CreatedBy),
	})
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	a := accountToDomain(row)
	return &a, nil
}

func (r *AccountRepository) Update(ctx context.Context, account *domain.Account) (*domain.Account, error) {
	row, err := r.q.UpdateAccount(ctx, queries.UpdateAccountParams{
		ID:       utils.ParseUUID(account.ID),
		Code:     account.Code,
		Name:     account.Name,
		ParentID: utils.StringPtrToUUID(account.ParentID),
		Level:    int32(account.Level),
		Type:     queries.AccountType(account.Type),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("update account: %w", err)
	}
	a := accountToDomain(row)
	return &a, nil
}

func (r *AccountRepository) Delete(ctx context.Context, id string) error {
	err := r.q.DeleteAccount(ctx, utils.ParseUUID(id))
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

func (r *AccountRepository) Search(ctx context.Context, filter domain.AccountFilter) ([]domain.Account, error) {
	rows, err := r.q.SearchAccounts(ctx, queries.SearchAccountsParams{
		Search: utils.TextFromString(filter.Search),
		AccountType: queries.NullAccountType{
			AccountType: queries.AccountType(filter.Type),
			Valid:       filter.Type != "",
		},
		ParentID: utils.StringPtrToUUID(func() *string {
			if filter.ParentID == "" {
				return nil
			}
			return &filter.ParentID
		}()),
	})
	if err != nil {
		return nil, fmt.Errorf("search accounts: %w", err)
	}
	accounts := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, accountToDomain(row))
	}
	return accounts, nil
}

// CreateOpeningBalance creates a posted journal entry that records the opening
// balance for a newly-created account. It debits or credits the new account and
// uses the system "Saldo Awal (Ekuitas)" account (code 313.000) as the contra.
//
// For asset/expense accounts  → debit new account, credit 313.000
// For liability/equity/revenue → credit new account, debit 313.000
func (r *AccountRepository) CreateOpeningBalance(ctx context.Context, accountID string, accountType domain.AccountType, amount float64, createdBy string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := r.q.WithTx(tx)

	// 1. Look up the "Saldo Awal (Ekuitas)" system account by code
	contraID, err := qtx.GetAccountIDByCode(ctx, "313.000")
	if err != nil {
		return fmt.Errorf("lookup opening balance account (313.000): %w", err)
	}

	// 2. Create a posted journal entry
	entryNumberRaw, err := qtx.GenerateJournalEntryNumber(ctx)
	if err != nil {
		return fmt.Errorf("generate entry number: %w", err)
	}
	entryNumber := fmt.Sprintf("OB-%v", entryNumberRaw)

	entry, err := qtx.CreateJournalEntry(ctx, queries.CreateJournalEntryParams{
		EntryNumber: entryNumber,
		Date:        utils.CurrentDate(),
		Description: "Opening Balance",
		Source:      pgtype.Text{String: "opening_balance", Valid: true},
		Status:      queries.JournalStatusPosted,
		CreatedBy:   utils.ParseUUID(createdBy),
	})
	if err != nil {
		return fmt.Errorf("insert journal entry: %w", err)
	}

	// 3. Determine debit/credit accounts based on account type
	acctUUID := utils.ParseUUID(accountID)
	amountNum := utils.Float64ToNumeric(amount)
	zeroNum := utils.Float64ToNumeric(0)

	var debitAcctID, creditAcctID pgtype.UUID
	switch accountType {
	case domain.AccountTypeAsset, domain.AccountTypeExpense:
		debitAcctID = acctUUID
		creditAcctID = contraID
	default: // liability, equity, revenue
		debitAcctID = contraID
		creditAcctID = acctUUID
	}

	// Debit line
	_, err = qtx.CreateJournalEntryLine(ctx, queries.CreateJournalEntryLineParams{
		JournalEntryID: entry.ID,
		AccountID:      debitAcctID,
		Debit:          amountNum,
		Credit:         zeroNum,
		LineOrder:      1,
	})
	if err != nil {
		return fmt.Errorf("insert debit line: %w", err)
	}

	// Credit line
	_, err = qtx.CreateJournalEntryLine(ctx, queries.CreateJournalEntryLineParams{
		JournalEntryID: entry.ID,
		AccountID:      creditAcctID,
		Debit:          zeroNum,
		Credit:         amountNum,
		LineOrder:      2,
	})
	if err != nil {
		return fmt.Errorf("insert credit line: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// AdjustBalance creates a posted journal entry that adjusts an account's balance
// from currentBalance to newBalance. The delta is computed and recorded against
// the system "Saldo Awal (Ekuitas)" account (code 313.000) as the contra.
//
// If the delta increases the normal balance:
//   - asset/expense → debit the account, credit 313.000
//   - liability/equity/revenue → credit the account, debit 313.000
//
// If the delta decreases the normal balance, the sides are flipped.
func (r *AccountRepository) AdjustBalance(ctx context.Context, accountID string, accountType domain.AccountType, currentBalance, newBalance float64, createdBy string) error {
	delta := newBalance - currentBalance
	if delta == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := r.q.WithTx(tx)

	// 1. Look up the "Saldo Awal (Ekuitas)" contra account
	contraID, err := qtx.GetAccountIDByCode(ctx, "313.000")
	if err != nil {
		return fmt.Errorf("lookup opening balance account (313.000): %w", err)
	}

	// 2. Create a posted journal entry
	entryNumberRaw, err := qtx.GenerateJournalEntryNumber(ctx)
	if err != nil {
		return fmt.Errorf("generate entry number: %w", err)
	}
	entryNumber := fmt.Sprintf("ADJ-%v", entryNumberRaw)

	entry, err := qtx.CreateJournalEntry(ctx, queries.CreateJournalEntryParams{
		EntryNumber: entryNumber,
		Date:        utils.CurrentDate(),
		Description: "Balance Adjustment",
		Source:      pgtype.Text{String: "adjustment", Valid: true},
		Status:      queries.JournalStatusPosted,
		CreatedBy:   utils.ParseUUID(createdBy),
	})
	if err != nil {
		return fmt.Errorf("insert journal entry: %w", err)
	}

	// 3. Determine debit/credit sides
	absDelta := delta
	if absDelta < 0 {
		absDelta = -absDelta
	}

	var acctDebit, acctCredit, contraDebit, contraCredit float64
	switch accountType {
	case domain.AccountTypeAsset, domain.AccountTypeExpense:
		if delta > 0 {
			acctDebit = absDelta
			contraCredit = absDelta
		} else {
			acctCredit = absDelta
			contraDebit = absDelta
		}
	default: // liability, equity, revenue
		if delta > 0 {
			acctCredit = absDelta
			contraDebit = absDelta
		} else {
			acctDebit = absDelta
			contraCredit = absDelta
		}
	}

	acctUUID := utils.ParseUUID(accountID)

	// Account line
	_, err = qtx.CreateJournalEntryLine(ctx, queries.CreateJournalEntryLineParams{
		JournalEntryID: entry.ID,
		AccountID:      acctUUID,
		Debit:          utils.Float64ToNumeric(acctDebit),
		Credit:         utils.Float64ToNumeric(acctCredit),
		LineOrder:      1,
	})
	if err != nil {
		return fmt.Errorf("insert account adjustment line: %w", err)
	}

	// Contra (313.000) line
	_, err = qtx.CreateJournalEntryLine(ctx, queries.CreateJournalEntryLineParams{
		JournalEntryID: entry.ID,
		AccountID:      contraID,
		Debit:          utils.Float64ToNumeric(contraDebit),
		Credit:         utils.Float64ToNumeric(contraCredit),
		LineOrder:      2,
	})
	if err != nil {
		return fmt.Errorf("insert contra adjustment line: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
