SELECT dashboard_id, version_id, version_no, root_revision, version_revision,
       lifecycle_state, version_state, name, visibility, layout_document,
       layout_sha256, manifest_sha256, widgets, created_at, updated_at
FROM analytics_dashboard_get_default_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,
    $2::text
);
