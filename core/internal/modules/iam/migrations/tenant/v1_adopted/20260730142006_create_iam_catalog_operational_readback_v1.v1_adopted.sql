-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_operational_readback_v1(
    p_component_id uuid
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_expected_resources bigint;
    v_expected_grants bigint;
    v_actual_resources bigint;
    v_actual_grants bigint;
    v_invalid bigint;
BEGIN
    SELECT pg_catalog.count(*) INTO v_expected_resources
    FROM iam_catalog_resource_projections AS resource
    WHERE resource.component_id=p_component_id;
    SELECT COALESCE(pg_catalog.sum(cardinality(grant_row.actions)),0)
    INTO v_expected_grants
    FROM iam_catalog_grant_projections AS grant_row
    WHERE grant_row.component_id=p_component_id;
    SELECT pg_catalog.count(*) INTO v_actual_resources
    FROM iam_catalog_operational_resources AS resource
    WHERE resource.component_id=p_component_id;
    SELECT pg_catalog.count(*) INTO v_actual_grants
    FROM iam_catalog_operational_grants AS grant_row
    WHERE grant_row.component_id=p_component_id;
    IF v_expected_resources < 2 OR v_expected_grants < 1 OR
       v_actual_resources <> v_expected_resources OR
       v_actual_grants <> v_expected_grants THEN
        RETURN false;
    END IF;
    SELECT pg_catalog.count(*) INTO v_invalid
    FROM iam_catalog_operational_resources AS projected
    JOIN iam_catalog_resource_projections AS source
      ON source.component_id=projected.component_id
     AND source.resource_slug=projected.resource_slug
    JOIN iam_resources AS resource ON resource.id=projected.resource_id
    WHERE projected.component_id=p_component_id
      AND ROW(resource.resource_key,resource.resource_kind,
              resource.display_name,COALESCE(resource.description,''),
              resource.is_protected,resource.state)
          IS DISTINCT FROM
          ROW(source.resource_slug,source.resource_kind,
              source.name,source.description,false,'active'::text);
    IF v_invalid <> 0 THEN
        RETURN false;
    END IF;
    SELECT pg_catalog.count(*) INTO v_invalid
    FROM iam_catalog_operational_grants AS projected
    JOIN iam_roles AS role ON role.id=projected.role_id
    JOIN iam_resources AS resource ON resource.id=projected.resource_id
    JOIN iam_actions AS action ON action.id=projected.action_id
    JOIN iam_permissions AS permission ON permission.id=projected.permission_id
    JOIN iam_role_permission_bindings AS binding
      ON binding.id=projected.binding_id
    WHERE projected.component_id=p_component_id
      AND (
        role.name IS DISTINCT FROM projected.role_name OR
        role.status IS DISTINCT FROM 'active' OR
        role.is_protected IS DISTINCT FROM false OR
        resource.resource_key IS DISTINCT FROM projected.resource_slug OR
        resource.state IS DISTINCT FROM 'active' OR
        action.action_key IS DISTINCT FROM projected.action_key OR
        action.state IS DISTINCT FROM 'active' OR
        permission.match_kind IS DISTINCT FROM 'exact' OR
        permission.resource_id IS DISTINCT FROM projected.resource_id OR
        permission.action_id IS DISTINCT FROM projected.action_id OR
        permission.state IS DISTINCT FROM 'active' OR
        binding.role_id IS DISTINCT FROM projected.role_id OR
        binding.permission_id IS DISTINCT FROM projected.permission_id OR
        binding.permission_match_kind IS DISTINCT FROM 'exact' OR
        binding.state IS DISTINCT FROM 'active'
      );
    RETURN v_invalid=0;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_catalog_operational_readback_v1(uuid) FROM PUBLIC;
