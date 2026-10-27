-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_role_binding_grant_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_id uuid, p_role_id uuid, p_permission_id uuid, p_expected_revision bigint,
    p_actor_id uuid, p_required_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (result_id uuid, result_revision bigint, authorization_clock bigint, replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_clock bigint; v_revision bigint; v_replayed boolean; v_match text;
    v_role_protected boolean; v_count bigint; v_finished boolean; v_now timestamptz;
BEGIN
    SELECT b.authorization_clock, b.result_revision, b.replayed
    INTO v_clock, v_revision, v_replayed FROM iam_rbac_mutation_begin_v1(
        p_scope_schema, p_operation_id, 'iam.rbac.role_binding.grant', p_request_sha256,
        p_id, p_actor_id, p_required_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) AS b;
    IF v_replayed THEN RETURN QUERY SELECT p_id, v_revision, v_clock, true; RETURN; END IF;
    SELECT is_protected INTO v_role_protected FROM iam_roles
    WHERE id = p_role_id AND status = 'active';
    SELECT match_kind INTO v_match FROM iam_permissions
    WHERE id = p_permission_id AND state = 'active';
    IF v_role_protected IS NULL OR v_match IS NULL OR
       (v_match <> 'exact' AND NOT v_role_protected) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'IAM RBAC role binding is ineligible';
    END IF;
    v_now := pg_catalog.transaction_timestamp();
    IF p_expected_revision = 0 THEN
        INSERT INTO iam_role_permission_bindings (
            id, role_id, role_is_protected, permission_id, permission_match_kind,
            state, revision, created_at, updated_at, activated_at, revoked_at
        ) VALUES (p_id, p_role_id, v_role_protected, p_permission_id, v_match,
            'active', 1, v_now, v_now, v_now, NULL) RETURNING revision INTO v_revision;
    ELSE
        UPDATE iam_role_permission_bindings SET state = 'active', revision = revision + 1,
            updated_at = v_now, activated_at = v_now, revoked_at = NULL
        WHERE id = p_id AND role_id = p_role_id AND permission_id = p_permission_id
          AND state = 'revoked' AND revision = p_expected_revision
        RETURNING revision INTO v_revision;
    END IF;
    IF v_revision IS NULL THEN RAISE EXCEPTION USING ERRCODE = '40001',
        MESSAGE = 'IAM RBAC role binding revision changed'; END IF;
    v_count := iam_rbac_bump_role_users_v1(p_role_id, v_clock);
    v_finished := iam_rbac_mutation_finish_v1(p_operation_id,
        'iam.rbac.role_binding.grant', p_request_sha256, v_clock,
        'iam.rbac.role_binding.granted', p_id, v_revision);
    RETURN QUERY SELECT p_id, v_revision, v_clock, false;
END;
$function$;

CREATE FUNCTION iam_rbac_role_binding_revoke_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_id uuid, p_role_id uuid, p_permission_id uuid, p_expected_revision bigint,
    p_actor_id uuid, p_required_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (result_id uuid, result_revision bigint, authorization_clock bigint, replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_clock bigint; v_revision bigint; v_replayed boolean;
    v_count bigint; v_finished boolean; v_now timestamptz;
BEGIN
    SELECT b.authorization_clock, b.result_revision, b.replayed
    INTO v_clock, v_revision, v_replayed FROM iam_rbac_mutation_begin_v1(
        p_scope_schema, p_operation_id, 'iam.rbac.role_binding.revoke', p_request_sha256,
        p_id, p_actor_id, p_required_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) AS b;
    IF v_replayed THEN RETURN QUERY SELECT p_id, v_revision, v_clock, true; RETURN; END IF;
    v_now := pg_catalog.transaction_timestamp();
    UPDATE iam_role_permission_bindings SET state = 'revoked', revision = revision + 1,
        updated_at = v_now, revoked_at = v_now
    WHERE id = p_id AND role_id = p_role_id AND permission_id = p_permission_id
      AND state = 'active' AND revision = p_expected_revision
    RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN RAISE EXCEPTION USING ERRCODE = '40001',
        MESSAGE = 'IAM RBAC role binding revision changed'; END IF;
    v_count := iam_rbac_bump_role_users_v1(p_role_id, v_clock);
    v_finished := iam_rbac_mutation_finish_v1(p_operation_id,
        'iam.rbac.role_binding.revoke', p_request_sha256, v_clock,
        'iam.rbac.role_binding.revoked', p_id, v_revision);
    RETURN QUERY SELECT p_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_rbac_role_binding_grant_v1(
    name,uuid,bytea,uuid,uuid,uuid,bigint,uuid,uuid,bigint,bigint
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_rbac_role_binding_revoke_v1(
    name,uuid,bytea,uuid,uuid,uuid,bigint,uuid,uuid,bigint,bigint
) FROM PUBLIC;
