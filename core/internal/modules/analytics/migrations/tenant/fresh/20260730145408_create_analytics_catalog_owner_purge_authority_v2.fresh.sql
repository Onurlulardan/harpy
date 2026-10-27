-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_owner_purge_authority_v2(
    p_scope_schema name,p_operation_id uuid,p_component_id uuid,
    p_package_id uuid,p_generation bigint,p_contract text,
    p_content_sha256 text,p_reference_epoch bigint,
    p_authority_sha256 text,p_resource_id uuid,p_version_id uuid,
    p_head_id uuid,p_recorded_at timestamptz
)
RETURNS SETOF analytics_catalog_owner_purge_authorities_v2
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_resource_id uuid;
    v_version_id uuid;
    v_head_id uuid;
    v_lifecycle_state text;
    v_existing analytics_catalog_owner_purge_authorities_v2%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       p_reference_epoch<1 OR
       p_authority_sha256!~'^[0-9a-f]{64}$' OR
       pg_catalog.uuid_extract_version(p_resource_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_version_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_head_id) IS DISTINCT FROM 7 OR
       p_recorded_at IS NULL OR NOT pg_catalog.isfinite(p_recorded_at) OR
       date_trunc('microseconds',p_recorded_at)<>p_recorded_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog purge authority input is invalid';
    END IF;
    SELECT binding.resource_id,binding.version_id,binding.head_id,
        binding.lifecycle_state
    INTO STRICT v_resource_id,v_version_id,v_head_id,v_lifecycle_state
    FROM analytics_catalog_owner_binding_v2(
        p_scope_schema,p_component_id,p_package_id,p_generation,
        p_contract,p_content_sha256
    ) AS binding;
    IF v_lifecycle_state<>'RETIRED' OR
       v_resource_id<>p_resource_id OR
       v_version_id<>p_version_id OR v_head_id<>p_head_id THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge authority target differs';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_operation_id::text||':'||p_component_id::text,145408
    ));
    SELECT * INTO v_existing
    FROM analytics_catalog_owner_purge_authorities_v2 AS authority
    WHERE authority.operation_id=p_operation_id
      AND authority.component_id=p_component_id
    FOR SHARE;
    IF FOUND THEN
        IF v_existing.reference_epoch<>p_reference_epoch OR
           v_existing.authority_sha256<>p_authority_sha256 OR
           v_existing.resource_id<>p_resource_id OR
           v_existing.version_id<>p_version_id OR
           v_existing.head_id<>p_head_id OR
           v_existing.recorded_at<>p_recorded_at THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='Analytics Catalog purge authority differs';
        END IF;
    ELSE
        INSERT INTO analytics_catalog_owner_purge_authorities_v2(
            operation_id,component_id,reference_epoch,authority_sha256,
            resource_id,version_id,head_id,recorded_at
        ) VALUES (
            p_operation_id,p_component_id,p_reference_epoch,
            p_authority_sha256,p_resource_id,p_version_id,p_head_id,
            p_recorded_at
        ) RETURNING * INTO v_existing;
    END IF;
    RETURN NEXT v_existing;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_owner_purge_authority_v2(
    name,uuid,uuid,uuid,bigint,text,text,bigint,text,uuid,uuid,uuid,
    timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_catalog_owner_purge_authority_v2(
    name,uuid,uuid,uuid,bigint,text,text,bigint,text,uuid,uuid,uuid,
    timestamptz
) TO forgecrud_runtime;
