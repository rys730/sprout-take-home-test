package usecase

import (
	"context"
	"fmt"
	"testing"

	"sprout-backend/internal/domain"
)

// ---------------------------------------------------------------------------
// Mock Payment Repository
// ---------------------------------------------------------------------------

type mockPaymentRepo struct {
	payments   map[string]*domain.Payment
	nextNumber int
}

func newMockPaymentRepo() *mockPaymentRepo {
	return &mockPaymentRepo{
		payments:   make(map[string]*domain.Payment),
		nextNumber: 1,
	}
}

func (m *mockPaymentRepo) GetByID(_ context.Context, id string) (*domain.Payment, error) {
	p, ok := m.payments[id]
	if !ok {
		return nil, nil
	}
	return p, nil
}

func (m *mockPaymentRepo) List(_ context.Context, filter domain.PaymentFilter) ([]domain.Payment, int64, error) {
	var result []domain.Payment
	for _, p := range m.payments {
		if filter.CustomerID != "" && p.CustomerID != filter.CustomerID {
			continue
		}
		result = append(result, *p)
	}
	return result, int64(len(result)), nil
}

func (m *mockPaymentRepo) RecordPayment(_ context.Context, req domain.CreatePaymentRequest, createdBy string) (*domain.Payment, error) {
	id := fmt.Sprintf("pay-%d", m.nextNumber)
	m.nextNumber++

	var allocations []domain.PaymentAllocation
	for i, a := range req.Allocations {
		allocations = append(allocations, domain.PaymentAllocation{
			ID:        fmt.Sprintf("alloc-%d", i+1),
			PaymentID: id,
			InvoiceID: a.InvoiceID,
			Amount:    a.Amount,
		})
	}

	journalID := fmt.Sprintf("je-%d", m.nextNumber)
	payment := &domain.Payment{
		ID:                 id,
		PaymentNumber:      fmt.Sprintf("PAY-2026-%03d", m.nextNumber),
		CustomerID:         req.CustomerID,
		PaymentDate:        req.PaymentDate,
		Amount:             req.Amount,
		DepositToAccountID: req.DepositToAccountID,
		JournalEntryID:     &journalID,
		Notes:              req.Notes,
		Allocations:        allocations,
		CreatedBy:          &createdBy,
	}
	m.payments[id] = payment
	return payment, nil
}

// ---------------------------------------------------------------------------
// Mock Customer Repository
// ---------------------------------------------------------------------------

type mockCustomerRepo struct {
	customers map[string]*domain.Customer
}

func newMockCustomerRepo() *mockCustomerRepo {
	return &mockCustomerRepo{customers: make(map[string]*domain.Customer)}
}

func (m *mockCustomerRepo) GetByID(_ context.Context, id string) (*domain.Customer, error) {
	c, ok := m.customers[id]
	if !ok {
		return nil, nil
	}
	return c, nil
}

func (m *mockCustomerRepo) List(_ context.Context, _ domain.CustomerFilter) ([]domain.Customer, int64, error) {
	var result []domain.Customer
	for _, c := range m.customers {
		result = append(result, *c)
	}
	return result, int64(len(result)), nil
}

func (m *mockCustomerRepo) Create(_ context.Context, customer *domain.Customer) (*domain.Customer, error) {
	m.customers[customer.ID] = customer
	return customer, nil
}

func (m *mockCustomerRepo) Update(_ context.Context, customer *domain.Customer) (*domain.Customer, error) {
	m.customers[customer.ID] = customer
	return customer, nil
}

func (m *mockCustomerRepo) Delete(_ context.Context, id string) error {
	delete(m.customers, id)
	return nil
}

// ---------------------------------------------------------------------------
// Mock Invoice Repository
// ---------------------------------------------------------------------------

type mockInvoiceRepo struct {
	invoices map[string]*domain.Invoice
	summary  *domain.ReceivablesSummary
}

func newMockInvoiceRepo() *mockInvoiceRepo {
	return &mockInvoiceRepo{invoices: make(map[string]*domain.Invoice)}
}

func (m *mockInvoiceRepo) GetByID(_ context.Context, id string) (*domain.Invoice, error) {
	inv, ok := m.invoices[id]
	if !ok {
		return nil, nil
	}
	return inv, nil
}

