-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_identity_directory_tombstone_v1(
    p_tenant_id uuid, p_group_generation bigint, p_iam_generation bigint,
    p_user_id uuid, p_source_revision bigint
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_tenant_id IS NULL OR p_user_id IS NULL OR
       p_source_revision IS NULL OR p_source_revision < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM identity directory tombstone input is invalid';
    END IF;
    PERFORM platform.iam_assert_current_tenant_authority_v1(
        p_tenant_id, p_group_generation, p_iam_generation
    );
    IF EXISTS (
        SELECT 1 FROM platform.iam_login_directory AS directory
        WHERE directory.tenant_id OPERATOR(pg_catalog.=) p_tenant_id
          AND directory.user_id OPERATOR(pg_catalog.=) p_user_id
          AND directory.source_revision OPERATOR(pg_catalog.>) p_source_revision
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM identity directory tombstone revision is stale';
    END IF;
    UPDATE platform.iam_login_directory AS directory
    SET state = 'tombstoned', source_revision = p_source_revision,
        projected_at = v_now, tombstoned_at = v_now
    WHERE directory.tenant_id OPERATOR(pg_catalog.=) p_tenant_id
      AND directory.user_id OPERATOR(pg_catalog.=) p_user_id
      AND directory.state OPERATOR(pg_catalog.=) 'active';
    RETURN p_source_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_identity_directory_tombstone_v1(
    uuid,bigint,bigint,uuid,bigint
) FROM PUBLIC;
