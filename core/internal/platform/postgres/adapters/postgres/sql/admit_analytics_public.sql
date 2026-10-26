SELECT platform.analytics_public_admit_v1(
    $1::uuid, $2::uuid, $3::bigint, $4::inet, $5::timestamptz
);
