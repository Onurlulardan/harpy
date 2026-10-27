SELECT * FROM iam_rbac_action_list_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1::text, $2::integer, $3::uuid, $4::uuid
);
