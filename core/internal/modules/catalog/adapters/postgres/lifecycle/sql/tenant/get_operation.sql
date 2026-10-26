SELECT operation_id, installation_id, package_id, kind, state, plan,
       plan_sha256, retry_of_operation_id, started_at,
       destructive_started_at, failure_code, revision
FROM catalog_operation_runs
WHERE operation_id = $1::uuid