func (m *mockInvoiceRepo) List(_ context.Context, _ domain.InvoiceFilter) ([]domain.Invoice, int64, error) {
	var result []domain.Invoice
	for _, inv := range m.invoices {
		result = append(result, *inv)
	}
	return result, int64(len(result)), nil
}

func (m *mockInvoiceRepo) ListUnpaidByCustomer(_ context.Context, customerID string) ([]domain.Invoice, error) {
	var result []domain.Invoice
	for _, inv := range m.invoices {
		if inv.CustomerID == customerID && inv.Status != domain.InvoiceStatusPaid {
			result = append(result, *inv)
		}
	}
	return result, nil
}

func (m *mockInvoiceRepo) ListAging(_ context.Context, _, _ int) ([]domain.Invoice, int64, error) {
	return nil, 0, nil
}

func (m *mockInvoiceRepo) GetReceivablesSummary(_ context.Context) (*domain.ReceivablesSummary, error) {
	if m.summary != nil {
		return m.summary, nil
	}
	// Calculate from invoices
	var totalOutstanding, totalOverdue float64
	for _, inv := range m.invoices {
		remaining := inv.TotalAmount - inv.AmountPaid
		if remaining > 0 {
			totalOutstanding += remaining
			if inv.DaysOverdue > 0 {
				totalOverdue += remaining
			}
		}
	}
	return &domain.ReceivablesSummary{
		TotalOutstanding: totalOutstanding,
		TotalOverdue:     totalOverdue,
	}, nil
}

func (m *mockInvoiceRepo) Create(_ context.Context, invoice *domain.Invoice) (*domain.Invoice, error) {
	m.invoices[invoice.ID] = invoice
	return invoice, nil
}

func (m *mockInvoiceRepo) Update(_ context.Context, invoice *domain.Invoice) (*domain.Invoice, error) {
	m.invoices[invoice.ID] = invoice
	return invoice, nil
}

func (m *mockInvoiceRepo) Delete(_ context.Context, id string) error {
	delete(m.invoices, id)
	return nil
}

func (m *mockInvoiceRepo) GenerateInvoiceNumber(_ context.Context) (string, error) {
	return fmt.Sprintf("INV-2026-%03d", len(m.invoices)+1), nil
}

// ---------------------------------------------------------------------------
// Payment Use Case Tests
// ---------------------------------------------------------------------------

func TestRecordPayment_Success(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()

	customerRepo.customers["cust-1"] = &domain.Customer{ID: "cust-1", Name: "PT Maju Jaya", IsActive: true}
	invoiceRepo.invoices["inv-1"] = &domain.Invoice{
		ID:          "inv-1",
		CustomerID:  "cust-1",
		TotalAmount: 1000000,
		AmountPaid:  0,
		Status:      domain.InvoiceStatusUnpaid,
	}

	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	req := domain.CreatePaymentRequest{
		CustomerID:         "cust-1",
		PaymentDate:        "2026-03-10",
		Amount:             1000000,
		DepositToAccountID: "bank-account-id",
		Allocations: []domain.PaymentAllocationLine{
			{InvoiceID: "inv-1", Amount: 1000000},
		},
	}

	payment, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if payment.Amount != 1000000 {
		t.Errorf("expected amount 1000000, got %.2f", payment.Amount)
	}
	if len(payment.Allocations) != 1 {
		t.Errorf("expected 1 allocation, got %d", len(payment.Allocations))
	}
	if payment.JournalEntryID == nil {
		t.Error("expected journal_entry_id to be set")
	}
}

func TestRecordPayment_MultipleInvoices(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()

	customerRepo.customers["cust-1"] = &domain.Customer{ID: "cust-1", Name: "PT Maju Jaya"}
	invoiceRepo.invoices["inv-1"] = &domain.Invoice{
		ID: "inv-1", CustomerID: "cust-1", TotalAmount: 500000, AmountPaid: 0, Status: domain.InvoiceStatusUnpaid,
	}
	invoiceRepo.invoices["inv-2"] = &domain.Invoice{
		ID: "inv-2", CustomerID: "cust-1", TotalAmount: 300000, AmountPaid: 0, Status: domain.InvoiceStatusUnpaid,
	}

	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	req := domain.CreatePaymentRequest{
		CustomerID:         "cust-1",
		PaymentDate:        "2026-03-10",
		Amount:             800000,
		DepositToAccountID: "bank-account-id",
		Allocations: []domain.PaymentAllocationLine{
			{InvoiceID: "inv-1", Amount: 500000},
			{InvoiceID: "inv-2", Amount: 300000},
		},
	}

	payment, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(payment.Allocations) != 2 {
		t.Errorf("expected 2 allocations, got %d", len(payment.Allocations))
	}
}

