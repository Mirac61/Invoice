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
	invoice, err := s.repo.GetByID(id)
	if err != nil{
		return err
	}
	if invoice.Status != StatusDraft{
		return ErrNotDeletable
	}
	return s.repo.Delete(id)
}

func (s *Service) Update(id string, existing Invoice) (Invoice, error){
	invoice, err := s.repo.GetByID(id)
	if err != nil{
		return Invoice{}, err
	}
	if invoice.Status != StatusDraft{
		return Invoice{}, ErrNotUpdatable
	}
	return s.repo.Update(id, existing)
}
