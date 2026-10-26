SELECT resource_id, actual_schema_version_id, source_revision, profile_sha256
FROM iam_maintenance_profile_lineage_resources_read_v3(
    pg_catalog.current_schema(), $1, $2, $3, $4, $5, $6, $7, $8
)
