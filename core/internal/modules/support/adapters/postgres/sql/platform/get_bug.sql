SELECT response_document FROM platform.support_get_bug_v1(
$1::uuid,$2::bigint,$3::bigint,$4::uuid,$5::uuid,$6::uuid)
