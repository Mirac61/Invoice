package invoice

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService() *Service {
	return NewService(NewRepository())
}

func seedDraftInvoice(s *Service) Invoice {
	invoice := Invoice{
		Status:       StatusDraft,
		PaymentDueAt: time.Now().Add(14 * 24 * time.Hour),
		Sender:       Contact{Name: "Sender GmbH", Street: "Hauptstr. 1", Zip: "70173", City: "Stuttgart", Country: "DE"},
		Recipient:    Contact{Name: "Recipient GmbH", Street: "Nebenstr. 2", Zip: "70174", City: "Stuttgart", Country: "DE"},
		Items: []LineItem{
			{Description: "Beratung", Quantity: 2, UnitPrice: 100},
		},
		VATRate: 0.19,
	}
	return s.Create(invoice)
}

func TestPartialUpdate_NotesOnly_LeavesOtherFieldsUnchanged(t *testing.T) {
	s := newTestService()
	created := seedDraftInvoice(s)

	newNotes := "Bitte bis Ende des Monats zahlen"
	updated, err := s.PartialUpdate(created.ID, InvoicePatch{Notes: &newNotes})

	require.NoError(t, err)
	assert.Equal(t, newNotes, updated.Notes)
	assert.Equal(t, created.Recipient, updated.Recipient)
	assert.Equal(t, created.Items, updated.Items)
	assert.Equal(t, created.GrossTotal, updated.GrossTotal)
}

func TestPartialUpdate_RecalculatesTotals(t *testing.T) {
	tests := []struct {
		name      string
		items     []LineItem
		vatRate   float64
		wantNet   float64
		wantVAT   float64
		wantGross float64
	}{
		{
			name:      "standard VAT rate",
			items:     []LineItem{{Description: "Beratung", Quantity: 3, UnitPrice: 150}},
			vatRate:   0.19,
			wantNet:   450,
			wantVAT:   85.5,
			wantGross: 535.5,
		},
		{
			name:      "reduced VAT rate",
			items:     []LineItem{{Description: "Buch", Quantity: 1, UnitPrice: 20}},
			vatRate:   0.07,
			wantNet:   20,
			wantVAT:   1.4,
			wantGross: 21.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestService()
			created := seedDraftInvoice(s)

			patch := InvoicePatch{Items: &tt.items, VATRate: &tt.vatRate}
			updated, err := s.PartialUpdate(created.ID, patch)

			require.NoError(t, err)
			assert.Equal(t, tt.wantNet, updated.NetTotal)
			assert.Equal(t, tt.wantVAT, updated.VATAmount)
			assert.Equal(t, tt.wantGross, updated.GrossTotal)
		})
	}
}

func TestPartialUpdate_UnknownID_ReturnsNotFound(t *testing.T) {
	s := newTestService()

	notes := "egal"
	_, err := s.PartialUpdate("does-not-exist", InvoicePatch{Notes: &notes})

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestUpdate_PreservesServerManagedFields(t *testing.T) {
	s := newTestService()
	created := seedDraftInvoice(s)

	tampered := created
	tampered.Status = StatusPaid
	tampered.InvoiceNumber = "HACKED-001"
	tampered.CreatedAt = time.Time{}

	updated, err := s.Update(created.ID, tampered)

	require.NoError(t, err)
	assert.Equal(t, created.Status, updated.Status)
	assert.Equal(t, created.InvoiceNumber, updated.InvoiceNumber)
	assert.Equal(t, created.CreatedAt, updated.CreatedAt)
}

func seedIssuedInvoice(repo *Repository) Invoice {
	return repo.Create(Invoice{
		ID:      "issued-1",
		Status:  StatusIssued,
		VATRate: 0.19,
		Items:   []LineItem{{Description: "Beratung", Quantity: 1, UnitPrice: 100}},
	})
}

func TestUpdate_NonDraft_ReturnsNotUpdatable(t *testing.T) {
	repo := NewRepository()
	s := NewService(repo)
	issued := seedIssuedInvoice(repo)

	_, err := s.Update(issued.ID, issued)

	assert.ErrorIs(t, err, ErrNotUpdatable)
}

func TestPartialUpdate_NonDraft_ReturnsNotUpdatable(t *testing.T) {
	repo := NewRepository()
	s := NewService(repo)
	issued := seedIssuedInvoice(repo)

	notes := "egal"
	_, err := s.PartialUpdate(issued.ID, InvoicePatch{Notes: &notes})

	assert.ErrorIs(t, err, ErrNotUpdatable)
}

func TestDelete_NonDraft_ReturnsNotDeletable(t *testing.T) {
	repo := NewRepository()
	s := NewService(repo)
	issued := seedIssuedInvoice(repo)

	err := s.Delete(issued.ID)

	assert.ErrorIs(t, err, ErrNotDeletable)
}

func TestDelete_Draft_Succeeds(t *testing.T) {
	s := newTestService()
	created := seedDraftInvoice(s)

	err := s.Delete(created.ID)
	require.NoError(t, err)

	_, getErr := s.GetByID(created.ID)
	assert.ErrorIs(t, getErr, ErrNotFound)
}

func TestPartialUpdate_InvalidData_ReturnsInvalidInput(t *testing.T) {
	s := newTestService()
	created := seedDraftInvoice(s)

	items := []LineItem{{Description: "X", Quantity: -1, UnitPrice: 10}}
	_, err := s.PartialUpdate(created.ID, InvoicePatch{Items: &items})

	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestUpdate_InvalidData_ReturnsInvalidInput(t *testing.T) {
	s := newTestService()
	created := seedDraftInvoice(s)

	replacement := created
	replacement.VATRate = 1.5

	_, err := s.Update(created.ID, replacement)

	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestIssue_Draft_SetsNumberAndTimestamp(t *testing.T) {
	s := newTestService()
	created := seedDraftInvoice(s)

	issued, err := s.Issue(created.ID)

	require.NoError(t, err)
	assert.Equal(t, StatusIssued, issued.Status)
	assert.False(t, issued.IssuedAt.IsZero())
	assert.Regexp(t, `^\d{4}-\d{4}$`, issued.InvoiceNumber)
}

func TestIssue_AssignsSequentialNumbers(t *testing.T) {
	s := newTestService()
	first := seedDraftInvoice(s)
	second := seedDraftInvoice(s)

	a, err := s.Issue(first.ID)
	require.NoError(t, err)
	b, err := s.Issue(second.ID)
	require.NoError(t, err)

	year := time.Now().Year()
	assert.Equal(t, fmt.Sprintf("%d-0001", year), a.InvoiceNumber)
	assert.Equal(t, fmt.Sprintf("%d-0002", year), b.InvoiceNumber)
}

func TestIssue_AlreadyIssued_ReturnsInvalidTransition(t *testing.T) {
	s := newTestService()
	created := seedDraftInvoice(s)

	_, err := s.Issue(created.ID)
	require.NoError(t, err)

	_, err = s.Issue(created.ID)
	assert.ErrorIs(t, err, ErrInvalidTransition)
}

func TestIssue_UnknownID_ReturnsNotFound(t *testing.T) {
	s := newTestService()

	_, err := s.Issue("does-not-exist")

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestIssue_ThenUpdate_ReturnsNotUpdatable(t *testing.T) {
	s := newTestService()
	created := seedDraftInvoice(s)

	issued, err := s.Issue(created.ID)
	require.NoError(t, err)

	notes := "zu spät"
	_, err = s.PartialUpdate(issued.ID, InvoicePatch{Notes: &notes})

	assert.ErrorIs(t, err, ErrNotUpdatable)
}
