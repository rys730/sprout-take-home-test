package domain

import (
	"context"
	"time"
)

// ---------------------------------------------------------------------------
// Invoice Status
// ---------------------------------------------------------------------------

// InvoiceStatus represents the payment state of an invoice.
type InvoiceStatus string

const (
	InvoiceStatusUnpaid        InvoiceStatus = "unpaid"
	InvoiceStatusPartiallyPaid InvoiceStatus = "partially_paid"
	InvoiceStatusPaid          InvoiceStatus = "paid"
)

// ---------------------------------------------------------------------------
// Customer
// ---------------------------------------------------------------------------

// Customer represents a customer (pelanggan) who owes us money.
type Customer struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     *string   `json:"email,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	Address   *string   `json:"address,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedBy *string   `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCustomerRequest holds the data needed to create a new customer.
type CreateCustomerRequest struct {
	Name    string  `json:"name" validate:"required"`
	Email   *string `json:"email,omitempty"`
	Phone   *string `json:"phone,omitempty"`
	Address *string `json:"address,omitempty"`
}

// UpdateCustomerRequest holds the data allowed to update on a customer.
type UpdateCustomerRequest struct {
	Name    *string `json:"name,omitempty"`
	Email   *string `json:"email,omitempty"`
	Phone   *string `json:"phone,omitempty"`
	Address *string `json:"address,omitempty"`
}

// CustomerFilter holds optional filter/pagination parameters for listing customers.
type CustomerFilter struct {
	Search string
	Limit  int
	Offset int
}

// ---------------------------------------------------------------------------
// Invoice
// ---------------------------------------------------------------------------

