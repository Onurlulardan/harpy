SELECT platform.iam_maintenance_directory_apply_v1(
    $1::uuid, $2::uuid, $3::uuid, $4::text, $5::text,
    $6::bigint, $7::bytea, $8::jsonb
)
