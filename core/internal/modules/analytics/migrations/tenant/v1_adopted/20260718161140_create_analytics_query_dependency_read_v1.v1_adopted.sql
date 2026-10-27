-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_query_dependencies_v1(
    p_scope_schema name, p_query_id uuid, p_version_id uuid
)
RETURNS TABLE (
    owner_domain text, resource_kind text, resource_id uuid,
    resource_version_id uuid, owner_generation bigint,
    catalog_version bigint, query_contract_version bigint,
    policy_version bigint, data_revision_capability text,
    dependency_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM 1
    FROM analytics_saved_query_versions AS version_row
    WHERE version_row.id = p_version_id
      AND version_row.query_id = p_query_id
      AND version_row.version_state = 'published';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics query version was not found';
    END IF;
    RETURN QUERY
    SELECT dependency.owner_domain, dependency.resource_kind,
        dependency.resource_id, dependency.resource_version_id,
        dependency.owner_generation, dependency.catalog_version,
        dependency.query_contract_version, dependency.policy_version,
        dependency.data_revision_capability, dependency.dependency_sha256
    FROM analytics_query_dependencies AS dependency
    WHERE dependency.query_version_id = p_version_id
    ORDER BY dependency.ordinal;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics query dependencies were not found';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_query_dependencies_v1(name,uuid,uuid) FROM PUBLIC;
