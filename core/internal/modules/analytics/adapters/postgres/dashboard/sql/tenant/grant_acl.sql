SELECT dashboard_id, principal_kind, principal_id, access_level, row_revision, replayed
FROM analytics_dashboard_grant_acl_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid, $3::text, $4::bytea, $5::uuid,
    $6::text, $7::uuid, $8::text, $9::bigint, $10::timestamptz
);
