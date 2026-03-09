package usecase

import (
	"context"
	"fmt"
	"testing"

	"sprout-backend/internal/domain"
)

// ---------------------------------------------------------------------------
// Mock Journal Repository
// ---------------------------------------------------------------------------

type mockJournalRepo struct {
	entries       map[string]*domain.JournalEntry
	nextNumber    int
	generateError error
}

func newMockJournalRepo() *mockJournalRepo {
	return &mockJournalRepo{
		entries:    make(map[string]*domain.JournalEntry),
		nextNumber: 1,
	}
}

func (m *mockJournalRepo) List(_ context.Context, filter domain.JournalFilter) ([]domain.JournalEntry, int64, error) {
	var result []domain.JournalEntry
	for _, e := range m.entries {
		if filter.Status != "" && string(e.Status) != filter.Status {
			continue
		}
		result = append(result, *e)
	}
	return result, int64(len(result)), nil
}

func (m *mockJournalRepo) GetByID(_ context.Context, id string) (*domain.JournalEntry, error) {
	e, ok := m.entries[id]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockJournalRepo) Create(_ context.Context, entry *domain.JournalEntry) (*domain.JournalEntry, error) {
	entry.ID = fmt.Sprintf("je-%d", m.nextNumber)
	m.entries[entry.ID] = entry
	return entry, nil
}

func (m *mockJournalRepo) Update(_ context.Context, entry *domain.JournalEntry, _ bool) (*domain.JournalEntry, error) {
	m.entries[entry.ID] = entry
	return entry, nil
}

