-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_number VARCHAR(50) UNIQUE NOT NULL,          -- e.g., "PAY-2025-001"
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    payment_date DATE NOT NULL,
    amount NUMERIC(18, 2) NOT NULL,                      -- total amount received
    deposit_to_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT, -- bank/cash account to debit
    journal_entry_id UUID REFERENCES journal_entries(id), -- auto-generated journal entry
    notes TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT chk_payment_amount CHECK (amount > 0)
);

CREATE INDEX idx_payments_customer_id ON payments(customer_id);
CREATE INDEX idx_payments_payment_date ON payments(payment_date);
CREATE INDEX idx_payments_journal_entry_id ON payments(journal_entry_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_payments_journal_entry_id;
DROP INDEX IF EXISTS idx_payments_payment_date;
DROP INDEX IF EXISTS idx_payments_customer_id;

DROP TABLE IF EXISTS payments;
-- +goose StatementEnd
