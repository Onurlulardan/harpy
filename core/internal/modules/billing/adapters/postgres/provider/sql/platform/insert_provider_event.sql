INSERT INTO platform.billing_provider_events (
    id, authority_id, tenant_id, webhook_receipt_id,
    provider_object_mapping_id, normalized_schema, normalized_version,
    ordering_mode, ordering_sequence, ordering_previous,
    canonical_fact, occurred_at, normalized_at, normalized_sha256
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9, $10,
    $11, $12, $13, $14
)
RETURNING id;