func TestRecordPayment_PartialPayment(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()

	customerRepo.customers["cust-1"] = &domain.Customer{ID: "cust-1", Name: "PT Maju Jaya"}
	invoiceRepo.invoices["inv-1"] = &domain.Invoice{
		ID: "inv-1", CustomerID: "cust-1", TotalAmount: 1000000, AmountPaid: 0, Status: domain.InvoiceStatusUnpaid,
	}

	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	req := domain.CreatePaymentRequest{
		CustomerID:         "cust-1",
		PaymentDate:        "2026-03-10",
		Amount:             400000,
		DepositToAccountID: "bank-account-id",
		Allocations: []domain.PaymentAllocationLine{
			{InvoiceID: "inv-1", Amount: 400000},
		},
	}

	payment, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if payment.Amount != 400000 {
		t.Errorf("expected amount 400000, got %.2f", payment.Amount)
	}
}

func TestRecordPayment_MissingCustomerID(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()
	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	req := domain.CreatePaymentRequest{
		PaymentDate:        "2026-03-10",
		Amount:             1000000,
		DepositToAccountID: "bank-account-id",
		Allocations: []domain.PaymentAllocationLine{
			{InvoiceID: "inv-1", Amount: 1000000},
		},
	}
	_, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for missing customer_id")
	}
}

func TestRecordPayment_MissingDepositAccount(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()
	customerRepo.customers["cust-1"] = &domain.Customer{ID: "cust-1", Name: "PT Maju Jaya"}
	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	req := domain.CreatePaymentRequest{
		CustomerID:  "cust-1",
		PaymentDate: "2026-03-10",
		Amount:      1000000,
		Allocations: []domain.PaymentAllocationLine{
			{InvoiceID: "inv-1", Amount: 1000000},
		},
	}
	_, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for missing deposit_to_account_id")
	}
}

func TestRecordPayment_ZeroAmount(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()
	customerRepo.customers["cust-1"] = &domain.Customer{ID: "cust-1", Name: "PT Maju Jaya"}
	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	req := domain.CreatePaymentRequest{
		CustomerID:         "cust-1",
		PaymentDate:        "2026-03-10",
		Amount:             0,
		DepositToAccountID: "bank-account-id",
		Allocations: []domain.PaymentAllocationLine{
			{InvoiceID: "inv-1", Amount: 0},
		},
	}
	_, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestRecordPayment_NoAllocations(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()
	customerRepo.customers["cust-1"] = &domain.Customer{ID: "cust-1", Name: "PT Maju Jaya"}
	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	req := domain.CreatePaymentRequest{
		CustomerID:         "cust-1",
		PaymentDate:        "2026-03-10",
		Amount:             1000000,
		DepositToAccountID: "bank-account-id",
		Allocations:        []domain.PaymentAllocationLine{},
	}
	_, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for no allocations")
	}
}

func TestRecordPayment_CustomerNotFound(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()
	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	req := domain.CreatePaymentRequest{
		CustomerID:         "nonexistent",
		PaymentDate:        "2026-03-10",
		Amount:             1000000,
		DepositToAccountID: "bank-account-id",
		Allocations: []domain.PaymentAllocationLine{
			{InvoiceID: "inv-1", Amount: 1000000},
		},
	}
	_, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for customer not found")
	}
}

func TestRecordPayment_InvoiceNotFound(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()
	customerRepo.customers["cust-1"] = &domain.Customer{ID: "cust-1", Name: "PT Maju Jaya"}
	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	req := domain.CreatePaymentRequest{
		CustomerID:         "cust-1",
		PaymentDate:        "2026-03-10",
		Amount:             1000000,
		DepositToAccountID: "bank-account-id",
		Allocations: []domain.PaymentAllocationLine{
			{InvoiceID: "nonexistent", Amount: 1000000},
		},
	}
	_, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for invoice not found")
	}
}

