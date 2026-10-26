SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_identity_assignment_update_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::bigint, $5::text,
    $6::uuid, $7::uuid, $8::bigint, $9::bigint
);
