SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_identity_password_enroll_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::bytea, $5::bytea,
    $6::uuid, $7::uuid, $8::bigint, $9::bigint
);
