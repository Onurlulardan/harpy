WITH stale_candidates AS MATERIALIZED (
    SELECT event.event_id
    FROM platform.outbox AS event
    WHERE event.scope_kind = $1
      AND event.scope_id = $2
      AND event.authority_group_id = $3
      AND event.authority_domain = $5
      AND event.parked_at IS NULL
      AND event.authority_group_generation <= $4
      AND event.cutover_generation <= $6
      AND (event.authority_group_generation < $4 OR
           event.cutover_generation < $6)
      AND (
          event.delivery_state = 'pending' OR
          (event.delivery_state = 'publishing' AND
           event.lease_until <= pg_catalog.clock_timestamp())
      )
    ORDER BY event.created_at, event.event_id
    LIMIT $7
)
SELECT pg_catalog.count(*)::bigint
FROM stale_candidates AS candidate
WHERE platform.platform_async_park_stale_outbox_v1(
    candidate.event_id, pg_catalog.uuidv7()
);
