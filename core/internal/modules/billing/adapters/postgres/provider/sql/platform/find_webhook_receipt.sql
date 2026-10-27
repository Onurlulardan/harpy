SELECT receipt.id,
       receipt.tenant_id,
       receipt.provider_object_mapping_id,
       receipt.body_sha256,
       receipt.provider_event_id
FROM platform.billing_webhook_receipts AS receipt
WHERE receipt.provider_account_id = $1
  AND receipt.provider_event_id = $2
FOR UPDATE;
