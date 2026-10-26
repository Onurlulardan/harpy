SELECT * FROM iam_rbac_assignment_create_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::uuid, $5::text, $6::uuid,
    $7::jsonb, $8::uuid, $9::uuid, $10::bigint, $11::bigint,
    $12::bytea, $13::timestamptz, $14::boolean, $15::boolean, $16::boolean
);
