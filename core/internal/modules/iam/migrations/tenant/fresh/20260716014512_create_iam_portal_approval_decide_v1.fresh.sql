-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_portal_approval_decide_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_binding_id uuid, p_expected_revision bigint,
    p_decision text, p_actor_id uuid, p_permission_id uuid,
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
    v_access_mode_id uuid;
    v_revision bigint;
    v_identifier_id uuid;
    v_token_version bigint;
    v_authorization_version bigint;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF p_binding_id IS NULL OR p_expected_revision < 1 OR
       p_decision NOT IN ('approved', 'rejected', 'revision_required') OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND
                tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.portal_approval.decide',
        p_request_sha256, p_binding_id
    );
    IF FOUND THEN RETURN; END IF;
    IF NOT iam_identity_actor_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    SELECT setup.user_id, setup.access_mode_id, setup.revision
    INTO v_user_id, v_access_mode_id, v_revision
    FROM iam_portal_access_setups AS setup
    WHERE setup.binding_id = p_binding_id AND setup.revision = p_expected_revision
      AND setup.setup_state = 'completed' AND setup.approval_state = 'pending'
    FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = '40001'; END IF;
    UPDATE iam_portal_access_setups SET approval_state = p_decision,
        decision_at = v_now, decision_actor_id = p_actor_id,
        revision = revision + 1, updated_at = v_now
    WHERE binding_id = p_binding_id RETURNING revision INTO v_revision;
    IF p_decision = 'approved' THEN
        UPDATE iam_portal_subject_bindings SET state = 'active',
            activated_at = COALESCE(activated_at, v_now),
            suspended_at = NULL, revision = revision + 1, updated_at = v_now
        WHERE id = p_binding_id AND state IN ('pending', 'suspended');
        UPDATE iam_user_access_modes SET state = 'active',
            activated_at = COALESCE(activated_at, v_now),
            suspended_at = NULL, revision = revision + 1, updated_at = v_now
        WHERE id = v_access_mode_id AND state IN ('pending', 'suspended');
        SELECT identifier.id INTO v_identifier_id
        FROM iam_user_identifiers AS identifier
        WHERE identifier.user_id = v_user_id AND identifier.kind = 'portal_v1'
          AND identifier.state IN ('pending', 'verified', 'active') FOR UPDATE;
        UPDATE iam_user_identifiers SET state = 'active',
            verified_at = COALESCE(verified_at, v_now),
            activated_at = COALESCE(activated_at, v_now),
            revision = revision + 1, updated_at = v_now
        WHERE id = v_identifier_id AND state <> 'active';
        UPDATE iam_users SET status = 'active', revision = revision + 1,
            updated_at = v_now WHERE id = v_user_id AND status = 'pending';
    END IF;
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.portal_approval.decide', p_request_sha256,
        p_binding_id, v_user_id, v_revision,
        'actor_tenant', p_actor_id, 'portal_approval', true, true
    ) AS result;
    RETURN QUERY SELECT p_binding_id, v_user_id, v_revision,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_approval_decide_v1(
    name,uuid,uuid,bytea,uuid,bigint,text,uuid,uuid,bigint,bigint
) FROM PUBLIC;
