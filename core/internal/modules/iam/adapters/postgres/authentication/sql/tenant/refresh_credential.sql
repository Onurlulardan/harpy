SELECT token_id, token_revision, token_state, family_id, family_revision,
       family_state, session_id, session_revision, session_state, user_id,
       user_revision, user_status, identity_kind, token_version,
       authorization_version, access_mode_id, access_mode, access_mode_state,
       access_mode_revision, portal_binding_id, portal_binding_state,
       portal_binding_revision, hash_key_id, secret_hmac, sequence, auth_time,
       acr, amr_password, amr_webauthn, amr_totp, amr_recovery,
       step_up_purpose_sha256, idle_expires_at, absolute_expires_at, retain_until,
       platform_authorization_clock, security_epoch
FROM iam_refresh_credential_v3(
    (pg_catalog.current_schemas(false))[2]::name,
    $1, $2, $3, $4, $5
);
