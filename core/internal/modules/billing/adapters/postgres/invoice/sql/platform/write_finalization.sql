WITH updated AS (
    UPDATE platform.billing_invoices AS invoice
    SET state = 'finalized',
        invoice_total_minor_units = $6,
        settled_minor_units = $7,
        amount_due_minor_units = $8,
        credit_balance_minor_units = $9,
        updated_at = $12,
        finalized_at = $12,
        revision = $5
    WHERE invoice.authority_id = $1 AND invoice.tenant_id = $2
      AND invoice.id = $3 AND invoice.state = 'draft'
      AND invoice.revision = $4 AND invoice.policy_sha256 = $11
      AND invoice.currency_registry_version = $10
    RETURNING invoice.id, invoice.currency, invoice.currency_registry_version,
              invoice.policy_sha256
), inserted AS (
    INSERT INTO platform.billing_invoice_finalizations (
        id, authority_id, tenant_id, invoice_id, source_operation_id,
        expected_revision, result_revision, invoice_total_minor_units,
        settled_minor_units, amount_due_minor_units, credit_balance_minor_units,
        currency, currency_registry_version, policy_sha256, finalized_at
    )
    SELECT $13, $1, $2, updated.id, $13, $4, $5, $6, $7, $8, $9,
           updated.currency, updated.currency_registry_version,
           updated.policy_sha256, $12
    FROM updated
    RETURNING invoice_id
)
SELECT invoice_id FROM inserted;
