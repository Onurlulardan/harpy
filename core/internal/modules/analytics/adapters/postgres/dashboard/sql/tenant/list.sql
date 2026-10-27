SELECT dashboard_id, version_id, version_no, root_revision, version_revision,
       lifecycle_state, version_state, name, visibility, layout_document,
       layout_sha256, manifest_sha256, widgets, created_at, updated_at
FROM analytics_dashboard_list_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1::integer, $2::uuid
);
