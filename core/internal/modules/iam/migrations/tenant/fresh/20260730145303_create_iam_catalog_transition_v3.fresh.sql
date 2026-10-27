-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_component_transition_v3(
    p_receipt_id uuid,p_operation_id uuid,p_component_id uuid,p_owner text,
    p_generation bigint,p_contract text,p_content_sha256 text,
    p_action text,p_reference_epoch bigint,p_idempotency_key text,
    p_completed_at timestamptz
)
RETURNS TABLE(
    receipt_id uuid,operation_id uuid,component_id uuid,owner_module text,
    owner_generation bigint,contract text,content_sha256 text,action text,
    ready boolean,completed_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component iam_catalog_components%ROWTYPE;
    v_head iam_catalog_component_heads%ROWTYPE;
    v_receipt iam_catalog_component_receipts%ROWTYPE;
    v_authorization iam_authorization_operation_receipts%ROWTYPE;
    v_expected_state text;
BEGIN
    IF pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_receipt_id=p_operation_id OR
       p_owner IS DISTINCT FROM 'iam' OR p_generation<1 OR
       p_contract IS NULL OR p_content_sha256!~'^[0-9a-f]{64}$' OR
       p_action NOT IN ('disable','enable','retire','restore') OR
       p_reference_epoch<>0 OR
       length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM Catalog transition command is invalid';
    END IF;
    SELECT * INTO v_receipt
    FROM iam_catalog_component_receipts AS receipt
    WHERE receipt.operation_id=p_operation_id
      AND receipt.component_id=p_component_id
      AND receipt.action=p_action
    FOR SHARE;
    IF FOUND THEN
        SELECT * INTO v_authorization
        FROM iam_authorization_operation_receipts AS receipt
        WHERE receipt.operation_id=p_operation_id;
        IF v_receipt.owner_module<>p_owner OR
           v_receipt.owner_generation<>p_generation OR
           v_receipt.contract<>p_contract OR
           v_receipt.content_sha256<>p_content_sha256 OR
           v_receipt.idempotency_key<>p_idempotency_key OR
           NOT v_receipt.ready OR
           v_authorization.operation_id IS NULL OR
           v_authorization.command_key<>'iam.catalog.'||p_action OR
           v_authorization.request_sha256<>
             pg_catalog.decode(p_content_sha256,'hex') OR
           v_authorization.result_id<>p_component_id THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='IAM Catalog transition receipt differs';
        END IF;
        RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
            v_receipt.component_id,v_receipt.owner_module,
            v_receipt.owner_generation,v_receipt.contract,
            v_receipt.content_sha256,v_receipt.action,
            v_receipt.ready,v_receipt.completed_at;
        RETURN;
    END IF;
    SELECT * INTO v_component
    FROM iam_catalog_components AS component
    WHERE component.component_id=p_component_id FOR UPDATE;
    SELECT * INTO v_head
    FROM iam_catalog_component_heads AS head
    WHERE head.component_id=p_component_id FOR SHARE;
    v_expected_state:=CASE p_action
      WHEN 'disable' THEN 'ACTIVE'
      WHEN 'enable' THEN 'DISABLED'
      WHEN 'restore' THEN 'RETIRED'
      WHEN 'retire' THEN NULL
      ELSE NULL END;
    IF v_component.component_id IS NULL OR v_head.component_id IS NULL OR
       v_component.component_kind<>'IAM_TEMPLATE' OR
       v_component.owner_generation<>p_generation OR
       v_component.writer_state<>'v2' OR
       v_component.contract<>p_contract OR
       v_component.content_sha256<>p_content_sha256 OR
       v_head.owner_generation<>p_generation OR
       v_head.content_sha256<>p_content_sha256 OR
       p_completed_at<v_component.updated_at OR
       (p_action='retire' AND
        v_component.lifecycle_state NOT IN ('ACTIVE','DISABLED')) OR
       (p_action<>'retire' AND
        v_component.lifecycle_state<>v_expected_state) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM Catalog transition authority differs';
    END IF;
    PERFORM iam_catalog_grants_set_v2(
        p_component_id,p_action IN ('enable','restore'),p_completed_at
    );
    UPDATE iam_catalog_components AS component
    SET lifecycle_state=CASE p_action
          WHEN 'disable' THEN 'DISABLED'
          WHEN 'retire' THEN 'RETIRED'
          ELSE 'ACTIVE' END,
        updated_at=p_completed_at
    WHERE component.component_id=p_component_id;
    INSERT INTO iam_catalog_component_receipts(
        receipt_id,operation_id,component_id,owner_module,owner_generation,
        contract,content_sha256,action,idempotency_key,ready,completed_at
    ) VALUES (
        p_receipt_id,p_operation_id,p_component_id,p_owner,p_generation,
        p_contract,p_content_sha256,p_action,p_idempotency_key,true,
        p_completed_at
    ) RETURNING * INTO v_receipt;
    PERFORM iam_catalog_authorization_record_v2(
        p_operation_id,p_component_id,p_content_sha256,p_action,p_completed_at
    );
    RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
        v_receipt.component_id,v_receipt.owner_module,
        v_receipt.owner_generation,v_receipt.contract,
        v_receipt.content_sha256,v_receipt.action,
        v_receipt.ready,v_receipt.completed_at;
END;
$function$;
REVOKE ALL ON FUNCTION iam_catalog_component_transition_v3(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION iam_catalog_component_transition_v3(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) FROM harpy_runtime;
-- +goose StatementEnd
