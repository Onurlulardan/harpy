SELECT
    root.id,
    root.tenant_id,
    root.source_kind,
    root.source_id,
    root.consumer_name,
    root.consumer_version,
    root.source_contract,
    root.source_version,
    root.reason_class,
    root.reason_code,
    root.payload_canonical IS NOT NULL
        AND root.payload_redacted_at IS NULL
        AND pg_catalog.clock_timestamp() < root.full_payload_expires_at,
    root.redacted_envelope,
    root.recorded_at,
    root.full_payload_expires_at,
    root.metadata_expires_at,
    (
        SELECT pg_catalog.count(*)::integer
        FROM platform.dead_letters AS action
        WHERE action.root_id = root.id
          AND action.record_kind = 'retry_action'
    )
FROM platform.dead_letters AS root
WHERE root.record_kind = 'root'
  AND root.tenant_id = ANY($1::uuid[])
  AND (
      NOT $2::boolean OR
      ROW(root.recorded_at, root.id) < ROW($3::timestamptz, $4::uuid)
  )
ORDER BY root.recorded_at DESC, root.id DESC
LIMIT $5;
