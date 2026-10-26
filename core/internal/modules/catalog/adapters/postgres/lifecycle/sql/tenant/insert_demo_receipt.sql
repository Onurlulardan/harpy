INSERT INTO catalog_demo_receipts (
    receipt_id, installation_id, operation_id, component_id, action,
    stable_ids, content_sha256, actor_id, idempotency_key, completed_at
) VALUES (
    $1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::text,
    $6::jsonb, $7::text, $8::uuid, $9::text, $10::timestamptz
)
ON CONFLICT (installation_id, component_id, action, idempotency_key) DO NOTHING

