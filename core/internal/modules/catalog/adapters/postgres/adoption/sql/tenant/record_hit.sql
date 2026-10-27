INSERT INTO catalog_legacy_route_hits (
    route_key, hit_count, last_hit_at, consumer_ack_count,
    required_consumer_count, remove_after
) VALUES (
    $1::text, 1, $4::timestamptz, 0, $2::bigint, $3::timestamptz
)
ON CONFLICT (route_key) DO UPDATE
SET hit_count = catalog_legacy_route_hits.hit_count + 1,
    last_hit_at = EXCLUDED.last_hit_at,
    remove_after = GREATEST(catalog_legacy_route_hits.remove_after, EXCLUDED.remove_after)
WHERE catalog_legacy_route_hits.required_consumer_count = $2::bigint
  AND catalog_legacy_route_hits.retired_at IS NULL
RETURNING route_key
