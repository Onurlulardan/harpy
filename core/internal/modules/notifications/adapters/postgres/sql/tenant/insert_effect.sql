INSERT INTO notifications_effect_receipts (
    producer, effect_key_hash, recipient_id, canonical_digest, message_id,
    legacy_receipt_reference, outcome, created_at, retain_until
) VALUES ($1, $2, $3::uuid, $4, $5::uuid, $6, $7, $8, $9);
