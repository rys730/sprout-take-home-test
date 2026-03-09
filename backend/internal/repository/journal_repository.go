package repository

import (
	"context"
	"errors"
	"fmt"

	"sprout-backend/db/queries"
	"sprout-backend/internal/domain"
	"sprout-backend/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JournalRepository struct {
	pool *pgxpool.Pool
	q    *queries.Queries
}

func NewJournalRepository(pool *pgxpool.Pool) *JournalRepository {
	return &JournalRepository{
		pool: pool,
		q:    queries.New(pool),
	}
}

func (r *JournalRepository) List(ctx context.Context, filter domain.JournalFilter) ([]domain.JournalEntry, int64, error) {
	limit := int32(filter.Limit)
	if limit <= 0 {
		limit = 20
	}
	offset := int32(filter.Offset)
	if offset < 0 {
		offset = 0
	}

	params := queries.ListJournalEntriesParams{
		Limit:  limit,
		Offset: offset,
		Status: queries.NullJournalStatus{
			JournalStatus: queries.JournalStatus(filter.Status),
			Valid:         filter.Status != "",
		},
		Source:    utils.TextFromString(filter.Source),
		StartDate: utils.StringToDate(filter.StartDate),
		EndDate:   utils.StringToDate(filter.EndDate),
	}

	rows, err := r.q.ListJournalEntries(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("list journal entries: %w", err)
	}

	countParams := queries.CountJournalEntriesParams{
		Status:    params.Status,
		Source:    params.Source,
		StartDate: params.StartDate,
		EndDate:   params.EndDate,
	}
	total, err := r.q.CountJournalEntries(ctx, countParams)
	if err != nil {
		return nil, 0, fmt.Errorf("count journal entries: %w", err)
	}

	entries := make([]domain.JournalEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, listRowToDomain(row))
	}
	return entries, total, nil
}

func (r *JournalRepository) GetByID(ctx context.Context, id string) (*domain.JournalEntry, error) {
	row, err := r.q.GetJournalEntryByID(ctx, utils.ParseUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get journal entry: %w", err)
	}
	entry := getByIDRowToDomain(row)

	// Fetch lines
	lineRows, err := r.q.GetJournalEntryLinesByEntryID(ctx, row.ID)
	if err != nil {
		return nil, fmt.Errorf("get journal entry lines: %w", err)
	}
	for _, l := range lineRows {
		entry.Lines = append(entry.Lines, journalLineToDomain(l))
	}
	return &entry, nil
}

