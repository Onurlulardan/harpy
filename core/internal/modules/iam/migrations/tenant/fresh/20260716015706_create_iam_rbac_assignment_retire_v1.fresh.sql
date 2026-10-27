-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_assignment_retire_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_id uuid, p_expected_revision bigint,
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
    v_target_kind text;
    v_target_id uuid;
    v_finished boolean;
    v_bumped bigint;
    v_member_permission_id uuid;
    v_member_binding_id uuid;
    v_now timestamptz;
BEGIN
    v_now := pg_catalog.transaction_timestamp();
    SELECT begin_result.authorization_clock,
           begin_result.result_revision, begin_result.replayed
    INTO v_clock, v_revision, v_replayed
    FROM iam_rbac_mutation_begin_v1(
        p_scope_schema, p_operation_id, 'iam.rbac.assignment.retire',
        p_request_sha256, p_id, p_actor_id, p_required_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) AS begin_result;
    IF v_replayed THEN
        RETURN QUERY SELECT p_id, v_revision, v_clock, true;
        RETURN;
    END IF;
    v_asserted := iam_rbac_assignment_step_up_assert_v1(
        'iam.rbac.assignment.retire', p_id, p_step_up_purpose_sha256,
        p_step_up_auth_time, p_amr_password, p_amr_webauthn, p_amr_totp
    );
    SELECT target_kind, target_id, revision
    INTO v_target_kind, v_target_id, v_revision
    FROM iam_rbac_assignment_groups
    WHERE id = p_id AND state = 'active'
    FOR UPDATE;
    IF NOT FOUND OR v_revision IS DISTINCT FROM p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM RBAC assignment revision changed';
    END IF;
    FOR v_member_permission_id, v_member_binding_id IN
        SELECT permission_id, binding_id
        FROM iam_rbac_assignment_members
        WHERE group_id = p_id AND group_revision = v_revision
          AND state = 'active'
    LOOP
        v_finished := iam_rbac_assignment_binding_revoke_v1(
            v_target_kind, v_target_id,
            v_member_permission_id, v_member_binding_id
        );
    END LOOP;
    UPDATE iam_rbac_assignment_members
    SET state = 'revoked', revoked_at = v_now
    WHERE group_id = p_id AND group_revision = v_revision AND state = 'active';
    UPDATE iam_rbac_assignment_groups
    SET state = 'retired', revision = revision + 1,
        updated_at = v_now, retired_at = v_now
    WHERE id = p_id AND revision = v_revision
    RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM RBAC assignment revision changed';
    END IF;
    v_bumped := iam_rbac_assignment_bump_target_v1(
        v_target_kind, v_target_id, v_clock
    );
    v_finished := iam_rbac_mutation_finish_v1(
        p_operation_id, 'iam.rbac.assignment.retire', p_request_sha256,
        v_clock, 'iam.rbac.assignment.retired', p_id, v_revision
    );
    RETURN QUERY SELECT p_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_assignment_retire_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint,
    bytea,timestamptz,boolean,boolean,boolean
) FROM PUBLIC;
