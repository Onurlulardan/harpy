SELECT actor_principals, authorization_clock, target_exists, target_revision
FROM iam_work_principal_evidence_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,
    $2::uuid,
    $3::text,
    $4::uuid
);
