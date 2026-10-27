SELECT cutover_authority.record_full_rollout_completion_v1(
    $1::uuid,$2::uuid,$3::uuid,
    $4::text,$5::text,$6::text,$7::text,$8::text,$9::text,
    $10::text,$11::text,$12::text,
    $13::timestamptz,$14::timestamptz,$15::timestamptz
)
