SELECT response_document,replayed FROM platform.support_update_bug_v1(
$1::uuid,$2::bigint,$3::bigint,$4::boolean,$5::uuid,$6::uuid,$7::boolean,$8::text,
$9::boolean,$10::text,$11::boolean,$12::text,$13::boolean,$14::text,$15::bigint,
$16::text,$17::bytea,$18::timestamp with time zone)
