SELECT analytics_dashboard_revoke_acl_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::text, $4::bytea, $5::uuid,
    $6::text, $7::uuid, $8::bigint, $9::timestamptz
);
