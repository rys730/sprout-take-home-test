-- ============================================================================
-- PAYMENTS (Pembayaran)
-- ============================================================================

-- name: GetPaymentByID :one
SELECT p.id, p.payment_number, p.customer_id, p.payment_date, p.amount,
       p.deposit_to_account_id, p.journal_entry_id, p.notes,
       p.created_by, p.created_at, p.updated_at,
       c.name AS customer_name,
       a.code AS deposit_account_code, a.name AS deposit_account_name
FROM payments p
JOIN customers c ON c.id = p.customer_id
JOIN accounts a ON a.id = p.deposit_to_account_id
WHERE p.id = $1;

-- name: ListPayments :many
SELECT p.id, p.payment_number, p.customer_id, p.payment_date, p.amount,
       p.deposit_to_account_id, p.journal_entry_id, p.notes,
       p.created_by, p.created_at, p.updated_at,
       c.name AS customer_name,
       a.code AS deposit_account_code, a.name AS deposit_account_name
FROM payments p
JOIN customers c ON c.id = p.customer_id
JOIN accounts a ON a.id = p.deposit_to_account_id
WHERE (sqlc.narg('customer_id')::UUID IS NULL OR p.customer_id = sqlc.narg('customer_id')::UUID)
ORDER BY p.payment_date DESC
LIMIT $1 OFFSET $2;

-- name: CountPayments :one
SELECT COUNT(*) FROM payments
WHERE (sqlc.narg('customer_id')::UUID IS NULL OR customer_id = sqlc.narg('customer_id')::UUID);

-- name: CreatePayment :one
INSERT INTO payments (payment_number, customer_id, payment_date, amount, deposit_to_account_id, journal_entry_id, notes, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, payment_number, customer_id, payment_date, amount,
          deposit_to_account_id, journal_entry_id, notes,
          created_by, created_at, updated_at;

-- name: GeneratePaymentNumber :one
SELECT 'PAY-' || to_char(CURRENT_DATE, 'YYYY') || '-' ||
       LPAD((COALESCE(
           (SELECT COUNT(*) + 1 FROM payments
            WHERE payment_number LIKE 'PAY-' || to_char(CURRENT_DATE, 'YYYY') || '-%'),
           1
       ))::TEXT, 3, '0') AS payment_number;
