WITH retry_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS observed_at
), source_job AS MATERIALIZED (
    SELECT source.*
    FROM platform.dead_letters AS terminal_root
    JOIN platform.jobs AS source
      ON source.id = terminal_root.source_id
     AND source.tenant_id = terminal_root.tenant_id
    CROSS JOIN retry_clock
    WHERE terminal_root.id = $1
      AND terminal_root.record_kind = 'root'
      AND terminal_root.source_kind = 'job'
      AND source.state = 'dead'
      AND source.payload_ref IS NOT NULL
      AND source.payload_redacted_at IS NULL
      AND source.payload_sha256 = terminal_root.source_sha256
    FOR UPDATE OF source
), gap_reference AS MATERIALIZED (
    SELECT
        source.id AS source_job_id,
        pg_catalog.substr(
            source.payload_ref,
            pg_catalog.length('deadletter:event-gap:') + 1
        )::uuid AS gap_root_id
    FROM source_job AS source
    WHERE source.kind =
            'platform_async.event_gap_reconcile.' || source.authority_domain
      AND source.payload_ref ~
            '^deadletter:event-gap:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
), locked_gap_root AS MATERIALIZED (
    SELECT reference.source_job_id, gap_root.id, gap_root.source_id
    FROM gap_reference AS reference
    JOIN source_job AS source ON source.id = reference.source_job_id
    JOIN platform.dead_letters AS gap_root ON gap_root.id = reference.gap_root_id
    CROSS JOIN retry_clock
    WHERE gap_root.root_id = gap_root.id
      AND gap_root.entry_sequence = 0
      AND gap_root.record_kind = 'root'
      AND gap_root.source_kind = 'event'
      AND gap_root.reason_class = 'sequence_gap'
      AND gap_root.reason_code = 'EVENT_SEQUENCE_GAP'
      AND gap_root.payload_canonical IS NOT NULL
      AND gap_root.payload_redacted_at IS NULL
      AND gap_root.full_payload_expires_at > retry_clock.observed_at
      AND gap_root.metadata_expires_at > retry_clock.observed_at
      AND gap_root.source_sha256 = source.payload_sha256
      AND gap_root.payload_sha256 = source.payload_sha256
      AND ROW(
          gap_root.tenant_id,gap_root.scope_kind,gap_root.scope_id,
          gap_root.authority_group_id,gap_root.authority_group_generation,
          gap_root.authority_domain,gap_root.cutover_generation,gap_root.writer_origin
      ) = ROW(
          source.tenant_id,source.scope_kind,source.scope_id,
          source.authority_group_id,source.authority_group_generation,
          source.authority_domain,source.cutover_generation,'v2'
      )
), locked_gap_event AS MATERIALIZED (
    SELECT locked.source_job_id
    FROM locked_gap_root AS locked
    JOIN source_job AS source ON source.id = locked.source_job_id
    JOIN platform.dead_letters AS gap_root ON gap_root.id = locked.id
    JOIN platform.outbox AS event ON event.event_id = locked.source_id
    CROSS JOIN LATERAL (
        SELECT pg_catalog.convert_from(
            gap_root.payload_canonical, 'UTF8'
        )::jsonb AS value
    ) AS gap
    WHERE event.tenant_id = gap_root.tenant_id
      AND event.scope_kind = gap_root.scope_kind
      AND event.scope_id = gap_root.scope_id
      AND event.delivery_state = 'delivered'
      AND event.payload_canonical IS NOT NULL
      AND event.payload_redacted_at IS NULL
      AND event.parked_at IS NULL
      AND gap.value->>'contract_version' = 'event-sequence-gap.v1'
      AND gap.value->>'consumer' = gap_root.consumer_name
      AND (gap.value->>'consumer_version')::integer = gap_root.consumer_version
      AND (gap.value->>'received_sequence')::bigint = event.event_sequence
      AND gap.value->'envelope'->>'event_id' = event.event_id::text
      AND gap.value->'envelope'->>'event_type' = event.event_type
      AND (gap.value->'envelope'->>'event_version')::integer = event.event_version
      AND gap.value->'envelope'->>'aggregate_type' = event.aggregate_type
      AND gap.value->'envelope'->>'aggregate_id' = event.aggregate_id::text
      AND (gap.value->'envelope'->>'event_sequence')::bigint = event.event_sequence
      AND gap.value->'envelope'->>'payload_sha256' = event.payload_sha256
      AND gap.value->'envelope'->'payload' =
          pg_catalog.convert_from(event.payload_canonical, 'UTF8')::jsonb
      AND gap.value->'envelope'->'producer_authority'->>'scope_kind' =
          event.scope_kind
      AND gap.value->'envelope'->'producer_authority'->>'scope_id' =
          event.scope_id::text
      AND gap.value->'envelope'->'producer_authority'->>'group_id' =
          event.authority_group_id
      AND (gap.value->'envelope'->'producer_authority'->>'group_generation')::bigint =
          event.authority_group_generation
      AND gap.value->'envelope'->'producer_authority'->>'domain' =
          event.authority_domain
      AND (gap.value->'envelope'->'producer_authority'->>'cutover_generation')::bigint =
          event.cutover_generation
      AND gap.value->'envelope'->'producer_authority'->>'writer_origin' =
          event.writer_origin
    FOR UPDATE OF event
), current_gap_event AS MATERIALIZED (
    SELECT
        locked.source_job_id,
        platform.platform_async_assert_current_authority_v1(
            event.scope_kind,event.scope_id,event.authority_group_id,
            event.authority_group_generation,event.authority_domain,
            event.cutover_generation
        ) AS authority_assertion
    FROM locked_gap_event AS locked
    JOIN locked_gap_root AS gap_root
      ON gap_root.source_job_id = locked.source_job_id
    JOIN platform.outbox AS event ON event.event_id = gap_root.source_id
), eligible_source AS MATERIALIZED (
    SELECT source.*
    FROM source_job AS source
    WHERE pg_catalog.left(
        source.kind,
        pg_catalog.length('platform_async.event_gap_reconcile.')
    ) <> 'platform_async.event_gap_reconcile.'
    UNION ALL
    SELECT source.*
    FROM source_job AS source
    JOIN current_gap_event AS gap ON gap.source_job_id = source.id
)
INSERT INTO platform.jobs (
    id, tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin,
    kind, payload_revision, payload_ref, payload_sha256, dedupe_key,
    effect_class, effect_idempotency_key, priority, run_at,
    state, attempts, max_attempts, lease_generation, retry_of_job_id
)
SELECT
    $2, source.tenant_id, source.scope_kind, source.scope_id,
    source.authority_group_id, source.authority_group_generation,
    source.authority_domain, source.cutover_generation, 'v2',
    source.kind, source.payload_revision, source.payload_ref,
    source.payload_sha256, $3,
    source.effect_class, source.effect_idempotency_key,
    source.priority, retry_clock.observed_at,
    'queued', 0, source.max_attempts, 0, source.id
FROM eligible_source AS source
CROSS JOIN retry_clock
RETURNING id;
