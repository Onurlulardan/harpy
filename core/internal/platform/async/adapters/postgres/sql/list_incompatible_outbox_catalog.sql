SELECT event.event_type, event.event_version::bigint, event.authority_domain, event.payload_schema_sha256
FROM platform.outbox AS event
WHERE event.delivery_state IN ('pending', 'publishing')
  AND NOT EXISTS (
      SELECT 1
      FROM pg_catalog.generate_subscripts($1::text[], 1) AS policy(position)
      WHERE ($1::text[])[policy.position] = event.event_type
        AND ($2::integer[])[policy.position] = event.event_version
        AND ($3::text[])[policy.position] = event.authority_domain
        AND ($4::text[])[policy.position] = event.payload_schema_sha256
  )
ORDER BY event.event_type, event.event_version, event.event_id
LIMIT 8;
