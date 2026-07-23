BEGIN;

ALTER TABLE invoice_items
    ALTER COLUMN unit_price TYPE BIGINT USING ROUND(unit_price * 100),
    ALTER COLUMN total      TYPE BIGINT USING ROUND(total * 100);

ALTER TABLE invoices
    ALTER COLUMN net_total   TYPE BIGINT USING ROUND(net_total * 100),
    ALTER COLUMN vat_amount  TYPE BIGINT USING ROUND(vat_amount * 100),
    ALTER COLUMN gross_total TYPE BIGINT USING ROUND(gross_total * 100);

COMMIT;
