-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_v1_adoption_project_version_v1(
    p_adoption_run_id uuid,p_record jsonb
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_checkpoint documents_blob_copy_checkpoints%ROWTYPE;
    v_id uuid;
    v_plaintext_sha bytea;
    v_logical jsonb;
BEGIN
    v_id:=(p_record->>'id')::uuid;
    v_plaintext_sha:=pg_catalog.decode(
        pg_catalog.substr(p_record->>'plaintext_sha256',3),'hex'
    );
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Documents V1 version projection principal denied';
    END IF;
    SELECT checkpoint.* INTO STRICT v_checkpoint
    FROM documents_blob_copy_checkpoints AS checkpoint
    WHERE checkpoint.run_id=p_adoption_run_id
      AND checkpoint.target_version_id=v_id
      AND checkpoint.lifecycle_state='VERIFIED'
    FOR SHARE;
    IF v_checkpoint.source_bytes IS DISTINCT FROM
           (p_record->>'plaintext_bytes')::bigint OR
       v_checkpoint.source_sha256 IS DISTINCT FROM v_plaintext_sha OR
       v_checkpoint.target_plaintext_sha256 IS DISTINCT FROM v_plaintext_sha OR
       v_checkpoint.metadata_sha256 IS DISTINCT FROM
           v_checkpoint.source_metadata_sha256 OR
       v_checkpoint.applied_lsn IS DISTINCT FROM v_checkpoint.source_lsn THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 verified copy differs from signed version';
    END IF;
    INSERT INTO documents_versions (
        id,item_id,version_number,lifecycle_state,original_name,declared_mime,
        detected_mime,file_kind,plaintext_bytes,ciphertext_bytes,
        plaintext_sha256,ciphertext_sha256,cipher_format,object_key,
        object_version,object_generation,key_id,wrapped_dek,
        scan_signature_revision,scanned_at,created_at
    ) VALUES (
        v_id,(p_record->>'item_id')::uuid,(p_record->>'version_number')::bigint,
        'READY',p_record->>'original_name',p_record->>'declared_mime',
        p_record->>'detected_mime',p_record->>'file_kind',
        (p_record->>'plaintext_bytes')::bigint,
        v_checkpoint.target_ciphertext_bytes,v_plaintext_sha,
        v_checkpoint.target_ciphertext_sha256,'FCDOC1',
        v_checkpoint.target_object_key,v_checkpoint.target_object_version,
        v_checkpoint.target_object_generation,v_checkpoint.target_key_id,
        v_checkpoint.target_wrapped_dek,p_record->>'scan_signature_revision',
        (p_record->>'scanned_at')::timestamptz,
        (p_record->>'created_at')::timestamptz
    ) ON CONFLICT (id) DO NOTHING;
    SELECT pg_catalog.jsonb_build_object(
        'id',target.id,'item_id',target.item_id,
        'version_number',target.version_number,
        'lifecycle_state',target.lifecycle_state,
        'original_name',target.original_name,'declared_mime',target.declared_mime,
        'detected_mime',target.detected_mime,'file_kind',target.file_kind,
        'plaintext_bytes',target.plaintext_bytes,
        'plaintext_sha256',target.plaintext_sha256,
        'scan_signature_revision',target.scan_signature_revision,
        'scanned_at',target.scanned_at,'created_at',target.created_at
    ) INTO v_logical
    FROM documents_versions AS target
    WHERE target.id=v_id AND target.ciphertext_bytes=v_checkpoint.target_ciphertext_bytes
      AND target.ciphertext_sha256=v_checkpoint.target_ciphertext_sha256
      AND target.object_key=v_checkpoint.target_object_key
      AND target.object_version=v_checkpoint.target_object_version
      AND target.object_generation=v_checkpoint.target_object_generation
      AND target.key_id=v_checkpoint.target_key_id
      AND target.wrapped_dek=v_checkpoint.target_wrapped_dek;
    IF v_logical IS DISTINCT FROM p_record THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 version projection differs';
    END IF;
    RETURN v_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_v1_adoption_project_version_v1(uuid,jsonb)
FROM PUBLIC;
