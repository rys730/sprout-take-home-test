-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS payment_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE RESTRICT,
    amount NUMERIC(18, 2) NOT NULL,                 -- amount applied to this invoice
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT chk_allocation_amount CHECK (amount > 0),
    CONSTRAINT uq_payment_invoice UNIQUE (payment_id, invoice_id) -- one allocation per payment-invoice pair
);

CREATE INDEX idx_payment_allocations_payment_id ON payment_allocations(payment_id);
CREATE INDEX idx_payment_allocations_invoice_id ON payment_allocations(invoice_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_payment_allocations_invoice_id;
DROP INDEX IF EXISTS idx_payment_allocations_payment_id;

DROP TABLE IF EXISTS payment_allocations;
-- +goose StatementEnd
