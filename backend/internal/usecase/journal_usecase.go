package usecase

import (
	"context"
	"fmt"
	"math"

	"sprout-backend/internal/domain"
)

type journalUseCase struct {
	repo domain.JournalRepository
}

// NewJournalUseCase creates a new JournalUseCase.
func NewJournalUseCase(repo domain.JournalRepository) domain.JournalUseCase {
	return &journalUseCase{repo: repo}
}

func (uc *journalUseCase) List(ctx context.Context, filter domain.JournalFilter) ([]domain.JournalEntry, int64, error) {
	return uc.repo.List(ctx, filter)
}

func (uc *journalUseCase) GetByID(ctx context.Context, id string) (*domain.JournalEntry, error) {
	entry, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, fmt.Errorf("journal entry not found")
	}
	return entry, nil
}

func (uc *journalUseCase) Create(ctx context.Context, req domain.CreateJournalRequest, createdBy string) (*domain.JournalEntry, error) {
	// Validate: at least 2 lines
	if len(req.Lines) < 2 {
		return nil, fmt.Errorf("journal entry must have at least 2 lines")
	}

	// Validate: each line must have either debit > 0 or credit > 0, not both
	var totalDebit, totalCredit float64
	for i, line := range req.Lines {
		if line.Debit < 0 || line.Credit < 0 {
			return nil, fmt.Errorf("line %d: amounts must be non-negative", i+1)
		}
		if line.Debit > 0 && line.Credit > 0 {
			return nil, fmt.Errorf("line %d: a line cannot have both debit and credit", i+1)
		}
		if line.Debit == 0 && line.Credit == 0 {
			return nil, fmt.Errorf("line %d: debit or credit must be greater than zero", i+1)
		}
		if line.AccountID == "" {
			return nil, fmt.Errorf("line %d: account_id is required", i+1)
		}
		totalDebit += line.Debit
		totalCredit += line.Credit
	}

	// Validate: total debit == total credit
	if math.Abs(totalDebit-totalCredit) > 0.001 {
		return nil, fmt.Errorf("journal entry is unbalanced: total debit (%.2f) != total credit (%.2f)", totalDebit, totalCredit)
	}

	// Determine status
	status := domain.JournalStatusDraft
	if req.Status == "posted" {
		status = domain.JournalStatusPosted
	}

	// Generate entry number
	entryNumber, err := uc.repo.GenerateEntryNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate entry number: %w", err)
	}

	// Build domain entry
	entry := &domain.JournalEntry{
		EntryNumber: entryNumber,
		Date:        req.Date,
		Description: req.Description,
		Status:      status,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		CreatedBy:   &createdBy,
	}
	for _, l := range req.Lines {
		entry.Lines = append(entry.Lines, domain.JournalLine{
			AccountID:   l.AccountID,
			Debit:       l.Debit,
			Credit:      l.Credit,
		})
	}

	return uc.repo.Create(ctx, entry)
}

func (uc *journalUseCase) Update(ctx context.Context, id string, req domain.UpdateJournalRequest) (*domain.JournalEntry, error) {
	// Fetch existing
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("journal entry not found")
	}
	if existing.Status != domain.JournalStatusDraft {
		return nil, fmt.Errorf("only draft entries can be edited")
	}

	// Apply updates
	if req.Date != nil {
		existing.Date = *req.Date
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}

	replaceLines := false
	if len(req.Lines) > 0 {
		// Validate new lines
		if len(req.Lines) < 2 {
			return nil, fmt.Errorf("journal entry must have at least 2 lines")
		}

		var totalDebit, totalCredit float64
		for i, line := range req.Lines {
			if line.Debit < 0 || line.Credit < 0 {
				return nil, fmt.Errorf("line %d: amounts must be non-negative", i+1)
			}
			if line.Debit > 0 && line.Credit > 0 {
				return nil, fmt.Errorf("line %d: a line cannot have both debit and credit", i+1)
			}
			if line.Debit == 0 && line.Credit == 0 {
				return nil, fmt.Errorf("line %d: debit or credit must be greater than zero", i+1)
			}
			if line.AccountID == "" {
				return nil, fmt.Errorf("line %d: account_id is required", i+1)
			}
			totalDebit += line.Debit
			totalCredit += line.Credit
		}

		if math.Abs(totalDebit-totalCredit) > 0.001 {
			return nil, fmt.Errorf("journal entry is unbalanced: total debit (%.2f) != total credit (%.2f)", totalDebit, totalCredit)
		}

		existing.Lines = nil
		for _, l := range req.Lines {
			existing.Lines = append(existing.Lines, domain.JournalLine{
				AccountID:   l.AccountID,
				Debit:       l.Debit,
				Credit:      l.Credit,
			})
		}
		replaceLines = true
	}

	return uc.repo.Update(ctx, existing, replaceLines)
}

func (uc *journalUseCase) Post(ctx context.Context, id string) (*domain.JournalEntry, error) {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("journal entry not found")
	}
	if existing.Status != domain.JournalStatusDraft {
		return nil, fmt.Errorf("only draft entries can be posted")
	}
	if len(existing.Lines) < 2 {
		return nil, fmt.Errorf("journal entry must have at least 2 lines to be posted")
	}

	// Verify balance
	var totalDebit, totalCredit float64
	for _, line := range existing.Lines {
		totalDebit += line.Debit
		totalCredit += line.Credit
	}
	if math.Abs(totalDebit-totalCredit) > 0.001 {
		return nil, fmt.Errorf("journal entry is unbalanced: total debit (%.2f) != total credit (%.2f)", totalDebit, totalCredit)
	}

	return uc.repo.Post(ctx, id)
}

func (uc *journalUseCase) Delete(ctx context.Context, id string) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("journal entry not found")
	}
	if existing.Status != domain.JournalStatusDraft {
		return fmt.Errorf("only draft entries can be deleted")
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *journalUseCase) Reverse(ctx context.Context, id string, req domain.ReverseJournalRequest, createdBy string) (*domain.JournalEntry, error) {
	if req.Reason == "" {
		return nil, fmt.Errorf("reversal reason is required")
	}

	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("journal entry not found")
	}
	if existing.Status != domain.JournalStatusPosted {
		return nil, fmt.Errorf("only posted entries can be reversed")
	}

	return uc.repo.Reverse(ctx, id, req.Reason, createdBy)
}
