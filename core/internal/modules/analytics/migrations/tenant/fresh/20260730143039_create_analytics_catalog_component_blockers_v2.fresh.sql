-- +goose Up
-- +goose StatementBegin
DROP FUNCTION analytics_catalog_component_blockers_v1(
    uuid,uuid,uuid,bigint,bigint,timestamptz
);
CREATE FUNCTION analytics_catalog_component_blockers_v1(
    p_receipt_id uuid,p_operation_id uuid,p_component_id uuid,
    p_reference_epoch bigint,p_generation bigint,p_completed_at timestamptz
)
RETURNS TABLE(
    receipt_id uuid,operation_id uuid,component_id uuid,
    reference_epoch bigint,catalog_version text,zero_digest text,
    blockers jsonb,completed_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component analytics_catalog_components%ROWTYPE;
    v_receipt analytics_catalog_blocker_receipts%ROWTYPE;
    v_snapshot_blockers jsonb;
    v_snapshot_zero_digest text;
BEGIN
    IF pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_receipt_id=p_operation_id OR p_reference_epoch<1 OR p_generation<1 OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog blocker request is invalid';
    END IF;
    SELECT * INTO v_component
    FROM analytics_catalog_components AS component
    WHERE component.component_id=p_component_id
    FOR SHARE;
    IF NOT FOUND OR v_component.owner_generation<>p_generation OR
       v_component.writer_state<>'v2' OR
       v_component.lifecycle_state<>'RETIRED' OR
       p_completed_at<v_component.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog blocker authority differs';
    END IF;
    SELECT snapshot.blockers,snapshot.zero_digest
    INTO STRICT v_snapshot_blockers,v_snapshot_zero_digest
    FROM analytics_catalog_blocker_snapshot_v2(
        p_component_id,p_reference_epoch,p_generation
    ) AS snapshot;
    SELECT * INTO v_receipt
    FROM analytics_catalog_blocker_receipts AS receipt
    WHERE receipt.operation_id=p_operation_id
      AND receipt.component_id=p_component_id
      AND receipt.reference_epoch=p_reference_epoch
    FOR SHARE;
    IF FOUND THEN
        IF v_receipt.catalog_version IS DISTINCT FROM
               'analytics-catalog-blockers.v2' OR
           v_receipt.zero_digest IS DISTINCT FROM v_snapshot_zero_digest OR
           v_receipt.blockers IS DISTINCT FROM v_snapshot_blockers OR
           v_receipt.completed_at>p_completed_at THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog blocker receipt is stale';
        END IF;
    ELSE
        INSERT INTO analytics_catalog_blocker_receipts(
            receipt_id,operation_id,component_id,reference_epoch,
            catalog_version,zero_digest,blockers,completed_at
        ) VALUES (
            p_receipt_id,p_operation_id,p_component_id,p_reference_epoch,
            'analytics-catalog-blockers.v2',v_snapshot_zero_digest,
            v_snapshot_blockers,p_completed_at
        ) RETURNING * INTO v_receipt;
    END IF;
    RETURN QUERY SELECT v_receipt.receipt_id,v_receipt.operation_id,
        v_receipt.component_id,v_receipt.reference_epoch,
        v_receipt.catalog_version,v_receipt.zero_digest,
        v_receipt.blockers,v_receipt.completed_at;
END;
$function$;
REVOKE ALL ON FUNCTION analytics_catalog_component_blockers_v1(
    uuid,uuid,uuid,bigint,bigint,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_catalog_component_blockers_v1(
    uuid,uuid,uuid,bigint,bigint,timestamptz
) TO forgecrud_runtime;
-- +goose StatementEnd
