INSERT INTO catalog_operation_runs (
    operation_id, installation_id, package_id, kind, state, plan, plan_sha256,
    retry_of_operation_id, actor_id, idempotency_key, request_sha256,
    started_at, destructive_started_at, failure_code, revision
) VALUES (
    $1::uuid, $2::uuid, $3::uuid, $4::text, $5::text, $6::jsonb, $7::text,
    $8::uuid, $9::uuid, $10::text, $11::text, $12::timestamptz, NULL, NULL, $13::bigint
)
RETURNING operation_id, installation_id, package_id, kind, state, plan,
          plan_sha256, retry_of_operation_id, started_at,
          destructive_started_at, failure_code, revision
