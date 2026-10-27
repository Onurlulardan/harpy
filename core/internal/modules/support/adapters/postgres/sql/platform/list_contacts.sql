SELECT items,total FROM platform.support_list_contacts_v1(
$1::uuid,$2::bigint,$3::bigint,$4::text,$5::text,$6::integer,$7::integer)
