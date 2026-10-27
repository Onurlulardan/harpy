-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_operational_readback_v2(
    p_component_id uuid,p_package_id uuid,p_package_version text,
    p_release_id uuid,p_owner_generation bigint,p_content_sha256 text
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
    FROM iam_catalog_version_resources AS resource
    WHERE resource.component_id=p_component_id
      AND resource.package_id=p_package_id
      AND resource.package_version=p_package_version
      AND resource.release_id=p_release_id
      AND resource.owner_generation=p_owner_generation
      AND resource.content_sha256=p_content_sha256;
    SELECT COALESCE(pg_catalog.sum(cardinality(grant_row.actions)),0)
    INTO v_expected_grants
    FROM iam_catalog_version_grants AS grant_row
    WHERE grant_row.component_id=p_component_id
      AND grant_row.package_id=p_package_id
      AND grant_row.package_version=p_package_version
      AND grant_row.release_id=p_release_id
      AND grant_row.owner_generation=p_owner_generation
      AND grant_row.content_sha256=p_content_sha256;
    SELECT pg_catalog.count(*) INTO v_actual_resources
    FROM iam_catalog_version_operational_resources AS resource
    WHERE resource.component_id=p_component_id
      AND resource.package_id=p_package_id
      AND resource.package_version=p_package_version
      AND resource.release_id=p_release_id
      AND resource.owner_generation=p_owner_generation
      AND resource.content_sha256=p_content_sha256;
    SELECT pg_catalog.count(*) INTO v_actual_grants
    FROM iam_catalog_version_operational_grants AS grant_row
    WHERE grant_row.component_id=p_component_id
      AND grant_row.package_id=p_package_id
      AND grant_row.package_version=p_package_version
      AND grant_row.release_id=p_release_id
      AND grant_row.owner_generation=p_owner_generation
      AND grant_row.content_sha256=p_content_sha256;
    IF v_expected_resources<2 OR v_expected_grants<1 OR
       v_actual_resources<>v_expected_resources OR
       v_actual_grants<>v_expected_grants THEN
        RETURN false;
    END IF;
    SELECT pg_catalog.count(*) INTO v_invalid
    FROM iam_catalog_version_operational_resources AS projected
    JOIN iam_catalog_version_resources AS source
      ON source.component_id=projected.component_id
     AND source.package_id=projected.package_id
     AND source.package_version=projected.package_version
     AND source.release_id=projected.release_id
     AND source.owner_generation=projected.owner_generation
     AND source.content_sha256=projected.content_sha256
     AND source.resource_slug=projected.resource_slug
    JOIN iam_resources AS resource ON resource.id=projected.resource_id
    WHERE projected.component_id=p_component_id
      AND projected.package_id=p_package_id
      AND projected.package_version=p_package_version
      AND projected.release_id=p_release_id
      AND projected.owner_generation=p_owner_generation
      AND projected.content_sha256=p_content_sha256
      AND ROW(
        resource.resource_key,resource.resource_kind,
        resource.display_name,COALESCE(resource.description,''),
        resource.is_protected,resource.state
      ) IS DISTINCT FROM ROW(
        source.resource_slug,source.resource_kind,
        source.name,source.description,false,'active'::text
      );
    IF v_invalid<>0 THEN
        RETURN false;
    END IF;
    SELECT pg_catalog.count(*) INTO v_invalid
    FROM iam_catalog_version_operational_grants AS projected
    JOIN iam_catalog_version_grants AS source
      ON source.component_id=projected.component_id
     AND source.package_id=projected.package_id
     AND source.package_version=projected.package_version
     AND source.release_id=projected.release_id
     AND source.owner_generation=projected.owner_generation
     AND source.content_sha256=projected.content_sha256
     AND source.resource_slug=projected.resource_slug
     AND source.role_name=projected.role_name
    JOIN iam_roles AS role ON role.id=projected.role_id
    JOIN iam_resources AS resource ON resource.id=projected.resource_id
    JOIN iam_actions AS action ON action.id=projected.action_id
    JOIN iam_permissions AS permission ON permission.id=projected.permission_id
    JOIN iam_role_permission_bindings AS binding ON binding.id=projected.binding_id
    WHERE projected.component_id=p_component_id
      AND projected.package_id=p_package_id
      AND projected.package_version=p_package_version
      AND projected.release_id=p_release_id
      AND projected.owner_generation=p_owner_generation
      AND projected.content_sha256=p_content_sha256
      AND (
        NOT projected.action_key=ANY(source.actions) OR
        role.name IS DISTINCT FROM projected.role_name OR
        role.status<>'active' OR role.is_protected OR
        resource.resource_key IS DISTINCT FROM projected.resource_slug OR
        resource.state<>'active' OR
        action.action_key IS DISTINCT FROM projected.action_key OR
        action.state<>'active' OR permission.match_kind<>'exact' OR
        permission.resource_id IS DISTINCT FROM projected.resource_id OR
        permission.action_id IS DISTINCT FROM projected.action_id OR
        permission.state<>'active' OR
        binding.role_id IS DISTINCT FROM projected.role_id OR
        binding.permission_id IS DISTINCT FROM projected.permission_id OR
        binding.permission_match_kind<>'exact' OR binding.state<>'active'
      );
    RETURN v_invalid=0;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_catalog_operational_readback_v2(
    uuid,uuid,text,uuid,bigint,text
) FROM PUBLIC;