// Invoice represents a customer invoice (piutang / accounts receivable).
type Invoice struct {
	ID            string        `json:"id"`
	InvoiceNumber string        `json:"invoice_number"`
	CustomerID    string        `json:"customer_id"`
	CustomerName  string        `json:"customer_name,omitempty"`
	IssueDate     string        `json:"issue_date"` // YYYY-MM-DD
	DueDate       string        `json:"due_date"`   // YYYY-MM-DD
	TotalAmount   float64       `json:"total_amount"`
	AmountPaid    float64       `json:"amount_paid"`
	Status        InvoiceStatus `json:"status"`
	Description   *string       `json:"description,omitempty"`
	DaysOverdue   int           `json:"days_overdue,omitempty"` // positive = overdue, negative = days until due
	CreatedBy     *string       `json:"created_by,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// CreateInvoiceRequest holds the data needed to create a new invoice.
type CreateInvoiceRequest struct {
	CustomerID  string  `json:"customer_id" validate:"required"`
	IssueDate   string  `json:"issue_date" validate:"required"` // YYYY-MM-DD
	DueDate     string  `json:"due_date" validate:"required"`   // YYYY-MM-DD
	TotalAmount float64 `json:"total_amount" validate:"required,gt=0"`
	Description *string `json:"description,omitempty"`
}

// UpdateInvoiceRequest holds the data allowed to update on an unpaid invoice.
type UpdateInvoiceRequest struct {
	IssueDate   *string  `json:"issue_date,omitempty"`
	DueDate     *string  `json:"due_date,omitempty"`
	TotalAmount *float64 `json:"total_amount,omitempty"`
	Description *string  `json:"description,omitempty"`
}

// InvoiceFilter holds optional filter/pagination parameters for listing invoices.
type InvoiceFilter struct {
	CustomerID string
	Status     string // unpaid, partially_paid, paid
	Limit      int
	Offset     int
}

// ---------------------------------------------------------------------------
// Payment & Allocation
// ---------------------------------------------------------------------------

// Payment represents a customer payment (pembayaran).
type Payment struct {
	ID                 string              `json:"id"`
	PaymentNumber      string              `json:"payment_number"`
	CustomerID         string              `json:"customer_id"`
	CustomerName       string              `json:"customer_name,omitempty"`
	PaymentDate        string              `json:"payment_date"` // YYYY-MM-DD
	Amount             float64             `json:"amount"`
	DepositToAccountID string              `json:"deposit_to_account_id"`
	DepositAccountCode string              `json:"deposit_account_code,omitempty"`
	DepositAccountName string              `json:"deposit_account_name,omitempty"`
	JournalEntryID     *string             `json:"journal_entry_id,omitempty"`
	Notes              *string             `json:"notes,omitempty"`
	Allocations        []PaymentAllocation `json:"allocations,omitempty"`
	CreatedBy          *string             `json:"created_by,omitempty"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

// PaymentAllocation represents an allocation of a payment to a specific invoice.
type PaymentAllocation struct {
	ID            string    `json:"id"`
	PaymentID     string    `json:"payment_id"`
	InvoiceID     string    `json:"invoice_id"`
	InvoiceNumber string    `json:"invoice_number,omitempty"`
	PaymentNumber string    `json:"payment_number,omitempty"`
	Amount        float64   `json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
}

// PaymentAllocationLine represents a single allocation in the create payment request.
type PaymentAllocationLine struct {
	InvoiceID string  `json:"invoice_id" validate:"required"`
	Amount    float64 `json:"amount" validate:"required,gt=0"`
}

// CreatePaymentRequest holds the data needed to record a customer payment.
type CreatePaymentRequest struct {
	CustomerID         string                  `json:"customer_id" validate:"required"`
	PaymentDate        string                  `json:"payment_date" validate:"required"` // YYYY-MM-DD
	Amount             float64                 `json:"amount" validate:"required,gt=0"`
	DepositToAccountID string                  `json:"deposit_to_account_id" validate:"required"`
	Notes              *string                 `json:"notes,omitempty"`
	Allocations        []PaymentAllocationLine `json:"allocations" validate:"required,min=1"`
}

// PaymentFilter holds optional filter/pagination parameters for listing payments.
type PaymentFilter struct {
	CustomerID string
	Limit      int
	Offset     int
}

// ReceivablesSummary holds the dashboard summary cards data.
type ReceivablesSummary struct {
	TotalOutstanding float64 `json:"total_outstanding"` // Total Piutang
	TotalOverdue     float64 `json:"total_overdue"`     // Total Jatuh Tempo
}

// ---------------------------------------------------------------------------
// Repository interfaces
// ---------------------------------------------------------------------------

// CustomerRepository defines the persistence interface for customers.
type CustomerRepository interface {
	GetByID(ctx context.Context, id string) (*Customer, error)
	List(ctx context.Context, filter CustomerFilter) ([]Customer, int64, error)
	Create(ctx context.Context, customer *Customer) (*Customer, error)
	Update(ctx context.Context, customer *Customer) (*Customer, error)
	Delete(ctx context.Context, id string) error
}

// InvoiceRepository defines the persistence interface for invoices.
type InvoiceRepository interface {
	GetByID(ctx context.Context, id string) (*Invoice, error)
	List(ctx context.Context, filter InvoiceFilter) ([]Invoice, int64, error)
	ListUnpaidByCustomer(ctx context.Context, customerID string) ([]Invoice, error)
	ListAging(ctx context.Context, limit, offset int) ([]Invoice, int64, error)
	GetReceivablesSummary(ctx context.Context) (*ReceivablesSummary, error)
	Create(ctx context.Context, invoice *Invoice) (*Invoice, error)
	Update(ctx context.Context, invoice *Invoice) (*Invoice, error)
	Delete(ctx context.Context, id string) error
	GenerateInvoiceNumber(ctx context.Context) (string, error)
}

// PaymentRepository defines the persistence interface for payments.
type PaymentRepository interface {
	GetByID(ctx context.Context, id string) (*Payment, error)
	List(ctx context.Context, filter PaymentFilter) ([]Payment, int64, error)
	// RecordPayment performs the full payment flow in a single transaction:
	// 1. Generate payment number
	// 2. Create auto-journal entry (Debit bank, Credit Piutang Usaha 112.000)
	// 3. Create payment record
	// 4. Create payment allocations
	// 5. Update invoice amount_paid and status
	RecordPayment(ctx context.Context, req CreatePaymentRequest, createdBy string) (*Payment, error)
}

// ---------------------------------------------------------------------------
// Use-case interfaces
// ---------------------------------------------------------------------------

// PaymentUseCase defines the business logic interface for payments.
type PaymentUseCase interface {
	GetByID(ctx context.Context, id string) (*Payment, error)
	List(ctx context.Context, filter PaymentFilter) ([]Payment, int64, error)
	RecordPayment(ctx context.Context, req CreatePaymentRequest, createdBy string) (*Payment, error)
	GetReceivablesSummary(ctx context.Context) (*ReceivablesSummary, error)
}
