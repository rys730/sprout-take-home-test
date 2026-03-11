package usecase

import (
	"context"
	"fmt"
	"math"

	"sprout-backend/internal/domain"
)

type paymentUseCase struct {
	repo         domain.PaymentRepository
	invoiceRepo  domain.InvoiceRepository
	customerRepo domain.CustomerRepository
}

// NewPaymentUseCase creates a new PaymentUseCase.
func NewPaymentUseCase(
	repo domain.PaymentRepository,
	invoiceRepo domain.InvoiceRepository,
	customerRepo domain.CustomerRepository,
) domain.PaymentUseCase {
	return &paymentUseCase{
		repo:         repo,
		invoiceRepo:  invoiceRepo,
		customerRepo: customerRepo,
	}
}

func (uc *paymentUseCase) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	payment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return nil, fmt.Errorf("payment not found")
	}
	return payment, nil
}

func (uc *paymentUseCase) List(ctx context.Context, filter domain.PaymentFilter) ([]domain.Payment, int64, error) {
	return uc.repo.List(ctx, filter)
}

func (uc *paymentUseCase) RecordPayment(ctx context.Context, req domain.CreatePaymentRequest, createdBy string) (*domain.Payment, error) {
	// Validate required fields
	if req.CustomerID == "" {
		return nil, fmt.Errorf("customer_id is required")
	}
	if req.PaymentDate == "" {
		return nil, fmt.Errorf("payment_date is required")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	if req.DepositToAccountID == "" {
		return nil, fmt.Errorf("deposit_to_account_id is required")
	}
	if len(req.Allocations) == 0 {
		return nil, fmt.Errorf("at least one allocation is required")
	}

	// Validate customer exists
	customer, err := uc.customerRepo.GetByID(ctx, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("get customer: %w", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("customer not found")
	}

	// Validate each allocation
	var totalAllocated float64
	for i, alloc := range req.Allocations {
		if alloc.InvoiceID == "" {
			return nil, fmt.Errorf("allocation %d: invoice_id is required", i+1)
		}
		if alloc.Amount <= 0 {
			return nil, fmt.Errorf("allocation %d: amount must be greater than zero", i+1)
		}

		// Verify invoice exists and belongs to the customer
		invoice, err := uc.invoiceRepo.GetByID(ctx, alloc.InvoiceID)
		if err != nil {
			return nil, fmt.Errorf("allocation %d: get invoice: %w", i+1, err)
		}
		if invoice == nil {
			return nil, fmt.Errorf("allocation %d: invoice not found", i+1)
		}
		if invoice.CustomerID != req.CustomerID {
			return nil, fmt.Errorf("allocation %d: invoice does not belong to this customer", i+1)
		}
		if invoice.Status == domain.InvoiceStatusPaid {
			return nil, fmt.Errorf("allocation %d: invoice is already fully paid", i+1)
		}

		// Verify allocation amount does not exceed remaining balance
		remaining := invoice.TotalAmount - invoice.AmountPaid
		if alloc.Amount > remaining+0.001 {
			return nil, fmt.Errorf("allocation %d: amount (%.2f) exceeds invoice remaining balance (%.2f)", i+1, alloc.Amount, remaining)
		}

		totalAllocated += alloc.Amount
	}

	// Validate total allocations match payment amount
	if math.Abs(totalAllocated-req.Amount) > 0.01 {
		return nil, fmt.Errorf("total allocations (%.2f) must equal payment amount (%.2f)", totalAllocated, req.Amount)
	}

	// Delegate to repository for transactional execution
	return uc.repo.RecordPayment(ctx, req, createdBy)
}

func (uc *paymentUseCase) GetReceivablesSummary(ctx context.Context) (*domain.ReceivablesSummary, error) {
	return uc.invoiceRepo.GetReceivablesSummary(ctx)
}
