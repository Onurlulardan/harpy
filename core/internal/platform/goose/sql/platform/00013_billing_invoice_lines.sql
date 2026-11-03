-- +goose Up
CREATE TABLE platform.billing_invoice_lines (
    id uuid NOT NULL PRIMARY KEY,
    gross_minor_units text NOT NULL,
    discount_minor_units text NOT NULL,
    tax_minor_units text NOT NULL,
    debit_adjustment_minor_units text NOT NULL,
    credit_adjustment_minor_units text NOT NULL,
    line_total_minor_units text NOT NULL,
    line_sequence text NOT NULL,
    invoice_id uuid NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_invoice_lines;
