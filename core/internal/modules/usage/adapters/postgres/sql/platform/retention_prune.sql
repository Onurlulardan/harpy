SELECT platform.usage_retention_prune_v1(
    $1::uuid,$2::uuid,$3::timestamptz,$4::integer,$5::timestamptz
)
