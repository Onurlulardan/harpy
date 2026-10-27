INSERT INTO platform.dead_letters (
    id, root_id, entry_sequence, record_kind,
    tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin,
    source_kind, source_id, consumer_name, consumer_version,
    source_contract, source_version, source_sha256,
    reason_class, reason_code, redacted_error_sha256,
    payload_canonical, payload_sha256, payload_redacted_at,
    redacted_envelope,
    action_reason, action_ticket, action_actor_id, action_causation_id,
    full_payload_expires_at, metadata_expires_at, recorded_at
)
SELECT
    $2, root.id, $3, 'retry_action',
    root.tenant_id, root.scope_kind, root.scope_id,
    root.authority_group_id, root.authority_group_generation,
    root.authority_domain, root.cutover_generation, 'v2',
    root.source_kind, root.source_id, root.consumer_name, root.consumer_version,
    root.source_contract, root.source_version, root.source_sha256,
    root.reason_class, root.reason_code, root.redacted_error_sha256,
    NULL, root.payload_sha256, NULL,
    root.redacted_envelope,
    $4, $5, $6, $7,
    root.full_payload_expires_at, root.metadata_expires_at,
    pg_catalog.clock_timestamp()
FROM platform.dead_letters AS root
WHERE root.id = $1
  AND root.record_kind = 'root'
RETURNING recorded_at;
