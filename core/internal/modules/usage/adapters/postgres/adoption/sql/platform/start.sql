SELECT document,replayed FROM platform.usage_adoption_start_v1(
$1::uuid,$2::uuid,$3::text,$4::text,$5::text,$6::oid,$7::bigint,$8::bytea,
$9::text,$10::bytea,$11::timestamp with time zone)
