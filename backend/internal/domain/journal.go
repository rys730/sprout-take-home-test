package domain

import (
	"context"
	"time"
)

// JournalStatus represents the lifecycle state of a journal entry.
type JournalStatus string

const (
	JournalStatusDraft    JournalStatus = "draft"
	JournalStatusPosted   JournalStatus = "posted"
	JournalStatusReversed JournalStatus = "reversed"
)

// JournalEntry represents a general journal entry header (Jurnal Umum).
type JournalEntry struct {
	ID             string        `json:"id"`
	EntryNumber    string        `json:"entry_number"`
	Date           string        `json:"date"` // YYYY-MM-DD
	Description    string        `json:"description"`
	Status         JournalStatus `json:"status"`
	TotalDebit     float64       `json:"total_debit"`
	TotalCredit    float64       `json:"total_credit"`
	ReversalOf     *string       `json:"reversal_of,omitempty"`
	ReversalReason *string       `json:"reversal_reason,omitempty"`
	ReversedBy     *string       `json:"reversed_by,omitempty"` // ID of the reversing entry
	Source         string        `json:"source"`
	CreatedBy      *string       `json:"created_by,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Lines          []JournalLine `json:"lines,omitempty"`
}

// JournalLine represents a single debit/credit line inside a journal entry.
type JournalLine struct {
	ID             string    `json:"id"`
	JournalEntryID string    `json:"journal_entry_id"`
	AccountID      string    `json:"account_id"`
	AccountCode    string    `json:"account_code,omitempty"`
	AccountName    string    `json:"account_name,omitempty"`
	Description    string    `json:"description,omitempty"`
	Debit          float64   `json:"debit"`
	Credit         float64   `json:"credit"`
	LineOrder      int       `json:"line_order"`
	CreatedAt      time.Time `json:"created_at"`
}

// ---------------------------------------------------------------------------
// Request / Response types
// ---------------------------------------------------------------------------

// CreateJournalRequest is the payload for creating a new journal entry.
type CreateJournalRequest struct {
	Date        string              `json:"date" validate:"required"` // YYYY-MM-DD
	Description string              `json:"description" validate:"required"`
	InvoiceID   *string             `json:"invoice_id,omitempty"` // optional link to an invoice
	Status      string              `json:"status"` // "draft" or "posted"; defaults to "draft"
	Lines       []CreateJournalLine `json:"lines" validate:"required,min=2"`
}

// CreateJournalLine represents a single line in the create request.
type CreateJournalLine struct {
	AccountID   string  `json:"account_id" validate:"required"`
	Debit       float64 `json:"debit"`
	Credit      float64 `json:"credit"`
}

// UpdateJournalRequest is the payload for updating a draft journal entry.
type UpdateJournalRequest struct {
	Date        *string             `json:"date,omitempty"`
	Description *string             `json:"description,omitempty"`
	Lines       []CreateJournalLine `json:"lines,omitempty"` // if provided, replaces all lines
}

// ReverseJournalRequest is the payload for reversing a posted journal entry.
type ReverseJournalRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// JournalFilter holds optional filter/pagination parameters.
type JournalFilter struct {
	Status    string // draft, posted, reversed
	Source    string // manual, payment, opening_balance, adjustment
	StartDate string // YYYY-MM-DD
	EndDate   string // YYYY-MM-DD
	Limit     int
	Offset    int
}

// ---------------------------------------------------------------------------
// Repository interface
// ---------------------------------------------------------------------------

// JournalRepository defines the persistence interface for journal entries.
type JournalRepository interface {
	// List returns journal entries matching the filter, with count for pagination.
	List(ctx context.Context, filter JournalFilter) ([]JournalEntry, int64, error)
	// GetByID returns a journal entry with its lines.
	GetByID(ctx context.Context, id string) (*JournalEntry, error)
	// Create creates a journal entry with its lines in a single transaction.
	// If status == "posted", totals are computed and stored.
	Create(ctx context.Context, entry *JournalEntry) (*JournalEntry, error)
	// Update updates a draft journal entry header (and optionally replaces lines).
	Update(ctx context.Context, entry *JournalEntry, replaceLines bool) (*JournalEntry, error)
	// Post transitions a draft entry to posted, computing and storing totals.
	Post(ctx context.Context, id string) (*JournalEntry, error)
	// Delete deletes a draft journal entry (and cascading lines).
	Delete(ctx context.Context, id string) error
	// Reverse creates a new reversing entry and marks the original as reversed.
	Reverse(ctx context.Context, id string, reason string, createdBy string) (*JournalEntry, error)
	// GenerateEntryNumber returns the next sequential entry number.
	GenerateEntryNumber(ctx context.Context) (string, error)
}

// ---------------------------------------------------------------------------
// Use-case interface
// ---------------------------------------------------------------------------

// JournalUseCase defines the business logic interface for the General Journal.
type JournalUseCase interface {
	List(ctx context.Context, filter JournalFilter) ([]JournalEntry, int64, error)
	GetByID(ctx context.Context, id string) (*JournalEntry, error)
	Create(ctx context.Context, req CreateJournalRequest, createdBy string) (*JournalEntry, error)
	Update(ctx context.Context, id string, req UpdateJournalRequest) (*JournalEntry, error)
	Post(ctx context.Context, id string) (*JournalEntry, error)
	Delete(ctx context.Context, id string) error
	Reverse(ctx context.Context, id string, req ReverseJournalRequest, createdBy string) (*JournalEntry, error)
}
