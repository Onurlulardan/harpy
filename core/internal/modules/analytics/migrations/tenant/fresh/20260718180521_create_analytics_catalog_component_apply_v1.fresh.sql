-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_component_apply_v1(
    p_receipt_id uuid, p_operation_id uuid, p_component_id uuid,
    p_owner text, p_generation bigint, p_contract text, p_content_sha256 text,
    p_action text, p_reference_epoch bigint, p_idempotency_key text,
    p_completed_at timestamptz
) RETURNS TABLE (
    receipt_id uuid, operation_id uuid, component_id uuid, owner_module text,
    owner_generation bigint, contract text, content_sha256 text,
    action text, ready boolean, completed_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component analytics_catalog_components%ROWTYPE;
    v_receipt analytics_catalog_component_receipts%ROWTYPE;
    v_next text;
BEGIN
    SELECT * INTO v_receipt FROM analytics_catalog_component_receipts AS receipt
    WHERE receipt.operation_id=p_operation_id AND receipt.component_id=p_component_id
      AND receipt.action=p_action;
    IF FOUND THEN
        IF v_receipt.owner_generation<>p_generation OR v_receipt.contract<>p_contract
           OR v_receipt.content_sha256<>p_content_sha256
           OR v_receipt.idempotency_key<>p_idempotency_key THEN
            RAISE EXCEPTION USING ERRCODE='23505',MESSAGE='catalog component receipt differs';
        END IF;
        RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
            v_receipt.component_id,v_receipt.owner_module,v_receipt.owner_generation,
            v_receipt.contract,v_receipt.content_sha256,v_receipt.action,
            v_receipt.ready,v_receipt.completed_at;
        RETURN;
    END IF;
    IF pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_owner IS DISTINCT FROM 'analytics' OR p_generation<1 OR
       p_contract IS NULL OR p_content_sha256!~'^[0-9a-f]{64}$' OR
       p_action NOT IN ('stage','cleanup','activate','disable','enable','retire','restore','purge') OR
       length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_completed_at IS NULL OR date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='catalog component command invalid';
    END IF;
    SELECT * INTO v_component FROM analytics_catalog_components AS component
    WHERE component.component_id=p_component_id FOR UPDATE;
    IF NOT FOUND OR v_component.owner_generation<>p_generation OR
       v_component.writer_state<>'v2' OR v_component.contract<>p_contract OR
       v_component.content_sha256<>p_content_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='55000',MESSAGE='catalog component authority differs';
    END IF;
    v_next := CASE
      WHEN p_action='stage' AND v_component.lifecycle_state='AVAILABLE' THEN 'STAGED'
      WHEN p_action='cleanup' AND v_component.lifecycle_state='STAGED' THEN 'AVAILABLE'
      WHEN p_action='activate' AND v_component.lifecycle_state='STAGED' THEN 'ACTIVE'
      WHEN p_action='disable' AND v_component.lifecycle_state='ACTIVE' THEN 'DISABLED'
      WHEN p_action='enable' AND v_component.lifecycle_state='DISABLED' THEN 'ACTIVE'
      WHEN p_action='retire' AND v_component.lifecycle_state IN ('ACTIVE','DISABLED') THEN 'RETIRED'
      WHEN p_action='restore' AND v_component.lifecycle_state='RETIRED' THEN 'ACTIVE'
      WHEN p_action='purge' AND v_component.lifecycle_state='RETIRED' THEN 'PURGED'
      ELSE NULL END;
    IF v_next IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',MESSAGE='catalog component transition invalid';
    END IF;
    IF p_action='purge' AND (p_reference_epoch<1 OR NOT EXISTS (
        SELECT 1 FROM analytics_catalog_blocker_receipts AS blocker
        WHERE blocker.operation_id=p_operation_id AND blocker.component_id=p_component_id
          AND blocker.reference_epoch=p_reference_epoch AND blocker.blockers='[]'::jsonb
    )) THEN
        RAISE EXCEPTION USING ERRCODE='55000',MESSAGE='catalog purge receipt unavailable';
    END IF;
    UPDATE analytics_catalog_components SET lifecycle_state=v_next,updated_at=p_completed_at
    WHERE analytics_catalog_components.component_id=p_component_id;
    INSERT INTO analytics_catalog_component_receipts (
        receipt_id,operation_id,component_id,owner_module,owner_generation,
        contract,content_sha256,action,idempotency_key,ready,completed_at
    ) VALUES (p_receipt_id,p_operation_id,p_component_id,p_owner,p_generation,
        p_contract,p_content_sha256,p_action,p_idempotency_key,true,p_completed_at)
    RETURNING * INTO v_receipt;
    RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
        v_receipt.component_id,v_receipt.owner_module,v_receipt.owner_generation,
        v_receipt.contract,v_receipt.content_sha256,v_receipt.action,
        v_receipt.ready,v_receipt.completed_at;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_component_apply_v1(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_catalog_component_apply_v1(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) TO forgecrud_runtime;

