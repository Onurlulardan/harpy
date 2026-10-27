INSERT INTO notifications_operator_alerts (
    id, code, resource_id, created_at, acknowledged_at
) VALUES (
    pg_catalog.uuidv7(), $1, $2::uuid, pg_catalog.transaction_timestamp(), NULL
);
