SELECT run_document,replayed FROM platform.support_start_adoption_v1(
$1::uuid,$2::uuid,$3::text,$4::uuid,$5::text,$6::text,$7::oid,$8::text,$9::text,
$10::bytea,$11::bytea,$12::bytea,$13::text,$14::bigint,$15::bytea,$16::timestamp with time zone)
