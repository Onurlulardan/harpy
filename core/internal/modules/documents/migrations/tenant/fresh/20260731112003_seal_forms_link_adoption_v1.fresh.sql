-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_form_record_adoption_seal_v1(
    p_operation_id uuid,p_definition_id uuid,p_epoch bigint,p_cursor uuid,
    p_records bigint,p_links bigint,p_source bytea,p_catalog bytea,
    p_source_receipt bytea,p_issued_at timestamptz
) RETURNS TABLE(
    source_receipt_sha256 bytea,receipt_sha256 bytea,completed_at timestamptz
) LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $function$
DECLARE
    v_row documents_form_record_adoption_receipts%ROWTYPE;
    v_expected_source bytea;
    v_receipt bytea;
    v_issued_micro bigint;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_definition_id) IS DISTINCT FROM 7 OR
       p_epoch<2 OR p_records<0 OR p_links<0 OR
       pg_catalog.octet_length(p_source)<>32 OR
       pg_catalog.octet_length(p_catalog)<>32 OR
       pg_catalog.octet_length(p_source_receipt)<>32 OR
       p_issued_at IS NULL OR NOT pg_catalog.isfinite(p_issued_at) OR
       pg_catalog.date_trunc('microseconds',p_issued_at)<>p_issued_at OR
       (p_records=0 AND p_cursor<>
        '00000000-0000-0000-0000-000000000000'::uuid) OR
       (p_records>0 AND
        pg_catalog.uuid_extract_version(p_cursor) IS DISTINCT FROM 7) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Documents Form record adoption seal is invalid';
    END IF;
    v_issued_micro:=(pg_catalog.date_part('epoch',p_issued_at)*1000000)::bigint;
    v_expected_source:=pg_catalog.sha256(
        pg_catalog.convert_to(
            'forgecrud/forms/document-link-source-receipt/v1','UTF8') ||
        pg_catalog.uuid_send(p_operation_id) ||
        pg_catalog.uuid_send(p_definition_id) ||
        pg_catalog.int8send(p_epoch) || pg_catalog.uuid_send(p_cursor) ||
        pg_catalog.int8send(p_records) || pg_catalog.int8send(p_links) ||
        p_source || p_catalog || pg_catalog.int8send(v_issued_micro)
    );
    IF p_source_receipt<>v_expected_source THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Documents Form source receipt differs';
    END IF;
    SELECT * INTO v_row FROM documents_form_record_adoption_receipts
    WHERE operation_id=p_operation_id FOR UPDATE;
    IF v_row.lifecycle_state='SUCCEEDED' THEN
        IF v_row.source_receipt_sha256<>p_source_receipt OR
           v_row.completed_at<>p_issued_at THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='Documents Form adoption seal replay differs';
        END IF;
        RETURN QUERY SELECT v_row.source_receipt_sha256,
            v_row.receipt_sha256,v_row.completed_at;
        RETURN;
    END IF;
    IF v_row.lifecycle_state<>'READY' OR v_row.definition_id<>p_definition_id OR
       v_row.reference_epoch<>p_epoch OR
       (v_row.cursor_record_id,v_row.record_count,v_row.link_count,
        v_row.source_catalog_sha256,v_row.link_catalog_sha256) IS DISTINCT FROM
       (p_cursor,p_records,p_links,p_source,p_catalog) OR
       p_issued_at<v_row.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents Form adoption inventory differs';
    END IF;
    v_receipt:=pg_catalog.sha256(
        pg_catalog.convert_to(
            'forgecrud/documents/form-record-adoption-receipt/v1','UTF8') ||
        p_source_receipt || pg_catalog.uuid_send(p_operation_id) ||
        pg_catalog.int8send(p_epoch) || pg_catalog.int8send(v_issued_micro)
    );
    UPDATE documents_form_record_adoption_receipts SET
        lifecycle_state='SUCCEEDED',source_receipt_sha256=p_source_receipt,
        receipt_sha256=v_receipt,updated_at=p_issued_at,completed_at=p_issued_at
    WHERE operation_id=p_operation_id RETURNING * INTO v_row;
    RETURN QUERY SELECT v_row.source_receipt_sha256,
        v_row.receipt_sha256,v_row.completed_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_form_record_adoption_seal_v1(
    uuid,uuid,bigint,uuid,bigint,bigint,bytea,bytea,bytea,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_form_record_adoption_seal_v1(
    uuid,uuid,bigint,uuid,bigint,bigint,bytea,bytea,bytea,timestamptz
) TO harpy_runtime;
