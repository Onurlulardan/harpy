SELECT * FROM iam_rbac_assignment_update_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::uuid, $5::text, $6::uuid,
    $7::jsonb, $8::bigint, $9::uuid, $10::uuid, $11::bigint, $12::bigint,
    $13::bytea, $14::timestamptz, $15::boolean, $16::boolean, $17::boolean
);
