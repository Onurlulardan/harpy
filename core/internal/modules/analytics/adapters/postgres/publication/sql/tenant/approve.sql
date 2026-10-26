SELECT release_id, resource_kind, resource_version_id, audience, lifecycle_state,
       cache_policy, manifest_sha256, requester_id, approver_id, row_revision,
       generation, public_id, capability_expires_at, created_at, updated_at, replayed
FROM analytics_publication_approve_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::text, $4::bytea, $5::uuid,
    $6::bigint, $7::timestamptz, $8::timestamptz
);
