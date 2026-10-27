-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_permission_create_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_id uuid, p_match_kind text, p_resource_id uuid, p_action_id uuid,
    p_is_protected boolean, p_actor_id uuid, p_required_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (result_id uuid, result_revision bigint, authorization_clock bigint, replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_clock bigint; v_revision bigint; v_replayed boolean;
    v_count bigint; v_finished boolean;
BEGIN
    SELECT b.authorization_clock, b.result_revision, b.replayed
    INTO v_clock, v_revision, v_replayed FROM iam_rbac_mutation_begin_v1(
        p_scope_schema, p_operation_id, 'iam.rbac.permission.create', p_request_sha256,
        p_id, p_actor_id, p_required_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) AS b;
    IF v_replayed THEN RETURN QUERY SELECT p_id, v_revision, v_clock, true; RETURN; END IF;
    INSERT INTO iam_permissions (
        id, match_kind, resource_id, action_id, is_protected, state,
        revision, created_at, updated_at, retired_at
    ) VALUES (
        p_id, p_match_kind, p_resource_id, p_action_id, p_is_protected,
        'active', 1, pg_catalog.transaction_timestamp(),
        pg_catalog.transaction_timestamp(), NULL
    );
    v_count := iam_rbac_bump_all_users_v1(v_clock);
    v_finished := iam_rbac_mutation_finish_v1(p_operation_id, 'iam.rbac.permission.create',
        p_request_sha256, v_clock, 'iam.rbac.permission.created', p_id, 1);
    RETURN QUERY SELECT p_id, 1::bigint, v_clock, false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_rbac_permission_create_v1(
    name,uuid,bytea,uuid,text,uuid,uuid,boolean,uuid,uuid,bigint,bigint
) FROM PUBLIC;
