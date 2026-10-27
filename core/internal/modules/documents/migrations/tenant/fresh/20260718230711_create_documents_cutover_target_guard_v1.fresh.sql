-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_v1_adoption_target_record_valid_v1(
    p_target_kind text,p_target_identity text,p_target_id uuid,p_record jsonb
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_relation oid;
    v_expected jsonb;
    v_actual jsonb;
    v_folder documents_folders%ROWTYPE;
    v_item documents_items%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       pg_catalog.jsonb_typeof(p_record) IS DISTINCT FROM 'object' OR
       p_record='{}'::jsonb OR pg_catalog.octet_length(p_record::text)>1048576 OR
       p_target_id='00000000-0000-0000-0000-000000000000'::uuid OR
       p_target_identity IS DISTINCT FROM p_target_kind||':'||p_target_id::text THEN
        RETURN false;
    END IF;
    v_relation:=CASE p_target_kind
        WHEN 'folder' THEN 'documents_folders'::regclass::oid
        WHEN 'item' THEN 'documents_items'::regclass::oid
        ELSE NULL
    END;
    IF v_relation IS NOT NULL THEN
        SELECT pg_catalog.jsonb_agg(
            attribute.attname::text ORDER BY attribute.attname::text COLLATE "C"
        )
        INTO v_expected FROM pg_catalog.pg_attribute AS attribute
        WHERE attribute.attrelid=v_relation AND attribute.attnum>0
          AND NOT attribute.attisdropped;
    ELSIF p_target_kind='version' THEN
        v_expected:=pg_catalog.to_jsonb(ARRAY[
            'created_at','declared_mime','detected_mime','file_kind','id','item_id',
            'lifecycle_state','original_name','plaintext_bytes','plaintext_sha256',
            'scan_signature_revision','scanned_at','version_number'
        ]::text[]);
    ELSIF p_target_kind='inventory' THEN
        v_expected:=pg_catalog.to_jsonb(ARRAY[
            'adoption_run_id','created_at','cutover_generation',
            'manifest_sha256','transform_version'
        ]::text[]);
    ELSIF p_target_kind='checkpoint' THEN
        v_expected:=pg_catalog.to_jsonb(ARRAY[
            'adoption_run_id','enqueued_at','job_id','source_bytes',
            'source_document_id','source_lsn','source_metadata_sha256',
            'source_object_key','source_object_revision','source_sha256',
            'source_version_id','target_document_id','target_version_id'
        ]::text[]);
    ELSE
        RETURN false;
    END IF;
    SELECT pg_catalog.jsonb_agg(key ORDER BY key COLLATE "C") INTO v_actual
    FROM pg_catalog.jsonb_object_keys(p_record) AS key;
    IF v_actual IS DISTINCT FROM v_expected THEN RETURN false; END IF;
    BEGIN
        CASE p_target_kind
            WHEN 'folder' THEN
                SELECT * INTO v_folder
                FROM pg_catalog.jsonb_populate_record(NULL::documents_folders,p_record);
                RETURN v_folder.id=p_target_id AND v_folder.row_version>0 AND
                    v_folder.lifecycle_state IN ('ACTIVE','TRASHED');
            WHEN 'item' THEN
                SELECT * INTO v_item
                FROM pg_catalog.jsonb_populate_record(NULL::documents_items,p_record);
                RETURN v_item.id=p_target_id AND v_item.row_version>0 AND
                    v_item.lifecycle_state IN ('ACTIVE','TRASHED','PURGE_COOLING');
            WHEN 'version' THEN
                RETURN (p_record->>'id')::uuid=p_target_id AND
                    (p_record->>'item_id')::uuid IS NOT NULL AND
                    p_record->>'lifecycle_state'='READY' AND
                    (p_record->>'version_number')::bigint>0 AND
                    (p_record->>'plaintext_bytes')::bigint BETWEEN 1 AND 104857600 AND
                    p_record->>'plaintext_sha256'~'^\\x[0-9a-f]{64}$' AND
                    pg_catalog.isfinite((p_record->>'created_at')::timestamptz) AND
                    pg_catalog.isfinite((p_record->>'scanned_at')::timestamptz);
            WHEN 'inventory' THEN
                RETURN (p_record->>'adoption_run_id')::uuid=p_target_id AND
                    pg_catalog.uuid_extract_version(p_target_id)=7 AND
                    (p_record->>'cutover_generation')::bigint>0 AND
                    p_record->>'manifest_sha256'~'^[0-9a-f]{64}$' AND
                    p_record->>'transform_version'='documents_adoption_v1' AND
                    pg_catalog.isfinite((p_record->>'created_at')::timestamptz);
            WHEN 'checkpoint' THEN
                RETURN (p_record->>'target_version_id')::uuid=p_target_id AND
                    pg_catalog.uuid_extract_version(
                        (p_record->>'adoption_run_id')::uuid)=7 AND
                    pg_catalog.uuid_extract_version((p_record->>'job_id')::uuid)=7 AND
                    pg_catalog.uuid_extract_version(
                        (p_record->>'source_version_id')::uuid)=7 AND
                    (p_record->>'source_bytes')::bigint BETWEEN 1 AND 104857600 AND
                    p_record->>'source_sha256'~'^[0-9a-f]{64}$' AND
                    p_record->>'source_metadata_sha256'~'^[0-9a-f]{64}$' AND
                    (p_record->>'source_lsn')::pg_lsn IS NOT NULL AND
                    pg_catalog.octet_length(p_record->>'source_object_key')
                        BETWEEN 1 AND 1024 AND
                    pg_catalog.octet_length(p_record->>'source_object_revision')
                        BETWEEN 1 AND 1024 AND
                    pg_catalog.isfinite((p_record->>'enqueued_at')::timestamptz);
        END CASE;
    EXCEPTION WHEN invalid_text_representation OR numeric_value_out_of_range OR
        datetime_field_overflow THEN
        RETURN false;
    END;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_v1_adoption_target_record_valid_v1(
    text,text,uuid,jsonb
) FROM PUBLIC;
