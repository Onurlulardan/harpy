-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_membership_update_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_membership_id uuid, p_expected_revision bigint, p_kind text, p_state text,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
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
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_organization_id uuid; v_user_id uuid; v_old_kind text; v_old_state text;
    v_revision bigint; v_clock bigint; v_user_revision bigint;
BEGIN
    IF p_membership_id IS NULL OR p_expected_revision < 1 OR
       p_kind NOT IN ('member', 'owner') OR p_state NOT IN ('active', 'suspended') OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '23514',
          MESSAGE = 'IAM identity membership update input is invalid'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.membership.update',
        p_request_sha256, p_membership_id
    );
    IF FOUND THEN RETURN; END IF;
    SELECT organization_id, user_id, membership_kind, state
    INTO v_organization_id, v_user_id, v_old_kind, v_old_state
    FROM iam_organization_memberships
    WHERE id = p_membership_id AND revision = p_expected_revision FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE = '40001',
        MESSAGE = 'IAM identity membership update compare-and-swap failed'; END IF;
    PERFORM 1 FROM iam_organizations WHERE id = v_organization_id FOR UPDATE;
    IF v_old_kind = 'owner' AND v_old_state = 'active' AND
       (p_kind <> 'owner' OR p_state <> 'active') AND NOT EXISTS (
          SELECT 1 FROM iam_organization_memberships AS other
          JOIN iam_users AS principal ON principal.id = other.user_id
          WHERE other.organization_id = v_organization_id
            AND other.id <> p_membership_id AND other.membership_kind = 'owner'
            AND other.state = 'active' AND principal.status = 'active'
       ) THEN RAISE EXCEPTION USING ERRCODE = '23514',
          MESSAGE = 'IAM identity organization requires an active owner'; END IF;
    UPDATE iam_organization_memberships SET membership_kind = p_kind,
        state = p_state, revision = revision + 1, updated_at = v_now
    WHERE id = p_membership_id AND revision = p_expected_revision
    RETURNING revision INTO v_revision;
    v_user_revision := iam_identity_user_bump_v1(v_user_id, false);
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.membership.update', p_request_sha256,
        p_membership_id, v_revision, p_request_sha256
    );
    RETURN QUERY SELECT p_membership_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_membership_update_v1(
    name,uuid,bytea,uuid,bigint,text,text,uuid,uuid,bigint,bigint
) FROM PUBLIC;
