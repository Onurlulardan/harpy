SELECT target_user_id, subject_id, role_id, subject_role_binding_id,
       role_permission_binding_id, target_authorization_version,
       platform_clock, replayed
FROM iam_platform_operator_revoke_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::bigint, $5::bigint, $6::bigint,
    $7::bytea, $8::timestamptz, $9::text, $10::boolean, $11::boolean,
    $12::boolean, $13::uuid, $14::bigint, $15::text
);
