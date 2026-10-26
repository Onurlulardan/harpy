UPDATE catalog_operation_runs
SET state = $2::text,
    destructive_started_at = $3::timestamptz,
    failure_code = $4::text,
    revision = $5::bigint
WHERE operation_id = $1::uuid AND revision = $6::bigint
RETURNING operation_id, installation_id, package_id, kind, state, plan,
          plan_sha256, retry_of_operation_id, started_at,
          destructive_started_at, failure_code, revision
