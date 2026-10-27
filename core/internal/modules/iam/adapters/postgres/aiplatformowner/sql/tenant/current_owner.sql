SELECT environment_id, deployment_mode, owner_subject_id, platform_clock
FROM platform.iam_platform_owner_context_lock_v1(
    $1::uuid, $2::uuid, $3::bigint
);
