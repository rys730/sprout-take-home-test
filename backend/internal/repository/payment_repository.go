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

// ---------------------------------------------------------------------------
// Payment Repository
// ---------------------------------------------------------------------------

type PaymentRepository struct {
	pool *pgxpool.Pool
	q    *queries.Queries
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{
		pool: pool,
		q:    queries.New(pool),
	}
}

func getByIDRowToPaymentDomain(r queries.GetPaymentByIDRow) domain.Payment {
	return domain.Payment{
		ID:                 utils.UUIDToString(r.ID),
		PaymentNumber:      r.PaymentNumber,
		CustomerID:         utils.UUIDToString(r.CustomerID),
		CustomerName:       r.CustomerName,
		PaymentDate:        utils.DateToString(r.PaymentDate),
		Amount:             utils.NumericToFloat64(r.Amount),
		DepositToAccountID: utils.UUIDToString(r.DepositToAccountID),
		DepositAccountCode: r.DepositAccountCode,
		DepositAccountName: r.DepositAccountName,
		JournalEntryID:     utils.UUIDToStringPtr(r.JournalEntryID),
		Notes:              utils.TextToStringPtr(r.Notes),
		CreatedBy:          utils.UUIDToStringPtr(r.CreatedBy),
		CreatedAt:          utils.TimestamptzToTime(r.CreatedAt),
		UpdatedAt:          utils.TimestamptzToTime(r.UpdatedAt),
	}
}

func listRowToPaymentDomain(r queries.ListPaymentsRow) domain.Payment {
	return domain.Payment{
		ID:                 utils.UUIDToString(r.ID),
		PaymentNumber:      r.PaymentNumber,
		CustomerID:         utils.UUIDToString(r.CustomerID),
		CustomerName:       r.CustomerName,
		PaymentDate:        utils.DateToString(r.PaymentDate),
		Amount:             utils.NumericToFloat64(r.Amount),
		DepositToAccountID: utils.UUIDToString(r.DepositToAccountID),
		DepositAccountCode: r.DepositAccountCode,
		DepositAccountName: r.DepositAccountName,
		JournalEntryID:     utils.UUIDToStringPtr(r.JournalEntryID),
		Notes:              utils.TextToStringPtr(r.Notes),
		CreatedBy:          utils.UUIDToStringPtr(r.CreatedBy),
		CreatedAt:          utils.TimestamptzToTime(r.CreatedAt),
		UpdatedAt:          utils.TimestamptzToTime(r.UpdatedAt),
	}
}

func (r *PaymentRepository) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	row, err := r.q.GetPaymentByID(ctx, utils.ParseUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get payment by id: %w", err)
	}
	p := getByIDRowToPaymentDomain(row)

	// Fetch allocations
	allocRows, err := r.q.GetPaymentAllocationsByPaymentID(ctx, row.ID)
	if err != nil {
		return nil, fmt.Errorf("get payment allocations: %w", err)
	}
	for _, a := range allocRows {
		p.Allocations = append(p.Allocations, domain.PaymentAllocation{
			ID:            utils.UUIDToString(a.ID),
			PaymentID:     utils.UUIDToString(a.PaymentID),
			InvoiceID:     utils.UUIDToString(a.InvoiceID),
			InvoiceNumber: a.InvoiceNumber,
			Amount:        utils.NumericToFloat64(a.Amount),
			CreatedAt:     utils.TimestamptzToTime(a.CreatedAt),
		})
	}

	return &p, nil
}

func (r *PaymentRepository) List(ctx context.Context, filter domain.PaymentFilter) ([]domain.Payment, int64, error) {
	limit := int32(filter.Limit)
	if limit <= 0 {
		limit = 20
	}
	offset := int32(filter.Offset)
	if offset < 0 {
		offset = 0
	}

	params := queries.ListPaymentsParams{
		Limit:      limit,
		Offset:     offset,
		CustomerID: utils.StringPtrToUUID(nilIfEmpty(filter.CustomerID)),
	}

	rows, err := r.q.ListPayments(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("list payments: %w", err)
	}

	total, err := r.q.CountPayments(ctx, params.CustomerID)
	if err != nil {
		return nil, 0, fmt.Errorf("count payments: %w", err)
	}

	payments := make([]domain.Payment, 0, len(rows))
	for _, row := range rows {
		payments = append(payments, listRowToPaymentDomain(row))
	}
	return payments, total, nil
}

