WITH dead_letter_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS recorded_at
), inserted AS (
    INSERT INTO platform.dead_letters (
        id, root_id, entry_sequence, record_kind,
        tenant_id, scope_kind, scope_id,
        authority_group_id, authority_group_generation,
        authority_domain, cutover_generation, writer_origin,
        source_kind, source_id, consumer_name, consumer_version,
        source_contract, source_version, source_sha256,
        reason_class, reason_code, redacted_error_sha256,
        payload_canonical, payload_sha256, redacted_envelope,
        full_payload_expires_at, metadata_expires_at, recorded_at
    )
    SELECT
        $1, $1, 0, 'root',
        $2, $3, $4,
        $5, $6,
        $7, $8, 'v2',
        'event', $9, $10, $11,
        $12, $13, $14,
        $15, $16, $17,
        $18, $19, $20::jsonb,
        dead_letter_clock.recorded_at + interval '30 days',
        dead_letter_clock.recorded_at + interval '365 days',
        dead_letter_clock.recorded_at
    FROM dead_letter_clock
    ON CONFLICT (id) DO NOTHING
    RETURNING id, recorded_at, true AS created
)
SELECT id, recorded_at, created
FROM inserted
UNION ALL
SELECT existing.id, existing.recorded_at, false
FROM platform.dead_letters AS existing
WHERE existing.id = $1
  AND NOT EXISTS (SELECT 1 FROM inserted)
  AND existing.root_id = $1
  AND existing.entry_sequence = 0
  AND existing.record_kind = 'root'
  AND existing.tenant_id IS NOT DISTINCT FROM $2::uuid
  AND ROW(
      existing.scope_kind, existing.scope_id,
      existing.authority_group_id, existing.authority_group_generation,
      existing.authority_domain, existing.cutover_generation,
      existing.writer_origin, existing.source_kind, existing.source_id,
      existing.consumer_name, existing.consumer_version,
      existing.source_contract, existing.source_version,
      existing.source_sha256, existing.reason_class, existing.reason_code,
      existing.redacted_error_sha256, existing.payload_sha256
  ) IS NOT DISTINCT FROM ROW(
      $3::text, $4::uuid, $5::text, $6::bigint,
      $7::text, $8::bigint, 'v2', 'event', $9::uuid,
      $10::text, $11::integer, $12::text, $13::integer,
      $14::text, $15::text, $16::text, $17::text, $19::text
  )
  AND existing.payload_canonical = $18::bytea
  AND existing.redacted_envelope = $20::jsonb
LIMIT 1;
