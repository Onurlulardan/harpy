SELECT code_id, authenticator_id, user_id, code_revision,
       authenticator_revision, user_revision, token_version, code_state,
       authenticator_state, user_status, identity_kind, hmac_key_id, code_hmac
FROM iam_recovery_credential_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1, $2, $3
);
