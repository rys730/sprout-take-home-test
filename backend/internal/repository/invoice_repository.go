package repository

import (
	"context"
	"errors"
	"fmt"

	"sprout-backend/db/queries"
	"sprout-backend/internal/domain"
	"sprout-backend/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ---------------------------------------------------------------------------
// Invoice Repository
// ---------------------------------------------------------------------------

type InvoiceRepository struct {
	pool *pgxpool.Pool
	q    *queries.Queries
}

func NewInvoiceRepository(pool *pgxpool.Pool) *InvoiceRepository {
	return &InvoiceRepository{
		pool: pool,
		q:    queries.New(pool),
	}
}

func invoiceToDomain(i queries.Invoice) domain.Invoice {
	return domain.Invoice{
		ID:            utils.UUIDToString(i.ID),
		InvoiceNumber: i.InvoiceNumber,
		CustomerID:    utils.UUIDToString(i.CustomerID),
		IssueDate:     utils.DateToString(i.IssueDate),
		DueDate:       utils.DateToString(i.DueDate),
		TotalAmount:   utils.NumericToFloat64(i.TotalAmount),
		AmountPaid:    utils.NumericToFloat64(i.AmountPaid),
		Status:        domain.InvoiceStatus(i.Status),
		Description:   utils.TextToStringPtr(i.Description),
		CreatedBy:     utils.UUIDToStringPtr(i.CreatedBy),
		CreatedAt:     utils.TimestamptzToTime(i.CreatedAt),
		UpdatedAt:     utils.TimestamptzToTime(i.UpdatedAt),
	}
}

func listInvoiceRowToDomain(r queries.ListInvoicesRow) domain.Invoice {
	return domain.Invoice{
		ID:            utils.UUIDToString(r.ID),
		InvoiceNumber: r.InvoiceNumber,
		CustomerID:    utils.UUIDToString(r.CustomerID),
		CustomerName:  r.CustomerName,
		IssueDate:     utils.DateToString(r.IssueDate),
		DueDate:       utils.DateToString(r.DueDate),
		TotalAmount:   utils.NumericToFloat64(r.TotalAmount),
		AmountPaid:    utils.NumericToFloat64(r.AmountPaid),
		Status:        domain.InvoiceStatus(r.Status),
		Description:   utils.TextToStringPtr(r.Description),
		CreatedBy:     utils.UUIDToStringPtr(r.CreatedBy),
		CreatedAt:     utils.TimestamptzToTime(r.CreatedAt),
		UpdatedAt:     utils.TimestamptzToTime(r.UpdatedAt),
	}
}

func agingRowToDomain(r queries.ListInvoicesAgingRow) domain.Invoice {
	return domain.Invoice{
		ID:            utils.UUIDToString(r.ID),
		InvoiceNumber: r.InvoiceNumber,
		CustomerID:    utils.UUIDToString(r.CustomerID),
		CustomerName:  r.CustomerName,
		IssueDate:     utils.DateToString(r.IssueDate),
		DueDate:       utils.DateToString(r.DueDate),
		TotalAmount:   utils.NumericToFloat64(r.TotalAmount),
		AmountPaid:    utils.NumericToFloat64(r.AmountPaid),
		Status:        domain.InvoiceStatus(r.Status),
		Description:   utils.TextToStringPtr(r.Description),
		DaysOverdue:   int(r.DaysOverdue),
		CreatedBy:     utils.UUIDToStringPtr(r.CreatedBy),
		CreatedAt:     utils.TimestamptzToTime(r.CreatedAt),
		UpdatedAt:     utils.TimestamptzToTime(r.UpdatedAt),
	}
}

func (r *InvoiceRepository) GetByID(ctx context.Context, id string) (*domain.Invoice, error) {
	row, err := r.q.GetInvoiceByID(ctx, utils.ParseUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get invoice by id: %w", err)
	}
	inv := invoiceToDomain(row)
	return &inv, nil
}

func (r *InvoiceRepository) List(ctx context.Context, filter domain.InvoiceFilter) ([]domain.Invoice, int64, error) {
	limit := int32(filter.Limit)
	if limit <= 0 {
		limit = 20
	}
	offset := int32(filter.Offset)
	if offset < 0 {
		offset = 0
	}

	params := queries.ListInvoicesParams{
		Limit:      limit,
		Offset:     offset,
		CustomerID: utils.StringPtrToUUID(nilIfEmpty(filter.CustomerID)),
		Status: queries.NullInvoiceStatus{
			InvoiceStatus: queries.InvoiceStatus(filter.Status),
			Valid:         filter.Status != "",
		},
	}

	rows, err := r.q.ListInvoices(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("list invoices: %w", err)
	}

	total, err := r.q.CountInvoices(ctx, queries.CountInvoicesParams{
		CustomerID: params.CustomerID,
		Status:     params.Status,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count invoices: %w", err)
	}

	invoices := make([]domain.Invoice, 0, len(rows))
	for _, row := range rows {
		invoices = append(invoices, listInvoiceRowToDomain(row))
	}
	return invoices, total, nil
}

func (r *InvoiceRepository) ListUnpaidByCustomer(ctx context.Context, customerID string) ([]domain.Invoice, error) {
	rows, err := r.q.ListUnpaidInvoicesByCustomer(ctx, utils.ParseUUID(customerID))
	if err != nil {
		return nil, fmt.Errorf("list unpaid invoices: %w", err)
	}
	invoices := make([]domain.Invoice, 0, len(rows))
	for _, row := range rows {
		invoices = append(invoices, invoiceToDomain(row))
	}
	return invoices, nil
}

func (r *InvoiceRepository) ListAging(ctx context.Context, limit, offset int) ([]domain.Invoice, int64, error) {
	l := int32(limit)
	if l <= 0 {
		l = 20
	}
	o := int32(offset)
	if o < 0 {
		o = 0
	}

	rows, err := r.q.ListInvoicesAging(ctx, queries.ListInvoicesAgingParams{
		Limit:  l,
		Offset: o,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list aging invoices: %w", err)
	}

	total, err := r.q.CountInvoicesAging(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count aging invoices: %w", err)
	}

	invoices := make([]domain.Invoice, 0, len(rows))
	for _, row := range rows {
		invoices = append(invoices, agingRowToDomain(row))
	}
	return invoices, total, nil
}

func (r *InvoiceRepository) GetReceivablesSummary(ctx context.Context) (*domain.ReceivablesSummary, error) {
	row, err := r.q.GetReceivablesSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("get receivables summary: %w", err)
	}
	return &domain.ReceivablesSummary{
		TotalOutstanding: utils.NumericToFloat64(row.TotalOutstanding),
		TotalOverdue:     utils.NumericToFloat64(row.TotalOverdue),
	}, nil
}

func (r *InvoiceRepository) Create(ctx context.Context, invoice *domain.Invoice) (*domain.Invoice, error) {
	row, err := r.q.CreateInvoice(ctx, queries.CreateInvoiceParams{
		InvoiceNumber: invoice.InvoiceNumber,
		CustomerID:    utils.ParseUUID(invoice.CustomerID),
		IssueDate:     utils.StringToDate(invoice.IssueDate),
		DueDate:       utils.StringToDate(invoice.DueDate),
		TotalAmount:   utils.Float64ToNumeric(invoice.TotalAmount),
		Description:   utils.TextFromString(ptrToString(invoice.Description)),
		CreatedBy:     utils.StringPtrToUUID(invoice.CreatedBy),
	})
	if err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}
	inv := invoiceToDomain(row)
	return &inv, nil
}

func (r *InvoiceRepository) Update(ctx context.Context, invoice *domain.Invoice) (*domain.Invoice, error) {
	row, err := r.q.UpdateInvoice(ctx, queries.UpdateInvoiceParams{
		ID:            utils.ParseUUID(invoice.ID),
		InvoiceNumber: invoice.InvoiceNumber,
		IssueDate:     utils.StringToDate(invoice.IssueDate),
		DueDate:       utils.StringToDate(invoice.DueDate),
		TotalAmount:   utils.Float64ToNumeric(invoice.TotalAmount),
		Description:   utils.TextFromString(ptrToString(invoice.Description)),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("invoice not found or not in unpaid status")
		}
		return nil, fmt.Errorf("update invoice: %w", err)
	}
	inv := invoiceToDomain(row)
	return &inv, nil
}

func (r *InvoiceRepository) Delete(ctx context.Context, id string) error {
	err := r.q.DeleteInvoice(ctx, utils.ParseUUID(id))
	if err != nil {
		return fmt.Errorf("delete invoice: %w", err)
	}
	return nil
}

func (r *InvoiceRepository) GenerateInvoiceNumber(ctx context.Context) (string, error) {
	raw, err := r.q.GenerateInvoiceNumber(ctx)
	if err != nil {
		return "", fmt.Errorf("generate invoice number: %w", err)
	}
	return fmt.Sprintf("%v", raw), nil
}

// nilIfEmpty returns nil if the string is empty, otherwise returns a pointer to it.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
