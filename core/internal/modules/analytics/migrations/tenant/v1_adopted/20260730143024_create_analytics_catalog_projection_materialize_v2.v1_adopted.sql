-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_projection_materialize_v2(
    p_scope_schema name,p_attempt_receipt_id uuid,p_occurred_at timestamptz
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_attempt analytics_catalog_projection_attempts%ROWTYPE;
    v_projection analytics_catalog_version_projections%ROWTYPE;
    v_component analytics_catalog_components%ROWTYPE;
    v_original jsonb;
    v_plan jsonb;
    v_state text;
    v_blocker text;
    v_resource_id uuid;
    v_published_id uuid;
    v_draft_id uuid;
    v_version_seed uuid;
    v_was_activated boolean;
    v_candidate_projection_id uuid;
BEGIN
    IF pg_catalog.uuid_extract_version(p_attempt_receipt_id)
           IS DISTINCT FROM 7 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) OR
       date_trunc('microseconds',p_occurred_at)<>p_occurred_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog materialization input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT attempt.* INTO v_attempt
    FROM analytics_catalog_projection_attempts AS attempt
    JOIN analytics_catalog_version_projections AS projection
      ON projection.receipt_id=attempt.projection_receipt_id
    WHERE attempt.receipt_id=p_attempt_receipt_id
    FOR UPDATE OF attempt,projection;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog materialization attempt is unavailable';
    END IF;
    SELECT * INTO v_projection
    FROM analytics_catalog_version_projections AS projection
    WHERE projection.receipt_id=v_attempt.projection_receipt_id;
    IF NOT FOUND OR p_occurred_at<v_attempt.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog materialization attempt is unavailable';
    END IF;
    SELECT * INTO v_component
    FROM analytics_catalog_components AS component
    WHERE component.component_id=v_projection.component_id FOR UPDATE;
    IF NOT FOUND OR
       v_component.owner_generation<>v_projection.owner_generation OR
       v_component.contract<>v_projection.contract OR
       v_component.writer_state<>'v2' OR
       v_component.lifecycle_state NOT IN
           ('AVAILABLE','STAGED','ACTIVE','DISABLED') THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog materialization authority differs';
    END IF;
    IF v_projection.operational_state='ready' THEN
        v_state:='ready';
        v_resource_id:=v_projection.operational_resource_id;
        v_published_id:=v_projection.published_version_id;
        v_draft_id:=v_projection.current_draft_version_id;
    ELSE
        v_original:=pg_catalog.convert_from(
            v_projection.document_bytes,'UTF8'
        )::jsonb;
        v_plan:=pg_catalog.convert_from(
            v_projection.runtime_plan_bytes,'UTF8'
        )::jsonb;
        v_version_seed:=analytics_catalog_projection_seed_v2(
            v_projection.component_id,v_projection.package_id,
            v_projection.package_version,v_projection.release_id,
            v_projection.content_sha256
        );
        IF v_projection.analytics_kind='endpoint' THEN
            SELECT result.operational_state,result.blocker_code,
                result.operational_resource_id,result.published_version_id,
                result.current_draft_version_id
            INTO v_state,v_blocker,v_resource_id,v_published_id,v_draft_id
            FROM analytics_catalog_materialize_endpoint_v1(
                p_scope_schema,v_version_seed,v_attempt.actor_id,
                v_original,v_plan,p_occurred_at
            ) AS result;
        ELSIF v_projection.analytics_kind='dashboard' THEN
            SELECT result.operational_state,result.blocker_code,
                result.operational_resource_id,result.published_version_id,
                result.current_draft_version_id
            INTO v_state,v_blocker,v_resource_id,v_published_id,v_draft_id
            FROM analytics_catalog_materialize_dashboard_v2(
                p_scope_schema,v_projection.component_id,v_version_seed,
                v_projection.package_id,v_projection.package_version,
                v_projection.release_id,v_attempt.actor_id,
                v_original,v_plan,p_occurred_at
            ) AS result;
        ELSE
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog materialization kind differs';
        END IF;
    END IF;
    IF v_state='ready' THEN
        IF pg_catalog.uuid_extract_version(v_resource_id) IS DISTINCT FROM 7 OR
           pg_catalog.uuid_extract_version(v_published_id) IS DISTINCT FROM 7 OR
           pg_catalog.uuid_extract_version(v_draft_id) IS DISTINCT FROM 7 OR
           v_blocker IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog ready result is invalid';
        END IF;
        UPDATE analytics_catalog_version_projections AS projection
        SET operational_state='ready',blocker_code=NULL,
            operational_resource_id=v_resource_id,
            published_version_id=v_published_id,
            current_draft_version_id=v_draft_id,
            ready_at=COALESCE(projection.ready_at,p_occurred_at),
            updated_at=CASE WHEN projection.ready_at IS NULL
                THEN p_occurred_at ELSE projection.updated_at END
        WHERE projection.receipt_id=v_projection.receipt_id;
        UPDATE analytics_catalog_projection_attempts AS attempt
        SET ready_at=COALESCE(attempt.ready_at,p_occurred_at),
            updated_at=CASE WHEN attempt.ready_at IS NULL
                THEN p_occurred_at ELSE attempt.updated_at END
        WHERE attempt.receipt_id=p_attempt_receipt_id;
        SELECT EXISTS(
            SELECT 1 FROM analytics_catalog_projection_activations AS activation
            WHERE activation.projection_receipt_id=v_projection.receipt_id
        ) INTO v_was_activated;
        IF NOT v_was_activated THEN
            SELECT candidate.projection_receipt_id
            INTO v_candidate_projection_id
            FROM analytics_catalog_component_candidates AS candidate
            WHERE candidate.component_id=v_projection.component_id
            FOR UPDATE;
            IF FOUND AND
               v_candidate_projection_id<>v_projection.receipt_id THEN
                RAISE EXCEPTION USING ERRCODE='55000',
                    MESSAGE='Analytics Catalog component candidate is unresolved';
            END IF;
            PERFORM analytics_catalog_operational_stage_v2(
                p_scope_schema,p_attempt_receipt_id,p_occurred_at
            );
            INSERT INTO analytics_catalog_component_candidates(
                component_id,projection_receipt_id,attempt_receipt_id,
                package_id,package_version,release_id,owner_generation,
                content_sha256,staged_operation_id,prepared_at
            ) VALUES (
                v_projection.component_id,v_projection.receipt_id,
                p_attempt_receipt_id,v_projection.package_id,
                v_projection.package_version,v_projection.release_id,
                v_projection.owner_generation,v_projection.content_sha256,
                NULL,p_occurred_at
            ) ON CONFLICT ON CONSTRAINT
                analytics_catalog_component_candidates_pkey DO UPDATE SET
                projection_receipt_id=EXCLUDED.projection_receipt_id,
                attempt_receipt_id=EXCLUDED.attempt_receipt_id,
                package_id=EXCLUDED.package_id,
                package_version=EXCLUDED.package_version,
                release_id=EXCLUDED.release_id,
                owner_generation=EXCLUDED.owner_generation,
                content_sha256=EXCLUDED.content_sha256,
                staged_operation_id=CASE
                  WHEN analytics_catalog_component_candidates.attempt_receipt_id=
                       EXCLUDED.attempt_receipt_id
                    THEN analytics_catalog_component_candidates.staged_operation_id
                  ELSE NULL END,
                prepared_at=EXCLUDED.prepared_at
              WHERE analytics_catalog_component_candidates.
                    projection_receipt_id=EXCLUDED.projection_receipt_id;
            IF NOT FOUND THEN
                RAISE EXCEPTION USING ERRCODE='55000',
                    MESSAGE='Analytics Catalog candidate refresh was not applied';
            END IF;
        END IF;
    ELSIF v_state='pending' AND
          v_blocker ~ '^[A-Z][A-Z0-9_]{0,127}$' THEN
        UPDATE analytics_catalog_version_projections
        SET blocker_code=v_blocker,updated_at=p_occurred_at
        WHERE receipt_id=v_projection.receipt_id;
        UPDATE analytics_catalog_projection_attempts
        SET updated_at=p_occurred_at
        WHERE receipt_id=p_attempt_receipt_id;
    ELSE
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog materialization result is invalid';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_projection_materialize_v2(
    name,uuid,timestamptz
) FROM PUBLIC;
