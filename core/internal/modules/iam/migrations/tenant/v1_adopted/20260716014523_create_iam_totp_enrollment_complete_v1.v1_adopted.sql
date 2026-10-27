-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_totp_enrollment_complete_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_authenticator_id uuid, p_challenge_id uuid,
    p_hmac_key_id text, p_challenge_hmac bytea, p_context_sha256 bytea,
    p_expected_authenticator_revision bigint, p_expected_challenge_revision bigint,
    p_accepted_time_step bigint, p_actor_id uuid, p_permission_id uuid,
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
    v_revision bigint;
    v_token_version bigint;
    v_authorization_version bigint;
    v_current_step bigint;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    v_current_step := pg_catalog.floor(
        pg_catalog.date_part('epoch', v_now) / 30
    )::bigint;
    IF p_authenticator_id IS NULL OR p_challenge_id IS NULL OR
       p_expected_authenticator_revision < 1 OR p_expected_challenge_revision < 1 OR
       p_accepted_time_step NOT BETWEEN v_current_step - 1 AND v_current_step + 1 OR
       pg_catalog.octet_length(p_challenge_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_context_sha256) IS DISTINCT FROM 32 OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.totp_enrollment.complete',
        p_request_sha256, p_authenticator_id
    );
    IF FOUND THEN RETURN; END IF;
    IF p_actor_id IS NULL OR NOT iam_identity_actor_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    SELECT authenticator.user_id INTO v_user_id
    FROM iam_authenticators AS authenticator
    WHERE authenticator.id = p_authenticator_id
      AND authenticator.user_id = p_actor_id
      AND authenticator.kind = 'totp_v1'
      AND authenticator.state = 'pending'
      AND authenticator.revision = p_expected_authenticator_revision
    FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = '40001'; END IF;
    PERFORM 1 FROM iam_auth_challenges AS challenge
    WHERE challenge.id = p_challenge_id
      AND challenge.user_id = v_user_id
      AND challenge.authenticator_id = p_authenticator_id
      AND challenge.purpose = 'totp_enroll_v1'
      AND challenge.state = 'pending'
      AND challenge.revision = p_expected_challenge_revision
      AND challenge.expires_at > v_now
      AND challenge.hmac_key_id = p_hmac_key_id
      AND challenge.challenge_hmac = p_challenge_hmac
      AND challenge.context_sha256 = p_context_sha256
    FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = '28000'; END IF;
    UPDATE iam_totp_authenticators SET verified_at = v_now,
        last_accepted_time_step = p_accepted_time_step,
        revision = revision + 1, updated_at = v_now
    WHERE authenticator_id = p_authenticator_id
      AND verified_at IS NULL AND enrollment_expires_at > v_now;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = '40001'; END IF;
    UPDATE iam_auth_challenges SET state = 'consumed', terminal_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE id = p_challenge_id;
    UPDATE iam_authenticators SET state = 'active', activated_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE id = p_authenticator_id RETURNING revision INTO v_revision;
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.totp_enrollment.complete', p_request_sha256,
        p_authenticator_id, v_user_id, v_revision,
        'actor_tenant', p_actor_id, 'authenticator', true, true
    ) AS result;
    RETURN QUERY SELECT p_authenticator_id, v_user_id, v_revision,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_totp_enrollment_complete_v1(
    name,uuid,uuid,bytea,uuid,uuid,text,bytea,bytea,bigint,bigint,bigint,
    uuid,uuid,bigint,bigint
) FROM PUBLIC;
