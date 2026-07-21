package invoice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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

func (r *PostgresRepository) GetAll() ([]Invoice, error) {
	ctx := context.Background()

	rows, err := r.pool.Query(ctx, `
		SELECT
			id, invoice_number, status, created_at, issued_at, payment_due_at,
			sender_name, sender_street, sender_zip, sender_city, sender_country, sender_email, sender_phone, sender_tax_id,
			recipient_name, recipient_street, recipient_zip, recipient_city, recipient_country, recipient_email, recipient_phone, recipient_tax_id,
			vat_rate, net_total, vat_amount, gross_total, notes
		FROM invoices
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []Invoice
	for rows.Next() {
		var invoice Invoice
		err := rows.Scan(
			&invoice.ID, &invoice.InvoiceNumber, &invoice.Status, &invoice.CreatedAt, &invoice.IssuedAt, &invoice.PaymentDueAt,
			&invoice.Sender.Name, &invoice.Sender.Street, &invoice.Sender.Zip, &invoice.Sender.City, &invoice.Sender.Country, &invoice.Sender.Email, &invoice.Sender.Phone, &invoice.Sender.TaxID,
			&invoice.Recipient.Name, &invoice.Recipient.Street, &invoice.Recipient.Zip, &invoice.Recipient.City, &invoice.Recipient.Country, &invoice.Recipient.Email, &invoice.Recipient.Phone, &invoice.Recipient.TaxID,
			&invoice.VATRate, &invoice.NetTotal, &invoice.VATAmount, &invoice.GrossTotal, &invoice.Notes,
		)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invoices, nil
}

func (r *PostgresRepository) Create(invoice Invoice) (Invoice, error) {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Invoice{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO invoices (
			id, invoice_number, status, created_at, issued_at, payment_due_at,
			sender_name, sender_street, sender_zip, sender_city, sender_country, sender_email, sender_phone, sender_tax_id,
			recipient_name, recipient_street, recipient_zip, recipient_city, recipient_country, recipient_email, recipient_phone, recipient_tax_id,
			vat_rate, net_total, vat_amount, gross_total, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)
	`, invoice.ID, invoice.InvoiceNumber, invoice.Status, invoice.CreatedAt, invoice.IssuedAt, invoice.PaymentDueAt,
		invoice.Sender.Name, invoice.Sender.Street, invoice.Sender.Zip, invoice.Sender.City, invoice.Sender.Country, invoice.Sender.Email, invoice.Sender.Phone, invoice.Sender.TaxID,
		invoice.Recipient.Name, invoice.Recipient.Street, invoice.Recipient.Zip, invoice.Recipient.City, invoice.Recipient.Country, invoice.Recipient.Email, invoice.Recipient.Phone, invoice.Recipient.TaxID,
		invoice.VATRate, invoice.NetTotal, invoice.VATAmount, invoice.GrossTotal, invoice.Notes)
	if err != nil {
		return Invoice{}, err
	}

	for _, item := range invoice.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO invoice_items (id, invoice_id, position, description, quantity, unit_price, unit, total)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, item.ID, invoice.ID, item.Position, item.Description, item.Quantity, item.UnitPrice, item.Unit, item.Total)
		if err != nil {
			return Invoice{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Invoice{}, err
	}
	return invoice, nil
}

