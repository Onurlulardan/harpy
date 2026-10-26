INSERT INTO notifications_delivery_attempts (
    id, delivery_id, attempt_number, lease_generation, provider_kind,
    response_class, provider_reference_digest, status_code, error_code,
    started_at, finished_at
) VALUES (
    $1::uuid, $2::uuid, 1, 1, 'smtp', 'legacy_v1_accepted', $3::bytea,
    '', 'V1_ACCEPTED', $4, $4
);
