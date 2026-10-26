SELECT response_document,replayed,admitted FROM platform.support_submit_contact_v1(
$1::uuid,$2::text,$3::text,$4::text,$5::text,$6::text,$7::text,$8::text,$9::text,
$10::bytea,$11::text,$12::bytea,$13::timestamp with time zone)
