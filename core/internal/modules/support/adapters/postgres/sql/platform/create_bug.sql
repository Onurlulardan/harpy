SELECT response_document,replayed FROM platform.support_create_bug_v1(
$1::uuid,$2::uuid,$3::text,$4::text,$5::uuid,$6::text,$7::text,$8::bigint,$9::bigint,
$10::text,$11::text,$12::text,$13::text,$14::text,$15::text,$16::jsonb,$17::jsonb,
$18::jsonb,$19::text,$20::text,$21::bytea,$22::timestamp with time zone)
