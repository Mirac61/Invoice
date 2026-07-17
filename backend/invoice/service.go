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

func (s *Service) Create(invoice Invoice) Invoice {
	invoice.ID = uuid.NewString()
	invoice.CreatedAt = time.Now()

	var net float64
	for i := range invoice.Items {
		invoice.Items[i].Total = float64(invoice.Items[i].Quantity) * invoice.Items[i].UnitPrice
		net += invoice.Items[i].Total
	}

	invoice.NetTotal = math.Round(net*100) / 100
	invoice.VATAmount = math.Round(net*invoice.VATRate*100) / 100
	invoice.GrossTotal = math.Round((invoice.NetTotal+invoice.VATAmount)*100) / 100

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
