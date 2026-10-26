SELECT result_revision, replay, outcome
FROM platform.tenancy_settings_update_v1(
    $1::uuid,
    $2::text,
    $3::bigint,
    $4::text,
    $5::bytea
);
