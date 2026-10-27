-- +goose Up
CREATE TABLE documents_form_record_adoption_receipts (
    operation_id uuid PRIMARY KEY,
    definition_id uuid NOT NULL,
    reference_epoch bigint NOT NULL CHECK (reference_epoch > 1),
    cursor_record_id uuid NOT NULL DEFAULT
        '00000000-0000-0000-0000-000000000000'::uuid,
    record_count bigint NOT NULL DEFAULT 0 CHECK (record_count >= 0),
    link_count bigint NOT NULL DEFAULT 0 CHECK (link_count >= 0),
    source_catalog_sha256 bytea NOT NULL
        CHECK (pg_catalog.octet_length(source_catalog_sha256) = 32),
    link_catalog_sha256 bytea NOT NULL
        CHECK (pg_catalog.octet_length(link_catalog_sha256) = 32),
    lifecycle_state text NOT NULL
        CHECK (lifecycle_state IN ('RUNNING','READY','SUCCEEDED')),
    source_receipt_sha256 bytea,
    receipt_sha256 bytea,
    started_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    completed_at timestamptz,
    UNIQUE (definition_id, reference_epoch),
    CHECK (pg_catalog.isfinite(started_at) AND pg_catalog.isfinite(updated_at)),
    CHECK (
        (lifecycle_state IN ('RUNNING','READY') AND
         source_receipt_sha256 IS NULL AND receipt_sha256 IS NULL AND
         completed_at IS NULL) OR
        (lifecycle_state = 'SUCCEEDED' AND
         pg_catalog.octet_length(source_receipt_sha256) = 32 AND
         pg_catalog.octet_length(receipt_sha256) = 32 AND
         pg_catalog.isfinite(completed_at))
    ),
    CHECK (
        (record_count = 0 AND cursor_record_id =
            '00000000-0000-0000-0000-000000000000'::uuid) OR
        (record_count > 0 AND
         pg_catalog.uuid_extract_version(cursor_record_id) IS NOT DISTINCT FROM 7)
    )
);

-- +goose StatementBegin
CREATE FUNCTION documents_form_record_adoption_load_v1(
    p_operation_id uuid,p_definition_id uuid,p_epoch bigint,
    p_source_seed bytea,p_link_seed bytea,p_started_at timestamptz
) RETURNS SETOF documents_form_record_adoption_receipts
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $function$
DECLARE
    v_row documents_form_record_adoption_receipts%ROWTYPE;
    v_expected_source bytea;
    v_expected_link bytea;
    v_source_label bytea;
    v_link_label bytea;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_definition_id) IS DISTINCT FROM 7 OR
       p_epoch < 2 OR pg_catalog.octet_length(p_source_seed) <> 32 OR
       pg_catalog.octet_length(p_link_seed) <> 32 OR p_started_at IS NULL OR
       NOT pg_catalog.isfinite(p_started_at) OR
       pg_catalog.date_trunc('microseconds',p_started_at) <> p_started_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Documents Form record adoption input is invalid';
    END IF;
    v_source_label:=pg_catalog.convert_to(
        'forgecrud/forms/document-link-source/v1','UTF8');
    v_link_label:=pg_catalog.convert_to(
        'forgecrud/forms/document-link-catalog/v1','UTF8');
    v_expected_source:=pg_catalog.sha256(
        pg_catalog.int4send(pg_catalog.octet_length(v_source_label))||v_source_label||
        pg_catalog.int4send(16)||pg_catalog.uuid_send(p_definition_id)||
        pg_catalog.int4send(8)||pg_catalog.int8send(p_epoch));
    v_expected_link:=pg_catalog.sha256(
        pg_catalog.int4send(pg_catalog.octet_length(v_link_label))||v_link_label||
        pg_catalog.int4send(16)||pg_catalog.uuid_send(p_definition_id)||
        pg_catalog.int4send(8)||pg_catalog.int8send(p_epoch));
    IF p_source_seed<>v_expected_source OR p_link_seed<>v_expected_link THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Documents Form record adoption seed differs';
    END IF;
    INSERT INTO documents_form_record_adoption_receipts(
        operation_id,definition_id,reference_epoch,source_catalog_sha256,
        link_catalog_sha256,lifecycle_state,started_at,updated_at
    ) VALUES(
        p_operation_id,p_definition_id,p_epoch,p_source_seed,p_link_seed,
        'RUNNING',p_started_at,p_started_at
    ) ON CONFLICT (operation_id) DO NOTHING;
    SELECT * INTO v_row FROM documents_form_record_adoption_receipts
    WHERE operation_id=p_operation_id FOR UPDATE;
    IF v_row.definition_id<>p_definition_id OR v_row.reference_epoch<>p_epoch OR
       (v_row.record_count=0 AND
        (v_row.source_catalog_sha256<>p_source_seed OR
         v_row.link_catalog_sha256<>p_link_seed)) THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='Documents Form record adoption replay differs';
    END IF;
    RETURN NEXT v_row;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_form_record_adoption_load_v1(
    uuid,uuid,bigint,bytea,bytea,timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_form_record_adoption_load_v1(
    uuid,uuid,bigint,bytea,bytea,timestamptz) TO harpy_runtime;
