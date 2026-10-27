-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_authenticator_retire_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_authenticator_id uuid,
    p_expected_revision bigint, p_actor_id uuid, p_permission_id uuid,
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
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF p_authenticator_id IS NULL OR p_expected_revision < 1 OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND
                tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.authenticator.retire',
        p_request_sha256, p_authenticator_id
    );
    IF FOUND THEN RETURN; END IF;
    IF NOT iam_identity_actor_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    UPDATE iam_authenticators AS authenticator SET state = 'retired',
        retired_at = v_now, revision = authenticator.revision + 1,
        updated_at = v_now
    WHERE authenticator.id = p_authenticator_id
      AND authenticator.revision = p_expected_revision
      AND authenticator.state IN (
          'pending', 'suspended', 'compromised', 'renewal_required'
      ) RETURNING authenticator.user_id, authenticator.revision
        INTO v_user_id, v_revision;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = '40001'; END IF;
    UPDATE iam_auth_challenges SET state = 'cancelled', terminal_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE authenticator_id = p_authenticator_id AND state = 'pending';
    UPDATE iam_recovery_codes SET state = 'revoked', revoked_at = v_now,
        purge_after = v_now + INTERVAL '30 days', revision = revision + 1,
        updated_at = v_now
    WHERE authenticator_id = p_authenticator_id AND state = 'active';
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.authenticator.retire', p_request_sha256,
        p_authenticator_id, v_user_id, v_revision,
        'actor_tenant', p_actor_id, 'authenticator', true, true
    ) AS result;
    RETURN QUERY SELECT p_authenticator_id, v_user_id, v_revision,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_authenticator_retire_v1(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) FROM PUBLIC;
