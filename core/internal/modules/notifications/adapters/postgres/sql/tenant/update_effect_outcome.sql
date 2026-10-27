UPDATE notifications_effect_receipts
SET outcome = $4
WHERE producer = $1 AND effect_key_hash = $2 AND recipient_id = $3::uuid;
