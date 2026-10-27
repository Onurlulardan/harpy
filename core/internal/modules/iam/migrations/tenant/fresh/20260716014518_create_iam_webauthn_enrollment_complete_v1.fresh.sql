-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_webauthn_enrollment_complete_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_authenticator_id uuid, p_challenge_id uuid,
    p_hmac_key_id text, p_challenge_hmac bytea, p_context_sha256 bytea,
    p_credential_id bytea, p_public_key_cose bytea, p_rp_id text,
    p_rp_id_sha256 bytea, p_origin_policy_version bigint, p_aaguid uuid,
    p_sign_count bigint, p_backup_eligible boolean, p_backup_state boolean,
    p_discoverable boolean, p_uv_required boolean,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid, subject_user_id uuid, result_revision bigint,
    token_version bigint, authorization_version bigint, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_user_id uuid;
    v_challenge_state text;
    v_challenge_expires_at timestamptz;
    v_challenge_rp_id_sha256 bytea;
    v_challenge_origin_policy_version bigint;
    v_challenge_uv_required boolean;
    v_revision bigint;
    v_token_version bigint;
    v_authorization_version bigint;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF p_authenticator_id IS NULL OR p_challenge_id IS NULL OR p_aaguid IS NULL OR
       pg_catalog.octet_length(p_challenge_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_context_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_credential_id) NOT BETWEEN 1 AND 1024 OR
       pg_catalog.octet_length(p_public_key_cose) NOT BETWEEN 1 AND 4096 OR
       pg_catalog.sha256(pg_catalog.convert_to(p_rp_id, 'UTF8'))
          IS DISTINCT FROM p_rp_id_sha256 OR p_origin_policy_version < 1 OR
       p_sign_count < 0 OR (p_backup_state AND NOT p_backup_eligible) OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND
                tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.webauthn_enrollment.complete',
        p_request_sha256, p_authenticator_id
    );
    IF FOUND THEN RETURN; END IF;
    IF NOT iam_identity_actor_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    SELECT challenge.user_id, challenge.state, challenge.expires_at,
        challenge.rp_id_sha256, challenge.origin_policy_version,
        challenge.user_verification_required
    INTO v_user_id, v_challenge_state, v_challenge_expires_at,
        v_challenge_rp_id_sha256, v_challenge_origin_policy_version,
        v_challenge_uv_required
    FROM iam_auth_challenges AS challenge
    WHERE challenge.id = p_challenge_id
      AND challenge.authenticator_id = p_authenticator_id
      AND challenge.purpose = 'webauthn_register_v1'
      AND challenge.hmac_key_id = p_hmac_key_id
      AND challenge.challenge_hmac = p_challenge_hmac
      AND challenge.context_sha256 = p_context_sha256 FOR UPDATE;
    IF NOT FOUND OR v_challenge_state <> 'pending' OR
       v_challenge_expires_at <= v_now OR
       v_challenge_rp_id_sha256 <> p_rp_id_sha256 OR
       v_challenge_origin_policy_version <> p_origin_policy_version OR
       v_challenge_uv_required <> p_uv_required THEN
        RAISE EXCEPTION USING ERRCODE = '28000';
    END IF;
    PERFORM 1 FROM iam_authenticators AS authenticator
    WHERE authenticator.id = p_authenticator_id
      AND authenticator.user_id = v_user_id
      AND authenticator.kind = 'webauthn_v1' AND authenticator.state = 'pending'
    FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = '40001'; END IF;
    INSERT INTO iam_webauthn_credentials (
        authenticator_id, user_id, authenticator_kind, credential_id,
        public_key_cose, rp_id, rp_id_sha256, origin_policy_version,
        aaguid, sign_count, backup_eligible, backup_state, discoverable,
        uv_required, revision, created_at, updated_at
    ) VALUES (
        p_authenticator_id, v_user_id, 'webauthn_v1',
        p_credential_id, p_public_key_cose, p_rp_id, p_rp_id_sha256,
        p_origin_policy_version, p_aaguid, p_sign_count, p_backup_eligible,
        p_backup_state, p_discoverable, p_uv_required, 1, v_now, v_now
    );
    UPDATE iam_auth_challenges SET state = 'consumed', terminal_at = v_now,
        revision = revision + 1, updated_at = v_now WHERE id = p_challenge_id;
    UPDATE iam_authenticators SET state = 'active', activated_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE id = p_authenticator_id RETURNING revision INTO v_revision;
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.webauthn_enrollment.complete', p_request_sha256,
        p_authenticator_id, v_user_id, v_revision,
        'actor_tenant', p_actor_id, 'authenticator', true, true
    ) AS result;
    RETURN QUERY SELECT p_authenticator_id, v_user_id, v_revision,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_webauthn_enrollment_complete_v1(
    name,uuid,uuid,bytea,uuid,uuid,text,bytea,bytea,bytea,bytea,text,bytea,
    bigint,uuid,bigint,boolean,boolean,boolean,boolean,uuid,uuid,bigint,bigint
) FROM PUBLIC;
