-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_projection_materialize_v1(
    p_scope_schema name, p_receipt_id uuid, p_occurred_at timestamptz
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_provision analytics_catalog_provisions%ROWTYPE;
    v_definition analytics_catalog_definitions%ROWTYPE;
    v_component analytics_catalog_components%ROWTYPE;
    v_original jsonb;
    v_plan jsonb;
    v_state text;
    v_blocker text;
    v_resource_id uuid;
    v_published_id uuid;
    v_draft_id uuid;
BEGIN
    IF pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) OR
       date_trunc('microseconds',p_occurred_at)<>p_occurred_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog materialization input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT * INTO v_provision FROM analytics_catalog_provisions AS provision
    WHERE provision.receipt_id=p_receipt_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog provision is unavailable';
    END IF;
    SELECT * INTO v_definition FROM analytics_catalog_definitions AS definition
    WHERE definition.component_id=v_provision.component_id FOR UPDATE;
    SELECT * INTO v_component FROM analytics_catalog_components AS component
    WHERE component.component_id=v_provision.component_id FOR UPDATE;
    IF NOT FOUND OR v_component.owner_generation<>v_provision.owner_generation OR
       v_component.contract<>v_provision.contract OR
       v_component.content_sha256<>v_provision.content_sha256 OR
       v_component.writer_state<>'v2' OR
       v_component.lifecycle_state NOT IN ('AVAILABLE','STAGED','ACTIVE') OR
       p_occurred_at<v_provision.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog materialization authority differs';
    END IF;
    IF v_provision.ready_at IS NOT NULL THEN
        IF v_definition.operational_state<>'ready' OR
           v_definition.operational_resource_id IS NULL OR
           v_definition.published_version_id IS NULL OR
           v_definition.current_draft_version_id IS NULL THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog ready evidence differs';
        END IF;
        RETURN;
    END IF;
    v_original := pg_catalog.convert_from(v_definition.document_bytes,'UTF8')::jsonb;
    v_plan := pg_catalog.convert_from(v_definition.runtime_plan_bytes,'UTF8')::jsonb;
    IF v_definition.analytics_kind='endpoint' THEN
        SELECT result.operational_state,result.blocker_code,
            result.operational_resource_id,result.published_version_id,
            result.current_draft_version_id
        INTO v_state,v_blocker,v_resource_id,v_published_id,v_draft_id
        FROM analytics_catalog_materialize_endpoint_v1(
            p_scope_schema,v_definition.component_id,v_provision.actor_id,
            v_original,v_plan,p_occurred_at
        ) AS result;
    ELSIF v_definition.analytics_kind='dashboard' THEN
        SELECT result.operational_state,result.blocker_code,
            result.operational_resource_id,result.published_version_id,
            result.current_draft_version_id
        INTO v_state,v_blocker,v_resource_id,v_published_id,v_draft_id
        FROM analytics_catalog_materialize_dashboard_v1(
            p_scope_schema,v_definition.component_id,v_provision.actor_id,
            v_original,v_plan,p_occurred_at
        ) AS result;
    ELSE
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog materialization kind differs';
    END IF;
    IF v_state='ready' THEN
        IF pg_catalog.uuid_extract_version(v_resource_id) IS DISTINCT FROM 7 OR
           pg_catalog.uuid_extract_version(v_published_id) IS DISTINCT FROM 7 OR
           pg_catalog.uuid_extract_version(v_draft_id) IS DISTINCT FROM 7 OR
           v_blocker IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog ready result is invalid';
        END IF;
        UPDATE analytics_catalog_definitions AS definition
        SET operational_state='ready',blocker_code=NULL,
            operational_resource_id=v_resource_id,
            published_version_id=v_published_id,
            current_draft_version_id=v_draft_id,updated_at=p_occurred_at
        WHERE definition.component_id=v_definition.component_id;
        UPDATE analytics_catalog_provisions AS provision
        SET ready_at=p_occurred_at,updated_at=p_occurred_at
        WHERE provision.receipt_id=p_receipt_id;
    ELSIF v_state='pending' AND
          v_blocker ~ '^[A-Z][A-Z0-9_]{0,127}$' THEN
        UPDATE analytics_catalog_definitions AS definition
        SET blocker_code=v_blocker,updated_at=p_occurred_at
        WHERE definition.component_id=v_definition.component_id;
        UPDATE analytics_catalog_provisions AS provision
        SET updated_at=p_occurred_at
        WHERE provision.receipt_id=p_receipt_id;
    ELSE
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog materialization result is invalid';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_catalog_projection_materialize_v1(
    name,uuid,timestamptz
) FROM PUBLIC;
