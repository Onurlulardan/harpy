-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_organization_delete_v2(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_organization_id uuid, p_expected_revision bigint,
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
    v_status text; v_revision bigint; v_clock bigint; v_user_id uuid;
BEGIN
    IF p_organization_id IS NULL OR p_expected_revision < 1 OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE =
          'IAM identity organization deletion input is invalid'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.organization.delete',
        p_request_sha256, p_organization_id
    );
    IF FOUND THEN RETURN; END IF;
    PERFORM membership.id FROM iam_organization_memberships AS membership
    WHERE membership.organization_id = p_organization_id
    ORDER BY membership.user_id, membership.id FOR UPDATE;
    SELECT organization.status, organization.revision INTO v_status, v_revision
    FROM iam_organizations AS organization
    WHERE organization.id = p_organization_id FOR UPDATE;
    IF NOT FOUND OR v_status <> 'suspended' OR v_revision <> p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE =
            'IAM identity organization delete compare-and-swap failed';
    END IF;
    IF EXISTS (SELECT 1 FROM iam_organizations
               WHERE parent_id = p_organization_id AND status <> 'tombstoned') OR
       EXISTS (SELECT 1 FROM iam_roles
               WHERE organization_id = p_organization_id AND status = 'active') OR
       EXISTS (SELECT 1 FROM iam_organization_permission_bindings
               WHERE organization_id = p_organization_id AND state = 'active') OR
       EXISTS (SELECT 1 FROM iam_user_role_assignments
               WHERE organization_id = p_organization_id AND state <> 'revoked') THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE =
            'IAM identity organization deletion has dependencies';
    END IF;
    FOR v_user_id IN
        WITH tombstoned AS (
            UPDATE iam_organization_memberships AS membership
            SET state = 'tombstoned', tombstoned_at = v_now,
                revision = membership.revision + 1, updated_at = v_now
            WHERE membership.organization_id = p_organization_id
              AND membership.state <> 'tombstoned'
            RETURNING membership.user_id
        ) SELECT DISTINCT user_id FROM tombstoned ORDER BY user_id
    LOOP
        PERFORM iam_identity_user_bump_v1(v_user_id, false);
    END LOOP;
    v_revision := NULL;
    UPDATE iam_organizations SET status = 'tombstoned', tombstoned_at = v_now,
        revision = revision + 1, updated_at = v_now
    WHERE id = p_organization_id AND revision = p_expected_revision
      AND status = 'suspended' RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE =
        'IAM identity organization delete compare-and-swap failed'; END IF;
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.organization.delete', p_request_sha256,
        p_organization_id, v_revision, p_request_sha256
    );
    RETURN QUERY SELECT p_organization_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_organization_delete_v2(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) FROM PUBLIC;
