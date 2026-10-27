-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_assignment_create_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_id uuid, p_resource_id uuid, p_target_kind text, p_target_id uuid,
    p_items jsonb, p_actor_id uuid, p_required_permission_id uuid,
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
    v_permission_id uuid;
    v_binding_id uuid;
    v_bumped bigint;
    v_finished boolean;
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
        p_scope_schema, p_operation_id, 'iam.rbac.assignment.create',
        p_request_sha256, p_id, p_actor_id, p_required_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) AS begin_result;
    IF v_replayed THEN
        RETURN QUERY SELECT p_id, v_revision, v_clock, true;
        RETURN;
    END IF;
    v_asserted := iam_rbac_assignment_step_up_assert_v1(
        'iam.rbac.assignment.create', NULL, p_step_up_purpose_sha256,
        p_step_up_auth_time, p_amr_password, p_amr_webauthn, p_amr_totp
    );
    IF p_id IS NULL OR p_resource_id IS NULL OR p_target_id IS NULL OR
       NOT iam_rbac_assignment_target_active_v1(p_target_kind, p_target_id) OR
       NOT iam_rbac_assignment_items_valid_v1(p_items) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM RBAC assignment create input is invalid';
    END IF;
    INSERT INTO iam_rbac_assignment_groups (
        id, resource_id, target_kind, target_id, state, revision,
        created_at, updated_at, retired_at
    ) VALUES (
        p_id, p_resource_id, p_target_kind, p_target_id,
        'active', 1, v_now, v_now, NULL
    );
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
            p_id, 1, v_item_ordinal, v_item_action_id, v_permission_id,
            'exact', v_binding_id, 'active', v_now, NULL
        );
    END LOOP;
    v_revision := 1;
    v_bumped := iam_rbac_assignment_bump_target_v1(
        p_target_kind, p_target_id, v_clock
    );
    v_finished := iam_rbac_mutation_finish_v1(
        p_operation_id, 'iam.rbac.assignment.create', p_request_sha256,
        v_clock, 'iam.rbac.assignment.created', p_id, v_revision
    );
    RETURN QUERY SELECT p_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_assignment_create_v1(
    name,uuid,bytea,uuid,uuid,text,uuid,jsonb,uuid,uuid,bigint,bigint,
    bytea,timestamptz,boolean,boolean,boolean
) FROM PUBLIC;
