-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_owner_preflight_v2(
    p_scope_schema name,p_component_id uuid,p_package_id uuid,
    p_generation bigint,p_contract text,p_content_sha256 text,
    p_reference_epoch bigint
)
RETURNS TABLE(
    blockers jsonb,resource_id uuid,version_id uuid,head_id uuid
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_resource_id uuid;
    v_version_id uuid;
    v_head_id uuid;
    v_lifecycle_state text;
    v_resource_reference_epoch bigint;
    v_blockers jsonb;
    v_snapshot_resource_id uuid;
    v_snapshot_version_id uuid;
    v_snapshot_reference_epoch bigint;
BEGIN
    IF p_reference_epoch<1 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog owner preflight input is invalid';
    END IF;
    SELECT binding.resource_id,binding.version_id,binding.head_id,
        binding.lifecycle_state,binding.resource_reference_epoch
    INTO STRICT v_resource_id,v_version_id,v_head_id,
        v_lifecycle_state,v_resource_reference_epoch
    FROM analytics_catalog_owner_binding_v2(
        p_scope_schema,p_component_id,p_package_id,p_generation,
        p_contract,p_content_sha256
    ) AS binding;
    IF v_lifecycle_state<>'RETIRED' THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog owner preflight lifecycle differs';
    END IF;
    SELECT snapshot.blockers,snapshot.root_id,
        snapshot.head_version_id,snapshot.root_reference_epoch
    INTO STRICT v_blockers,v_snapshot_resource_id,
        v_snapshot_version_id,v_snapshot_reference_epoch
    FROM analytics_catalog_blocker_snapshot_v2(
        p_component_id,p_reference_epoch,p_generation
    ) AS snapshot;
    IF v_snapshot_resource_id<>v_resource_id OR
       v_snapshot_version_id<>v_version_id OR
       v_snapshot_reference_epoch<>v_resource_reference_epoch THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog owner preflight target differs';
    END IF;
    RETURN QUERY SELECT v_blockers,v_resource_id,v_version_id,v_head_id;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_owner_preflight_v2(
    name,uuid,uuid,bigint,text,text,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_catalog_owner_preflight_v2(
    name,uuid,uuid,bigint,text,text,bigint
) TO forgecrud_runtime;
