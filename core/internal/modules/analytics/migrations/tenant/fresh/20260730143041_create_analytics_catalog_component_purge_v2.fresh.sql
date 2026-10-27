-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_component_purge_v2(
    p_receipt_id uuid,p_operation_id uuid,p_component_id uuid,p_owner text,
    p_generation bigint,p_contract text,p_content_sha256 text,
    p_action text,p_reference_epoch bigint,p_idempotency_key text,
    p_completed_at timestamptz
)
RETURNS TABLE(
    receipt_id uuid,operation_id uuid,component_id uuid,owner_module text,
    owner_generation bigint,contract text,content_sha256 text,action text,ready boolean,
    completed_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component analytics_catalog_components%ROWTYPE;
    v_receipt analytics_catalog_component_receipts%ROWTYPE;
    v_blocker analytics_catalog_blocker_receipts%ROWTYPE;
    v_scope_schema text;
    v_snapshot_blockers jsonb;
    v_snapshot_zero_digest text;
    v_analytics_kind text;
    v_root_id uuid;
    v_head_version_id uuid;
    v_root_reference_epoch bigint;
    v_receipt_count bigint;
    v_changed bigint;
BEGIN
    IF pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_owner IS DISTINCT FROM 'analytics' OR p_generation IS NULL OR p_generation<1 OR
       p_contract IS NULL OR
       p_content_sha256 IS NULL OR
       p_content_sha256!~'^[0-9a-f]{64}$' OR
       p_action IS DISTINCT FROM 'purge' OR p_reference_epoch IS NULL OR p_reference_epoch<0 OR
       p_idempotency_key IS NULL OR
       length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog purge command is invalid';
    END IF;
    SELECT schema_name INTO STRICT v_scope_schema FROM tenancy_schema_identity WHERE singleton;
    PERFORM analytics_assert_current_authority_v1(v_scope_schema::name);
    SELECT * INTO v_receipt
    FROM analytics_catalog_component_receipts AS receipt
    WHERE receipt.operation_id=p_operation_id
      AND receipt.component_id=p_component_id
      AND receipt.action='purge'
    FOR SHARE;
    IF FOUND THEN
        IF v_receipt.owner_module IS DISTINCT FROM p_owner OR
           v_receipt.owner_generation IS DISTINCT FROM p_generation OR
           v_receipt.contract IS DISTINCT FROM p_contract OR
           v_receipt.content_sha256 IS DISTINCT FROM p_content_sha256 OR
           v_receipt.idempotency_key IS DISTINCT FROM p_idempotency_key OR
           NOT v_receipt.ready OR NOT EXISTS (
               SELECT 1 FROM analytics_catalog_components AS component
               WHERE component.component_id=p_component_id
                 AND component.lifecycle_state='PURGED'
           ) THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='Analytics Catalog purge receipt differs';
        END IF;
        RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
            v_receipt.component_id,v_receipt.owner_module,
            v_receipt.owner_generation,v_receipt.contract,
            v_receipt.content_sha256,v_receipt.action,
            v_receipt.ready,v_receipt.completed_at;
        RETURN;
    END IF;
    SELECT * INTO v_component
    FROM analytics_catalog_components AS component
    WHERE component.component_id=p_component_id FOR UPDATE;
    IF NOT FOUND OR
       v_component.component_kind NOT IN ('ENDPOINT','DASHBOARD') OR
       v_component.owner_generation<>p_generation OR
       v_component.writer_state<>'v2' OR
       v_component.lifecycle_state<>'RETIRED' OR
       v_component.contract<>p_contract OR
       v_component.content_sha256<>p_content_sha256 OR
       p_completed_at<v_component.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge authority differs';
    END IF;
    SELECT pg_catalog.count(*) INTO v_receipt_count
    FROM analytics_catalog_blocker_receipts AS blocker
    WHERE blocker.operation_id=p_operation_id
      AND blocker.component_id=p_component_id;
    IF v_receipt_count<>1 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge blocker receipt is ambiguous';
    END IF;
    SELECT * INTO v_blocker
    FROM analytics_catalog_blocker_receipts AS blocker
    WHERE blocker.operation_id=p_operation_id
      AND blocker.component_id=p_component_id
    FOR SHARE;
    IF v_blocker.catalog_version<>'analytics-catalog-blockers.v2' OR
       v_blocker.blockers<>'[]'::jsonb OR
       v_blocker.completed_at>p_completed_at OR
       (p_reference_epoch<>0 AND
        p_reference_epoch<>v_blocker.reference_epoch) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge blocker receipt differs';
    END IF;
    SELECT snapshot.blockers,snapshot.zero_digest,
        snapshot.analytics_kind,snapshot.root_id,snapshot.head_version_id,
        snapshot.root_reference_epoch
    INTO STRICT v_snapshot_blockers,v_snapshot_zero_digest,
        v_analytics_kind,v_root_id,v_head_version_id,
        v_root_reference_epoch
    FROM analytics_catalog_blocker_snapshot_v2(
        p_component_id,v_blocker.reference_epoch,p_generation
    ) AS snapshot;
    IF v_snapshot_blockers<>'[]'::jsonb OR
       v_blocker.blockers IS DISTINCT FROM v_snapshot_blockers OR
       v_blocker.zero_digest IS DISTINCT FROM v_snapshot_zero_digest THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge blockers changed';
    END IF;
    IF v_analytics_kind='endpoint' THEN
        PERFORM 1 FROM analytics_saved_queries AS root
        WHERE root.id=v_root_id AND root.lifecycle_state='retired'
          AND root.reference_epoch=v_root_reference_epoch
        FOR UPDATE;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog purge query changed';
        END IF;
        PERFORM 1 FROM analytics_saved_query_versions AS version_row
        WHERE version_row.query_id=v_root_id
        ORDER BY version_row.id FOR UPDATE;
    ELSE
        PERFORM 1 FROM analytics_dashboards AS root
        WHERE root.id=v_root_id AND root.lifecycle_state='retired'
          AND root.reference_epoch=v_root_reference_epoch
        FOR UPDATE;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog purge dashboard changed';
        END IF;
        PERFORM 1 FROM analytics_dashboard_versions AS version_row
        WHERE version_row.dashboard_id=v_root_id
        ORDER BY version_row.id FOR UPDATE;
    END IF;
    UPDATE analytics_catalog_projection_activations AS activation
    SET superseded_at=p_completed_at
    FROM analytics_catalog_component_heads AS head
    WHERE head.component_id=p_component_id
      AND head.projection_receipt_id=activation.projection_receipt_id
      AND activation.component_id=head.component_id
      AND activation.superseded_at IS NULL;
    GET DIAGNOSTICS v_changed=ROW_COUNT;
    IF v_changed<>1 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge activation differs';
    END IF;
    DELETE FROM analytics_catalog_component_heads AS head
    WHERE head.component_id=p_component_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge head differs';
    END IF;
    INSERT INTO analytics_catalog_component_receipts(
        receipt_id,operation_id,component_id,owner_module,owner_generation,
        contract,content_sha256,action,idempotency_key,ready,completed_at
    ) VALUES (
        p_receipt_id,p_operation_id,p_component_id,p_owner,p_generation,
        p_contract,p_content_sha256,'purge',p_idempotency_key,true,
        p_completed_at
    ) RETURNING * INTO v_receipt;
    PERFORM analytics_catalog_purge_data_v2(
        v_analytics_kind,v_root_id,p_operation_id,p_completed_at
    );
    UPDATE analytics_catalog_components AS component
    SET lifecycle_state='PURGED',updated_at=p_completed_at
    WHERE component.component_id=p_component_id
      AND component.owner_generation=p_generation
      AND component.writer_state='v2'
      AND component.lifecycle_state='RETIRED'
      AND component.contract=p_contract
      AND component.content_sha256=p_content_sha256;
    GET DIAGNOSTICS v_changed=ROW_COUNT;
    IF v_changed<>1 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge final CAS differs';
    END IF;
    RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
        v_receipt.component_id,v_receipt.owner_module,
        v_receipt.owner_generation,v_receipt.contract,
        v_receipt.content_sha256,v_receipt.action,
        v_receipt.ready,v_receipt.completed_at;
END;
$function$;
REVOKE ALL ON FUNCTION analytics_catalog_component_purge_v2(uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION analytics_catalog_component_purge_v2(uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz) FROM forgecrud_runtime;
-- +goose StatementEnd
