INSERT INTO platform.outbox (
    event_id, tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin,
    envelope_contract_version, event_type, event_version,
    aggregate_type, aggregate_id, aggregate_version, event_sequence,
    occurred_at, actor_id, correlation_id, causation_id,
    idempotency_key, business_effect_idempotency_key,
    payload_canonical, payload_sha256, payload_schema_sha256,
    delivery_state, delivery_attempts, lease_generation, next_attempt_at
) VALUES (
    $1, $2, $3, $4,
    $5, $6,
    $7, $8, $9,
    $10, $11, $12,
    $13, $14, $15, $16,
    $17, $18, $19, $20,
    $21, $22,
    $23, $24, $25,
    'pending', 0, 0, $17
)
ON CONFLICT DO NOTHING;
