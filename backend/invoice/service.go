package invoice

import (
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

func (s *Service) Create(invoice Invoice) Invoice {
	invoice.ID = uuid.NewString()
	invoice.CreatedAt = time.Now()

	var net float64
	for i := range invoice.Items {
		invoice.Items[i].Total = float64(invoice.Items[i].Quantity) * invoice.Items[i].UnitPrice
		net += invoice.Items[i].Total
	}

	invoice.NetTotal = net
	invoice.VATAmount = net * invoice.VATRate
	invoice.GrossTotal = net + invoice.VATAmount

	return s.repo.Create(invoice)
}

func (s *Service) GetByID(id string) (Invoice, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetAll() []Invoice {
	return s.repo.GetAll()
}

func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}
