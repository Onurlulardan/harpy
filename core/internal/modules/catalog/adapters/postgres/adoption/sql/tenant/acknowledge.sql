WITH candidate AS (
    INSERT INTO catalog_legacy_consumer_acks (
        route_key, consumer_id, evidence_sha256, acknowledged_at
    ) VALUES (
        $1::text, $2::text, $3::text, $4::timestamptz
    ) ON CONFLICT DO NOTHING
    RETURNING route_key, consumer_id, evidence_sha256
), exact AS (
    SELECT route_key, consumer_id, evidence_sha256 FROM candidate
    UNION ALL
    SELECT route_key, consumer_id, evidence_sha256
    FROM catalog_legacy_consumer_acks
    WHERE route_key = $1::text AND consumer_id = $2::text
    LIMIT 1
), refreshed AS (
    UPDATE catalog_legacy_route_hits AS route
    SET consumer_ack_count = (
        SELECT count(*) FROM catalog_legacy_consumer_acks AS ack
        WHERE ack.route_key = route.route_key
    )
    WHERE route.route_key = $1::text AND route.retired_at IS NULL
      AND (SELECT evidence_sha256 FROM exact) = $3::text
      AND (SELECT count(*) FROM catalog_legacy_consumer_acks AS ack
           WHERE ack.route_key = route.route_key) <= route.required_consumer_count
    RETURNING route.route_key
)
SELECT exact.route_key, exact.consumer_id, exact.evidence_sha256
FROM exact JOIN refreshed ON refreshed.route_key = exact.route_key
