UPDATE catalog_legacy_route_hits
SET retired_at = $2::timestamptz
WHERE route_key = $1::text AND retired_at IS NULL
  AND remove_after <= $2::timestamptz
  AND consumer_ack_count = required_consumer_count
  AND (last_hit_at IS NULL OR last_hit_at < remove_after)
RETURNING route_key
