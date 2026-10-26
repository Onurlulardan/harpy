WITH publish_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS observed_at
), renewed AS (
    UPDATE platform.outbox AS event
    SET lease_heartbeat_at = publish_clock.observed_at,
        lease_until = publish_clock.observed_at + interval '60 seconds'
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
    RETURNING event.lease_heartbeat_at, event.lease_until
)
SELECT lease_heartbeat_at, lease_until
FROM renewed;
