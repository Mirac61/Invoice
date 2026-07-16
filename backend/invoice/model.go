package main

import "time"

type Invoice struct {
	ID            string     `json:"id"`
	InvoiceNumber string     `json:"invoiceNumber"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	IssuedAt      time.Time  `json:"issuedAt"`
	PaymentDueAt  time.Time  `json:"paymentDueAt"`
	Sender        Contact    `json:"sender"`
	Recipient     Contact    `json:"recipient"`
	Items         []LineItem `json:"items"`
	VATRate       float64    `json:"vatRate"`
	NetTotal      float64    `json:"netTotal"`
	VATAmount     float64    `json:"vatAmount"`
	GrossTotal    float64    `json:"grossTotal"`
	Notes         string     `json:"notes"`
}

type Contact struct {
	Name    string `json:"name"`
	Street  string `json:"street"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Country string `json:"country"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	TaxID   string `json:"taxId"`
}

type LineItem struct {
	ID          string  `json:"id"`
	InvoiceID   string  `json:"invoiceId"`
	Position    int     `json:"position"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unitPrice"`
	Total       float64 `json:"total"`
}
