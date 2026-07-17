package invoice

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func calculateTotals(items []LineItem, vatRate float64) (net, vat, gross float64) {
	for i := range items {
		items[i].Total = float64(items[i].Quantity) * items[i].UnitPrice
		net += items[i].Total
	}
	net = math.Round(net*100) / 100
	vat = math.Round(net*vatRate*100) / 100
	gross = math.Round((net+vat)*100) / 100
	return
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

func (s *Service) Update(id string, existing Invoice) (Invoice, error) {
	invoice, err := s.repo.GetByID(id)
	if err != nil {
		return Invoice{}, err
	}
	if invoice.Status != StatusDraft {
		return Invoice{}, ErrNotUpdatable
	}
	existing.ID = invoice.ID
	existing.InvoiceNumber = invoice.InvoiceNumber
	existing.Status = invoice.Status
	existing.CreatedAt = invoice.CreatedAt
	existing.IssuedAt = invoice.IssuedAt
	existing.NetTotal, existing.VATAmount, existing.GrossTotal = calculateTotals(existing.Items, existing.VATRate)
	return s.repo.Update(id, existing)
}

func (s *Service) PartialUpdate(id string, patch InvoicePatch) (Invoice, error) {
	invoice, err := s.repo.GetByID(id)
	if err != nil {
		return Invoice{}, err
	}
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

	invoice.NetTotal, invoice.VATAmount, invoice.GrossTotal = calculateTotals(invoice.Items, invoice.VATRate)

	return s.repo.Update(id, invoice)
}
