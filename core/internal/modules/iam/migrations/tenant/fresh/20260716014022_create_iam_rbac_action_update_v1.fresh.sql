-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_action_update_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_id uuid, p_display_name text, p_description text, p_expected_revision bigint,
    p_actor_id uuid, p_required_permission_id uuid,
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
        p_scope_schema, p_operation_id, 'iam.rbac.action.update', p_request_sha256,
        p_id, p_actor_id, p_required_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) AS b;
    IF v_replayed THEN RETURN QUERY SELECT p_id, v_revision, v_clock, true; RETURN; END IF;
    UPDATE iam_actions SET display_name = p_display_name, description = p_description,
        revision = revision + 1, updated_at = pg_catalog.transaction_timestamp()
    WHERE id = p_id AND state = 'active' AND revision = p_expected_revision
    RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN RAISE EXCEPTION USING ERRCODE = '40001',
        MESSAGE = 'IAM RBAC action revision changed'; END IF;
    v_count := iam_rbac_bump_all_users_v1(v_clock);
    v_finished := iam_rbac_mutation_finish_v1(p_operation_id, 'iam.rbac.action.update',
        p_request_sha256, v_clock, 'iam.rbac.action.updated', p_id, v_revision);
    RETURN QUERY SELECT p_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_rbac_action_update_v1(
    name,uuid,bytea,uuid,text,text,bigint,uuid,uuid,bigint,bigint
) FROM PUBLIC;
