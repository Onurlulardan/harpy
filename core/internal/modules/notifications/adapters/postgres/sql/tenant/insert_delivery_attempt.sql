INSERT INTO notifications_delivery_attempts (
    id, delivery_id, attempt_number, lease_generation, provider_kind,
    response_class, provider_reference_digest, status_code, error_code,
    started_at, finished_at
) VALUES (
    $1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9, $10, $11
);
