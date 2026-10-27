SELECT
    completed_at,
    action_id,
    result_job_id,
    result_root_id,
    created
FROM platform.platform_async_dispose_stale_job_v1(
    $1::uuid,$2::text,$3::uuid,$4::uuid,$5::uuid,$6::uuid,
    $7::text,$8::uuid,$9::text,$10::bigint,$11::text,$12::bigint,
    $13::text,$14::text,$15::bytea,$16::text,$17::text,$18::text,
    $19::uuid,$20::uuid,$21::text,$22::text,$23::text
);
