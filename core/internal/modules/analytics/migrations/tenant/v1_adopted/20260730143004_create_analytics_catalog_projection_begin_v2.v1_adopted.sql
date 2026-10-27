-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_projection_begin_v1(
    p_scope_schema name, p_receipt_id uuid,
    p_package_id uuid, p_package_slug text, p_package_version text,
    p_release_id uuid, p_component_id uuid, p_owner_generation bigint,
    p_contract text, p_content_sha256 text, p_component_kind text,
    p_module_slug text, p_semantic_identity text, p_semantic_sha256 bytea,
    p_item_count integer, p_auxiliary_count integer, p_document_bytes bytea,
    p_runtime_plan_bytes bytea, p_runtime_plan_sha256 bytea, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea, p_occurred_at timestamptz
)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_original jsonb;
    v_plan jsonb;
    v_kind text;
    v_auxiliary integer;
    v_blocker text;
    v_existing analytics_catalog_provisions%ROWTYPE;
    v_definition analytics_catalog_definitions%ROWTYPE;
    v_component analytics_catalog_components%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       p_package_id IS NULL OR
       p_package_id='00000000-0000-0000-0000-000000000000'::uuid OR
       p_package_slug !~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$' OR
       p_module_slug IS DISTINCT FROM p_package_slug OR
       p_package_version !~
         '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$' OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_owner_generation < 1 OR
       p_contract IS DISTINCT FROM 'analytics.catalog-definition.v1' OR
       p_content_sha256 !~ '^[0-9a-f]{64}$' OR
       p_component_kind NOT IN ('ENDPOINT','DASHBOARD') OR
       p_semantic_identity !~ '^[a-z][a-z0-9-]{1,62}$' OR
       pg_catalog.octet_length(p_semantic_sha256) IS DISTINCT FROM 32 OR
       p_item_count IS DISTINCT FROM 1 OR p_auxiliary_count < 2 OR
       pg_catalog.octet_length(p_runtime_plan_sha256) IS DISTINCT FROM 32 OR
       p_actor_id IS NULL OR
       p_actor_id='00000000-0000-0000-0000-000000000000'::uuid OR
       length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_idempotency_key !~ '^[A-Za-z0-9_.:-]+$' OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) OR
       date_trunc('microseconds',p_occurred_at)<>p_occurred_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog projection input is invalid';
    END IF;
    IF pg_catalog.encode(pg_catalog.sha256(p_document_bytes),'hex') <>
           p_content_sha256 OR
       pg_catalog.sha256(p_runtime_plan_bytes) <> p_runtime_plan_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog projection digest differs';
    END IF;
    BEGIN
        v_original := pg_catalog.convert_from(p_document_bytes,'UTF8')::jsonb;
        v_plan := pg_catalog.convert_from(p_runtime_plan_bytes,'UTF8')::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog projection document is invalid';
    END;
    v_kind := v_original->>'kind';
    IF pg_catalog.jsonb_typeof(v_original) IS DISTINCT FROM 'object' OR
       v_original->>'contract_version' IS DISTINCT FROM p_contract OR
       v_original->>'component_id' IS DISTINCT FROM p_component_id::text OR
       v_original->>'analytics_id' IS DISTINCT FROM p_component_id::text OR
       v_original->>'slug' IS DISTINCT FROM p_semantic_identity OR
       pg_catalog.octet_length(v_original->>'name') NOT BETWEEN 1 AND 100 OR
       pg_catalog.octet_length(v_original->>'description') NOT BETWEEN 1 AND 2048 OR
       v_original->>'legacy_hash' !~ '^[0-9a-f]{64}$' OR
       v_original->>'legacy_hash' <>
           pg_catalog.encode(p_semantic_sha256,'hex') OR
       v_kind NOT IN ('endpoint','dashboard') OR
       (v_kind='endpoint' AND p_component_kind<>'ENDPOINT') OR
       (v_kind='dashboard' AND p_component_kind<>'DASHBOARD') OR
       pg_catalog.jsonb_typeof(v_original#>'{definition,dataSources}')
           IS DISTINCT FROM 'array' OR
       v_plan->>'contract' IS DISTINCT FROM
           'analytics.catalog-runtime-plan.v1' OR
       v_plan->>'kind' IS DISTINCT FROM v_kind OR
       v_plan->>'status' NOT IN ('candidate','pending') OR
       pg_catalog.jsonb_typeof(v_plan->'sources') IS DISTINCT FROM 'array' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog projection binding differs';
    END IF;
    IF v_kind='endpoint' THEN
        IF pg_catalog.jsonb_typeof(v_original#>'{definition,fields}')
               IS DISTINCT FROM 'array' THEN
            RAISE EXCEPTION USING ERRCODE='22023',
                MESSAGE='Analytics Catalog endpoint definition is invalid';
        END IF;
        v_auxiliary :=
            pg_catalog.jsonb_array_length(v_original#>'{definition,dataSources}') +
            pg_catalog.jsonb_array_length(v_original#>'{definition,fields}');
    ELSE
        IF pg_catalog.jsonb_typeof(v_original#>'{definition,widgets}')
               IS DISTINCT FROM 'array' THEN
            RAISE EXCEPTION USING ERRCODE='22023',
                MESSAGE='Analytics Catalog dashboard definition is invalid';
        END IF;
        v_auxiliary :=
            pg_catalog.jsonb_array_length(v_original#>'{definition,dataSources}') +
            pg_catalog.jsonb_array_length(v_original#>'{definition,widgets}');
    END IF;
    IF v_auxiliary <> p_auxiliary_count OR
       (v_plan->>'status'='pending' AND
        v_plan->>'blocker_code' !~ '^[A-Z][A-Z0-9_]{0,127}$') THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog projection composition differs';
    END IF;
    v_blocker := CASE WHEN v_plan->>'status'='pending'
        THEN v_plan->>'blocker_code'
        ELSE 'ANALYTICS_DEPENDENCY_NOT_READY' END;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:catalog-project:'||p_component_id::text,6301
    ));
    SELECT * INTO v_existing FROM analytics_catalog_provisions AS provision
    WHERE provision.component_id=p_component_id FOR UPDATE;
    IF FOUND THEN
        SELECT * INTO v_definition FROM analytics_catalog_definitions AS definition
        WHERE definition.component_id=p_component_id;
        IF ROW(v_existing.package_id,v_existing.package_slug,
               v_existing.package_version,v_existing.release_id,
               v_existing.owner_generation,v_existing.contract,
               v_existing.content_sha256,v_existing.component_kind,
               v_existing.module_slug,v_existing.semantic_identity,
               v_existing.semantic_sha256,v_existing.item_count,
               v_existing.auxiliary_count,v_existing.idempotency_key,
               v_existing.request_sha256,v_definition.document_bytes)
           IS DISTINCT FROM
           ROW(p_package_id,p_package_slug,p_package_version,p_release_id,
               p_owner_generation,p_contract,p_content_sha256,p_component_kind,
               p_module_slug,p_semantic_identity,p_semantic_sha256,p_item_count,
               p_auxiliary_count,p_idempotency_key,p_request_sha256,
               p_document_bytes) THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='Analytics Catalog projection replay differs';
        END IF;
        IF v_definition.operational_state='pending' AND
           v_definition.runtime_plan_sha256<>p_runtime_plan_sha256 THEN
            UPDATE analytics_catalog_definitions SET runtime_plan_bytes=p_runtime_plan_bytes,
                runtime_plan_sha256=p_runtime_plan_sha256,blocker_code=v_blocker,
                updated_at=p_occurred_at WHERE component_id=p_component_id;
        END IF;
        RETURN v_existing.receipt_id;
    END IF;
    SELECT * INTO v_component FROM analytics_catalog_components AS component
    WHERE component.component_id=p_component_id FOR UPDATE;
    IF NOT FOUND THEN
        INSERT INTO analytics_catalog_components(
            component_id,component_kind,contract,content_sha256,
            owner_generation,writer_state,lifecycle_state,updated_at
        ) VALUES (p_component_id,p_component_kind,p_contract,p_content_sha256,
            p_owner_generation,'v2','AVAILABLE',p_occurred_at);
    ELSIF ROW(v_component.component_kind,v_component.contract,
              v_component.content_sha256,v_component.owner_generation,
              v_component.writer_state)
          IS DISTINCT FROM ROW(p_component_kind,p_contract,p_content_sha256,
              p_owner_generation,'v2'::text) OR
          v_component.lifecycle_state NOT IN ('AVAILABLE','STAGED') THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog component authority differs';
    END IF;
    INSERT INTO analytics_catalog_definitions(
        component_id,analytics_kind,module_slug,slug,name,description,
        legacy_hash,definition_json,definition_sha256,document_bytes,
        content_sha256,semantic_sha256,runtime_plan_bytes,runtime_plan_sha256,
        package_id,package_slug,package_version,release_id,
        operational_state,blocker_code,installed_at,updated_at
    ) VALUES (p_component_id,v_kind,p_module_slug,p_semantic_identity,
        v_original->>'name',v_original->>'description',v_original->>'legacy_hash',
        v_original->'definition',pg_catalog.sha256(pg_catalog.convert_to(
            (v_original->'definition')::text,'UTF8')),p_document_bytes,
        p_content_sha256,p_semantic_sha256,p_runtime_plan_bytes,
        p_runtime_plan_sha256,p_package_id,p_package_slug,p_package_version,
        p_release_id,'pending',v_blocker,p_occurred_at,p_occurred_at);
    INSERT INTO analytics_catalog_provisions(
        receipt_id,package_id,package_slug,package_version,release_id,
        component_id,owner_generation,contract,content_sha256,component_kind,
        module_slug,semantic_identity,semantic_sha256,item_count,auxiliary_count,
        actor_id,idempotency_key,request_sha256,created_at,updated_at
    ) VALUES (p_receipt_id,p_package_id,p_package_slug,p_package_version,
        p_release_id,p_component_id,p_owner_generation,p_contract,
        p_content_sha256,p_component_kind,p_module_slug,p_semantic_identity,
        p_semantic_sha256,p_item_count,p_auxiliary_count,p_actor_id,
        p_idempotency_key,p_request_sha256,p_occurred_at,p_occurred_at);
    RETURN p_receipt_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_catalog_projection_begin_v1(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz
) FROM PUBLIC;
