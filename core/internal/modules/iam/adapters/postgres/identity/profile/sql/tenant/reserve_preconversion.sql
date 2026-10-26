SELECT iam_profile_preconversion_reserve_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,
    $2::text,
    $3::uuid,
    $4::uuid,
    $5::bigint,
    $6::bytea,
    $7::bytea,
    $8::bytea
);