func (m *mockJournalRepo) Post(_ context.Context, id string) (*domain.JournalEntry, error) {
	e, ok := m.entries[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	e.Status = domain.JournalStatusPosted
	return e, nil
}

func (m *mockJournalRepo) Delete(_ context.Context, id string) error {
	delete(m.entries, id)
	return nil
}

func (m *mockJournalRepo) Reverse(_ context.Context, id string, reason string, createdBy string) (*domain.JournalEntry, error) {
	original, ok := m.entries[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	original.Status = domain.JournalStatusReversed

	m.nextNumber++
	reversalID := fmt.Sprintf("je-%d", m.nextNumber)
	reversal := &domain.JournalEntry{
		ID:             reversalID,
		EntryNumber:    fmt.Sprintf("REV-JU-2026-%03d", m.nextNumber),
		Date:           original.Date,
		Description:    fmt.Sprintf("Reversal of %s: %s", original.EntryNumber, reason),
		Status:         domain.JournalStatusPosted,
		ReversalOf:     &id,
		ReversalReason: &reason,
		Source:         "manual",
		CreatedBy:      &createdBy,
	}
	// Swap lines
	for _, l := range original.Lines {
		reversal.Lines = append(reversal.Lines, domain.JournalLine{
			AccountID:   l.AccountID,
			Description: "Reversal: " + l.Description,
			Debit:       l.Credit,
			Credit:      l.Debit,
		})
	}

	reversalIDStr := reversal.ID
	original.ReversedBy = &reversalIDStr
	m.entries[reversalID] = reversal

	return reversal, nil
}

func (m *mockJournalRepo) GenerateEntryNumber(_ context.Context) (string, error) {
	if m.generateError != nil {
		return "", m.generateError
	}
	num := fmt.Sprintf("JU-2026-%03d", m.nextNumber)
	m.nextNumber++
	return num, nil
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func validCreateRequest() domain.CreateJournalRequest {
	return domain.CreateJournalRequest{
		Date:        "2026-03-09",
		Description: "Test journal entry",
		Lines: []domain.CreateJournalLine{
			{AccountID: "acct-1", Debit: 100000, Credit: 0},
			{AccountID: "acct-2", Debit: 0, Credit: 100000},
		},
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCreateJournal_Success_Draft(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := validCreateRequest()
	entry, err := uc.Create(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entry.Status != domain.JournalStatusDraft {
		t.Errorf("expected draft, got %s", entry.Status)
	}
	if entry.EntryNumber == "" {
		t.Error("expected entry number to be set")
	}
	if len(entry.Lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(entry.Lines))
	}
}

func TestCreateJournal_Success_PostedDirectly(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := validCreateRequest()
	req.Status = "posted"
	entry, err := uc.Create(context.Background(), req, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entry.Status != domain.JournalStatusPosted {
		t.Errorf("expected posted, got %s", entry.Status)
	}
}

func TestCreateJournal_TooFewLines(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := domain.CreateJournalRequest{
		Date:        "2026-03-09",
		Description: "Bad entry",
		Lines: []domain.CreateJournalLine{
			{AccountID: "acct-1", Debit: 100, Credit: 0},
		},
	}
	_, err := uc.Create(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for too few lines")
	}
}

func TestCreateJournal_Unbalanced(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := domain.CreateJournalRequest{
		Date:        "2026-03-09",
		Description: "Unbalanced entry",
		Lines: []domain.CreateJournalLine{
			{AccountID: "acct-1", Debit: 100, Credit: 0},
			{AccountID: "acct-2", Debit: 0, Credit: 200},
		},
	}
	_, err := uc.Create(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for unbalanced entry")
	}
}

func TestCreateJournal_BothDebitAndCredit(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := domain.CreateJournalRequest{
		Date:        "2026-03-09",
		Description: "Bad line",
		Lines: []domain.CreateJournalLine{
			{AccountID: "acct-1", Debit: 100, Credit: 50},
			{AccountID: "acct-2", Debit: 0, Credit: 50},
		},
	}
	_, err := uc.Create(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for line with both debit and credit")
	}
}

func TestCreateJournal_ZeroAmountLine(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := domain.CreateJournalRequest{
		Date:        "2026-03-09",
		Description: "Zero line",
		Lines: []domain.CreateJournalLine{
			{AccountID: "acct-1", Debit: 100, Credit: 0},
			{AccountID: "acct-2", Debit: 0, Credit: 0},
		},
	}
	_, err := uc.Create(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for zero amount line")
	}
}

func TestCreateJournal_NegativeAmount(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := domain.CreateJournalRequest{
		Date:        "2026-03-09",
		Description: "Negative",
		Lines: []domain.CreateJournalLine{
			{AccountID: "acct-1", Debit: -100, Credit: 0},
			{AccountID: "acct-2", Debit: 0, Credit: -100},
		},
	}
	_, err := uc.Create(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestCreateJournal_MissingAccountID(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := domain.CreateJournalRequest{
		Date:        "2026-03-09",
		Description: "Missing account",
		Lines: []domain.CreateJournalLine{
			{AccountID: "", Debit: 100, Credit: 0},
			{AccountID: "acct-2", Debit: 0, Credit: 100},
		},
	}
	_, err := uc.Create(context.Background(), req, "user-1")
	if err == nil {
		t.Fatal("expected error for missing account_id")
	}
}

func TestUpdateJournal_Success(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	// Seed a draft entry
	req := validCreateRequest()
	entry, _ := uc.Create(context.Background(), req, "user-1")

	newDesc := "Updated description"
	updated, err := uc.Update(context.Background(), entry.ID, domain.UpdateJournalRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Description != newDesc {
		t.Errorf("expected description %q, got %q", newDesc, updated.Description)
	}
}

func TestUpdateJournal_NotDraft(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	// Seed a posted entry
	req := validCreateRequest()
	req.Status = "posted"
	entry, _ := uc.Create(context.Background(), req, "user-1")

	newDesc := "Try update"
	_, err := uc.Update(context.Background(), entry.ID, domain.UpdateJournalRequest{
		Description: &newDesc,
	})
	if err == nil {
		t.Fatal("expected error updating a posted entry")
	}
}

func TestUpdateJournal_ReplaceLines(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := validCreateRequest()
	entry, _ := uc.Create(context.Background(), req, "user-1")

	updated, err := uc.Update(context.Background(), entry.ID, domain.UpdateJournalRequest{
		Lines: []domain.CreateJournalLine{
			{AccountID: "acct-3", Debit: 200000, Credit: 0},
			{AccountID: "acct-4", Debit: 0, Credit: 200000},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(updated.Lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(updated.Lines))
	}
}

func TestUpdateJournal_UnbalancedLines(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := validCreateRequest()
	entry, _ := uc.Create(context.Background(), req, "user-1")

	_, err := uc.Update(context.Background(), entry.ID, domain.UpdateJournalRequest{
		Lines: []domain.CreateJournalLine{
			{AccountID: "acct-3", Debit: 200, Credit: 0},
			{AccountID: "acct-4", Debit: 0, Credit: 100},
		},
	})
	if err == nil {
		t.Fatal("expected error for unbalanced lines on update")
	}
}

func TestPostJournal_Success(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := validCreateRequest()
	entry, _ := uc.Create(context.Background(), req, "user-1")

	posted, err := uc.Post(context.Background(), entry.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if posted.Status != domain.JournalStatusPosted {
		t.Errorf("expected posted, got %s", posted.Status)
	}
}

func TestPostJournal_AlreadyPosted(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := validCreateRequest()
	req.Status = "posted"
	entry, _ := uc.Create(context.Background(), req, "user-1")

	_, err := uc.Post(context.Background(), entry.ID)
	if err == nil {
		t.Fatal("expected error posting an already posted entry")
	}
}

func TestPostJournal_NotFound(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	_, err := uc.Post(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent entry")
	}
}

func TestDeleteJournal_Success(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := validCreateRequest()
	entry, _ := uc.Create(context.Background(), req, "user-1")

	err := uc.Delete(context.Background(), entry.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify deleted
	_, err = uc.GetByID(context.Background(), entry.ID)
	if err == nil {
		t.Fatal("expected error for deleted entry")
	}
}

func TestDeleteJournal_NotDraft(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := validCreateRequest()
	req.Status = "posted"
	entry, _ := uc.Create(context.Background(), req, "user-1")

	err := uc.Delete(context.Background(), entry.ID)
	if err == nil {
		t.Fatal("expected error deleting a posted entry")
	}
}

func TestReverseJournal_Success(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := validCreateRequest()
	req.Status = "posted"
	entry, _ := uc.Create(context.Background(), req, "user-1")

	reversal, err := uc.Reverse(context.Background(), entry.ID, domain.ReverseJournalRequest{
		Reason: "Incorrect entry",
	}, "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if reversal.Status != domain.JournalStatusPosted {
		t.Errorf("expected reversal to be posted, got %s", reversal.Status)
	}
	if reversal.ReversalOf == nil || *reversal.ReversalOf != entry.ID {
		t.Error("expected reversal to reference original entry")
	}

	// Check original is now reversed
	original, _ := uc.GetByID(context.Background(), entry.ID)
	if original.Status != domain.JournalStatusReversed {
		t.Errorf("expected original to be reversed, got %s", original.Status)
	}

	// Verify lines are swapped
	if len(reversal.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(reversal.Lines))
	}
	// Original: line 0 debit=100000, credit=0 → reversal: debit=0, credit=100000
	if reversal.Lines[0].Debit != 0 || reversal.Lines[0].Credit != 100000 {
		t.Errorf("expected reversed line 0: debit=0 credit=100000, got debit=%.2f credit=%.2f",
			reversal.Lines[0].Debit, reversal.Lines[0].Credit)
	}
}

func TestReverseJournal_NotPosted(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := validCreateRequest()
	entry, _ := uc.Create(context.Background(), req, "user-1")

	_, err := uc.Reverse(context.Background(), entry.ID, domain.ReverseJournalRequest{
		Reason: "Test",
	}, "user-1")
	if err == nil {
		t.Fatal("expected error reversing a draft entry")
	}
}

func TestReverseJournal_EmptyReason(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	req := validCreateRequest()
	req.Status = "posted"
	entry, _ := uc.Create(context.Background(), req, "user-1")

	_, err := uc.Reverse(context.Background(), entry.ID, domain.ReverseJournalRequest{
		Reason: "",
	}, "user-1")
	if err == nil {
		t.Fatal("expected error for empty reason")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	_, err := uc.GetByID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent entry")
	}
}

func TestListJournals_FilterByStatus(t *testing.T) {
	repo := newMockJournalRepo()
	uc := NewJournalUseCase(repo)

	// Create a draft and a posted entry
	req := validCreateRequest()
	uc.Create(context.Background(), req, "user-1")
	req.Status = "posted"
	uc.Create(context.Background(), req, "user-1")

	// Filter posted only
	entries, total, err := uc.List(context.Background(), domain.JournalFilter{Status: "posted"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 posted entry, got %d", total)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}
