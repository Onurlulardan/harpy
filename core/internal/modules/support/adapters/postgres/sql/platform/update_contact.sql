SELECT response_document,replayed FROM platform.support_update_contact_v1(
$1::uuid,$2::bigint,$3::bigint,$4::uuid,$5::boolean,$6::text,$7::boolean,$8::text,
$9::bigint,$10::text,$11::bytea,$12::timestamp with time zone)
