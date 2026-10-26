SELECT event_row.id,
       event_row.normalized_schema,
       event_row.normalized_version,
       event_row.ordering_mode,
       event_row.ordering_sequence,
       event_row.ordering_previous,
       event_row.canonical_fact,
       event_row.occurred_at,
       event_row.normalized_sha256,
       receipt.processor_state
FROM platform.billing_provider_events AS event_row
JOIN platform.billing_webhook_receipts AS receipt
  ON receipt.id = event_row.webhook_receipt_id
 AND receipt.tenant_id = event_row.tenant_id
WHERE event_row.authority_id = $1
  AND event_row.tenant_id = $2
  AND event_row.webhook_receipt_id = $3
FOR SHARE OF event_row, receipt;
