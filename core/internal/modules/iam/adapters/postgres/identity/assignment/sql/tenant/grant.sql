SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_identity_assignment_grant_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::uuid, $5::uuid, $6::text,
    $7::uuid, $8::uuid, $9::uuid, $10::uuid, $11::uuid, $12::bigint, $13::bigint
);
