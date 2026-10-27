INSERT INTO platform.tenant_migration_requirements (
    migration_run_id, ordinal, owner, logical_version, database_origin,
    selected_path, selected_sha256
)
SELECT $1, requirement.ordinal, requirement.owner, requirement.logical_version,
       requirement.database_origin, requirement.selected_path,
       requirement.selected_sha256
FROM pg_catalog.jsonb_to_recordset($2::jsonb) AS requirement(
    ordinal integer, owner text, logical_version bigint, database_origin text,
    selected_path text, selected_sha256 text
);
