SELECT producer, effect_key_hash, recipient_id::text, canonical_digest,
       message_id::text, legacy_receipt_reference, outcome, created_at, retain_until
FROM notifications_effect_receipts
WHERE producer = $1 AND effect_key_hash = $2 AND recipient_id = $3::uuid;
