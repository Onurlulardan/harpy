SELECT * FROM analytics_adoption_accept_v1(
    $1::name,$2::uuid,$3::uuid,$4::uuid,$5::bigint,$6::bigint,$7::bigint,
    $8::bigint,$9::bytea,$10::timestamptz
)
