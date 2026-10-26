INSERT INTO notifications_delivery_receipts (
    id, delivery_id, provider_reference_digest, outcome,
    signature_key_id, received_at
) VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6)
ON CONFLICT (delivery_id, provider_reference_digest, outcome) DO NOTHING;
