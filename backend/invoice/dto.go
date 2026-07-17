package invoice

import "time"

type InvoicePatch struct {
	PaymentDueAt *time.Time  `json:"paymentDueAt"`
	Recipient    *Contact    `json:"recipient"`
	Items        *[]LineItem `json:"items" binding:"omitempty,min=1,dive"`
	VATRate      *float64    `json:"vatRate" binding:"omitempty,gte=0,lte=1"`
	Notes        *string     `json:"notes"`
}
