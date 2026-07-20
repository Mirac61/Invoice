CREATE TABLE invoices (
    id UUID PRIMARY KEY,
    invoice_number TEXT,
    status TEXT NOT NULL CHECK (status IN ('draft', 'issued', 'paid', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL,
    issued_at TIMESTAMPTZ,
    payment_due_at TIMESTAMPTZ NOT NULL,

    sender_name TEXT NOT NULL,
    sender_street TEXT NOT NULL,
    sender_zip TEXT NOT NULL,
    sender_city TEXT NOT NULL,
    sender_country TEXT NOT NULL,
    sender_email TEXT,
    sender_phone TEXT,
    sender_tax_id TEXT,

    recipient_name TEXT NOT NULL,
    recipient_street TEXT NOT NULL,
    recipient_zip TEXT NOT NULL,
    recipient_city TEXT NOT NULL,
    recipient_country TEXT NOT NULL,
    recipient_email TEXT,
    recipient_phone TEXT,
    recipient_tax_id TEXT,

    vat_rate NUMERIC(5,4) NOT NULL CHECK (vat_rate >= 0 AND vat_rate <= 1),
    net_total NUMERIC(12,2) NOT NULL,
    vat_amount NUMERIC(12,2) NOT NULL,
    gross_total NUMERIC(12,2) NOT NULL,
    notes TEXT
);
