-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_recovery_codes_issue_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_authenticator_id uuid, p_user_id uuid,
    p_display_label text, p_codes jsonb, p_hmac_key_id text,
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
    v_count integer;
    v_token_version bigint;
    v_authorization_version bigint;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF pg_catalog.jsonb_typeof(p_codes) IS DISTINCT FROM 'array' THEN
        RAISE EXCEPTION USING ERRCODE = '22023';
    END IF;
    v_count := pg_catalog.jsonb_array_length(p_codes);
    IF p_authenticator_id IS NULL OR p_user_id IS NULL OR
       v_count NOT BETWEEN 8 AND 16 OR
       p_hmac_key_id !~ '^[a-z0-9][a-z0-9._-]{0,127}$' OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND tenant.schema_name = p_scope_schema::text) OR
       EXISTS (
          SELECT 1
          FROM pg_catalog.jsonb_array_elements(p_codes) AS code(value)
          WHERE pg_catalog.jsonb_typeof(code.value) <> 'object'
             OR (SELECT pg_catalog.count(*) FROM
                    pg_catalog.jsonb_object_keys(code.value)) <> 2
             OR NOT (code.value ? 'id' AND code.value ? 'hmac_sha256')
             OR code.value->>'id' !~
                '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
             OR code.value->>'hmac_sha256' !~ '^[0-9a-f]{64}$'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.recovery_codes.issue',
        p_request_sha256, p_authenticator_id
    );
    IF FOUND THEN RETURN; END IF;
    IF p_user_id <> p_actor_id OR NOT iam_identity_actor_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    PERFORM 1 FROM iam_users AS subject
    WHERE subject.id = p_user_id AND subject.status = 'active' FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = '23503'; END IF;
    UPDATE iam_recovery_codes SET state = 'revoked', revoked_at = v_now,
        purge_after = v_now + INTERVAL '30 days',
        revision = revision + 1, updated_at = v_now
    WHERE user_id = p_user_id AND state = 'active';
    UPDATE iam_authenticators SET state = 'retired', retired_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE user_id = p_user_id AND kind = 'recovery_set_v1'
      AND state <> 'retired';
    INSERT INTO iam_authenticators (
        id, user_id, kind, display_label, state, revision,
        created_at, updated_at, activated_at
    ) VALUES (
        p_authenticator_id, p_user_id, 'recovery_set_v1', p_display_label,
        'active', 1, v_now, v_now, v_now
    );
    INSERT INTO iam_recovery_codes (
        id, authenticator_id, user_id, authenticator_kind, position,
        hash_profile, hmac_key_id, code_hmac, state, revision,
        created_at, updated_at
    )
    SELECT (code.value->>'id')::uuid, p_authenticator_id, p_user_id,
        'recovery_set_v1', code.code_ordinal::smallint,
        'hmac_sha256_v1', p_hmac_key_id,
        pg_catalog.decode(code.value->>'hmac_sha256', 'hex'),
        'active', 1, v_now, v_now
    FROM pg_catalog.jsonb_array_elements(p_codes)
      WITH ORDINALITY AS code(value, code_ordinal);
    UPDATE iam_sessions SET state = 'revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE user_id = p_user_id AND state = 'active';
    UPDATE iam_refresh_families SET state = 'revoked',
        revoke_reason = 'recovery_codes_rotated', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE user_id = p_user_id AND state = 'active';
    UPDATE iam_refresh_tokens SET state = 'revoked', revoked_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE family_id IN (
        SELECT id FROM iam_refresh_families WHERE user_id = p_user_id
    ) AND state = 'active';
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.recovery_codes.issue', p_request_sha256,
        p_authenticator_id, p_user_id, 1, 'actor_tenant', p_actor_id,
        'recovery_set', true, true
    ) AS result;
    RETURN QUERY SELECT p_authenticator_id, p_user_id, 1::bigint,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_recovery_codes_issue_v1(
    name,uuid,uuid,bytea,uuid,uuid,text,jsonb,text,
    uuid,uuid,bigint,bigint
) FROM PUBLIC;