func (r *PostgresRepository) Delete(id string) error {
	ctx := context.Background()

	commandTag, err := r.pool.Exec(ctx, `DELETE FROM invoices WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) NextInvoiceNumber(now time.Time) (string, error) {
	ctx := context.Background()
	row := r.pool.QueryRow(ctx, `
		INSERT INTO invoice_counters (year, counter) VALUES ($1, 1)
		ON CONFLICT (year) DO UPDATE SET counter = invoice_counters.counter + 1
		RETURNING counter
	`, now.Year())

	var counter int
	if err := row.Scan(&counter); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d-%04d", now.Year(), counter), nil
}

func (r *PostgresRepository) Update(id string, fn UpdateFunc) (Invoice, error) {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Invoice{}, err
	}
	defer tx.Rollback(ctx)

	var existing Invoice
	row := tx.QueryRow(ctx, `
		SELECT
			id, invoice_number, status, created_at, issued_at, payment_due_at,
			sender_name, sender_street, sender_zip, sender_city, sender_country, sender_email, sender_phone, sender_tax_id,
			recipient_name, recipient_street, recipient_zip, recipient_city, recipient_country, recipient_email, recipient_phone, recipient_tax_id,
			vat_rate, net_total, vat_amount, gross_total, notes
		FROM invoices
		WHERE id = $1
		FOR UPDATE
	`, id)

	err = row.Scan(
		&existing.ID, &existing.InvoiceNumber, &existing.Status, &existing.CreatedAt, &existing.IssuedAt, &existing.PaymentDueAt,
		&existing.Sender.Name, &existing.Sender.Street, &existing.Sender.Zip, &existing.Sender.City, &existing.Sender.Country, &existing.Sender.Email, &existing.Sender.Phone, &existing.Sender.TaxID,
		&existing.Recipient.Name, &existing.Recipient.Street, &existing.Recipient.Zip, &existing.Recipient.City, &existing.Recipient.Country, &existing.Recipient.Email, &existing.Recipient.Phone, &existing.Recipient.TaxID,
		&existing.VATRate, &existing.NetTotal, &existing.VATAmount, &existing.GrossTotal, &existing.Notes,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Invoice{}, ErrNotFound
	}
	if err != nil {
		return Invoice{}, err
	}

	itemRows, err := tx.Query(ctx, `
		SELECT id, invoice_id, position, description, quantity, unit_price, unit, total
		FROM invoice_items
		WHERE invoice_id = $1
		ORDER BY position
	`, id)
	if err != nil {
		return Invoice{}, err
	}
	for itemRows.Next() {
		var item LineItem
		if err := itemRows.Scan(&item.ID, &item.InvoiceID, &item.Position, &item.Description, &item.Quantity, &item.UnitPrice, &item.Unit, &item.Total); err != nil {
			itemRows.Close()
			return Invoice{}, err
		}
		existing.Items = append(existing.Items, item)
	}
	itemRows.Close()
	if err := itemRows.Err(); err != nil {
		return Invoice{}, err
	}

	updated, err := fn(existing)
	if err != nil {
		return Invoice{}, err
	}
	updated.ID = existing.ID

	_, err = tx.Exec(ctx, `
		UPDATE invoices SET
			invoice_number = $1, status = $2, issued_at = $3, payment_due_at = $4,
			sender_name = $5, sender_street = $6, sender_zip = $7, sender_city = $8, sender_country = $9, sender_email = $10, sender_phone = $11, sender_tax_id = $12,
			recipient_name = $13, recipient_street = $14, recipient_zip = $15, recipient_city = $16, recipient_country = $17, recipient_email = $18, recipient_phone = $19, recipient_tax_id = $20,
			vat_rate = $21, net_total = $22, vat_amount = $23, gross_total = $24, notes = $25
		WHERE id = $26
	`, updated.InvoiceNumber, updated.Status, updated.IssuedAt, updated.PaymentDueAt,
		updated.Sender.Name, updated.Sender.Street, updated.Sender.Zip, updated.Sender.City, updated.Sender.Country, updated.Sender.Email, updated.Sender.Phone, updated.Sender.TaxID,
		updated.Recipient.Name, updated.Recipient.Street, updated.Recipient.Zip, updated.Recipient.City, updated.Recipient.Country, updated.Recipient.Email, updated.Recipient.Phone, updated.Recipient.TaxID,
		updated.VATRate, updated.NetTotal, updated.VATAmount, updated.GrossTotal, updated.Notes, updated.ID)
	if err != nil {
		return Invoice{}, err
	}

	_, err = tx.Exec(ctx, `DELETE FROM invoice_items WHERE invoice_id = $1`, updated.ID)
	if err != nil {
		return Invoice{}, err
	}
	for _, item := range updated.Items {
		if item.ID == "" {
			item.ID = uuid.NewString()
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO invoice_items (id, invoice_id, position, description, quantity, unit_price, unit, total)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, item.ID, updated.ID, item.Position, item.Description, item.Quantity, item.UnitPrice, item.Unit, item.Total)
		if err != nil {
			return Invoice{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Invoice{}, err
	}
	return updated, nil
}
