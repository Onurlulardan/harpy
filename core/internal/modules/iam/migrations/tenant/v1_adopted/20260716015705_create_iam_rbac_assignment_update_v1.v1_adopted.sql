-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_assignment_update_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_id uuid, p_resource_id uuid, p_target_kind text, p_target_id uuid,
    p_items jsonb, p_expected_revision bigint,
    p_actor_id uuid, p_required_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint,
    p_step_up_purpose_sha256 bytea, p_step_up_auth_time timestamptz,
    p_amr_password boolean, p_amr_webauthn boolean, p_amr_totp boolean
)
RETURNS TABLE (
    result_id uuid, result_revision bigint,
    authorization_clock bigint, replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_clock bigint;
    v_revision bigint;
    v_replayed boolean;
    v_asserted boolean;
    v_old_resource_id uuid;
    v_old_target_kind text;
    v_old_target_id uuid;
    v_permission_id uuid;
    v_binding_id uuid;
    v_bumped bigint;
    v_finished boolean;
    v_old_action_id uuid;
    v_old_permission_id uuid;
    v_old_binding_id uuid;
    v_item_ordinal integer;
    v_item_action_id uuid;
    v_item_permission_id uuid;
    v_item_binding_id uuid;
    v_now timestamptz;
BEGIN
    v_now := pg_catalog.transaction_timestamp();
    SELECT begin_result.authorization_clock,
           begin_result.result_revision, begin_result.replayed
    INTO v_clock, v_revision, v_replayed
    FROM iam_rbac_mutation_begin_v1(
        p_scope_schema, p_operation_id, 'iam.rbac.assignment.update',
        p_request_sha256, p_id, p_actor_id, p_required_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) AS begin_result;
    IF v_replayed THEN
        RETURN QUERY SELECT p_id, v_revision, v_clock, true;
        RETURN;
    END IF;
    v_asserted := iam_rbac_assignment_step_up_assert_v1(
        'iam.rbac.assignment.update', p_id, p_step_up_purpose_sha256,
        p_step_up_auth_time, p_amr_password, p_amr_webauthn, p_amr_totp
    );
    IF p_resource_id IS NULL OR p_target_id IS NULL OR
       NOT iam_rbac_assignment_target_active_v1(p_target_kind, p_target_id) OR
       NOT iam_rbac_assignment_items_valid_v1(p_items) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM RBAC assignment update input is invalid';
    END IF;
    SELECT resource_id, target_kind, target_id, revision
    INTO v_old_resource_id, v_old_target_kind, v_old_target_id, v_revision
    FROM iam_rbac_assignment_groups
    WHERE id = p_id AND state = 'active'
    FOR UPDATE;
    IF NOT FOUND OR v_revision IS DISTINCT FROM p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM RBAC assignment revision changed';
    END IF;
    FOR v_old_action_id, v_old_permission_id, v_old_binding_id IN
        SELECT member.action_id, member.permission_id, member.binding_id
        FROM iam_rbac_assignment_members AS member
        WHERE member.group_id = p_id
          AND member.group_revision = v_revision
          AND member.state = 'active'
    LOOP
        IF v_old_resource_id IS DISTINCT FROM p_resource_id OR
           v_old_target_kind IS DISTINCT FROM p_target_kind OR
           v_old_target_id IS DISTINCT FROM p_target_id OR
           NOT EXISTS (
               SELECT 1
               FROM pg_catalog.jsonb_to_recordset(p_items) AS item(action_id uuid)
               WHERE item.action_id = v_old_action_id
           ) THEN
            v_finished := iam_rbac_assignment_binding_revoke_v1(
                v_old_target_kind, v_old_target_id,
                v_old_permission_id, v_old_binding_id
            );
        END IF;
    END LOOP;
    UPDATE iam_rbac_assignment_members
    SET state = 'revoked', revoked_at = v_now
    WHERE group_id = p_id AND group_revision = v_revision AND state = 'active';
    UPDATE iam_rbac_assignment_groups
    SET resource_id = p_resource_id, target_kind = p_target_kind,
        target_id = p_target_id, revision = revision + 1, updated_at = v_now
    WHERE id = p_id AND revision = v_revision
    RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM RBAC assignment revision changed';
    END IF;
    FOR v_item_ordinal, v_item_action_id,
        v_item_permission_id, v_item_binding_id IN
        SELECT item.ordinal, item.action_id, item.permission_id, item.binding_id
        FROM pg_catalog.jsonb_to_recordset(p_items) AS item(
            ordinal integer, action_id uuid, permission_id uuid, binding_id uuid
        ) ORDER BY item.ordinal
    LOOP
        v_permission_id := iam_rbac_assignment_exact_permission_v1(
            p_resource_id, v_item_action_id, v_item_permission_id
        );
        v_binding_id := iam_rbac_assignment_binding_grant_v1(
            p_id, p_target_kind, p_target_id,
            v_permission_id, v_item_binding_id
        );
        INSERT INTO iam_rbac_assignment_members (
            group_id, group_revision, ordinal, action_id, permission_id,
            permission_match_kind, binding_id, state, created_at, revoked_at
        ) VALUES (
            p_id, v_revision, v_item_ordinal, v_item_action_id, v_permission_id,
            'exact', v_binding_id, 'active', v_now, NULL
        );
    END LOOP;
    IF v_old_target_kind IS DISTINCT FROM p_target_kind OR
       v_old_target_id IS DISTINCT FROM p_target_id THEN
        v_bumped := iam_rbac_assignment_bump_target_v1(
            v_old_target_kind, v_old_target_id, v_clock
        );
    END IF;
    v_bumped := iam_rbac_assignment_bump_target_v1(
        p_target_kind, p_target_id, v_clock
    );
    v_finished := iam_rbac_mutation_finish_v1(
        p_operation_id, 'iam.rbac.assignment.update', p_request_sha256,
        v_clock, 'iam.rbac.assignment.updated', p_id, v_revision
    );
    RETURN QUERY SELECT p_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_assignment_update_v1(
    name,uuid,bytea,uuid,uuid,text,uuid,jsonb,bigint,uuid,uuid,bigint,bigint,
    bytea,timestamptz,boolean,boolean,boolean
) FROM PUBLIC;
