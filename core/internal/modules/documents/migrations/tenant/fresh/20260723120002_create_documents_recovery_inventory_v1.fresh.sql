-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_recovery_inventory_v1(
    p_recovery_set_id uuid,
    p_tenant_id uuid
)
RETURNS TABLE(
    version_id uuid,
    object_key text,
    object_version text,
    object_generation bigint,
    ciphertext_sha256 bytea,
    ciphertext_bytes bigint,
    key_id text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_backup'
       OR NOT EXISTS (
           SELECT 1 FROM documents_dr_holds AS hold
           WHERE hold.recovery_set_id = p_recovery_set_id
             AND hold.tenant_id = p_tenant_id
             AND hold.lifecycle_state IN ('HOLDING','COPYING','VERIFIED')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'documents recovery inventory denied';
    END IF;
    RETURN QUERY
    SELECT version.id, version.object_key, version.object_version,
           version.object_generation, version.ciphertext_sha256,
           version.ciphertext_bytes, version.key_id
    FROM documents_versions AS version
    WHERE version.lifecycle_state = 'READY'
    ORDER BY version.id;
END;
$function$;

CREATE FUNCTION documents_recovery_record_copy_v1(
    p_recovery_set_id uuid,
    p_version_id uuid,
    p_source_object_key text,
    p_source_object_version text,
    p_target_object_key text,
    p_target_object_version text,
    p_ciphertext_sha256 bytea,
    p_ciphertext_bytes bigint,
    p_key_id text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_existing documents_dr_copy_receipts%ROWTYPE;
    v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_backup'
       OR NOT EXISTS (
           SELECT 1 FROM documents_dr_holds AS hold
           WHERE hold.recovery_set_id = p_recovery_set_id
             AND hold.lifecycle_state IN ('HOLDING','COPYING')
             AND p_target_object_key =
                 'environments/' || hold.environment_id::text ||
                 '/recovery-sets/' || hold.recovery_set_id::text ||
                 '/documents/tenants/' || hold.tenant_id::text ||
                 '/versions/' || p_version_id::text || '.fcdoc1'
       )
       OR NOT EXISTS (
           SELECT 1 FROM documents_versions AS version
           WHERE version.id = p_version_id
             AND version.lifecycle_state = 'READY'
             AND version.object_key = p_source_object_key
             AND version.object_version = p_source_object_version
             AND version.ciphertext_sha256 = p_ciphertext_sha256
             AND version.ciphertext_bytes = p_ciphertext_bytes
             AND version.key_id = p_key_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'documents recovery copy denied';
    END IF;
    SELECT * INTO v_existing
    FROM documents_dr_copy_receipts
    WHERE recovery_set_id = p_recovery_set_id AND version_id = p_version_id;
    IF FOUND THEN
        IF (v_existing.source_object_key, v_existing.source_object_version,
            v_existing.target_object_key, v_existing.target_object_version,
            v_existing.ciphertext_sha256, v_existing.ciphertext_bytes,
            v_existing.key_id)
           IS DISTINCT FROM
           (p_source_object_key, p_source_object_version,
            p_target_object_key, p_target_object_version,
            p_ciphertext_sha256, p_ciphertext_bytes, p_key_id) THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'documents recovery copy conflicts';
        END IF;
        RETURN;
    END IF;
    INSERT INTO documents_dr_copy_receipts(
        recovery_set_id, version_id, source_object_key, source_object_version,
        target_object_key, target_object_version, ciphertext_sha256,
        ciphertext_bytes, key_id, verified_at
    ) VALUES (
        p_recovery_set_id, p_version_id, p_source_object_key,
        p_source_object_version, p_target_object_key, p_target_object_version,
        p_ciphertext_sha256, p_ciphertext_bytes, p_key_id, v_now
    );
    UPDATE documents_dr_holds SET lifecycle_state = 'COPYING', updated_at = v_now
    WHERE recovery_set_id = p_recovery_set_id AND lifecycle_state = 'HOLDING';
END;
$function$;
-- +goose StatementEnd
