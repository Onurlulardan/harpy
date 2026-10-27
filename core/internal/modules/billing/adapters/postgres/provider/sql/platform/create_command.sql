SELECT command_id, finalizer_job_id, tenant_id, actor_id,
       provider_account_id, provider_object_mapping_id, subscription_id,
       provider_key, command_kind, effect_key, request_sha256, state,
       requested_at, revision, replay
FROM platform.billing_create_provider_command_v1(
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
);
