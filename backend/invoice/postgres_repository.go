package invoice

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetByID(id string) (Invoice, error) {
	ctx := context.Background()
	var invoice Invoice

	row := r.pool.QueryRow(ctx, `
		SELECT
			id, invoice_number, status, created_at, issued_at, payment_due_at,
			sender_name, sender_street, sender_zip, sender_city, sender_country, sender_email, sender_phone, sender_tax_id,
			recipient_name, recipient_street, recipient_zip, recipient_city, recipient_country, recipient_email, recipient_phone, recipient_tax_id,
			vat_rate, net_total, vat_amount, gross_total, notes
		FROM invoices
		WHERE id = $1
	`, id)

	err := row.Scan(
		&invoice.ID, &invoice.InvoiceNumber, &invoice.Status, &invoice.CreatedAt, &invoice.IssuedAt, &invoice.PaymentDueAt,
		&invoice.Sender.Name, &invoice.Sender.Street, &invoice.Sender.Zip, &invoice.Sender.City, &invoice.Sender.Country, &invoice.Sender.Email, &invoice.Sender.Phone, &invoice.Sender.TaxID,
		&invoice.Recipient.Name, &invoice.Recipient.Street, &invoice.Recipient.Zip, &invoice.Recipient.City, &invoice.Recipient.Country, &invoice.Recipient.Email, &invoice.Recipient.Phone, &invoice.Recipient.TaxID,
		&invoice.VATRate, &invoice.NetTotal, &invoice.VATAmount, &invoice.GrossTotal, &invoice.Notes,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Invoice{}, ErrNotFound
	}
	if err != nil {
		return Invoice{}, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, invoice_id, position, description, quantity, unit_price, unit, total
		FROM invoice_items
		WHERE invoice_id = $1
		ORDER BY position
	`, id)
	if err != nil {
		return Invoice{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var item LineItem
		if err := rows.Scan(&item.ID, &item.InvoiceID, &item.Position, &item.Description, &item.Quantity, &item.UnitPrice, &item.Unit, &item.Total); err != nil {
			return Invoice{}, err
		}
		invoice.Items = append(invoice.Items, item)
	}
	if err := rows.Err(); err != nil {
		return Invoice{}, err
	}

	return invoice, nil
}