func (r *JournalRepository) Create(ctx context.Context, entry *domain.JournalEntry) (*domain.JournalEntry, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := r.q.WithTx(tx)

	status := queries.JournalStatusDraft
	if entry.Status == domain.JournalStatusPosted {
		status = queries.JournalStatusPosted
	}

	row, err := qtx.CreateJournalEntry(ctx, queries.CreateJournalEntryParams{
		EntryNumber: entry.EntryNumber,
		Date:        utils.StringToDate(entry.Date),
		Description: entry.Description,
		Source:      pgtype.Text{String: "manual", Valid: true},
		Status:      status,
		CreatedBy:   utils.StringPtrToUUID(entry.CreatedBy),
	})
	if err != nil {
		return nil, fmt.Errorf("insert journal entry: %w", err)
	}

	// Insert lines
	for i, line := range entry.Lines {
		_, err := qtx.CreateJournalEntryLine(ctx, queries.CreateJournalEntryLineParams{
			JournalEntryID: row.ID,
			AccountID:      utils.ParseUUID(line.AccountID),
			Debit:          utils.Float64ToNumeric(line.Debit),
			Credit:         utils.Float64ToNumeric(line.Credit),
			LineOrder:      int32(i + 1),
		})
		if err != nil {
			return nil, fmt.Errorf("insert journal line %d: %w", i+1, err)
		}
	}

	// If the caller wants it posted, transition draft → posted with computed totals.
	entryID := row.ID
	if entry.Status == domain.JournalStatusPosted {
		totals, err := qtx.GetJournalEntryLinesTotals(ctx, entryID)
		if err != nil {
			return nil, fmt.Errorf("compute totals: %w", err)
		}
		_, err = qtx.PostJournalEntry(ctx, queries.PostJournalEntryParams{
			ID:          entryID,
			TotalDebit:  totals.TotalDebit,
			TotalCredit: totals.TotalCredit,
		})
		if err != nil {
			return nil, fmt.Errorf("post journal entry: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// Re-fetch with lines
	return r.GetByID(ctx, utils.UUIDToString(entryID))
}

func (r *JournalRepository) Update(ctx context.Context, entry *domain.JournalEntry, replaceLines bool) (*domain.JournalEntry, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := r.q.WithTx(tx)

	entryUUID := utils.ParseUUID(entry.ID)

	_, err = qtx.UpdateJournalEntry(ctx, queries.UpdateJournalEntryParams{
		ID:          entryUUID,
		Description: entry.Description,
		Date:        utils.StringToDate(entry.Date),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("journal entry not found or not in draft status")
		}
		return nil, fmt.Errorf("update journal entry: %w", err)
	}

	if replaceLines {
		// Delete existing lines
		if err := qtx.DeleteJournalEntryLinesByEntryID(ctx, entryUUID); err != nil {
			return nil, fmt.Errorf("delete existing lines: %w", err)
		}
		// Insert new lines
		for i, line := range entry.Lines {
			_, err := qtx.CreateJournalEntryLine(ctx, queries.CreateJournalEntryLineParams{
				JournalEntryID: entryUUID,
				AccountID:      utils.ParseUUID(line.AccountID),
				Debit:          utils.Float64ToNumeric(line.Debit),
				Credit:         utils.Float64ToNumeric(line.Credit),
				LineOrder:      int32(i + 1),
			})
			if err != nil {
				return nil, fmt.Errorf("insert journal line %d: %w", i+1, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return r.GetByID(ctx, entry.ID)
}

func (r *JournalRepository) Post(ctx context.Context, id string) (*domain.JournalEntry, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := r.q.WithTx(tx)

	entryUUID := utils.ParseUUID(id)

	// Compute totals from lines
	totals, err := qtx.GetJournalEntryLinesTotals(ctx, entryUUID)
	if err != nil {
		return nil, fmt.Errorf("compute totals: %w", err)
	}

	_, err = qtx.PostJournalEntry(ctx, queries.PostJournalEntryParams{
		ID:          entryUUID,
		TotalDebit:  totals.TotalDebit,
		TotalCredit: totals.TotalCredit,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("journal entry not found or not in draft status")
		}
		return nil, fmt.Errorf("post journal entry: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *JournalRepository) Delete(ctx context.Context, id string) error {
	err := r.q.DeleteJournalEntry(ctx, utils.ParseUUID(id))
	if err != nil {
		return fmt.Errorf("delete journal entry: %w", err)
	}
	return nil
}

func (r *JournalRepository) Reverse(ctx context.Context, id string, reason string, createdBy string) (*domain.JournalEntry, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := r.q.WithTx(tx)

	originalUUID := utils.ParseUUID(id)

	// 1. Fetch original entry
	original, err := qtx.GetJournalEntryByID(ctx, originalUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("journal entry not found")
		}
		return nil, fmt.Errorf("get original entry: %w", err)
	}

	if original.Status != queries.JournalStatusPosted {
		return nil, fmt.Errorf("only posted entries can be reversed")
	}

	// 2. Fetch original lines
	originalLines, err := qtx.GetJournalEntryLinesByEntryID(ctx, originalUUID)
	if err != nil {
		return nil, fmt.Errorf("get original lines: %w", err)
	}

	// 3. Generate entry number for the reversal
	entryNumberRaw, err := qtx.GenerateJournalEntryNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate entry number: %w", err)
	}
	reversalNumber := fmt.Sprintf("REV-%v", entryNumberRaw)

	// 4. Create the reversal entry (always posted)
	reversalEntry, err := qtx.CreateReversalJournalEntry(ctx, queries.CreateReversalJournalEntryParams{
		EntryNumber:    reversalNumber,
		Date:           utils.CurrentDate(),
		Description:    fmt.Sprintf("Reversal of %s: %s", original.EntryNumber, reason),
		Source:         pgtype.Text{String: "manual", Valid: true},
		ReversalOf:     originalUUID,
		ReversalReason: pgtype.Text{String: reason, Valid: true},
		CreatedBy:      utils.ParseUUID(createdBy),
	})
	if err != nil {
		return nil, fmt.Errorf("create reversal entry: %w", err)
	}

	// 5. Create reversed lines (swap debit/credit)
	for i, line := range originalLines {
		_, err := qtx.CreateJournalEntryLine(ctx, queries.CreateJournalEntryLineParams{
			JournalEntryID: reversalEntry.ID,
			AccountID:      line.AccountID,
			Debit:          line.Credit, // swap
			Credit:         line.Debit,  // swap
			LineOrder:      int32(i + 1),
		})
		if err != nil {
			return nil, fmt.Errorf("insert reversal line %d: %w", i+1, err)
		}
	}

	// 6. Mark original entry as reversed, linking to the reversal entry
	_, err = qtx.ReverseJournalEntry(ctx, queries.ReverseJournalEntryParams{
		ID:         originalUUID,
		ReversedBy: reversalEntry.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("mark original as reversed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return r.GetByID(ctx, utils.UUIDToString(reversalEntry.ID))
}

func (r *JournalRepository) GenerateEntryNumber(ctx context.Context) (string, error) {
	raw, err := r.q.GenerateJournalEntryNumber(ctx)
	if err != nil {
		return "", fmt.Errorf("generate entry number: %w", err)
	}
	return fmt.Sprintf("%v", raw), nil
}

type journalRowFields struct {
	ID             pgtype.UUID
	EntryNumber    string
	Date           pgtype.Date
	Description    string
	Status         queries.JournalStatus
	TotalDebit     pgtype.Numeric
	TotalCredit    pgtype.Numeric
	ReversalOf     pgtype.UUID
	ReversalReason pgtype.Text
	ReversedBy     pgtype.UUID
	Source         pgtype.Text
	CreatedBy      pgtype.UUID
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
}

func journalFieldsToDomain(f journalRowFields) domain.JournalEntry {
	var source string
	if f.Source.Valid {
		source = f.Source.String
	}
	return domain.JournalEntry{
		ID:             utils.UUIDToString(f.ID),
		EntryNumber:    f.EntryNumber,
		Date:           utils.DateToString(f.Date),
		Description:    f.Description,
		Status:         domain.JournalStatus(f.Status),
		TotalDebit:     utils.NumericToFloat64(f.TotalDebit),
		TotalCredit:    utils.NumericToFloat64(f.TotalCredit),
		ReversalOf:     utils.UUIDToStringPtr(f.ReversalOf),
		ReversalReason: utils.TextToStringPtr(f.ReversalReason),
		ReversedBy:     utils.UUIDToStringPtr(f.ReversedBy),
		Source:         source,
		CreatedBy:      utils.UUIDToStringPtr(f.CreatedBy),
		CreatedAt:      utils.TimestamptzToTime(f.CreatedAt),
		UpdatedAt:      utils.TimestamptzToTime(f.UpdatedAt),
	}
}

func listRowToDomain(r queries.ListJournalEntriesRow) domain.JournalEntry {
	return journalFieldsToDomain(
		journalRowFields{
			ID:             r.ID,
			EntryNumber:    r.EntryNumber,
			Date:           r.Date,
			Description:    r.Description,
			Status:         r.Status,
			TotalDebit:     r.TotalDebit,
			TotalCredit:    r.TotalCredit,
			ReversalOf:     r.ReversalOf,
			ReversalReason: r.ReversalReason,
			ReversedBy:     r.ReversedBy,
			Source:         r.Source,
			CreatedBy:      r.CreatedBy,
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
		})
}

func getByIDRowToDomain(r queries.GetJournalEntryByIDRow) domain.JournalEntry {
	return journalFieldsToDomain(
		journalRowFields{
			ID:             r.ID,
			EntryNumber:    r.EntryNumber,
			Date:           r.Date,
			Description:    r.Description,
			Status:         r.Status,
			TotalDebit:     r.TotalDebit,
			TotalCredit:    r.TotalCredit,
			ReversalOf:     r.ReversalOf,
			ReversalReason: r.ReversalReason,
			ReversedBy:     r.ReversedBy,
			Source:         r.Source,
			CreatedBy:      r.CreatedBy,
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
		})
}

func journalLineToDomain(l queries.GetJournalEntryLinesByEntryIDRow) domain.JournalLine {
	return domain.JournalLine{
		ID:             utils.UUIDToString(l.ID),
		JournalEntryID: utils.UUIDToString(l.JournalEntryID),
		AccountID:      utils.UUIDToString(l.AccountID),
		AccountCode:    l.AccountCode,
		AccountName:    l.AccountName,
		Debit:          utils.NumericToFloat64(l.Debit),
		Credit:         utils.NumericToFloat64(l.Credit),
		LineOrder:      int(l.LineOrder),
		CreatedAt:      utils.TimestamptzToTime(l.CreatedAt),
	}
}
