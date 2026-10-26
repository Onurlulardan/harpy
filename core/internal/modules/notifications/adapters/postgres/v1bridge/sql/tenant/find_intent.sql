SELECT producer, effect_key_hash, recipient_id::text, producer_generation,
       notification_generation, canonical_digest, payload_digest, payload,
       job_id::text, state, created_at, updated_at, row_version
FROM notifications_migration_intents
WHERE producer = $1 AND effect_key_hash = $2::bytea AND recipient_id = $3::uuid;
