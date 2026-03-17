-- +goose Up
-- +goose StatementBegin

-- Replace the loose invoice_number text column with a proper FK to invoices
ALTER TABLE journal_entries
    DROP COLUMN IF EXISTS invoice_number,
    ADD COLUMN invoice_id UUID REFERENCES invoices(id) ON DELETE SET NULL;

CREATE INDEX idx_journal_entries_invoice_id ON journal_entries(invoice_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_journal_entries_invoice_id;

ALTER TABLE journal_entries
    DROP COLUMN IF EXISTS invoice_id,
    ADD COLUMN invoice_number VARCHAR(50);

-- +goose StatementEnd
