BEGIN;

ALTER TABLE invoice_items
    ALTER COLUMN unit_price TYPE NUMERIC(12,2) USING (unit_price::numeric / 100),
    ALTER COLUMN total      TYPE NUMERIC(12,2) USING (total::numeric / 100);

ALTER TABLE invoices
    ALTER COLUMN net_total   TYPE NUMERIC(12,2) USING (net_total::numeric / 100),
    ALTER COLUMN vat_amount  TYPE NUMERIC(12,2) USING (vat_amount::numeric / 100),
    ALTER COLUMN gross_total TYPE NUMERIC(12,2) USING (gross_total::numeric / 100);

COMMIT;
