-- +goose Up
-- +goose StatementBegin
CREATE TYPE journal_status AS ENUM ('draft', 'posted', 'reversed');

CREATE TABLE IF NOT EXISTS journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_number VARCHAR(50) UNIQUE NOT NULL,       -- e.g., "JU-2025-001"
    invoice_number VARCHAR(50),                      -- optional, for payment entries, currently not implemented
    date DATE NOT NULL,
    description TEXT NOT NULL,
    status journal_status NOT NULL DEFAULT 'draft',
    total_debit NUMERIC(18, 2) NOT NULL DEFAULT 0,
    total_credit NUMERIC(18, 2) NOT NULL DEFAULT 0,
    reversal_of UUID REFERENCES journal_entries(id), -- if this is a reversal, points to original
    reversal_reason TEXT,                             -- reason for reversal
    reversed_by UUID REFERENCES journal_entries(id),  -- the reversing entry (set on original)
    source VARCHAR(50) DEFAULT 'manual',              -- 'manual' or 'payment' (auto-generated)
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_journal_entries_date ON journal_entries(date);
CREATE INDEX idx_journal_entries_status ON journal_entries(status);
CREATE INDEX idx_journal_entries_entry_number ON journal_entries(entry_number);
CREATE INDEX idx_journal_entries_reversal_of ON journal_entries(reversal_of);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_journal_entries_reversal_of;
DROP INDEX IF EXISTS idx_journal_entries_entry_number;
DROP INDEX IF EXISTS idx_journal_entries_status;
DROP INDEX IF EXISTS idx_journal_entries_date;

DROP TABLE IF EXISTS journal_entries;
DROP TYPE IF EXISTS journal_status;
-- +goose StatementEnd
