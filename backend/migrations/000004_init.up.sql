BEGIN;

CREATE TABLE invoices (
    id UUID PRIMARY KEY,
    invoice_number TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('draft', 'issued', 'paid', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT '0001-01-01 00:00:00+00',
    payment_due_at TIMESTAMPTZ NOT NULL,
    sender_name TEXT NOT NULL,
    sender_street TEXT NOT NULL,
    sender_zip TEXT NOT NULL,
    sender_city TEXT NOT NULL,
    sender_country TEXT NOT NULL,
    sender_email TEXT NOT NULL DEFAULT '',
    sender_phone TEXT NOT NULL DEFAULT '',
    sender_tax_id TEXT NOT NULL DEFAULT '',
    recipient_name TEXT NOT NULL,
    recipient_street TEXT NOT NULL,
    recipient_zip TEXT NOT NULL,
    recipient_city TEXT NOT NULL,
    recipient_country TEXT NOT NULL,
    recipient_email TEXT NOT NULL DEFAULT '',
    recipient_phone TEXT NOT NULL DEFAULT '',
    recipient_tax_id TEXT NOT NULL DEFAULT '',
    vat_rate NUMERIC(5,4) NOT NULL CHECK (vat_rate >= 0 AND vat_rate <= 1),
    net_total NUMERIC(12,2) NOT NULL,
    vat_amount NUMERIC(12,2) NOT NULL,
    gross_total NUMERIC(12,2) NOT NULL,
    notes TEXT NOT NULL DEFAULT ''
);

-- Drafts share the empty string as number, so the constraint has to skip them.
CREATE UNIQUE INDEX invoices_invoice_number_key
    ON invoices (invoice_number)
    WHERE invoice_number <> '';

CREATE TABLE invoice_items (
    id UUID PRIMARY KEY,
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    description TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0),
    unit TEXT NOT NULL DEFAULT '',
    total NUMERIC(12,2) NOT NULL,
    UNIQUE (invoice_id, position)
);

CREATE INDEX invoice_items_invoice_id_idx ON invoice_items (invoice_id);

CREATE TABLE invoice_counters (
    year INTEGER PRIMARY KEY,
    counter INTEGER NOT NULL DEFAULT 0
);

COMMIT;
