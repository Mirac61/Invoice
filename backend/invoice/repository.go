package invoice

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("No Invoice found")

type Repository struct {
	invoices map[string]Invoice
	mu       sync.RWMutex
}

func NewRepository() *Repository {
	return &Repository{
		invoices: make(map[string]Invoice),
	}
}

func (r *Repository) Create(invoice Invoice) Invoice {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invoices[invoice.ID] = invoice
	return invoice
}

func (r *Repository) GetByID(id string) (Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	invoice, ok := r.invoices[id]
	if !ok {
		return Invoice{}, ErrNotFound
	}
	return invoice, nil
}

func (r *Repository) GetAll() []Invoice {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Invoice, 0, len(r.invoices))
	for _, invoice := range r.invoices {
		result = append(result, invoice)
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
