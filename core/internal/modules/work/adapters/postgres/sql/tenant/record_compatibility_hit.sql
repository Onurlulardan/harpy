INSERT INTO work_compatibility_hits (
    route_contract, consumer_id, build_id, hit_count,
    first_hit_at, last_hit_at, retired_at
) VALUES ($1, $2, $3, 1, transaction_timestamp(), transaction_timestamp(), NULL)
ON CONFLICT (route_contract, consumer_id, build_id) DO UPDATE
SET hit_count = work_compatibility_hits.hit_count + 1,
    last_hit_at = transaction_timestamp()
WHERE work_compatibility_hits.retired_at IS NULL;
