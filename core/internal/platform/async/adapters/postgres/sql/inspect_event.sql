WITH producer AS MATERIALIZED (
    SELECT event_row.event_id
    FROM platform.outbox AS event_row
    WHERE $6::uuid = $2::uuid AND $7::text = $1::text AND $8::uuid = $2::uuid
      AND event_row.event_id = $5::uuid AND event_row.tenant_id = $6::uuid
      AND event_row.scope_kind = $7::text AND event_row.scope_id = $8::uuid
      AND event_row.authority_group_id = $9::text
      AND event_row.authority_group_generation = $10::bigint
      AND event_row.authority_domain = $11::text
      AND event_row.cutover_generation = $12::bigint
      AND event_row.writer_origin = $13::text
      AND event_row.envelope_contract_version = $14::text
      AND event_row.event_type = $15::text AND event_row.event_version = $16::integer
      AND event_row.aggregate_type = $17::text AND event_row.aggregate_id = $18::uuid
      AND event_row.aggregate_version = $19::bigint
      AND event_row.event_sequence = $20::bigint
      AND event_row.occurred_at = $21::timestamptz AND event_row.actor_id = $22::uuid
      AND event_row.correlation_id = $23::text AND event_row.causation_id = $24::uuid
      AND event_row.idempotency_key = $25::text
      AND event_row.business_effect_idempotency_key = $26::text
      AND event_row.payload_sha256 = $27::text
      AND event_row.payload_schema_sha256 = $28::text
),
stream AS MATERIALIZED (
    SELECT
        COALESCE(pg_catalog.max(row.through_sequence), 0::bigint) + 1 AS expected,
        pg_catalog.bool_or(
            row.record_kind = 'event_receipt' AND row.event_id = $5::uuid
        ) AS duplicate
    FROM platform.inbox AS row
    WHERE row.consumer_name = $3::text AND row.consumer_version = $4::integer
      AND row.scope_kind = $1::text AND row.scope_id = $2::uuid
      AND row.aggregate_type = $17::text AND row.aggregate_id = $18::uuid
)
SELECT CASE
    WHEN NOT EXISTS (SELECT 1 FROM producer) THEN 'producer_missing'
    WHEN COALESCE(stream.duplicate, false) THEN 'duplicate'
    WHEN stream.expected = $20::bigint THEN 'apply'
    ELSE 'gap'
END, stream.expected
FROM stream;
