-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_membership_add_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_membership_id uuid, p_organization_id uuid, p_user_id uuid, p_kind text,
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
    v_clock bigint; v_user_revision bigint;
BEGIN
    IF p_membership_id IS NULL OR p_organization_id IS NULL OR p_user_id IS NULL OR
       p_kind NOT IN ('member', 'owner') OR
       NOT EXISTS (SELECT 1 FROM iam_organizations
                   WHERE id = p_organization_id AND status = 'active') OR
       NOT EXISTS (SELECT 1 FROM iam_users WHERE id = p_user_id AND status = 'active') OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '23514',
          MESSAGE = 'IAM identity membership add input is invalid'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.membership.add', p_request_sha256, p_membership_id
    );
    IF FOUND THEN RETURN; END IF;
    INSERT INTO iam_organization_memberships (
        id, organization_id, user_id, membership_kind, state,
        revision, created_at, updated_at
    ) VALUES (
        p_membership_id, p_organization_id, p_user_id, p_kind,
        'active', 1, v_now, v_now
    );
    v_user_revision := iam_identity_user_bump_v1(p_user_id, false);
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.membership.add', p_request_sha256,
        p_membership_id, 1, p_request_sha256
    );
    RETURN QUERY SELECT p_membership_id, 1::bigint, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_membership_add_v1(
    name,uuid,bytea,uuid,uuid,uuid,text,uuid,uuid,bigint,bigint
) FROM PUBLIC;
