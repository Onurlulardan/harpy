-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_organization_disable_v1(
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
    v_revision bigint; v_clock bigint;
BEGIN
    IF p_organization_id IS NULL OR p_expected_revision < 1 OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.organization.disable',
        p_request_sha256, p_organization_id
    );
    IF FOUND THEN RETURN; END IF;
    UPDATE iam_organizations SET status = 'suspended', revision = revision + 1,
        updated_at = v_now WHERE id = p_organization_id
        AND revision = p_expected_revision AND status = 'active'
    RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN RAISE EXCEPTION USING ERRCODE = '40001',
        MESSAGE = 'IAM identity organization disable compare-and-swap failed'; END IF;
    UPDATE iam_users AS principal
    SET authorization_version = principal.authorization_version + 1,
        revision = principal.revision + 1, updated_at = v_now
    WHERE principal.status <> 'tombstoned' AND EXISTS (
        SELECT 1 FROM iam_organization_memberships AS membership
        WHERE membership.organization_id = p_organization_id
          AND membership.user_id = principal.id AND membership.state <> 'tombstoned'
    );
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.organization.disable', p_request_sha256,
        p_organization_id, v_revision, p_request_sha256
    );
    RETURN QUERY SELECT p_organization_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_organization_disable_v1(
    name,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) FROM PUBLIC;
