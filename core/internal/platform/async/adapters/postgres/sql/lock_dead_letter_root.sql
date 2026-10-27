WITH dead_letter_lock AS MATERIALIZED (
    SELECT pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'harpy:deadletter:' || $1::uuid::text,
            0
        )
    )
)
SELECT
    root.id,
    root.tenant_id,
    root.scope_kind,
    root.scope_id,
    root.authority_group_id,
    root.authority_group_generation,
    root.authority_domain,
    root.cutover_generation,
    root.source_kind,
    root.source_id,
    root.consumer_name,
    root.consumer_version,
    root.source_contract,
    root.source_version,
    root.source_sha256,
    root.payload_canonical,
    root.payload_sha256,
    root.payload_redacted_at,
    root.redacted_envelope,
    root.full_payload_expires_at,
    root.metadata_expires_at,
    pg_catalog.clock_timestamp(),
    (
        SELECT pg_catalog.max(entry.entry_sequence) + 1
        FROM platform.dead_letters AS entry
        WHERE entry.root_id = root.id
    )
FROM platform.dead_letters AS root
CROSS JOIN dead_letter_lock
WHERE root.id = $1::uuid
  AND root.tenant_id = $2::uuid
  AND root.record_kind = 'root';
