-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_portal_setup_start_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_token_id uuid, p_binding_id uuid,
    p_selector bytea, p_hash_key_id text, p_secret_hmac bytea,
    p_binding_sha256 bytea, p_actor_id uuid, p_permission_id uuid,
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
    v_setup_revision bigint;
    v_identifier_id uuid;
    v_identifier_kind text;
    v_identifier_revision bigint;
    v_issue bigint;
    v_token_version bigint;
    v_authorization_version bigint;
BEGIN
    IF p_token_id IS NULL OR p_binding_id IS NULL OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND
                tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.portal_setup.start', p_request_sha256, p_token_id
    );
    IF FOUND THEN RETURN; END IF;
    IF NOT iam_identity_actor_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    SELECT setup.user_id, setup.revision
    INTO v_user_id, v_setup_revision
    FROM iam_portal_access_setups AS setup
    JOIN iam_portal_subject_bindings AS binding
      ON binding.id = setup.binding_id AND binding.user_id = setup.user_id
    WHERE setup.binding_id = p_binding_id AND setup.setup_state = 'required'
      AND binding.state = 'pending' FOR UPDATE OF setup, binding;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = '23503'; END IF;
    SELECT identifier.id, identifier.kind, identifier.revision
    INTO v_identifier_id, v_identifier_kind, v_identifier_revision
    FROM iam_user_identifiers AS identifier
    WHERE identifier.user_id = v_user_id AND identifier.kind = 'portal_v1'
      AND identifier.state IN ('pending', 'verified', 'active') FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = '23503'; END IF;
    v_issue := iam_one_time_issue_v1(
        p_token_id, 'portal_setup_v1', v_user_id, v_identifier_id,
        v_identifier_kind, v_identifier_revision, p_binding_id,
        v_setup_revision, p_selector, p_hash_key_id, p_secret_hmac,
        p_binding_sha256
    );
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.portal_setup.start', p_request_sha256,
        p_token_id, v_user_id, 1, 'actor_tenant', p_actor_id,
        'one_time_token', false, false
    ) AS result;
    RETURN QUERY SELECT p_token_id, v_user_id, 1::bigint,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_setup_start_v1(
    name,uuid,uuid,bytea,uuid,uuid,bytea,text,bytea,bytea,
    uuid,uuid,bigint,bigint
) FROM PUBLIC;
