SELECT operation_id, operation_state, snapshot_sequence,
       snapshot_sha256, failure_code, created_at, completed_at
FROM platform.billing_platform_license_refresh_read_v1($1, $2);
