SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_identity_membership_update_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::bigint, $5::text, $6::text,
    $7::uuid, $8::uuid, $9::bigint, $10::bigint
);
