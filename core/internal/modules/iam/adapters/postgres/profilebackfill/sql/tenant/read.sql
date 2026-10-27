SELECT resource_id, source_schema_version_id, source_revision,
       profile_values, profile_sha256
FROM iam_profile_backfill_source_resources_plan_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::text, $2::uuid, $3::uuid, $4::integer,
    $5::uuid, $6::uuid, $7::bigint, $8::bigint
);
