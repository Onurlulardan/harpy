SELECT operation_id, predecessor_user_id, successor_user_id,
       predecessor_token_version, predecessor_authorization_version,
       successor_token_version, successor_authorization_version,
       owner_policy_revision, platform_clock, head_generation, replayed
FROM iam_owner_transfer_commit_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::bytea,$3::uuid,$4::bigint,$5::bigint,$6::bigint,$7::uuid,$8::bytea
);
