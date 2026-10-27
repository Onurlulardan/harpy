-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_component_blockers_v1(
    p_receipt_id uuid, p_operation_id uuid, p_component_id uuid,
    p_reference_epoch bigint, p_generation bigint, p_completed_at timestamptz
) RETURNS TABLE (
    receipt_id uuid, operation_id uuid, component_id uuid,
    reference_epoch bigint, catalog_version text, zero_digest text,
    blockers jsonb, completed_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component iam_catalog_components%ROWTYPE;
    v_receipt iam_catalog_blocker_receipts%ROWTYPE;
    v_blockers jsonb;
    v_digest text;
BEGIN
    SELECT * INTO v_receipt FROM iam_catalog_blocker_receipts AS receipt
    WHERE receipt.operation_id=p_operation_id AND receipt.component_id=p_component_id
      AND receipt.reference_epoch=p_reference_epoch;
    IF FOUND THEN
        RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
            v_receipt.component_id,v_receipt.reference_epoch,v_receipt.catalog_version,
            v_receipt.zero_digest,v_receipt.blockers,v_receipt.completed_at;
        RETURN;
    END IF;
    IF pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_reference_epoch<1 OR p_generation<1 OR p_completed_at IS NULL OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='catalog blocker request invalid';
    END IF;
    SELECT * INTO v_component FROM iam_catalog_components AS component
    WHERE component.component_id=p_component_id FOR SHARE;
    IF NOT FOUND OR v_component.owner_generation<>p_generation OR
       v_component.writer_state<>'v2' OR v_component.lifecycle_state<>'RETIRED' THEN
        RAISE EXCEPTION USING ERRCODE='55000',MESSAGE='catalog blocker authority differs';
    END IF;
    SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'Owner','iam','Code',blocker.code,'Digest',blocker.digest
    ) ORDER BY blocker.code),'[]'::jsonb) INTO v_blockers
    FROM iam_catalog_component_blockers AS blocker
    WHERE blocker.component_id=p_component_id AND blocker.active;
    v_digest := pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        'iam-catalog-blockers.v1'||E'\\000'||p_component_id::text||E'\\000'||
        p_reference_epoch::text||E'\\000'||v_blockers::text,'UTF8'
    )),'hex');
    INSERT INTO iam_catalog_blocker_receipts (
        receipt_id,operation_id,component_id,reference_epoch,catalog_version,
        zero_digest,blockers,completed_at
    ) VALUES (p_receipt_id,p_operation_id,p_component_id,p_reference_epoch,
        'iam-catalog-blockers.v1',v_digest,v_blockers,p_completed_at)
    RETURNING * INTO v_receipt;
    RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
        v_receipt.component_id,v_receipt.reference_epoch,v_receipt.catalog_version,
        v_receipt.zero_digest,v_receipt.blockers,v_receipt.completed_at;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_catalog_component_blockers_v1(
    uuid,uuid,uuid,bigint,bigint,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_catalog_component_blockers_v1(
    uuid,uuid,uuid,bigint,bigint,timestamptz
) TO harpy_runtime;
