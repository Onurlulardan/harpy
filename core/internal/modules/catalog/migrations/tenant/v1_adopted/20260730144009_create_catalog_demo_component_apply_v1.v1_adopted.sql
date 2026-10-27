-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_demo_component_apply_v1(
    p_receipt_id uuid, p_operation_id uuid, p_component_id uuid,
    p_owner text, p_generation bigint, p_contract text, p_content_sha256 text,
    p_action text, p_reference_epoch bigint, p_idempotency_key text,
    p_completed_at timestamptz
) RETURNS TABLE (
    receipt_id uuid, operation_id uuid, component_id uuid, owner_module text,
    owner_generation bigint, contract text, content_sha256 text, action text,
    ready boolean, completed_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component catalog_catalog_components%ROWTYPE;
    v_receipt catalog_catalog_component_receipts%ROWTYPE;
    v_provision catalog_demo_provisions%ROWTYPE;
    v_handoff catalog_demo_handoffs%ROWTYPE;
    v_has_head boolean;
    v_next text;
BEGIN
    SELECT * INTO v_receipt
    FROM catalog_catalog_component_receipts AS receipt
    WHERE receipt.operation_id=p_operation_id
      AND receipt.component_id=p_component_id AND receipt.action=p_action;
    IF FOUND THEN
        IF v_receipt.owner_generation<>p_generation OR v_receipt.contract<>p_contract OR
           v_receipt.content_sha256<>p_content_sha256 OR
           v_receipt.idempotency_key<>p_idempotency_key THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='catalog demo component receipt differs';
        END IF;
        RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
            v_receipt.component_id,v_receipt.owner_module,
            v_receipt.owner_generation,v_receipt.contract,
            v_receipt.content_sha256,v_receipt.action,
            v_receipt.ready,v_receipt.completed_at;
        RETURN;
    END IF;
    IF p_action NOT IN ('stage','cleanup','activate') THEN
        RETURN QUERY SELECT * FROM catalog_catalog_component_apply_v1(
            p_receipt_id,p_operation_id,p_component_id,p_owner,p_generation,
            p_contract,p_content_sha256,p_action,p_reference_epoch,
            p_idempotency_key,p_completed_at
        );
        RETURN;
    END IF;
    IF pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_owner IS DISTINCT FROM 'catalog' OR p_generation<1 OR
       p_contract IS NULL OR p_content_sha256!~'^[0-9a-f]{64}$' OR
       length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='catalog demo component command invalid';
    END IF;
    SELECT * INTO v_component
    FROM catalog_catalog_components AS component
    WHERE component.component_id=p_component_id FOR UPDATE;
    IF NOT FOUND OR v_component.component_kind<>'DEMO_DATA' OR
       v_component.owner_generation<>p_generation OR
       v_component.writer_state<>'v2' OR v_component.contract<>p_contract THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='catalog demo component authority differs';
    END IF;
    SELECT EXISTS(SELECT 1 FROM catalog_demo_component_heads AS head
        WHERE head.component_id=p_component_id) INTO v_has_head;
    IF p_action='cleanup' THEN
        IF NOT EXISTS (
            SELECT 1 FROM catalog_catalog_component_receipts AS stage
            WHERE stage.operation_id=p_operation_id
              AND stage.component_id=p_component_id
              AND stage.action='stage'
              AND stage.content_sha256=p_content_sha256
        ) THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='catalog demo cleanup lacks staged evidence';
        END IF;
        v_next:=CASE
          WHEN NOT v_has_head AND v_component.lifecycle_state='STAGED'
            AND v_component.content_sha256=p_content_sha256 THEN 'AVAILABLE'
          WHEN v_has_head AND v_component.lifecycle_state IN ('ACTIVE','DISABLED')
            THEN v_component.lifecycle_state
          ELSE NULL END;
    ELSE
        SELECT * INTO v_provision
        FROM catalog_demo_provisions AS provision
        WHERE provision.component_id=p_component_id
          AND provision.owner_generation=p_generation
          AND provision.contract=p_contract
          AND provision.content_sha256=p_content_sha256
          AND provision.ready AND provision.retry_at IS NULL
          AND (
            provision.idempotency_key=p_operation_id::text||':'||
                p_component_id::text||':provision' OR
            NOT v_has_head AND
                v_component.content_sha256=p_content_sha256
          )
        ORDER BY (
            provision.idempotency_key=p_operation_id::text||':'||
                p_component_id::text||':provision'
        ) DESC, provision.created_at DESC
        LIMIT 1 FOR SHARE;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='catalog demo exact ready provision is unavailable';
        END IF;
        SELECT * INTO STRICT v_handoff
        FROM catalog_demo_handoffs AS handoff
        WHERE handoff.provision_receipt_id=v_provision.receipt_id
          AND handoff.component_id=v_provision.component_id
          AND handoff.package_id=v_provision.package_id
          AND handoff.package_version=v_provision.package_version
          AND handoff.release_id=v_provision.release_id
          AND handoff.content_sha256=v_provision.content_sha256
          AND handoff.completed_at=v_provision.completed_at
        FOR SHARE;
        IF p_action='stage' THEN
            v_next:=CASE
              WHEN NOT v_has_head AND v_component.lifecycle_state='AVAILABLE'
                AND v_component.content_sha256=p_content_sha256 THEN 'STAGED'
              WHEN v_has_head AND
                v_component.lifecycle_state IN ('ACTIVE','DISABLED')
                THEN v_component.lifecycle_state
              ELSE NULL END;
        ELSIF NOT EXISTS (
            SELECT 1 FROM catalog_catalog_component_receipts AS stage
            WHERE stage.operation_id=p_operation_id
              AND stage.component_id=p_component_id
              AND stage.action='stage'
              AND stage.content_sha256=p_content_sha256
        ) THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='catalog demo activation lacks staged evidence';
        ELSE
            v_next:=CASE
              WHEN NOT v_has_head AND v_component.lifecycle_state='STAGED'
                THEN 'ACTIVE'
              WHEN v_has_head AND
                v_component.lifecycle_state IN ('ACTIVE','DISABLED')
                THEN v_component.lifecycle_state
              ELSE NULL END;
            INSERT INTO catalog_demo_component_heads(
                component_id,package_id,package_version,release_id,
                provision_receipt_id,handoff_id,forms_run_id,
                content_sha256,activated_at
            ) VALUES (
                p_component_id,v_provision.package_id,
                v_provision.package_version,v_provision.release_id,
                v_provision.receipt_id,v_handoff.handoff_id,
                v_handoff.forms_run_id,p_content_sha256,p_completed_at
            ) ON CONFLICT ON CONSTRAINT catalog_demo_component_heads_pkey DO UPDATE SET
                package_id=EXCLUDED.package_id,
                package_version=EXCLUDED.package_version,
                release_id=EXCLUDED.release_id,
                provision_receipt_id=EXCLUDED.provision_receipt_id,
                handoff_id=EXCLUDED.handoff_id,
                forms_run_id=EXCLUDED.forms_run_id,
                content_sha256=EXCLUDED.content_sha256,
                activated_at=EXCLUDED.activated_at;
        END IF;
    END IF;
    IF v_next IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='catalog demo component transition invalid';
    END IF;
    UPDATE catalog_catalog_components
    SET lifecycle_state=v_next,
        content_sha256=CASE WHEN p_action='activate' THEN p_content_sha256
            ELSE catalog_catalog_components.content_sha256 END,
        updated_at=CASE
          WHEN v_component.lifecycle_state IS DISTINCT FROM v_next
            OR p_action='activate' THEN p_completed_at
          ELSE v_component.updated_at END
    WHERE catalog_catalog_components.component_id=p_component_id;
    INSERT INTO catalog_catalog_component_receipts(
        receipt_id,operation_id,component_id,owner_module,owner_generation,
        contract,content_sha256,action,idempotency_key,ready,completed_at
    ) VALUES (
        p_receipt_id,p_operation_id,p_component_id,p_owner,p_generation,
        p_contract,p_content_sha256,p_action,p_idempotency_key,true,p_completed_at
    ) RETURNING * INTO v_receipt;
    RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
        v_receipt.component_id,v_receipt.owner_module,
        v_receipt.owner_generation,v_receipt.contract,
        v_receipt.content_sha256,v_receipt.action,
        v_receipt.ready,v_receipt.completed_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION catalog_demo_component_apply_v1(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION catalog_demo_component_apply_v1(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) TO harpy_runtime;
