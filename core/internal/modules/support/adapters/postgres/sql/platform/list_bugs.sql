SELECT items,total FROM platform.support_list_bugs_v1(
$1::uuid,$2::bigint,$3::bigint,$4::boolean,$5::uuid,$6::uuid,$7::text,$8::text,
$9::text,$10::integer,$11::integer)
