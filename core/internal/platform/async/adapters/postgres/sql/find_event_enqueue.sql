WITH candidates AS (
    SELECT *
    FROM platform.outbox
    WHERE event_id = $1 OR (
        scope_kind = 'tenant' AND scope_id = $2 AND aggregate_type = $3 AND
        aggregate_id = $4 AND event_sequence = $5
    )
)
SELECT
    COALESCE(pg_catalog.min(event_id::text), ''),
    pg_catalog.count(*) = 1 AND COALESCE(pg_catalog.bool_and(
        event_id = $1 AND tenant_id = $2 AND scope_kind = 'tenant' AND scope_id = $2 AND
        authority_group_id = $6 AND authority_group_generation = $7 AND
        authority_domain = $8 AND cutover_generation = $9 AND writer_origin = 'v2' AND
        envelope_contract_version = $10 AND event_type = $11 AND event_version = $12 AND
        aggregate_type = $3 AND aggregate_id = $4 AND aggregate_version = $13 AND
        event_sequence = $5 AND actor_id = $14 AND correlation_id = $15 AND
        causation_id = $16 AND idempotency_key = $17 AND
        business_effect_idempotency_key = $18 AND payload_sha256 = $19 AND
        payload_schema_sha256 = $20 AND occurred_at = $21
    ), false)
FROM candidates;
