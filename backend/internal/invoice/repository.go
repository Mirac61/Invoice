package invoice

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrNotFound          = errors.New("invoice not found")
	ErrNotDeletable      = errors.New("invoice isn't deletable")
	ErrNotUpdatable      = errors.New("invoice not updatable")
	ErrInvalidInput      = errors.New("invalid invoice data")
	ErrInvalidTransition = errors.New("invalid status transition")
)

type Repository struct {
	invoices    map[string]Invoice
	mu          sync.RWMutex
	counterMu   sync.Mutex
	counterYear int
	counter     int
}

func NewRepository() *Repository {
	return &Repository{
		invoices: make(map[string]Invoice),
	}
}

func cloneInvoice(invoice Invoice) Invoice {
	invoice.Items = append([]LineItem(nil), invoice.Items...)
	return invoice
}

func (r *Repository) nextInvoiceNumber(now time.Time) (string, error) {
	r.counterMu.Lock()
	defer r.counterMu.Unlock()

	year := now.Year()
	if year != r.counterYear {
		r.counterYear = year
		r.counter = 0
	}
	r.counter++
	return fmt.Sprintf("%d-%04d", year, r.counter), nil
}

func (r *Repository) Create(invoice Invoice) (Invoice, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	stored := cloneInvoice(invoice)
	r.invoices[stored.ID] = stored
	return cloneInvoice(stored), nil
}

func (r *Repository) GetByID(id string) (Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	invoice, ok := r.invoices[id]
	if !ok {
		return Invoice{}, ErrNotFound
	}
	return cloneInvoice(invoice), nil
}

func (r *Repository) GetAll() ([]Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Invoice, 0, len(r.invoices))
	for _, invoice := range r.invoices {
		result = append(result, cloneInvoice(invoice))
	}
	return result, nil
}

func (r *Repository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.invoices[id]; !ok {
		return ErrNotFound
	}
	delete(r.invoices, id)
	return nil
}

// UpdateFunc mutates an invoice. It runs inside the repository transaction so
// that read, modify and write happen atomically. nextNumber draws the next
// invoice number from the same transaction and must only be called when the
// number is actually used.
type UpdateFunc func(existing Invoice, nextNumber func() (string, error)) (Invoice, error)

func (r *Repository) Update(id string, fn UpdateFunc) (Invoice, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.invoices[id]
	if !ok {
		return Invoice{}, ErrNotFound
	}

	nextNumber := func() (string, error) {
		return r.nextInvoiceNumber(time.Now())
	}

	updated, err := fn(cloneInvoice(existing), nextNumber)
	if err != nil {
		return Invoice{}, err
	}
	updated.ID = existing.ID

	stored := cloneInvoice(updated)
	r.invoices[id] = stored
	return cloneInvoice(stored), nil
}
