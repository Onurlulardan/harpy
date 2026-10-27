SELECT allowed, capability_key, deployment_mode, subject_id,
       platform_clock, authorization_version, reason
FROM iam_platform_authorize_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bigint, $3::bigint, $4::bigint, $5::text
);
