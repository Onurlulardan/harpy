-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_organization_update_v2(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_organization_id uuid, p_expected_revision bigint,
    p_name text, p_slug text, p_parent_id uuid,
    p_profile_schema_id uuid, p_profile_document bytea, p_profile_sha256 bytea,
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
    v_values jsonb; v_revision bigint; v_clock bigint;
    v_cursor uuid; v_steps integer;
BEGIN
    IF p_organization_id IS NULL OR p_expected_revision < 1 OR
       p_parent_id IS NOT DISTINCT FROM p_organization_id OR
       pg_catalog.char_length(p_name) NOT BETWEEN 1 AND 255 OR
       p_name <> pg_catalog.btrim(p_name) OR
       p_slug !~ '^[a-z0-9]([a-z0-9-]{0,98}[a-z0-9])?$' OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '23514',
          MESSAGE = 'IAM identity organization update input is invalid'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, 'iam.identity.organization.update',
        p_request_sha256, p_organization_id
    );
    IF FOUND THEN RETURN; END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended('harpy:iam:organization-tree', 0)
    );
    IF p_parent_id IS NOT NULL THEN
        IF NOT EXISTS (SELECT 1 FROM iam_organizations
                       WHERE id = p_parent_id AND status = 'active') THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'IAM identity organization parent is not active';
        END IF;
        v_cursor := p_parent_id; v_steps := 0;
        WHILE v_cursor IS NOT NULL LOOP
            IF v_cursor = p_organization_id THEN
                RAISE EXCEPTION USING ERRCODE = '23514',
                    MESSAGE = 'IAM identity organization parent would create a cycle';
            END IF;
            v_steps := v_steps + 1;
            IF v_steps > 10000 THEN
                RAISE EXCEPTION USING ERRCODE = '54001',
                    MESSAGE = 'IAM identity organization hierarchy is too deep';
            END IF;
            SELECT parent_id INTO v_cursor FROM iam_organizations WHERE id = v_cursor;
        END LOOP;
    END IF;
    v_values := iam_identity_profile_assert_v2(
        'organization', p_profile_schema_id, p_profile_document, p_profile_sha256
    );
    UPDATE iam_organizations SET name = p_name, slug = p_slug, parent_id = p_parent_id,
        profile_schema_version_id = p_profile_schema_id, profile_values = v_values,
        profile_sha256 = p_profile_sha256, revision = revision + 1,
        updated_at = pg_catalog.transaction_timestamp()
    WHERE id = p_organization_id AND revision = p_expected_revision
      AND status <> 'tombstoned' RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN RAISE EXCEPTION USING ERRCODE = '40001',
        MESSAGE = 'IAM identity organization update compare-and-swap failed'; END IF;
    v_clock := iam_identity_record_v1(
        p_operation_id, 'iam.identity.organization.update', p_request_sha256,
        p_organization_id, v_revision, p_profile_sha256
    );
    RETURN QUERY SELECT p_organization_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_organization_update_v2(
    name,uuid,bytea,uuid,bigint,text,text,uuid,uuid,bytea,bytea,
    uuid,uuid,bigint,bigint
) FROM PUBLIC;
