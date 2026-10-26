INSERT INTO documents_compatibility_hits (
    route_contract, consumer_id, build_id, hit_count, first_hit_at, last_hit_at
)
VALUES ($1, $2, $3, 1, $4, $4)
ON CONFLICT (route_contract, consumer_id, build_id) DO UPDATE
SET hit_count = documents_compatibility_hits.hit_count + 1,
    last_hit_at = excluded.last_hit_at
WHERE documents_compatibility_hits.retired_at IS NULL
RETURNING route_contract, consumer_id, build_id, hit_count, first_hit_at, last_hit_at;
