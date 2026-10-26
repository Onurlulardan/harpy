UPDATE platform.billing_webhook_receipts AS receipt
SET processor_state = 'applied', processed_at = $5,
    revision = receipt.revision + 1
WHERE receipt.authority_id = $1 AND receipt.tenant_id = $2
  AND receipt.id = $3 AND receipt.processor_state = 'processing'
  AND EXISTS (
      SELECT 1 FROM platform.billing_provider_events AS event_row
      WHERE event_row.id = $4 AND event_row.webhook_receipt_id = receipt.id
        AND event_row.authority_id = receipt.authority_id
        AND event_row.tenant_id = receipt.tenant_id
  )
RETURNING receipt.id;
