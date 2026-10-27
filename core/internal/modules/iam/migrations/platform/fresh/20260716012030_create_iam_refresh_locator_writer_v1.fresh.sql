-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_refresh_locator_create_v1(
    p_selector bytea,
    p_tenant_id uuid,
    p_token_id uuid,
    p_family_id uuid,
    p_group_generation bigint,
    p_cutover_generation bigint,
    p_source_revision bigint,
    p_source_sha256 bytea,
    p_source_token_state text,
    p_issued_at timestamptz,
    p_projected_at timestamptz,
    p_family_absolute_expires_at timestamptz,
    p_retain_until timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM refresh locator create session is invalid';
    END IF;
    PERFORM platform.iam_assert_current_tenant_authority_v1(
        p_tenant_id, p_group_generation, p_cutover_generation
    );

    INSERT INTO platform.iam_refresh_locators (
        selector, tenant_id, token_id, family_id, scope_kind, scope_id,
        authority_group_id, authority_group_generation, authority_domain,
        cutover_generation, source_revision, source_sha256,
        source_token_state, issued_at, projected_at,
        family_absolute_expires_at, retain_until
    ) VALUES (
        p_selector, p_tenant_id, p_token_id, p_family_id, 'tenant', p_tenant_id,
        'iam_system_resources', p_group_generation, 'iam',
        p_cutover_generation, p_source_revision, p_source_sha256,
        p_source_token_state, p_issued_at, p_projected_at,
        p_family_absolute_expires_at, p_retain_until
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_refresh_locator_create_v1(
    bytea,uuid,uuid,uuid,bigint,bigint,bigint,bytea,text,
    timestamptz,timestamptz,timestamptz,timestamptz
) FROM PUBLIC;
