SELECT fact_id,replayed FROM platform.usage_record_fact_v1(
$1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::bytea,$6::bytea,$7::text,$8::text,
$9::integer,$10::bigint,$11::timestamp with time zone)
