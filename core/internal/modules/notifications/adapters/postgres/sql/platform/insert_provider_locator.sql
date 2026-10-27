INSERT INTO platform.notification_provider_locators (
    provider_kind, provider_account_digest, external_message_digest,
    tenant_id, delivery_id, accepted_at, expires_at
) VALUES (
    $1, $2::bytea, $3::bytea, $4::uuid, $5::uuid, $6,
    $6::timestamptz + interval '90 days'
)
ON CONFLICT (
    provider_kind, provider_account_digest, external_message_digest
) DO UPDATE
SET expires_at = platform.notification_provider_locators.expires_at
WHERE platform.notification_provider_locators.tenant_id = EXCLUDED.tenant_id
  AND platform.notification_provider_locators.delivery_id = EXCLUDED.delivery_id
RETURNING tenant_id::text, delivery_id::text;
