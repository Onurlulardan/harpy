SELECT iam_identity_bulk_delete_reserve_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::text, $4::bytea, $5::integer
);
