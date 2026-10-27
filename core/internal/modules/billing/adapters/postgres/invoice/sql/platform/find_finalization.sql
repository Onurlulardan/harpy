SELECT invoice_id, source_operation_id, expected_revision, result_revision,
       invoice_total_minor_units, settled_minor_units, amount_due_minor_units,
       credit_balance_minor_units, currency::text, currency_registry_version,
       policy_sha256, finalized_at
FROM platform.billing_invoice_finalizations
WHERE authority_id = $1 AND tenant_id = $2 AND source_operation_id = $3;
