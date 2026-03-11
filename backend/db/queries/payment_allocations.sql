-- ============================================================================
-- PAYMENT ALLOCATIONS (Alokasi Pembayaran ke Invoice)
-- ============================================================================

-- name: CreatePaymentAllocation :one
INSERT INTO payment_allocations (payment_id, invoice_id, amount)
VALUES ($1, $2, $3)
RETURNING id, payment_id, invoice_id, amount, created_at;

-- name: GetPaymentAllocationsByPaymentID :many
SELECT pa.id, pa.payment_id, pa.invoice_id, pa.amount, pa.created_at,
       i.invoice_number
FROM payment_allocations pa
JOIN invoices i ON i.id = pa.invoice_id
WHERE pa.payment_id = $1
ORDER BY pa.created_at ASC;

-- name: GetPaymentAllocationsByInvoiceID :many
SELECT pa.id, pa.payment_id, pa.invoice_id, pa.amount, pa.created_at,
       p.payment_number
FROM payment_allocations pa
JOIN payments p ON p.id = pa.payment_id
WHERE pa.invoice_id = $1
ORDER BY pa.created_at ASC;
