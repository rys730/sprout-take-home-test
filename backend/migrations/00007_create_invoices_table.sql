-- +goose Up
-- +goose StatementBegin
CREATE TYPE invoice_status AS ENUM ('unpaid', 'partially_paid', 'paid');

CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,     -- e.g., "INV-2025-001"
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    issue_date DATE NOT NULL,
    due_date DATE NOT NULL,
    total_amount NUMERIC(18, 2) NOT NULL DEFAULT 0, -- original invoice total
    amount_paid NUMERIC(18, 2) NOT NULL DEFAULT 0,  -- cumulative amount paid so far
    status invoice_status NOT NULL DEFAULT 'unpaid',
    description TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT chk_amount_paid CHECK (amount_paid >= 0 AND amount_paid <= total_amount),
    CONSTRAINT chk_total_amount CHECK (total_amount > 0)
);

CREATE INDEX idx_invoices_customer_id ON invoices(customer_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);
CREATE INDEX idx_invoices_invoice_number ON invoices(invoice_number);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_invoices_invoice_number;
DROP INDEX IF EXISTS idx_invoices_due_date;
DROP INDEX IF EXISTS idx_invoices_status;
DROP INDEX IF EXISTS idx_invoices_customer_id;

DROP TABLE IF EXISTS invoices;

DROP TYPE IF EXISTS invoice_status;
-- +goose StatementEnd
