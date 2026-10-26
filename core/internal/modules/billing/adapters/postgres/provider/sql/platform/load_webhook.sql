SELECT receipt.authority_id,
       mapping.id,
       account.id,
       account.provider_key,
       account.provider_account_key,
       mapping.object_kind,
       mapping.provider_object_id,
       mapping.subscription_id,
       mapping.invoice_id,
       receipt.provider_event_id,
       receipt.authentication_digest,
       receipt.verification_revision,
       receipt.event_occurred_at,
       receipt.signature_issued_at,
       receipt.algorithm,
       receipt.nonce,
       receipt.encrypted_body,
       receipt.wrapped_dek,
       receipt.key_id,
       receipt.body_sha256,
       receipt.received_at,
       receipt.processor_state
FROM platform.billing_webhook_receipts AS receipt
JOIN platform.billing_provider_accounts AS account
  ON account.id = receipt.provider_account_id
JOIN platform.billing_provider_objects AS mapping
  ON mapping.id = receipt.provider_object_mapping_id
 AND mapping.tenant_id = receipt.tenant_id
 AND mapping.provider_account_id = receipt.provider_account_id
WHERE receipt.tenant_id = $1
  AND receipt.id = $2
  AND account.lifecycle = 'active'
  AND account.supports_webhooks
  AND mapping.retired_at IS NULL
FOR UPDATE OF receipt, account, mapping;
