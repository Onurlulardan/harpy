SELECT result_id, result_revision, replayed
FROM iam_profile_backfill_plan_create_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::bytea, $4::bytea, $5::bytea,
    $6::uuid, $7::uuid, $8::bigint, $9::bigint
);
