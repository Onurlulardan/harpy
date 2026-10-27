-- +goose Up
ALTER TABLE documents_archives
    DROP CONSTRAINT documents_archives_lifecycle_state_check,
    DROP CONSTRAINT documents_archives_check,
    DROP CONSTRAINT documents_archives_cipher_ready_check,
    ADD CONSTRAINT documents_archives_lifecycle_state_check CHECK (
        lifecycle_state IN ('PENDING','CANCELLED','READY','EXPIRED','PURGED')
    ),
    ADD CONSTRAINT documents_archives_recovery_ready_check CHECK (
        (lifecycle_state IN ('PENDING','CANCELLED') AND
         ready_at IS NULL AND expires_at IS NULL)
        OR
        (lifecycle_state IN ('READY','EXPIRED','PURGED') AND entry_count > 0 AND
         entries_sha256 IS NOT NULL AND ciphertext_bytes IS NOT NULL AND
         artifact_sha256 IS NOT NULL AND object_key IS NOT NULL AND
         object_version IS NOT NULL AND object_generation IS NOT NULL AND
         key_id IS NOT NULL AND wrapped_dek IS NOT NULL AND ready_at IS NOT NULL AND
         expires_at = ready_at + interval '24 hours')
    ),
    ADD CONSTRAINT documents_archives_cipher_ready_check CHECK (
        (lifecycle_state IN ('PENDING','CANCELLED') AND
         ((artifact_plaintext_bytes IS NULL AND ciphertext_sha256 IS NULL AND
           cipher_format IS NULL) OR
          (artifact_plaintext_bytes IS NOT NULL AND ciphertext_sha256 IS NOT NULL AND
           cipher_format = 'FCDOC1')))
        OR
        (lifecycle_state IN ('READY','EXPIRED','PURGED') AND
         artifact_plaintext_bytes IS NOT NULL AND ciphertext_sha256 IS NOT NULL AND
         cipher_format = 'FCDOC1')
    );

-- +goose StatementBegin
CREATE FUNCTION documents_guard_archive_transition_restore_v1() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $function$
DECLARE allowed boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'DOCUMENT_ARCHIVE_IMMUTABLE';
    END IF;
    IF SESSION_USER = 'harpy_restore'
       AND OLD.lifecycle_state = 'PENDING'
       AND NEW.lifecycle_state = 'CANCELLED'
       AND NEW.updated_at >= OLD.updated_at
       AND (pg_catalog.to_jsonb(NEW) - ARRAY['lifecycle_state','updated_at'])
           IS NOT DISTINCT FROM
           (pg_catalog.to_jsonb(OLD) - ARRAY['lifecycle_state','updated_at']) THEN
        RETURN NEW;
    END IF;
    allowed :=
        (OLD.lifecycle_state = 'PENDING' AND NEW.lifecycle_state = 'PENDING' AND
         OLD.entry_count = 0 AND NEW.entry_count > 0 AND
         OLD.entries_sha256 IS NULL AND NEW.entries_sha256 IS NOT NULL)
        OR (OLD.lifecycle_state = 'PENDING' AND NEW.lifecycle_state = 'PENDING' AND
            OLD.entry_count > 0 AND NEW.entry_count = OLD.entry_count AND
            NEW.entries_sha256 = OLD.entries_sha256 AND OLD.object_version IS NULL AND
            NEW.object_version IS NOT NULL)
        OR (OLD.lifecycle_state = 'PENDING' AND NEW.lifecycle_state = 'READY' AND
            OLD.entry_count > 0 AND NEW.entry_count = OLD.entry_count AND
            NEW.entries_sha256 = OLD.entries_sha256)
        OR (OLD.lifecycle_state = 'READY' AND NEW.lifecycle_state = 'EXPIRED')
        OR (OLD.lifecycle_state = 'EXPIRED' AND NEW.lifecycle_state = 'PURGED' AND
            NEW.delete_receipt_sha256 IS NOT NULL AND NEW.purged_at IS NOT NULL);
    IF NOT allowed OR
       (NEW.id, NEW.actor_id, NEW.authorization_revision, NEW.job_id,
        NEW.retry_of_archive_id, NEW.request_sha256, NEW.request_document,
        NEW.billing_reservation_id, NEW.created_at)
       IS DISTINCT FROM
       (OLD.id, OLD.actor_id, OLD.authorization_revision, OLD.job_id,
        OLD.retry_of_archive_id, OLD.request_sha256, OLD.request_document,
        OLD.billing_reservation_id, OLD.created_at) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'DOCUMENT_ARCHIVE_IMMUTABLE';
    END IF;
    IF OLD.lifecycle_state <> 'PENDING' AND
       (NEW.entry_count, NEW.entries_sha256, NEW.plaintext_bytes,
        NEW.ciphertext_bytes, NEW.artifact_sha256, NEW.ciphertext_sha256,
        NEW.artifact_plaintext_bytes, NEW.cipher_format, NEW.object_key,
        NEW.object_version, NEW.object_generation, NEW.key_id, NEW.wrapped_dek,
        NEW.ready_at, NEW.expires_at)
       IS DISTINCT FROM
       (OLD.entry_count, OLD.entries_sha256, OLD.plaintext_bytes,
        OLD.ciphertext_bytes, OLD.artifact_sha256, OLD.ciphertext_sha256,
        OLD.artifact_plaintext_bytes, OLD.cipher_format, OLD.object_key,
        OLD.object_version, OLD.object_generation, OLD.key_id, OLD.wrapped_dek,
        OLD.ready_at, OLD.expires_at) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'DOCUMENT_ARCHIVE_IMMUTABLE';
    END IF;
    IF OLD.lifecycle_state = 'PENDING' AND OLD.object_version IS NOT NULL AND
       (NEW.artifact_plaintext_bytes, NEW.ciphertext_bytes, NEW.artifact_sha256,
        NEW.ciphertext_sha256, NEW.cipher_format, NEW.object_key,
        NEW.object_version, NEW.object_generation, NEW.key_id, NEW.wrapped_dek)
       IS DISTINCT FROM
       (OLD.artifact_plaintext_bytes, OLD.ciphertext_bytes, OLD.artifact_sha256,
        OLD.ciphertext_sha256, OLD.cipher_format, OLD.object_key,
        OLD.object_version, OLD.object_generation, OLD.key_id, OLD.wrapped_dek) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'DOCUMENT_ARCHIVE_IMMUTABLE';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE OR REPLACE TRIGGER documents_archives_transition_guard
BEFORE UPDATE OR DELETE ON documents_archives
FOR EACH ROW EXECUTE FUNCTION documents_guard_archive_transition_restore_v1();
ALTER TABLE documents_archives ENABLE ALWAYS TRIGGER documents_archives_transition_guard;
