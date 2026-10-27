SELECT platform.iam_installation_identity_directory_upsert_v1(
    $1::uuid, $2::uuid, $3::uuid, $4::name,
    1::bigint, 1::bigint, $5::uuid, $6::uuid, $7::text,
    $8::bytea, 1::bigint, $9::bytea
);
