-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_component_apply_core_v2(
    p_receipt_id uuid,p_operation_id uuid,p_component_id uuid,
    p_owner text,p_generation bigint,p_contract text,p_content_sha256 text,
    p_action text,p_reference_epoch bigint,p_idempotency_key text,p_completed_at timestamptz
)
RETURNS TABLE(receipt_id uuid,operation_id uuid,component_id uuid,owner_module text,
    owner_generation bigint,contract text,content_sha256 text,action text,
    ready boolean,completed_at timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component analytics_catalog_components%ROWTYPE;
    v_receipt analytics_catalog_component_receipts%ROWTYPE;
    v_attempt analytics_catalog_projection_attempts%ROWTYPE;
    v_scope_schema text;
    v_has_head boolean;
    v_next text;
BEGIN
    SELECT schema_name INTO STRICT v_scope_schema FROM tenancy_schema_identity WHERE singleton;
    PERFORM analytics_assert_current_authority_v1(v_scope_schema::name);
    IF pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_owner IS DISTINCT FROM 'analytics' OR
       p_generation IS NULL OR p_generation<1 OR p_contract IS NULL OR
       p_content_sha256 IS NULL OR p_content_sha256!~'^[0-9a-f]{64}$' OR
       p_action IS NULL OR p_action NOT IN ('stage','cleanup','activate','disable','enable',
           'retire','restore') OR
       p_idempotency_key IS NULL OR
       length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog v2 component command is invalid';
    END IF;
    SELECT * INTO v_receipt FROM analytics_catalog_component_receipts AS receipt
    WHERE receipt.operation_id=p_operation_id AND receipt.component_id=p_component_id AND receipt.action=p_action;
    IF FOUND THEN
        IF v_receipt.owner_module IS DISTINCT FROM p_owner OR
           v_receipt.owner_generation IS DISTINCT FROM p_generation OR
           v_receipt.contract IS DISTINCT FROM p_contract OR
           v_receipt.content_sha256 IS DISTINCT FROM p_content_sha256 OR
           v_receipt.idempotency_key IS DISTINCT FROM p_idempotency_key THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='Analytics Catalog v2 component receipt differs';
        END IF;
        RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
            v_receipt.component_id,v_receipt.owner_module,v_receipt.owner_generation,
            v_receipt.contract,v_receipt.content_sha256,v_receipt.action,
            v_receipt.ready,v_receipt.completed_at;
        RETURN;
    END IF;
    SELECT * INTO v_component FROM analytics_catalog_components AS component
    WHERE component.component_id=p_component_id FOR UPDATE;
    IF NOT FOUND OR
       v_component.component_kind NOT IN ('ENDPOINT','DASHBOARD') OR
       v_component.owner_generation<>p_generation OR
       v_component.writer_state<>'v2' OR
       v_component.contract<>p_contract OR
       p_completed_at<v_component.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog v2 component authority differs';
    END IF;
    IF p_action IN ('disable','enable','retire','restore','purge') THEN
        PERFORM analytics_catalog_component_lifecycle_v2(
            v_scope_schema::name,p_operation_id,p_component_id,p_generation,p_contract,
            p_content_sha256,p_action,p_reference_epoch,p_completed_at);
    ELSE
        SELECT EXISTS(
            SELECT 1 FROM analytics_catalog_component_heads AS head
            WHERE head.component_id=p_component_id
        ) INTO v_has_head;
        IF p_action='cleanup' THEN
            IF NOT EXISTS (
                SELECT 1 FROM analytics_catalog_component_receipts AS stage
                WHERE stage.operation_id=p_operation_id
                  AND stage.component_id=p_component_id
                  AND stage.action='stage'
                  AND stage.content_sha256=p_content_sha256
            ) THEN
                RAISE EXCEPTION USING ERRCODE='55000',
                    MESSAGE='Analytics Catalog cleanup lacks staged evidence';
            END IF;
            v_next:=CASE
              WHEN NOT v_has_head AND v_component.lifecycle_state='STAGED'
                AND v_component.content_sha256=p_content_sha256
                THEN 'AVAILABLE'
              WHEN v_has_head AND
                   v_component.lifecycle_state IN ('ACTIVE','DISABLED')
                THEN v_component.lifecycle_state
              ELSE NULL END;
            IF v_next IS NOT NULL THEN
                PERFORM analytics_catalog_candidate_cleanup_v2(
                    v_scope_schema::name,p_component_id,p_operation_id,
                    p_content_sha256,p_completed_at
                );
            END IF;
        ELSE
            SELECT attempt.* INTO v_attempt
            FROM analytics_catalog_component_candidates AS candidate
            JOIN analytics_catalog_version_projections AS projection
              ON projection.receipt_id=candidate.projection_receipt_id
             AND projection.component_id=candidate.component_id
             AND projection.package_id=candidate.package_id
             AND projection.package_version=candidate.package_version
             AND projection.release_id=candidate.release_id
             AND projection.owner_generation=candidate.owner_generation
             AND projection.content_sha256=candidate.content_sha256
            JOIN analytics_catalog_projection_attempts AS attempt
              ON attempt.receipt_id=candidate.attempt_receipt_id
             AND attempt.projection_receipt_id=projection.receipt_id
            WHERE candidate.component_id=p_component_id
              AND candidate.owner_generation=p_generation
              AND candidate.content_sha256=p_content_sha256
              AND projection.contract=p_contract
              AND projection.operational_state='ready'
              AND attempt.ready_at IS NOT NULL
              AND attempt.idempotency_key=
                  p_operation_id::text||':'||p_component_id::text||':provision'
              AND ((p_action='stage' AND (
                    candidate.staged_operation_id IS NULL OR
                    candidate.staged_operation_id=p_operation_id
                  )) OR (p_action='activate' AND
                    candidate.staged_operation_id=p_operation_id))
            FOR UPDATE OF candidate,projection,attempt;
            IF NOT FOUND OR p_completed_at<v_attempt.updated_at THEN
                RAISE EXCEPTION USING ERRCODE='55000',
                    MESSAGE='Analytics Catalog exact candidate is unavailable';
            END IF;
            IF p_action='stage' THEN
                UPDATE analytics_catalog_component_candidates AS candidate
                SET staged_operation_id=p_operation_id,
                    prepared_at=p_completed_at
                WHERE candidate.component_id=p_component_id
                  AND candidate.attempt_receipt_id=v_attempt.receipt_id;
                v_next:=CASE
                  WHEN NOT v_has_head AND
                       v_component.lifecycle_state='AVAILABLE' AND
                       v_component.content_sha256=p_content_sha256
                    THEN 'STAGED'
                  WHEN v_has_head AND
                       v_component.lifecycle_state IN ('ACTIVE','DISABLED')
                    THEN v_component.lifecycle_state
                  ELSE NULL END;
            ELSIF NOT EXISTS (
                SELECT 1 FROM analytics_catalog_component_receipts AS stage
                WHERE stage.operation_id=p_operation_id
                  AND stage.component_id=p_component_id
                  AND stage.action='stage'
                  AND stage.content_sha256=p_content_sha256
            ) THEN
                RAISE EXCEPTION USING ERRCODE='55000',
                    MESSAGE='Analytics Catalog activation lacks staged evidence';
            ELSE
                v_next:=CASE
                  WHEN NOT v_has_head AND v_component.lifecycle_state='STAGED'
                    THEN 'ACTIVE'
                  WHEN v_has_head AND
                       v_component.lifecycle_state IN ('ACTIVE','DISABLED')
                    THEN v_component.lifecycle_state
                  ELSE NULL END;
                IF v_next IS NOT NULL THEN
                    PERFORM analytics_catalog_projection_activate_v2(
                        v_scope_schema::name,v_attempt.projection_receipt_id,
                        p_operation_id,
                        v_next='ACTIVE',p_completed_at
                    );
                END IF;
            END IF;
        END IF;
        IF v_next IS NULL THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog v2 component transition is invalid';
        END IF;
        UPDATE analytics_catalog_components AS component
        SET lifecycle_state=v_next,
            content_sha256=CASE WHEN p_action='activate'
                THEN p_content_sha256 ELSE component.content_sha256 END,
            updated_at=CASE WHEN component.lifecycle_state IS DISTINCT FROM v_next
                OR p_action='activate' THEN p_completed_at
                ELSE component.updated_at END
        WHERE component.component_id=p_component_id;
    END IF;
    INSERT INTO analytics_catalog_component_receipts(receipt_id,operation_id,
        component_id,owner_module,owner_generation,contract,content_sha256,
        action,idempotency_key,ready,completed_at) VALUES (p_receipt_id,
        p_operation_id,p_component_id,p_owner,p_generation,p_contract,
        p_content_sha256,p_action,p_idempotency_key,true,p_completed_at)
    RETURNING * INTO v_receipt;
    RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
        v_receipt.component_id,v_receipt.owner_module,v_receipt.owner_generation,
        v_receipt.contract,v_receipt.content_sha256,v_receipt.action,
        v_receipt.ready,v_receipt.completed_at;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_component_apply_core_v2(uuid,uuid,uuid,
    text,bigint,text,text,text,bigint,text,timestamptz) FROM PUBLIC;
