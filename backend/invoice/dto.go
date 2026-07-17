package invoice

import "time"

type InvoicePatch struct {
	PaymentDueAt *time.Time  `json:"paymentDueAt"`
	Recipient    *Contact    `json:"recipient"`
	Items        *[]LineItem `json:"items"`
	VATRate      *float64    `json:"vatRate"`
	Notes        *string     `json:"notes"`
}
