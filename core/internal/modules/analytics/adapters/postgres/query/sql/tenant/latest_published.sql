SELECT query_id, version_id, version_no, root_revision, reference_epoch,
       version_revision, lifecycle_state, version_state, name, query_document,
       parameter_document, output_document, manifest_sha256, cache_policy,
       created_at, updated_at
FROM analytics_query_latest_published_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1::uuid
);
