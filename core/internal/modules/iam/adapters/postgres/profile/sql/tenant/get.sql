SELECT result_id, email, status, revision, created_at, updated_at,
       roles, organizations
FROM iam_self_profile_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bigint, $3::bigint
);
