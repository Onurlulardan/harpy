-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_component_purge_v3(
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
    v_blocker iam_catalog_blocker_receipts%ROWTYPE;
    v_authorization iam_authorization_operation_receipts%ROWTYPE;
    v_blockers jsonb;
    v_digest text;
    v_legacy_digest text;
    v_count bigint;
    v_changed bigint;
BEGIN
    IF pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_receipt_id=p_operation_id OR p_owner IS DISTINCT FROM 'iam' OR
       p_generation<1 OR p_contract IS NULL OR
       p_content_sha256!~'^[0-9a-f]{64}$' OR
       p_action IS DISTINCT FROM 'purge' OR p_reference_epoch<1 OR
       length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM Catalog purge command is invalid';
    END IF;
    SELECT * INTO v_receipt
    FROM iam_catalog_component_receipts AS receipt
    WHERE receipt.operation_id=p_operation_id
      AND receipt.component_id=p_component_id
      AND receipt.action='purge'
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
           v_authorization.command_key<>'iam.catalog.purge' OR
           v_authorization.request_sha256<>
             pg_catalog.decode(p_content_sha256,'hex') OR
           v_authorization.result_id<>p_component_id OR NOT EXISTS (
               SELECT 1 FROM iam_catalog_components AS component
               WHERE component.component_id=p_component_id
                 AND component.lifecycle_state='PURGED'
           ) THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='IAM Catalog purge receipt differs';
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
    WHERE head.component_id=p_component_id FOR UPDATE;
    IF v_component.component_id IS NULL OR v_head.component_id IS NULL OR
       v_component.component_kind<>'IAM_TEMPLATE' OR
       v_component.owner_generation<>p_generation OR
       v_component.writer_state<>'v2' OR
       v_component.lifecycle_state<>'RETIRED' OR
       v_component.contract<>p_contract OR
       v_component.content_sha256<>p_content_sha256 OR
       v_head.owner_generation<>p_generation OR
       v_head.content_sha256<>p_content_sha256 OR
       p_completed_at<v_component.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM Catalog purge authority differs';
    END IF;
    SELECT pg_catalog.count(*) INTO v_count
    FROM iam_catalog_blocker_receipts AS blocker
    WHERE blocker.operation_id=p_operation_id
      AND blocker.component_id=p_component_id;
    IF v_count<>1 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM Catalog purge blocker receipt is ambiguous';
    END IF;
    SELECT * INTO STRICT v_blocker
    FROM iam_catalog_blocker_receipts AS blocker
    WHERE blocker.operation_id=p_operation_id
      AND blocker.component_id=p_component_id
    FOR SHARE;
    IF v_blocker.reference_epoch<>p_reference_epoch OR
       v_blocker.blockers<>'[]'::jsonb OR
       v_blocker.completed_at>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM Catalog purge blocker receipt differs';
    END IF;
    SELECT snapshot.blockers,snapshot.zero_digest
    INTO STRICT v_blockers,v_digest
    FROM iam_catalog_blocker_snapshot_v2(
        p_component_id,p_reference_epoch,p_generation
    ) AS snapshot;
    IF v_blockers<>'[]'::jsonb THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM Catalog purge blockers changed';
    END IF;
    IF v_blocker.catalog_version='iam-catalog-blockers.v1' THEN
        v_legacy_digest:=pg_catalog.encode(pg_catalog.sha256(
            pg_catalog.convert_to(
                'iam-catalog-blockers.v1'||E'\\000'||
                p_component_id::text||E'\\000'||
                p_reference_epoch::text||E'\\000'||'[]','UTF8'
            )
        ),'hex');
        IF v_blocker.zero_digest<>v_legacy_digest THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='IAM Catalog legacy purge proof differs';
        END IF;
    ELSIF v_blocker.catalog_version<>'iam-catalog-blockers.v2' OR
          v_blocker.zero_digest<>v_digest THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM Catalog purge proof differs';
    END IF;
    PERFORM iam_catalog_purge_data_v2(
        p_operation_id,p_component_id,p_generation,p_content_sha256,
        p_reference_epoch,p_completed_at
    );
    INSERT INTO iam_catalog_component_receipts(
        receipt_id,operation_id,component_id,owner_module,owner_generation,
        contract,content_sha256,action,idempotency_key,ready,completed_at
    ) VALUES (
        p_receipt_id,p_operation_id,p_component_id,p_owner,p_generation,
        p_contract,p_content_sha256,'purge',p_idempotency_key,true,
        p_completed_at
    ) RETURNING * INTO v_receipt;
    UPDATE iam_catalog_components AS component
    SET lifecycle_state='PURGED',updated_at=p_completed_at
    WHERE component.component_id=p_component_id
      AND component.lifecycle_state='RETIRED'
      AND component.owner_generation=p_generation
      AND component.content_sha256=p_content_sha256;
    GET DIAGNOSTICS v_changed=ROW_COUNT;
    IF v_changed<>1 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM Catalog purge final CAS differs';
    END IF;
    PERFORM iam_catalog_authorization_record_v2(
        p_operation_id,p_component_id,p_content_sha256,'purge',p_completed_at
    );
    RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
        v_receipt.component_id,v_receipt.owner_module,
        v_receipt.owner_generation,v_receipt.contract,
        v_receipt.content_sha256,v_receipt.action,
        v_receipt.ready,v_receipt.completed_at;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_catalog_component_purge_v3(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION iam_catalog_component_purge_v3(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) FROM forgecrud_runtime;
