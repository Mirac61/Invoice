package invoice

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type invoiceRepository interface {
	Create(invoice Invoice) Invoice
	GetByID(id string) (Invoice, error)
	GetAll() []Invoice
	Delete(id string) error
	Update(id string, fn UpdateFunc) (Invoice, error)
	NextInvoiceNumber(now time.Time) string
}

type Service struct {
	repo invoiceRepository
}

func NewService(repo invoiceRepository) *Service {
	return &Service{repo: repo}
}

func calculateTotals(items []LineItem, vatRate float64) (net, vat, gross float64) {
	for i := range items {
		items[i].Total = math.Round(float64(items[i].Quantity)*items[i].UnitPrice*100) / 100
		net += items[i].Total
	}
	net = math.Round(net*100) / 100
	vat = math.Round(net*vatRate*100) / 100
	gross = math.Round((net+vat)*100) / 100
	return
}

func validateInvoiceData(items []LineItem, vatRate float64) error {
	if vatRate < 0 || vatRate > 1 {
		return ErrInvalidInput
	}
	if len(items) == 0 {
		return ErrInvalidInput
	}
	for _, item := range items {
		if item.Description == "" || item.Quantity <= 0 || item.UnitPrice < 0 {
			return ErrInvalidInput
		}
	}
	return nil
}

func (s *Service) Create(invoice Invoice) Invoice {
	invoice.ID = uuid.NewString()
	invoice.CreatedAt = time.Now()
	invoice.Status = StatusDraft

	invoice.NetTotal, invoice.VATAmount, invoice.GrossTotal = calculateTotals(invoice.Items, invoice.VATRate)
	return s.repo.Create(invoice)
}

func (s *Service) GetByID(id string) (Invoice, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetAll() []Invoice {
	return s.repo.GetAll()
}

func (s *Service) Delete(id string) error {
	invoice, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if invoice.Status != StatusDraft {
		return ErrNotDeletable
	}
	return s.repo.Delete(id)
}

func (s *Service) Update(id string, replacement Invoice) (Invoice, error) {
	mutate := func(invoice Invoice) (Invoice, error) {
		if invoice.Status != StatusDraft {
			return Invoice{}, ErrNotUpdatable
		}
		replacement.ID = invoice.ID
		replacement.InvoiceNumber = invoice.InvoiceNumber
		replacement.Status = invoice.Status
		replacement.CreatedAt = invoice.CreatedAt
		replacement.IssuedAt = invoice.IssuedAt
		if err := validateInvoiceData(replacement.Items, replacement.VATRate); err != nil {
			return Invoice{}, err
		}
		replacement.NetTotal, replacement.VATAmount, replacement.GrossTotal = calculateTotals(replacement.Items, replacement.VATRate)
		return replacement, nil
	}

	return s.repo.Update(id, mutate)
}

func (s *Service) PartialUpdate(id string, patch InvoicePatch) (Invoice, error) {
	mutate := func(invoice Invoice) (Invoice, error) {
		if invoice.Status != StatusDraft {
			return Invoice{}, ErrNotUpdatable
		}
		if patch.Items != nil {
			invoice.Items = *patch.Items
		}
		if patch.Notes != nil {
			invoice.Notes = *patch.Notes
		}
		if patch.PaymentDueAt != nil {
			invoice.PaymentDueAt = *patch.PaymentDueAt
		}
		if patch.Recipient != nil {
			invoice.Recipient = *patch.Recipient
		}
		if patch.VATRate != nil {
			invoice.VATRate = *patch.VATRate
		}
		if err := validateInvoiceData(invoice.Items, invoice.VATRate); err != nil {
			return Invoice{}, err
		}
		invoice.NetTotal, invoice.VATAmount, invoice.GrossTotal = calculateTotals(invoice.Items, invoice.VATRate)
		return invoice, nil
	}

	return s.repo.Update(id, mutate)
}

func (s *Service) Issue(id string) (Invoice, error) {
	return s.repo.Update(id, func(invoice Invoice) (Invoice, error) {
		if invoice.Status != StatusDraft {
			return Invoice{}, ErrInvalidTransition
		}
		now := time.Now()
		invoice.Status = StatusIssued
		invoice.IssuedAt = now
		invoice.InvoiceNumber = s.repo.NextInvoiceNumber(now)
		return invoice, nil
	})
}
