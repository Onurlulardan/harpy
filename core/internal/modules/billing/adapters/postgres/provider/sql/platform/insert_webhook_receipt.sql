INSERT INTO platform.billing_webhook_receipts (
    id, authority_id, tenant_id, provider_account_id,
    provider_object_mapping_id, provider_event_id,
    body_sha256, authentication_digest, verification_revision,
    event_occurred_at, signature_issued_at,
    algorithm, nonce, encrypted_body, wrapped_dek, key_id,
    verified_at, received_at, processor_state, processed_at,
    raw_body_expires_at, retain_until, revision
) VALUES (
    $1, $2, $3, $4,
    $5, $6,
    $7, $8, $9,
    $10, $11,
    $12, $13, $14, $15, $16,
    $17, $18, 'received', NULL,
    $19, $20, 1
)
ON CONFLICT (provider_account_id, provider_event_id) DO NOTHING;
