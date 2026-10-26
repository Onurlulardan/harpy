SELECT event_row.id, receipt.processor_state
FROM platform.billing_provider_events AS event_row
JOIN platform.billing_webhook_receipts AS receipt
  ON receipt.id = event_row.webhook_receipt_id
 AND receipt.tenant_id = event_row.tenant_id
JOIN platform.billing_provider_objects AS mapping
  ON mapping.id = event_row.provider_object_mapping_id
 AND mapping.tenant_id = event_row.tenant_id
JOIN platform.billing_subscriptions AS subscription
  ON subscription.id = mapping.subscription_id
 AND subscription.tenant_id = mapping.tenant_id
WHERE event_row.authority_id = $1 AND event_row.tenant_id = $2
  AND event_row.id = $3 AND event_row.webhook_receipt_id = $4
  AND event_row.normalized_sha256 = $5
  AND subscription.id = $6
  AND receipt.processor_state IN ('processing', 'applied')
FOR SHARE OF event_row, receipt, mapping;
