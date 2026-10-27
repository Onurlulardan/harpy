-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_guard_version_mutation_v1() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
DECLARE allowed boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'DOCUMENT_VERSION_IMMUTABLE';
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
$$;
-- +goose StatementEnd

CREATE TRIGGER documents_versions_mutation_guard
BEFORE UPDATE OR DELETE ON documents_versions
FOR EACH ROW EXECUTE FUNCTION documents_guard_version_mutation_v1();
ALTER TABLE documents_versions ENABLE ALWAYS TRIGGER documents_versions_mutation_guard;

