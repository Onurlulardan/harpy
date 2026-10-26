SELECT operation_id,anchor_id,challenge,rp_id,allowed_origins,expires_at,replayed
FROM platform.iam_recovery_anchor_options_v1(
    $1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::uuid,$6::text,$7::bytea,
    $8::text,$9::text,$10::jsonb,$11::bytea,$12::timestamptz,$13::timestamptz
);
