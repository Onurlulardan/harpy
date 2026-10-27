-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_projection_begin_v2(
    p_scope_schema name,p_receipt_id uuid,
    p_package_id uuid,p_package_slug text,p_package_version text,
    p_release_id uuid,p_component_id uuid,p_owner_generation bigint,
    p_contract text,p_content_sha256 text,p_component_kind text,
    p_module_slug text,p_semantic_identity text,p_semantic_sha256 bytea,
    p_item_count integer,p_auxiliary_count integer,p_document_bytes bytea,
    p_runtime_plan_bytes bytea,p_runtime_plan_sha256 bytea,p_actor_id uuid,
    p_idempotency_key text,p_request_sha256 bytea,p_occurred_at timestamptz
)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_analytics_kind text;
    v_name text;
    v_description text;
    v_legacy_hash text;
    v_definition jsonb;
    v_definition_sha256 bytea;
    v_blocker_code text;
    v_existing analytics_catalog_version_projections%ROWTYPE;
    v_component analytics_catalog_components%ROWTYPE;
    v_has_head boolean;
BEGIN
    IF pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       p_package_id IS NULL OR
       p_package_id='00000000-0000-0000-0000-000000000000'::uuid OR
       p_package_slug IS NULL OR
       p_package_slug !~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$' OR
       p_module_slug IS DISTINCT FROM p_package_slug OR
       p_package_version IS NULL OR p_package_version !~
         '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$' OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       p_owner_generation<1 OR p_actor_id IS NULL OR
       p_actor_id='00000000-0000-0000-0000-000000000000'::uuid OR
       p_idempotency_key IS NULL OR
       length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_idempotency_key !~ '^[A-Za-z0-9_.:-]+$' OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) OR
       date_trunc('microseconds',p_occurred_at)<>p_occurred_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog projection input is invalid';
    END IF;
    SELECT valid.analytics_kind,valid.name,valid.description,
        valid.legacy_hash,valid.definition_json,valid.definition_sha256,
        valid.blocker_code
    INTO STRICT v_analytics_kind,v_name,v_description,v_legacy_hash,
        v_definition,v_definition_sha256,v_blocker_code
    FROM analytics_catalog_projection_validate_v2(
        p_contract,p_content_sha256,p_component_id,p_component_kind,
        p_module_slug,p_semantic_identity,p_semantic_sha256,p_item_count,
        p_auxiliary_count,p_document_bytes,p_runtime_plan_bytes,
        p_runtime_plan_sha256
    ) AS valid;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:catalog-project:'||p_component_id::text,6301
    ));
    SELECT * INTO v_existing
    FROM analytics_catalog_version_projections AS projection
    WHERE projection.package_id=p_package_id
      AND projection.package_version=p_package_version
      AND projection.component_id=p_component_id FOR UPDATE;
    IF FOUND THEN
        IF ROW(
            v_existing.package_slug,v_existing.release_id,
            v_existing.owner_generation,v_existing.contract,
            v_existing.content_sha256,v_existing.component_kind,
            v_existing.module_slug,v_existing.semantic_identity,
            v_existing.semantic_sha256,v_existing.item_count,
            v_existing.auxiliary_count,v_existing.document_bytes
        ) IS DISTINCT FROM ROW(
            p_package_slug,p_release_id,p_owner_generation,p_contract,
            p_content_sha256,p_component_kind,p_module_slug,
            p_semantic_identity,p_semantic_sha256,p_item_count,
            p_auxiliary_count,p_document_bytes
        ) THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='Analytics Catalog projection artifact differs';
        END IF;
        IF v_existing.runtime_plan_sha256<>p_runtime_plan_sha256 THEN
            IF v_existing.operational_state<>'pending' THEN
                RAISE EXCEPTION USING ERRCODE='23505',
                    MESSAGE='Analytics Catalog ready runtime plan differs';
            END IF;
            UPDATE analytics_catalog_version_projections
            SET runtime_plan_bytes=p_runtime_plan_bytes,
                runtime_plan_sha256=p_runtime_plan_sha256,
                blocker_code=v_blocker_code,updated_at=p_occurred_at
            WHERE receipt_id=v_existing.receipt_id;
        END IF;
        RETURN v_existing.receipt_id;
    END IF;
    SELECT * INTO v_component
    FROM analytics_catalog_components AS component
    WHERE component.component_id=p_component_id FOR UPDATE;
    IF NOT FOUND THEN
        INSERT INTO analytics_catalog_components(
            component_id,component_kind,contract,content_sha256,
            owner_generation,writer_state,lifecycle_state,updated_at
        ) VALUES (
            p_component_id,p_component_kind,p_contract,p_content_sha256,
            p_owner_generation,'v2','AVAILABLE',p_occurred_at
        );
    ELSE
        SELECT EXISTS(
            SELECT 1 FROM analytics_catalog_component_heads AS head
            WHERE head.component_id=p_component_id
        ) INTO v_has_head;
        IF v_component.component_kind<>p_component_kind OR
           v_component.contract<>p_contract OR
           v_component.owner_generation<>p_owner_generation OR
           v_component.writer_state<>'v2' OR
           v_component.lifecycle_state NOT IN
               ('AVAILABLE','STAGED','ACTIVE','DISABLED','RETIRED') OR
           (NOT v_has_head AND
            v_component.content_sha256<>p_content_sha256) THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog component authority differs';
        END IF;
    END IF;
    INSERT INTO analytics_catalog_version_projections(
        receipt_id,package_id,package_slug,package_version,release_id,
        component_id,owner_generation,contract,content_sha256,component_kind,
        analytics_kind,module_slug,semantic_identity,semantic_sha256,
        item_count,auxiliary_count,name,description,legacy_hash,
        definition_json,definition_sha256,document_bytes,runtime_plan_bytes,
        runtime_plan_sha256,operational_state,blocker_code,created_at,updated_at
    ) VALUES (
        p_receipt_id,p_package_id,p_package_slug,p_package_version,p_release_id,
        p_component_id,p_owner_generation,p_contract,p_content_sha256,
        p_component_kind,v_analytics_kind,p_module_slug,
        p_semantic_identity,p_semantic_sha256,p_item_count,p_auxiliary_count,
        v_name,v_description,v_legacy_hash,
        v_definition,v_definition_sha256,p_document_bytes,
        p_runtime_plan_bytes,p_runtime_plan_sha256,'pending',
        v_blocker_code,p_occurred_at,p_occurred_at
    );
    RETURN p_receipt_id;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_projection_begin_v2(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz
) FROM PUBLIC;
