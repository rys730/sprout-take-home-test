package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"

	"sprout-backend/db/queries"
	"sprout-backend/internal/domain"

	"github.com/jackc/pgx/v5"
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

// ---------------------------------------------------------------------------
// pgtype ↔ domain conversion helpers
// ---------------------------------------------------------------------------

func parseUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	_ = u.Scan(s)
	return u
}

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}

func uuidToStringPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuidToString(u)
	return &s
}

func stringPtrToUUID(s *string) pgtype.UUID {
	if s == nil {
		return pgtype.UUID{Valid: false}
	}
	return parseUUID(*s)
}

func timestamptzToTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	if !f.Valid {
		return 0
	}
	return f.Float64
}

func float64ToNumeric(f float64) pgtype.Numeric {
	// Convert to an integer-based representation to avoid floating-point issues.
	// We store 2 decimal places, so multiply by 100.
	scaled := int64(math.Round(f * 100))
	return pgtype.Numeric{
		Int:   big.NewInt(scaled),
		Exp:   -2,
		Valid: true,
	}
}

func textFromString(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func currentDate() pgtype.Date {
	now := time.Now()
	return pgtype.Date{Time: now, Valid: true}
}

// accountToDomain converts a sqlc Account model to domain.Account.
func accountToDomain(a queries.Account) domain.Account {
	return domain.Account{
		ID:        uuidToString(a.ID),
		Code:      a.Code,
		Name:      a.Name,
		Type:      domain.AccountType(a.Type),
		ParentID:  uuidToStringPtr(a.ParentID),
		Level:     int(a.Level),
		IsSystem:  a.IsSystem,
		IsControl: a.IsControl,
		IsActive:  a.IsActive,
		CreatedBy: uuidToStringPtr(a.CreatedBy),
		CreatedAt: timestamptzToTime(a.CreatedAt),
		UpdatedAt: timestamptzToTime(a.UpdatedAt),
	}
}

// accountWithBalanceToDomain converts a sqlc GetAllAccountsWithBalancesRow to domain.Account.
func accountWithBalanceToDomain(a queries.GetAllAccountsWithBalancesRow) domain.Account {
	return domain.Account{
		ID:        uuidToString(a.ID),
		Code:      a.Code,
		Name:      a.Name,
		Type:      domain.AccountType(a.Type),
		ParentID:  uuidToStringPtr(a.ParentID),
		Level:     int(a.Level),
		IsSystem:  a.IsSystem,
		IsControl: a.IsControl,
		IsActive:  a.IsActive,
		Balance:   numericToFloat64(a.Balance),
		CreatedBy: uuidToStringPtr(a.CreatedBy),
		CreatedAt: timestamptzToTime(a.CreatedAt),
		UpdatedAt: timestamptzToTime(a.UpdatedAt),
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
	row, err := r.q.GetAccountByID(ctx, parseUUID(id))
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
	rows, err := r.q.GetAccountChildren(ctx, parseUUID(parentID))
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
	has, err := r.q.HasAccountChildren(ctx, parseUUID(id))
	if err != nil {
		return false, fmt.Errorf("check children: %w", err)
	}
	return has, nil
}

func (r *AccountRepository) IsReferencedInJournalLines(ctx context.Context, id string) (bool, error) {
	ref, err := r.q.IsAccountReferencedInJournalLines(ctx, parseUUID(id))
	if err != nil {
		return false, fmt.Errorf("check journal references: %w", err)
	}
	return ref, nil
}

func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) (*domain.Account, error) {
	row, err := r.q.CreateAccount(ctx, queries.CreateAccountParams{
		Code:      account.Code,
		Name:      account.Name,
		Type:      queries.AccountType(account.Type),
		ParentID:  stringPtrToUUID(account.ParentID),
		Level:     int32(account.Level),
		IsSystem:  account.IsSystem,
		IsControl: account.IsControl,
		IsActive:  account.IsActive,
		CreatedBy: stringPtrToUUID(account.CreatedBy),
	})
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	a := accountToDomain(row)
	return &a, nil
}

func (r *AccountRepository) Update(ctx context.Context, account *domain.Account) (*domain.Account, error) {
	row, err := r.q.UpdateAccount(ctx, queries.UpdateAccountParams{
		ID:   parseUUID(account.ID),
		Code: account.Code,
		Name: account.Name,
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
	err := r.q.DeleteAccount(ctx, parseUUID(id))
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

func (r *AccountRepository) Search(ctx context.Context, filter domain.AccountFilter) ([]domain.Account, error) {
	rows, err := r.q.SearchAccounts(ctx, queries.SearchAccountsParams{
		Search: textFromString(filter.Search),
		AccountType: queries.NullAccountType{
			AccountType: queries.AccountType(filter.Type),
			Valid:       filter.Type != "",
		},
		ParentID: stringPtrToUUID(func() *string {
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
		Date:        currentDate(),
		Description: "Opening Balance",
		Source:      pgtype.Text{String: "opening_balance", Valid: true},
		Status:      queries.JournalStatusPosted,
		CreatedBy:   parseUUID(createdBy),
	})
	if err != nil {
		return fmt.Errorf("insert journal entry: %w", err)
	}

	// 3. Determine debit/credit accounts based on account type
	acctUUID := parseUUID(accountID)
	amountNum := float64ToNumeric(amount)
	zeroNum := float64ToNumeric(0)

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
		Description:    pgtype.Text{String: "Opening Balance", Valid: true},
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
		Description:    pgtype.Text{String: "Opening Balance", Valid: true},
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
		Date:        currentDate(),
		Description: "Balance Adjustment",
		Source:      pgtype.Text{String: "adjustment", Valid: true},
		Status:      queries.JournalStatusPosted,
		CreatedBy:   parseUUID(createdBy),
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

	acctUUID := parseUUID(accountID)

	// Account line
	_, err = qtx.CreateJournalEntryLine(ctx, queries.CreateJournalEntryLineParams{
		JournalEntryID: entry.ID,
		AccountID:      acctUUID,
		Description:    pgtype.Text{String: "Balance Adjustment", Valid: true},
		Debit:          float64ToNumeric(acctDebit),
		Credit:         float64ToNumeric(acctCredit),
		LineOrder:      1,
	})
	if err != nil {
		return fmt.Errorf("insert account adjustment line: %w", err)
	}

	// Contra (313.000) line
	_, err = qtx.CreateJournalEntryLine(ctx, queries.CreateJournalEntryLineParams{
		JournalEntryID: entry.ID,
		AccountID:      contraID,
		Description:    pgtype.Text{String: "Balance Adjustment", Valid: true},
		Debit:          float64ToNumeric(contraDebit),
		Credit:         float64ToNumeric(contraCredit),
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
