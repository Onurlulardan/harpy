SELECT platform.iam_identity_directory_upsert_v2(
    $1::uuid, $2::bigint, $3::bigint, $4::uuid, $5::uuid,
    $6::text, $7::bytea, $8::bigint, $9::bytea
)
