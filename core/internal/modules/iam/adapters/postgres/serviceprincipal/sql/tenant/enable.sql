SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_service_principal_state_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::bigint, 'active'::text,
    $5::uuid, $6::uuid, $7::bigint, $8::bigint, $9::bytea,
    $10::timestamptz, $11::boolean, $12::boolean, $13::boolean
);
