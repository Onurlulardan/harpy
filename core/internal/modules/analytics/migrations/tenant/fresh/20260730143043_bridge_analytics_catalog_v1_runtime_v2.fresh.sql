-- +goose Up
DROP FUNCTION analytics_catalog_project_v1(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz
);
DROP FUNCTION analytics_catalog_component_apply_v1(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
);

-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_canonical_json_v1(p_value jsonb)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_kind text;
    v_result text;
BEGIN
    v_kind:=pg_catalog.jsonb_typeof(p_value);
    IF v_kind='object' THEN
        SELECT '{'||COALESCE(pg_catalog.string_agg(
            pg_catalog.to_json(entry.key)::text||':'||
            analytics_catalog_canonical_json_v1(entry.value),
            ',' ORDER BY entry.key COLLATE "C"
        ),'')||'}' INTO v_result
        FROM pg_catalog.jsonb_each(p_value) AS entry(key,value);
        RETURN v_result;
    ELSIF v_kind='array' THEN
        SELECT '['||COALESCE(pg_catalog.string_agg(
            analytics_catalog_canonical_json_v1(entry.value),
            ',' ORDER BY entry.ordinality
        ),'')||']' INTO v_result
        FROM pg_catalog.jsonb_array_elements(p_value)
            WITH ORDINALITY AS entry(value,ordinality);
        RETURN v_result;
    ELSIF v_kind IN ('string','number','boolean','null') THEN
        RETURN p_value::text;
    END IF;
    RAISE EXCEPTION USING ERRCODE='22023',
        MESSAGE='Analytics Catalog canonical JSON input is invalid';
END;
$function$;

CREATE FUNCTION analytics_catalog_project_v1(
    p_scope_schema name,p_receipt_id uuid,
    p_package_id uuid,p_package_slug text,p_package_version text,
    p_release_id uuid,p_component_id uuid,p_owner_generation bigint,
    p_contract text,p_content_sha256 text,p_component_kind text,
    p_module_slug text,p_semantic_identity text,p_semantic_sha256 bytea,
    p_item_count integer,p_auxiliary_count integer,p_document_bytes bytea,
    p_runtime_plan_bytes bytea,p_runtime_plan_sha256 bytea,p_actor_id uuid,
    p_idempotency_key text,p_request_sha256 bytea,p_occurred_at timestamptz
)
RETURNS TABLE(
    receipt_id uuid,package_id uuid,package_slug text,
    package_version text,release_id uuid,component_id uuid,
    owner_generation bigint,contract text,content_sha256 text,
    component_kind text,module_slug text,semantic_identity text,
    semantic_sha256 bytea,item_count integer,auxiliary_count integer,
    document_bytes bytea,idempotency_key text,request_sha256 bytea,
    lifecycle_state text,ready_at timestamptz,
    created_at timestamptz,updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_document jsonb;
    v_plan jsonb;
    v_source_contract text;
    v_runtime_plan_bytes bytea;
    v_runtime_plan_sha256 bytea;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    IF p_runtime_plan_bytes IS NULL OR
       pg_catalog.octet_length(p_runtime_plan_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.sha256(p_runtime_plan_bytes)
           IS DISTINCT FROM p_runtime_plan_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog legacy runtime plan digest differs';
    END IF;
    BEGIN
        v_document:=pg_catalog.convert_from(p_document_bytes,'UTF8')::jsonb;
        v_plan:=pg_catalog.convert_from(p_runtime_plan_bytes,'UTF8')::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog legacy runtime plan is invalid';
    END;
    v_source_contract:=v_document->>'source_contract';
    IF pg_catalog.jsonb_typeof(v_document) IS DISTINCT FROM 'object' OR
       pg_catalog.jsonb_typeof(v_plan) IS DISTINCT FROM 'object' OR
       v_source_contract IS NULL OR
       (v_plan ? 'component_id' AND v_plan->>'component_id'
           IS DISTINCT FROM p_component_id::text) OR
       (v_plan ? 'source_contract' AND v_plan->>'source_contract'
           IS DISTINCT FROM v_source_contract) OR
       (v_plan ? 'document_sha256' AND v_plan->>'document_sha256'
           IS DISTINCT FROM p_content_sha256) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog legacy runtime plan binding differs';
    END IF;
    v_plan:=v_plan||pg_catalog.jsonb_build_object(
        'component_id',p_component_id::text,
        'source_contract',v_source_contract,
        'document_sha256',p_content_sha256
    );
    v_runtime_plan_bytes:=pg_catalog.convert_to(
        analytics_catalog_canonical_json_v1(v_plan),'UTF8'
    );
    v_runtime_plan_sha256:=pg_catalog.sha256(v_runtime_plan_bytes);
    RETURN QUERY
    SELECT result.receipt_id,result.package_id,result.package_slug,
           result.package_version,result.release_id,result.component_id,
           result.owner_generation,result.contract,result.content_sha256,
           result.component_kind,result.module_slug,result.semantic_identity,
           result.semantic_sha256,result.item_count,result.auxiliary_count,
           result.document_bytes,result.idempotency_key,result.request_sha256,
           result.lifecycle_state,result.ready_at,
           result.created_at,result.updated_at
    FROM analytics_catalog_project_v2(
        p_scope_schema,p_receipt_id,p_package_id,p_package_slug,
        p_package_version,p_release_id,p_component_id,p_owner_generation,
        p_contract,p_content_sha256,p_component_kind,p_module_slug,
        p_semantic_identity,p_semantic_sha256,p_item_count,p_auxiliary_count,
        p_document_bytes,v_runtime_plan_bytes,v_runtime_plan_sha256,p_actor_id,
        p_idempotency_key,p_request_sha256,p_occurred_at
    ) AS result;
END;
$function$;

CREATE FUNCTION analytics_catalog_component_apply_v1(
    p_receipt_id uuid,p_operation_id uuid,p_component_id uuid,
    p_owner text,p_generation bigint,p_contract text,p_content_sha256 text,
    p_action text,p_reference_epoch bigint,p_idempotency_key text,
    p_completed_at timestamptz
)
RETURNS TABLE(
    receipt_id uuid,operation_id uuid,component_id uuid,owner_module text,
    owner_generation bigint,contract text,content_sha256 text,
    action text,ready boolean,completed_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    RETURN QUERY
    SELECT result.receipt_id,result.operation_id,result.component_id,
           result.owner_module,result.owner_generation,result.contract,
           result.content_sha256,result.action,result.ready,result.completed_at
    FROM analytics_catalog_component_apply_v2(
        p_receipt_id,p_operation_id,p_component_id,p_owner,p_generation,
        p_contract,p_content_sha256,p_action,p_reference_epoch,
        p_idempotency_key,p_completed_at
    ) AS result;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_catalog_project_v1(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_catalog_component_apply_v1(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_catalog_canonical_json_v1(jsonb)
    FROM PUBLIC;

GRANT EXECUTE ON FUNCTION analytics_catalog_project_v1(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION analytics_catalog_component_apply_v1(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) TO harpy_runtime;
