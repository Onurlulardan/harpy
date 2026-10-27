-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_guard_version_mutation_restore_v1() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $function$
DECLARE
    allowed boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'DOCUMENT_VERSION_IMMUTABLE';
    END IF;
    IF SESSION_USER = 'harpy_restore'
       AND OLD.lifecycle_state = 'READY'
       AND NEW.lifecycle_state = OLD.lifecycle_state
       AND NEW.object_key = OLD.object_key
       AND NEW.object_version <> ''
       AND NEW.object_generation > 0
       AND (NEW.id, NEW.item_id, NEW.version_number, NEW.original_name,
            NEW.declared_mime, NEW.detected_mime, NEW.file_kind,
            NEW.plaintext_bytes, NEW.ciphertext_bytes, NEW.plaintext_sha256,
            NEW.ciphertext_sha256, NEW.cipher_format, NEW.key_id, NEW.wrapped_dek,
            NEW.scan_signature_revision, NEW.scanned_at, NEW.delete_receipt_sha256,
            NEW.purged_at, NEW.created_at)
           IS NOT DISTINCT FROM
           (OLD.id, OLD.item_id, OLD.version_number, OLD.original_name,
            OLD.declared_mime, OLD.detected_mime, OLD.file_kind,
            OLD.plaintext_bytes, OLD.ciphertext_bytes, OLD.plaintext_sha256,
            OLD.ciphertext_sha256, OLD.cipher_format, OLD.key_id, OLD.wrapped_dek,
            OLD.scan_signature_revision, OLD.scanned_at, OLD.delete_receipt_sha256,
            OLD.purged_at, OLD.created_at) THEN
        RETURN NEW;
    END IF;
    allowed := (OLD.lifecycle_state = 'READY' AND
                NEW.lifecycle_state = 'DELETE_PENDING' AND
                NEW.delete_receipt_sha256 IS NULL AND NEW.purged_at IS NULL)
            OR (OLD.lifecycle_state = 'DELETE_PENDING' AND
                NEW.lifecycle_state = 'PURGED' AND
                NEW.delete_receipt_sha256 IS NOT NULL AND NEW.purged_at IS NOT NULL);
    IF NOT allowed OR
       (NEW.id, NEW.item_id, NEW.version_number, NEW.original_name,
        NEW.declared_mime, NEW.detected_mime, NEW.file_kind,
        NEW.plaintext_bytes, NEW.ciphertext_bytes, NEW.plaintext_sha256,
        NEW.ciphertext_sha256, NEW.cipher_format, NEW.object_key,
        NEW.object_version, NEW.object_generation, NEW.key_id, NEW.wrapped_dek,
        NEW.scan_signature_revision, NEW.scanned_at, NEW.created_at)
       IS DISTINCT FROM
       (OLD.id, OLD.item_id, OLD.version_number, OLD.original_name,
        OLD.declared_mime, OLD.detected_mime, OLD.file_kind,
        OLD.plaintext_bytes, OLD.ciphertext_bytes, OLD.plaintext_sha256,
        OLD.ciphertext_sha256, OLD.cipher_format, OLD.object_key,
        OLD.object_version, OLD.object_generation, OLD.key_id, OLD.wrapped_dek,
        OLD.scan_signature_revision, OLD.scanned_at, OLD.created_at) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'DOCUMENT_VERSION_IMMUTABLE';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE OR REPLACE TRIGGER documents_versions_mutation_guard
BEFORE UPDATE OR DELETE ON documents_versions
FOR EACH ROW EXECUTE FUNCTION documents_guard_version_mutation_restore_v1();
ALTER TABLE documents_versions ENABLE ALWAYS TRIGGER documents_versions_mutation_guard;

CREATE FUNCTION documents_recovery_apply_locator_v1(
    p_restore_operation_id uuid,
    p_recovery_set_id uuid,
    p_version_id uuid,
    p_source_object_key text,
    p_source_object_version text,
    p_target_object_key text,
    p_target_object_version text,
    p_target_object_generation bigint,
    p_ciphertext_sha256 bytea
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_existing documents_restore_locator_remaps%ROWTYPE;
    v_live documents_versions%ROWTYPE;
    v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_restore'
       OR p_restore_operation_id IS NULL OR p_target_object_generation < 1
       OR p_target_object_version = '' OR p_target_object_key <> p_source_object_key
       OR pg_catalog.octet_length(p_ciphertext_sha256) <> 32
       OR NOT EXISTS (
           SELECT 1 FROM documents_dr_holds AS hold
           JOIN documents_dr_copy_receipts AS receipt
             ON receipt.recovery_set_id = hold.recovery_set_id
           WHERE hold.recovery_set_id = p_recovery_set_id
             AND hold.lifecycle_state IN ('VERIFIED','RELEASED')
             AND receipt.version_id = p_version_id
             AND receipt.source_object_key = p_source_object_key
             AND receipt.source_object_version = p_source_object_version
             AND receipt.ciphertext_sha256 = p_ciphertext_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'documents restore locator denied';
    END IF;
    SELECT * INTO v_existing FROM documents_restore_locator_remaps
    WHERE restore_operation_id = p_restore_operation_id AND version_id = p_version_id;
    IF FOUND THEN
        SELECT * INTO v_live
        FROM documents_versions
        WHERE id = p_version_id
        FOR UPDATE;
        IF NOT FOUND THEN
            RETURN false;
        END IF;
        RETURN (v_existing.recovery_set_id, v_existing.source_object_key,
                v_existing.source_object_version, v_existing.target_object_key,
                v_existing.target_object_version, v_existing.target_object_generation,
                v_existing.ciphertext_sha256, v_live.lifecycle_state,
                v_live.object_key, v_live.object_version,
                v_live.object_generation, v_live.ciphertext_sha256)
            IS NOT DISTINCT FROM
               (p_recovery_set_id, p_source_object_key, p_source_object_version,
                p_target_object_key, p_target_object_version,
                p_target_object_generation, p_ciphertext_sha256, 'READY',
                p_target_object_key, p_target_object_version,
                p_target_object_generation, p_ciphertext_sha256);
    END IF;
    UPDATE documents_versions SET object_key = p_target_object_key,
        object_version = p_target_object_version,
        object_generation = p_target_object_generation
    WHERE id = p_version_id AND lifecycle_state = 'READY'
      AND object_key = p_source_object_key AND object_version = p_source_object_version
      AND ciphertext_sha256 = p_ciphertext_sha256;
    IF NOT FOUND THEN RETURN false; END IF;
    INSERT INTO documents_restore_locator_remaps VALUES (
        p_restore_operation_id, p_recovery_set_id, p_version_id,
        p_source_object_key, p_source_object_version, p_target_object_key,
        p_target_object_version, p_target_object_generation,
        p_ciphertext_sha256, v_now
    );
    RETURN true;
END;
$function$;
-- +goose StatementEnd
