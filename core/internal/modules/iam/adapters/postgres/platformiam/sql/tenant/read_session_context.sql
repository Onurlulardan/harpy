SELECT deployment_mode, tenant_purpose, subject_id,
       platform_clock, authorization_version
FROM iam_platform_session_context_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bigint, $3::bigint
);
