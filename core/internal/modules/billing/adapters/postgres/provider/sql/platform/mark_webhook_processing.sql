UPDATE platform.billing_webhook_receipts
SET processor_state = 'processing', revision = revision + 1
WHERE authority_id = $1
  AND tenant_id = $2
  AND id = $3
  AND processor_state = 'received'
RETURNING id;
