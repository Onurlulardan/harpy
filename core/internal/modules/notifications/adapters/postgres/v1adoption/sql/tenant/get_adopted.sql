SELECT receipt.canonical_digest
FROM notifications_messages message
JOIN notifications_effect_receipts receipt
  ON receipt.message_id = message.id
 AND receipt.producer = 'legacy_v1'
 AND receipt.recipient_id = message.recipient_id
WHERE message.id = $1::uuid AND message.recipient_id = $2::uuid
  AND receipt.effect_key_hash = $3::bytea;
