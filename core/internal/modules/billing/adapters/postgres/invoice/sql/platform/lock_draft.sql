WITH locked AS MATERIALIZED (
    SELECT invoice.id, invoice.revision, invoice.state, invoice.currency::text,
           invoice.currency_registry_version, invoice.settled_minor_units,
           invoice.policy_sha256
    FROM platform.billing_invoices AS invoice
    WHERE invoice.authority_id = $1 AND invoice.tenant_id = $2 AND invoice.id = $3
    FOR UPDATE OF invoice
)
SELECT locked.id, locked.revision, locked.currency,
       locked.currency_registry_version, locked.settled_minor_units,
       locked.policy_sha256, locked.state = 'finalized',
       COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
           'id', line.id,
           'gross', line.gross_minor_units,
           'discount', line.discount_minor_units,
           'tax', line.tax_minor_units,
           'debit', line.debit_adjustment_minor_units,
           'credit', line.credit_adjustment_minor_units,
           'total', line.line_total_minor_units
       ) ORDER BY line.line_sequence) FILTER (WHERE line.id IS NOT NULL), '[]'::jsonb)
FROM locked
LEFT JOIN platform.billing_invoice_lines AS line ON line.invoice_id = locked.id
GROUP BY locked.id, locked.revision, locked.currency,
         locked.currency_registry_version, locked.settled_minor_units,
         locked.policy_sha256, locked.state;
