SELECT decision, command_id, actor_id, provider_key, active_job_id,
       finalizer_job_id, request_sha256, state, revision, occurred_at,
       provider_reference, error_code
FROM platform.billing_finalize_provider_command_v1(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
);
