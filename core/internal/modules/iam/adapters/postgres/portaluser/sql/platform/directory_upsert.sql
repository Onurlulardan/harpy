SELECT platform.iam_login_directory_upsert_v2(
    $1::uuid,$2::bigint,$3::bigint,$4::uuid,$5::uuid,
    $6::text,$7::text,$8::bytea,$9::bigint,$10::bytea
);
