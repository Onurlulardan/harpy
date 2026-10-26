WITH publish_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS observed_at
)
SELECT
    group_row.scope_kind,
    group_row.scope_id,
    group_row.group_id,
    group_row.generation,
    domain_row.domain,
    domain_row.generation
FROM platform.cutover_group_members AS member
JOIN platform.cutover_groups AS group_row
  ON group_row.scope_kind = member.scope_kind
 AND group_row.scope_id = member.scope_id
 AND group_row.group_id = member.group_id
JOIN platform.domain_cutovers AS domain_row
  ON domain_row.scope_kind = member.scope_kind
 AND domain_row.scope_id = member.scope_id
 AND domain_row.domain = member.domain
CROSS JOIN publish_clock
WHERE group_row.state = 'v2'
  AND group_row.writer = 'v2'
  AND domain_row.state = 'v2'
  AND domain_row.writer = 'v2'
  AND EXISTS (
      SELECT 1
      FROM platform.outbox AS event
      WHERE event.scope_kind = group_row.scope_kind
        AND event.scope_id = group_row.scope_id
        AND event.authority_group_id = group_row.group_id
        AND event.authority_domain = domain_row.domain
        AND event.parked_at IS NULL
        AND event.authority_group_generation <= group_row.generation
        AND event.cutover_generation <= domain_row.generation
        AND (
            (
                event.authority_group_generation = group_row.generation AND
                event.cutover_generation = domain_row.generation AND
                ((event.delivery_state = 'pending' AND
                    event.next_attempt_at <= publish_clock.observed_at) OR
                 (event.delivery_state = 'publishing' AND
                    event.lease_until <= publish_clock.observed_at))
            ) OR (
                (event.authority_group_generation < group_row.generation OR
                 event.cutover_generation < domain_row.generation) AND
                (event.delivery_state = 'pending' OR
                 (event.delivery_state = 'publishing' AND
                    event.lease_until <= publish_clock.observed_at))
            )
        )
  )
  AND (
      NOT $1::boolean OR
      ROW(
          group_row.scope_kind COLLATE "C", group_row.scope_id,
          group_row.group_id COLLATE "C", domain_row.domain COLLATE "C"
      ) > ROW(
          $2::text COLLATE "C", $3::uuid,
          $4::text COLLATE "C", $5::text COLLATE "C"
      )
  )
ORDER BY
    group_row.scope_kind COLLATE "C", group_row.scope_id,
    group_row.group_id COLLATE "C", domain_row.domain COLLATE "C"
LIMIT $6;
