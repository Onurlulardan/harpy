-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_form_record_adoption_advance_v1(
    p_operation_id uuid,p_definition_id uuid,p_epoch bigint,
    p_expected_cursor uuid,p_expected_records bigint,p_expected_links bigint,
    p_expected_source bytea,p_expected_catalog bytea,
    p_new_cursor uuid,p_new_records bigint,p_new_links bigint,
    p_new_source bytea,p_new_catalog bytea,p_done boolean,
    p_updated_at timestamptz
) RETURNS SETOF documents_form_record_adoption_receipts
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $function$
DECLARE v_row documents_form_record_adoption_receipts%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_definition_id) IS DISTINCT FROM 7 OR
       p_epoch < 2 OR p_expected_records < 0 OR p_expected_links < 0 OR
       p_new_records < p_expected_records OR
       p_new_records-p_expected_records>100 OR p_new_links < p_expected_links OR
       pg_catalog.octet_length(p_expected_source) <> 32 OR
       pg_catalog.octet_length(p_expected_catalog) <> 32 OR
       pg_catalog.octet_length(p_new_source) <> 32 OR
       pg_catalog.octet_length(p_new_catalog) <> 32 OR
       p_updated_at IS NULL OR NOT pg_catalog.isfinite(p_updated_at) OR
       pg_catalog.date_trunc('microseconds',p_updated_at) <> p_updated_at OR
       (p_expected_records=0 AND p_expected_cursor<>
        '00000000-0000-0000-0000-000000000000'::uuid) OR
       (p_expected_records>0 AND
        pg_catalog.uuid_extract_version(p_expected_cursor) IS DISTINCT FROM 7) OR
       (p_new_records=0 AND p_new_cursor<>
        '00000000-0000-0000-0000-000000000000'::uuid) OR
       (p_new_records>0 AND
        pg_catalog.uuid_extract_version(p_new_cursor) IS DISTINCT FROM 7) OR
       (p_new_records>p_expected_records AND p_new_cursor<=p_expected_cursor) OR
       (p_new_records=p_expected_records AND
        (NOT p_done OR p_new_cursor<>p_expected_cursor OR
         p_new_links<>p_expected_links OR p_new_source<>p_expected_source OR
         p_new_catalog<>p_expected_catalog)) OR
       (p_new_records>p_expected_records AND p_new_source=p_expected_source) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Documents Form record adoption page is invalid';
    END IF;
    SELECT * INTO v_row FROM documents_form_record_adoption_receipts
    WHERE operation_id=p_operation_id FOR UPDATE;
    IF v_row.operation_id IS NULL OR v_row.definition_id<>p_definition_id OR
       v_row.reference_epoch<>p_epoch THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents Form record adoption is unavailable';
    END IF;
    IF (v_row.cursor_record_id,v_row.record_count,v_row.link_count,
        v_row.source_catalog_sha256,v_row.link_catalog_sha256,
        v_row.lifecycle_state) IS NOT DISTINCT FROM
       (p_new_cursor,p_new_records,p_new_links,p_new_source,p_new_catalog,
        CASE WHEN p_done THEN 'READY' ELSE 'RUNNING' END) THEN
        RETURN NEXT v_row;
        RETURN;
    END IF;
    IF v_row.lifecycle_state<>'RUNNING' OR
       (v_row.cursor_record_id,v_row.record_count,v_row.link_count,
        v_row.source_catalog_sha256,v_row.link_catalog_sha256) IS DISTINCT FROM
       (p_expected_cursor,p_expected_records,p_expected_links,
        p_expected_source,p_expected_catalog) OR p_updated_at<v_row.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='Documents Form record adoption checkpoint differs';
    END IF;
    UPDATE documents_form_record_adoption_receipts SET
        cursor_record_id=p_new_cursor,record_count=p_new_records,
        link_count=p_new_links,source_catalog_sha256=p_new_source,
        link_catalog_sha256=p_new_catalog,
        lifecycle_state=CASE WHEN p_done THEN 'READY' ELSE 'RUNNING' END,
        updated_at=p_updated_at
    WHERE operation_id=p_operation_id RETURNING * INTO v_row;
    RETURN NEXT v_row;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_form_record_adoption_advance_v1(
    uuid,uuid,bigint,uuid,bigint,bigint,bytea,bytea,uuid,bigint,bigint,
    bytea,bytea,boolean,timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_form_record_adoption_advance_v1(
    uuid,uuid,bigint,uuid,bigint,bigint,bytea,bytea,uuid,bigint,bigint,
    bytea,bytea,boolean,timestamptz) TO forgecrud_runtime;
