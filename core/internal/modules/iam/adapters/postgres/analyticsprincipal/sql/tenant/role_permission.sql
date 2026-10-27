SELECT iam_analytics_role_permission_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,
    $2::uuid,
    $3::text,
    $4::text
);
