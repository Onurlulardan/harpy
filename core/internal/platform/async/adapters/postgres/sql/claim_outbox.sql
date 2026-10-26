WITH publish_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS started_at
), candidate AS MATERIALIZED (
    SELECT
        event.event_id,
        GREATEST(
            publish_clock.started_at,
            event.updated_at + interval '1 microsecond'
        ) AS started_at
    FROM platform.outbox AS event
    CROSS JOIN publish_clock
    WHERE event.scope_kind = $1
      AND event.scope_id = $2
      AND event.authority_group_id = $3
      AND event.authority_group_generation = $4
      AND event.authority_domain = $5
      AND event.cutover_generation = $6
      AND event.delivery_state = 'pending'
      AND event.parked_at IS NULL
      AND event.payload_canonical IS NOT NULL
      AND event.next_attempt_at <= publish_clock.started_at
      AND (
          event.event_sequence = 1 OR EXISTS (
              SELECT 1
              FROM platform.outbox AS predecessor
              WHERE predecessor.scope_kind = event.scope_kind
                AND predecessor.scope_id = event.scope_id
                AND predecessor.aggregate_type = event.aggregate_type
                AND predecessor.aggregate_id = event.aggregate_id
                AND predecessor.event_sequence = event.event_sequence - 1
                AND predecessor.delivery_state = 'delivered'
          )
      )
      AND NOT EXISTS (
          SELECT 1
          FROM platform.outbox AS prior
          WHERE prior.scope_kind = event.scope_kind
            AND prior.scope_id = event.scope_id
            AND prior.aggregate_type = event.aggregate_type
            AND prior.aggregate_id = event.aggregate_id
            AND prior.event_sequence < event.event_sequence
            AND prior.delivery_state <> 'delivered'
      )
    ORDER BY event.next_attempt_at, event.event_sequence, event.event_id
    FOR UPDATE OF event SKIP LOCKED
    LIMIT $8
), claimed AS (
    UPDATE platform.outbox AS event
    SET delivery_state = 'publishing',
        delivery_attempts = event.delivery_attempts + 1,
        lease_owner_id = $7,
        lease_generation = event.lease_generation + 1,
        lease_heartbeat_at = candidate.started_at,
        lease_until = candidate.started_at + interval '60 seconds',
        updated_at = candidate.started_at
    FROM candidate
    WHERE event.event_id = candidate.event_id
    RETURNING
        event.event_id, event.tenant_id, event.scope_kind, event.scope_id,
        event.authority_group_id, event.authority_group_generation,
        event.authority_domain, event.cutover_generation, event.writer_origin,
        event.envelope_contract_version, event.event_type, event.event_version,
        event.aggregate_type, event.aggregate_id, event.aggregate_version,
        event.event_sequence, event.occurred_at, event.actor_id,
        event.correlation_id, event.causation_id, event.idempotency_key,
        event.business_effect_idempotency_key, event.payload_canonical,
        event.payload_sha256, event.payload_schema_sha256,
        event.delivery_attempts, event.lease_owner_id,
        event.lease_generation, event.lease_heartbeat_at, event.lease_until,
        event.next_attempt_at
)
SELECT *
FROM claimed
ORDER BY next_attempt_at, event_sequence, event_id;
