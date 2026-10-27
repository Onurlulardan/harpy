INSERT INTO notifications_effect_receipts (
    producer, effect_key_hash, recipient_id, canonical_digest, message_id,
    legacy_receipt_reference, outcome, created_at, retain_until
) VALUES (
    'legacy_v1', $1::bytea, $2::uuid, $3::bytea, $4::uuid,
    'v1:notifications:' || $4::text, 'created', $5,
    GREATEST(
        $5::timestamptz + interval '365 days',
        pg_catalog.transaction_timestamp() + interval '365 days'
    )
);
