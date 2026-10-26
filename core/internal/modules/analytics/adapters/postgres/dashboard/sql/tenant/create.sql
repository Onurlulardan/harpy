SELECT dashboard_id, version_id, version_no, root_revision, version_revision,
       lifecycle_state, version_state, name, visibility, layout_document,
       layout_sha256, manifest_sha256, widgets, created_at, updated_at, replayed
FROM analytics_dashboard_create_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::text, $4::bytea, $5::uuid, $6::uuid,
    $7::text, $8::text, $9::bytea, $10::jsonb, $11::timestamptz
);
