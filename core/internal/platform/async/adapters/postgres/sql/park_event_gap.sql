WITH gap_document AS MATERIALIZED (
    SELECT pg_catalog.convert_from($19::bytea, 'UTF8')::jsonb AS value
), locked_source AS MATERIALIZED (
    SELECT event.event_id
    FROM platform.outbox AS event
    CROSS JOIN gap_document AS gap
    WHERE event.event_id = $10::uuid
      AND event.tenant_id = $3::uuid
      AND event.scope_kind = $4::text
      AND event.scope_id = $5::uuid
      AND event.delivery_state = 'delivered'
      AND event.payload_canonical IS NOT NULL
      AND event.payload_redacted_at IS NULL
      AND event.event_id::text = gap.value->'envelope'->>'event_id'
      AND event.event_type = gap.value->'envelope'->>'event_type'
      AND event.event_version =
          (gap.value->'envelope'->>'event_version')::integer
      AND event.aggregate_type = gap.value->'envelope'->>'aggregate_type'
      AND event.aggregate_id::text = gap.value->'envelope'->>'aggregate_id'
      AND event.event_sequence =
          (gap.value->'envelope'->>'event_sequence')::bigint
      AND event.payload_sha256 = gap.value->'envelope'->>'payload_sha256'
      AND pg_catalog.convert_from(event.payload_canonical, 'UTF8')::jsonb =
          gap.value->'envelope'->'payload'
    FOR UPDATE OF event
), gap_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS recorded_at
), inserted_root AS (
    INSERT INTO platform.dead_letters (
        id,root_id,entry_sequence,record_kind,tenant_id,scope_kind,scope_id,
        authority_group_id,authority_group_generation,authority_domain,
        cutover_generation,writer_origin,source_kind,source_id,consumer_name,
        consumer_version,source_contract,source_version,source_sha256,
        reason_class,reason_code,redacted_error_sha256,payload_canonical,
        payload_sha256,redacted_envelope,full_payload_expires_at,
        metadata_expires_at,recorded_at
    ) SELECT
        $1,$1,0,'root',$3,$4,$5,$6,$7,$8,$9,'v2','event',$10,$11,$12,
        $13,$14,$15,$16,$17,$18,$19,$20,$21::jsonb,
        gap_clock.recorded_at + interval '30 days',
        gap_clock.recorded_at + interval '365 days',gap_clock.recorded_at
    FROM gap_clock CROSS JOIN locked_source
    ON CONFLICT (id) DO NOTHING
    RETURNING id
), exact_root AS MATERIALIZED (
    SELECT inserted_root.id
    FROM inserted_root
    UNION ALL
    SELECT root.id
    FROM platform.dead_letters AS root
    WHERE root.id = $1 AND root.root_id = $1 AND root.entry_sequence = 0
      AND root.record_kind = 'root' AND root.tenant_id = $3
      AND ROW(root.scope_kind,root.scope_id,root.authority_group_id,
          root.authority_group_generation,root.authority_domain,
          root.cutover_generation,root.writer_origin)
          = ROW($4::text,$5::uuid,$6::text,$7::bigint,$8::text,$9::bigint,'v2')
      AND ROW(root.source_kind,root.source_id,root.consumer_name,
          root.consumer_version,root.source_contract,root.source_version,
          root.source_sha256,root.reason_class,root.reason_code,
          root.redacted_error_sha256,root.payload_sha256)
          = ROW('event',$10::uuid,$11::text,$12::integer,$13::text,$14::integer,
              $15::text,$16::text,$17::text,$18::text,$20::text)
      AND root.payload_canonical = $19::bytea
      AND root.payload_redacted_at IS NULL
      AND root.redacted_envelope = $21::jsonb
      AND NOT EXISTS (SELECT 1 FROM inserted_root)
), inserted_job AS (
    INSERT INTO platform.jobs (
        id,tenant_id,scope_kind,scope_id,authority_group_id,
        authority_group_generation,authority_domain,cutover_generation,
        writer_origin,kind,payload_revision,payload_ref,payload_sha256,
        dedupe_key,effect_class,priority,run_at,state,attempts,max_attempts,
        lease_generation
    ) SELECT
        $2,$3,$4,$5,$6,$7,$8,$9,'v2',$22,$23,$24,$20,$25,
        'database_transactional',100,gap_clock.recorded_at,'queued',0,$26,0
    FROM gap_clock CROSS JOIN exact_root
    ON CONFLICT DO NOTHING
    RETURNING id
), exact_job AS (
    SELECT inserted_job.id
    FROM inserted_job
    UNION ALL
    SELECT job.id
    FROM platform.jobs AS job
    WHERE job.id = $2 AND job.tenant_id = $3
      AND ROW(job.scope_kind,job.scope_id,job.authority_group_id,
          job.authority_group_generation,job.authority_domain,
          job.cutover_generation,job.writer_origin)
          = ROW($4::text,$5::uuid,$6::text,$7::bigint,$8::text,$9::bigint,'v2')
      AND ROW(job.kind,job.payload_revision,job.payload_ref,job.payload_sha256,
          job.dedupe_key,job.effect_class,job.max_attempts)
          = ROW($22::text,$23::integer,$24::text,$20::text,$25::text,
              'database_transactional',$26::integer)
      AND NOT EXISTS (SELECT 1 FROM inserted_job)
)
SELECT exact_root.id,exact_job.id,EXISTS (SELECT 1 FROM inserted_job)
FROM exact_root CROSS JOIN exact_job;
