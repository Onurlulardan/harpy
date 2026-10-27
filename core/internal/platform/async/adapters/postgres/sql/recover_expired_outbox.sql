WITH recovery_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS observed_at
), candidate AS MATERIALIZED (
    SELECT event.event_id
    FROM platform.outbox AS event
    CROSS JOIN recovery_clock
    WHERE event.scope_kind = $1
      AND event.scope_id = $2
      AND event.authority_group_id = $3
      AND event.authority_group_generation = $4
      AND event.authority_domain = $5
      AND event.cutover_generation = $6
      AND event.delivery_state = 'publishing'
      AND event.parked_at IS NULL
      AND event.lease_until <= recovery_clock.observed_at
    ORDER BY event.lease_until, event.event_id
    FOR UPDATE OF event SKIP LOCKED
    LIMIT $7
), recovered AS (
    UPDATE platform.outbox AS event
    SET delivery_state = 'pending',
        lease_owner_id = NULL,
        lease_heartbeat_at = NULL,
        lease_until = NULL,
        next_attempt_at = recovery_clock.observed_at,
        last_error_code = 'OUTBOX_LEASE_EXPIRED',
        last_error_sha256 = pg_catalog.encode(pg_catalog.sha256(
            pg_catalog.convert_to(
                'outbox-lease-expired.v1' || pg_catalog.chr(10) ||
                event.event_id::text || pg_catalog.chr(10) ||
                event.lease_generation::text,
                'UTF8'
            )
        ), 'hex')
    FROM candidate, recovery_clock
    WHERE event.event_id = candidate.event_id
    RETURNING event.event_id
)
SELECT pg_catalog.count(*)::bigint
FROM recovered;