func TestRecordPayment_InvoiceWrongCustomer(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()
	customerRepo.customers["cust-1"] = &domain.Customer{ID: "cust-1", Name: "PT Maju Jaya"}
	invoiceRepo.invoices["inv-1"] = &domain.Invoice{
		ID: "inv-1", CustomerID: "cust-2", TotalAmount: 1000000, Status: domain.InvoiceStatusUnpaid,
	}
	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	req := domain.CreatePaymentRequest{
		CustomerID:         "cust-1",
		PaymentDate:        "2026-03-10",
		Amount:             1000000,
		DepositToAccountID: "bank-account-id",
		Allocations: []domain.PaymentAllocationLine{
			{InvoiceID: "inv-1", Amount: 1000000},
		},
	}
	_, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for invoice belonging to different customer")
	}
}

func TestRecordPayment_InvoiceAlreadyPaid(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()
	customerRepo.customers["cust-1"] = &domain.Customer{ID: "cust-1", Name: "PT Maju Jaya"}
	invoiceRepo.invoices["inv-1"] = &domain.Invoice{
		ID: "inv-1", CustomerID: "cust-1", TotalAmount: 1000000, AmountPaid: 1000000, Status: domain.InvoiceStatusPaid,
	}
	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	req := domain.CreatePaymentRequest{
		CustomerID:         "cust-1",
		PaymentDate:        "2026-03-10",
		Amount:             1000000,
		DepositToAccountID: "bank-account-id",
		Allocations: []domain.PaymentAllocationLine{
			{InvoiceID: "inv-1", Amount: 1000000},
		},
	}
	_, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for already paid invoice")
	}
}

func TestRecordPayment_AllocationExceedsRemaining(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()
	customerRepo.customers["cust-1"] = &domain.Customer{ID: "cust-1", Name: "PT Maju Jaya"}
	invoiceRepo.invoices["inv-1"] = &domain.Invoice{
		ID: "inv-1", CustomerID: "cust-1", TotalAmount: 1000000, AmountPaid: 800000, Status: domain.InvoiceStatusPartiallyPaid,
	}
	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	req := domain.CreatePaymentRequest{
		CustomerID:         "cust-1",
		PaymentDate:        "2026-03-10",
		Amount:             300000,
		DepositToAccountID: "bank-account-id",
		Allocations: []domain.PaymentAllocationLine{
			{InvoiceID: "inv-1", Amount: 300000}, // remaining is only 200000
		},
	}
	_, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for allocation exceeding remaining balance")
	}
}

func TestRecordPayment_AllocationsMismatchAmount(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()
	customerRepo.customers["cust-1"] = &domain.Customer{ID: "cust-1", Name: "PT Maju Jaya"}
	invoiceRepo.invoices["inv-1"] = &domain.Invoice{
		ID: "inv-1", CustomerID: "cust-1", TotalAmount: 1000000, AmountPaid: 0, Status: domain.InvoiceStatusUnpaid,
	}
	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	req := domain.CreatePaymentRequest{
		CustomerID:         "cust-1",
		PaymentDate:        "2026-03-10",
		Amount:             1000000,
		DepositToAccountID: "bank-account-id",
		Allocations: []domain.PaymentAllocationLine{
			{InvoiceID: "inv-1", Amount: 500000}, // Only allocating 500k of 1M payment
		},
	}
	_, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for allocations not matching payment amount")
	}
}

func TestRecordPayment_GetByID_Success(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()

	customerRepo.customers["cust-1"] = &domain.Customer{ID: "cust-1", Name: "PT Maju Jaya"}
	invoiceRepo.invoices["inv-1"] = &domain.Invoice{
		ID: "inv-1", CustomerID: "cust-1", TotalAmount: 500000, AmountPaid: 0, Status: domain.InvoiceStatusUnpaid,
	}

	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	// Record a payment first
	req := domain.CreatePaymentRequest{
		CustomerID:         "cust-1",
		PaymentDate:        "2026-03-10",
		Amount:             500000,
		DepositToAccountID: "bank-account-id",
		Allocations: []domain.PaymentAllocationLine{
			{InvoiceID: "inv-1", Amount: 500000},
		},
	}
	created, err := uc.RecordPayment(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Retrieve it
	payment, err := uc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if payment.ID != created.ID {
		t.Errorf("expected id '%s', got '%s'", created.ID, payment.ID)
	}
}

func TestRecordPayment_GetByID_NotFound(t *testing.T) {
	paymentRepo := newMockPaymentRepo()
	invoiceRepo := newMockInvoiceRepo()
	customerRepo := newMockCustomerRepo()
	uc := NewPaymentUseCase(paymentRepo, invoiceRepo, customerRepo)

	_, err := uc.GetByID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}
