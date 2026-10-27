SELECT * FROM iam_profile_recovery_parent_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::uuid, $4::bigint, $5::bigint
);