// RecordPayment performs the full payment recording flow in a single transaction:
// 1. Generate payment number
// 2. Create auto-journal entry: Debit bank account, Credit Piutang Usaha (112.000)
// 3. Create payment record
// 4. Create payment allocations
// 5. Update each invoice's amount_paid and status
func (r *PaymentRepository) RecordPayment(ctx context.Context, req domain.CreatePaymentRequest, createdBy string) (*domain.Payment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := r.q.WithTx(tx)

	// 1. Generate payment number
	payNumRaw, err := qtx.GeneratePaymentNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate payment number: %w", err)
	}
	paymentNumber := fmt.Sprintf("%v", payNumRaw)

	// 2. Create auto-journal entry: Debit bank, Credit Piutang Usaha (112.000)
	entryNumRaw, err := qtx.GenerateJournalEntryNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate journal entry number: %w", err)
	}
	entryNumber := fmt.Sprintf("PAY-%v", entryNumRaw)

	journalEntry, err := qtx.CreateJournalEntry(ctx, queries.CreateJournalEntryParams{
		EntryNumber: entryNumber,
		Date:        utils.StringToDate(req.PaymentDate),
		Description: fmt.Sprintf("Payment %s received", paymentNumber),
		Source:      pgtype.Text{String: "payment", Valid: true},
		Status:      queries.JournalStatusDraft,
		CreatedBy:   utils.ParseUUID(createdBy),
	})
	if err != nil {
		return nil, fmt.Errorf("create journal entry: %w", err)
	}

	// Debit line: Bank/Cash account (deposit_to_account_id)
	amountNum := utils.Float64ToNumeric(req.Amount)
	zeroNum := utils.Float64ToNumeric(0)

	_, err = qtx.CreateJournalEntryLine(ctx, queries.CreateJournalEntryLineParams{
		JournalEntryID: journalEntry.ID,
		AccountID:      utils.ParseUUID(req.DepositToAccountID),
		Debit:          amountNum,
		Credit:         zeroNum,
		LineOrder:      1,
	})
	if err != nil {
		return nil, fmt.Errorf("create debit journal line: %w", err)
	}

	// Credit line: Piutang Usaha (112.000)
	piutangID, err := qtx.GetAccountIDByCode(ctx, "112.000")
	if err != nil {
		return nil, fmt.Errorf("lookup Piutang Usaha account (112.000): %w", err)
	}

	_, err = qtx.CreateJournalEntryLine(ctx, queries.CreateJournalEntryLineParams{
		JournalEntryID: journalEntry.ID,
		AccountID:      piutangID,
		Debit:          zeroNum,
		Credit:         amountNum,
		LineOrder:      2,
	})
	if err != nil {
		return nil, fmt.Errorf("create credit journal line: %w", err)
	}

	// Post the journal entry
	totals, err := qtx.GetJournalEntryLinesTotals(ctx, journalEntry.ID)
	if err != nil {
		return nil, fmt.Errorf("compute journal totals: %w", err)
	}
	_, err = qtx.PostJournalEntry(ctx, queries.PostJournalEntryParams{
		ID:          journalEntry.ID,
		TotalDebit:  totals.TotalDebit,
		TotalCredit: totals.TotalCredit,
	})
	if err != nil {
		return nil, fmt.Errorf("post journal entry: %w", err)
	}

	// 3. Create payment record
	payment, err := qtx.CreatePayment(ctx, queries.CreatePaymentParams{
		PaymentNumber:      paymentNumber,
		CustomerID:         utils.ParseUUID(req.CustomerID),
		PaymentDate:        utils.StringToDate(req.PaymentDate),
		Amount:             amountNum,
		DepositToAccountID: utils.ParseUUID(req.DepositToAccountID),
		JournalEntryID:     journalEntry.ID,
		Notes:              utils.TextFromString(ptrToString(req.Notes)),
		CreatedBy:          utils.ParseUUID(createdBy),
	})
	if err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}

	// 4. Create allocations and update invoices
	for i, alloc := range req.Allocations {
		_, err := qtx.CreatePaymentAllocation(ctx, queries.CreatePaymentAllocationParams{
			PaymentID: payment.ID,
			InvoiceID: utils.ParseUUID(alloc.InvoiceID),
			Amount:    utils.Float64ToNumeric(alloc.Amount),
		})
		if err != nil {
			return nil, fmt.Errorf("create allocation %d: %w", i+1, err)
		}

		// 5. Update invoice amount_paid and status
		err = qtx.UpdateInvoiceAmountPaid(ctx, queries.UpdateInvoiceAmountPaidParams{
			Addition: utils.Float64ToNumeric(alloc.Amount),
			ID:       utils.ParseUUID(alloc.InvoiceID),
		})
		if err != nil {
			return nil, fmt.Errorf("update invoice %d amount_paid: %w", i+1, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// Re-fetch the payment with all joins
	return r.GetByID(ctx, utils.UUIDToString(payment.ID))
}
