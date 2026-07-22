package invoice

import "time"

type InvoicePatch struct {
	PaymentDueAt *time.Time  `json:"paymentDueAt" binding:"omitempty,notzero"`
	Recipient    *Contact    `json:"recipient"`
	Items        *[]LineItem `json:"items" binding:"omitempty,min=1,dive"`
	VATRate      *float64    `json:"vatRate" binding:"omitempty,gte=0,lte=1"`
	Notes        *string     `json:"notes"`
}

type InvoiceStatus string

const (
	StatusDraft     InvoiceStatus = "draft"
	StatusIssued    InvoiceStatus = "issued"
	StatusPaid      InvoiceStatus = "paid"
	StatusCancelled InvoiceStatus = "cancelled"
)

type Invoice struct {
	ID            string        `json:"id"`
	InvoiceNumber string        `json:"invoiceNumber"`
	Status        InvoiceStatus `json:"status"`
	CreatedAt     time.Time     `json:"createdAt"`
	IssuedAt      time.Time     `json:"issuedAt"`
	PaymentDueAt  time.Time     `json:"paymentDueAt" binding:"required"`
	Sender        Contact       `json:"sender" binding:"required"`
	Recipient     Contact       `json:"recipient" binding:"required"`
	Items         []LineItem    `json:"items" binding:"required,min=1,dive"`
	VATRate       float64       `json:"vatRate" binding:"gte=0,lte=1"`
	NetTotal      float64       `json:"netTotal"`
	VATAmount     float64       `json:"vatAmount"`
	GrossTotal    float64       `json:"grossTotal"`
	Notes         string        `json:"notes"`
}

type Contact struct {
	Name    string `json:"name" binding:"required"`
	Street  string `json:"street" binding:"required"`
	Zip     string `json:"zip" binding:"required"`
	City    string `json:"city" binding:"required"`
	Country string `json:"country" binding:"required"`
	Email   string `json:"email" binding:"omitempty,email"`
	Phone   string `json:"phone"`
	TaxID   string `json:"taxId"`
}

type LineItem struct {
	ID          string  `json:"id"`
	InvoiceID   string  `json:"invoiceId"`
	Position    int     `json:"position"`
	Description string  `json:"description" binding:"required"`
	Quantity    int     `json:"quantity" binding:"gt=0"`
	UnitPrice   float64 `json:"unitPrice" binding:"gte=0"`
	Unit        string  `json:"unit"`
	Total       float64 `json:"total"`
}
