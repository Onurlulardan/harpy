WITH publish_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS observed_at
), released AS (
    UPDATE platform.outbox AS event
    SET delivery_state = 'pending',
        lease_owner_id = NULL,
        lease_heartbeat_at = NULL,
        lease_until = NULL,
        next_attempt_at = publish_clock.observed_at +
            ($10 * interval '1 microsecond'),
        last_error_code = $11,
        last_error_sha256 = $12
    FROM publish_clock
    WHERE event.event_id = $7
      AND event.scope_kind = $1
      AND event.scope_id = $2
      AND event.authority_group_id = $3
      AND event.authority_group_generation = $4
      AND event.authority_domain = $5
      AND event.cutover_generation = $6
      AND event.delivery_state = 'publishing'
      AND event.lease_owner_id = $8
      AND event.lease_generation = $9
      AND event.lease_until > publish_clock.observed_at
    RETURNING event.next_attempt_at
)
SELECT next_attempt_at
FROM released;
