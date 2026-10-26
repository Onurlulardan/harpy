SELECT result_user_id, result_session_id, result_client_id,
       result_access_mode, result_portal_binding_id,
       result_token_version, result_authorization_version,
       result_deployment_mode, result_tenant_purpose,
       result_platform_subject_id, result_platform_clock,
       result_security_epoch
FROM iam_current_session_read_v6(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::uuid
) AS current_session
WHERE result_token_version = $3::bigint
  AND result_authorization_version = $4::bigint;
