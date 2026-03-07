-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS journal_entry_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    description TEXT,
    debit NUMERIC(18, 2) NOT NULL DEFAULT 0,
    credit NUMERIC(18, 2) NOT NULL DEFAULT 0,
    line_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_journal_entry_lines_journal_id ON journal_entry_lines(journal_entry_id);
CREATE INDEX idx_journal_entry_lines_account_id ON journal_entry_lines(account_id);

ALTER TABLE journal_entry_lines
    ADD CONSTRAINT chk_debit_or_credit
    CHECK (
        (debit > 0 AND credit = 0) OR (debit = 0 AND credit > 0)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_journal_entry_lines_journal_id;
DROP INDEX IF EXISTS idx_journal_entry_lines_account_id;

DROP TABLE IF EXISTS journal_entry_lines;

DROP CONSTRAINT IF EXISTS chk_debit_or_credit ON journal_entry_lines;
-- +goose StatementEnd
