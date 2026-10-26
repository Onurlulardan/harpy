UPDATE platform.billing_webhook_receipts
SET processor_state = 'reconcile_required',
    processed_at = $4,
    revision = revision + 1
WHERE authority_id = $1
  AND tenant_id = $2
  AND id = $3
  AND processor_state = 'processing'
RETURNING id;
