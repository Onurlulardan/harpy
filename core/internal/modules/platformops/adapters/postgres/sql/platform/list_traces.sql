SELECT * FROM platform.platformops_list_traces_v1(
    $1::uuid,$2::text,$3::text,$4::text,$5::integer,$6::integer,$7::integer,
    $8::numeric,$9::numeric,$10::timestamptz,$11::timestamptz,$12::text,
    $13::text,$14::text,$15::integer,$16::integer
)
