-- ============================================================================
-- INVOICES (Penagihan / AR)
-- ============================================================================

-- name: GetInvoiceByID :one
SELECT id, invoice_number, customer_id, issue_date, due_date,
       total_amount, amount_paid, status, description,
       created_by, created_at, updated_at
FROM invoices
WHERE id = $1;

-- name: GetInvoiceByNumber :one
SELECT id, invoice_number, customer_id, issue_date, due_date,
       total_amount, amount_paid, status, description,
       created_by, created_at, updated_at
FROM invoices
WHERE invoice_number = $1;

-- name: ListInvoices :many
SELECT i.id, i.invoice_number, i.customer_id, i.issue_date, i.due_date,
       i.total_amount, i.amount_paid, i.status, i.description,
       i.created_by, i.created_at, i.updated_at,
       c.name AS customer_name
FROM invoices i
JOIN customers c ON c.id = i.customer_id
WHERE (sqlc.narg('customer_id')::UUID IS NULL OR i.customer_id = sqlc.narg('customer_id')::UUID)
  AND (sqlc.narg('status')::invoice_status IS NULL OR i.status = sqlc.narg('status')::invoice_status)
ORDER BY i.due_date ASC
LIMIT $1 OFFSET $2;

-- name: CountInvoices :one
SELECT COUNT(*) FROM invoices
WHERE (sqlc.narg('customer_id')::UUID IS NULL OR customer_id = sqlc.narg('customer_id')::UUID)
  AND (sqlc.narg('status')::invoice_status IS NULL OR status = sqlc.narg('status')::invoice_status);

-- name: ListUnpaidInvoicesByCustomer :many
SELECT id, invoice_number, customer_id, issue_date, due_date,
       total_amount, amount_paid, status, description,
       created_by, created_at, updated_at
FROM invoices
WHERE customer_id = $1
  AND status IN ('unpaid', 'partially_paid')
ORDER BY due_date ASC;

-- name: CreateInvoice :one
INSERT INTO invoices (invoice_number, customer_id, issue_date, due_date, total_amount, description, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, invoice_number, customer_id, issue_date, due_date,
          total_amount, amount_paid, status, description,
          created_by, created_at, updated_at;

-- name: UpdateInvoice :one
UPDATE invoices
SET invoice_number = $2, issue_date = $3, due_date = $4,
    total_amount = $5, description = $6, updated_at = NOW()
WHERE id = $1 AND status = 'unpaid'
RETURNING id, invoice_number, customer_id, issue_date, due_date,
          total_amount, amount_paid, status, description,
          created_by, created_at, updated_at;

-- name: UpdateInvoiceAmountPaid :exec
UPDATE invoices
SET amount_paid = amount_paid + @addition::NUMERIC(18,2),
    status = CASE
        WHEN amount_paid + @addition::NUMERIC(18,2) >= total_amount THEN 'paid'::invoice_status
        ELSE 'partially_paid'::invoice_status
    END,
    updated_at = NOW()
WHERE id = @id;

-- name: DeleteInvoice :exec
DELETE FROM invoices WHERE id = $1 AND status = 'unpaid';

-- name: GenerateInvoiceNumber :one
SELECT 'INV-' || to_char(CURRENT_DATE, 'YYYY') || '-' ||
       LPAD((COALESCE(
           (SELECT COUNT(*) + 1 FROM invoices
            WHERE invoice_number LIKE 'INV-' || to_char(CURRENT_DATE, 'YYYY') || '-%'),
           1
       ))::TEXT, 3, '0') AS invoice_number;

-- name: GetReceivablesSummary :one
SELECT
    COALESCE(SUM(total_amount - amount_paid), 0)::NUMERIC(18,2) AS total_outstanding,
    COALESCE(SUM(CASE WHEN due_date < CURRENT_DATE THEN total_amount - amount_paid ELSE 0 END), 0)::NUMERIC(18,2) AS total_overdue
FROM invoices
WHERE status IN ('unpaid', 'partially_paid');

-- name: ListInvoicesAging :many
SELECT i.id, i.invoice_number, i.customer_id, i.issue_date, i.due_date,
       i.total_amount, i.amount_paid, i.status, i.description,
       i.created_by, i.created_at, i.updated_at,
       c.name AS customer_name,
       (CURRENT_DATE - i.due_date) AS days_overdue
FROM invoices i
JOIN customers c ON c.id = i.customer_id
WHERE i.status IN ('unpaid', 'partially_paid')
ORDER BY i.due_date ASC
LIMIT $1 OFFSET $2;

-- name: CountInvoicesAging :one
SELECT COUNT(*) FROM invoices
WHERE status IN ('unpaid', 'partially_paid');
