INSERT INTO notifications_migration_intents (
    producer, effect_key_hash, recipient_id, producer_generation,
    notification_generation, canonical_digest, payload_digest, payload,
    job_id, state, created_at, updated_at, row_version
) VALUES (
    $1, $2::bytea, $3::uuid, $4, $5, $6::bytea, $7::bytea, $8::jsonb,
    $9::uuid, $10, $11, $12, $13
);
