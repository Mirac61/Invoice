package invoice

import (
	"errors"
	"sync"
)

var (
	ErrNotFound     = errors.New("invoice not found")
	ErrNotDeletable = errors.New("invoice isn't deletable")
	ErrNotUpdatable = errors.New("invoice not updatable")
)

type Repository struct {
	invoices map[string]Invoice
	mu       sync.RWMutex
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

func (r *Repository) Create(invoice Invoice) Invoice {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := cloneInvoice(invoice)
	r.invoices[stored.ID] = stored
	return cloneInvoice(stored)
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

func (r *Repository) GetAll() []Invoice {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Invoice, 0, len(r.invoices))
	for _, invoice := range r.invoices {
		result = append(result, cloneInvoice(invoice))
	}
	return result
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

// UpdateFunc verändert eine Rechnung. Sie läuft unter dem Repository-Lock,
// damit Lesen, Ändern und Schreiben atomar sind.
type UpdateFunc func(existing Invoice) (Invoice, error)

func (r *Repository) Update(id string, fn UpdateFunc) (Invoice, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.invoices[id]
	if !ok {
		return Invoice{}, ErrNotFound
	}
	updated, err := fn(cloneInvoice(existing))
	if err != nil {
		return Invoice{}, err
	}
	updated.ID = existing.ID
	stored := cloneInvoice(updated)
	r.invoices[id] = stored
	return cloneInvoice(stored), nil
}
