-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_guard_upload_transition_restore_v1() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $function$
DECLARE allowed boolean;
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'DOCUMENT_UPLOAD_IMMUTABLE';
    END IF;
    IF SESSION_USER = 'harpy_restore'
       AND OLD.lifecycle_state NOT IN ('READY','CANCELLED','EXPIRED')
       AND NEW.lifecycle_state = 'CANCELLED'
       AND NEW.lease_owner_id IS NULL AND NEW.lease_expires_at IS NULL
       AND NEW.cancel_requested_at IS NOT NULL AND NOT NEW.cleanup_required
       AND (NEW.id, NEW.actor_id, NEW.document_id, NEW.candidate_version_id,
            NEW.expected_current_version_id, NEW.folder_id, NEW.request_sha256,
            NEW.expected_content_sha256, NEW.expected_plaintext_bytes,
            NEW.ciphertext_upper_bound, NEW.original_name, NEW.declared_mime,
            NEW.description, NEW.tags, NEW.billing_reservation_id,
            NEW.expires_at, NEW.created_at)
           IS NOT DISTINCT FROM
           (OLD.id, OLD.actor_id, OLD.document_id, OLD.candidate_version_id,
            OLD.expected_current_version_id, OLD.folder_id, OLD.request_sha256,
            OLD.expected_content_sha256, OLD.expected_plaintext_bytes,
            OLD.ciphertext_upper_bound, OLD.original_name, OLD.declared_mime,
            OLD.description, OLD.tags, OLD.billing_reservation_id,
            OLD.expires_at, OLD.created_at) THEN
        RETURN NEW;
    END IF;
    allowed :=
        (OLD.lifecycle_state IN ('CREATED','RETRYABLE') AND
         NEW.lifecycle_state = 'RECEIVING' AND
         NEW.attempt_generation = OLD.attempt_generation + 1 AND
         NOT OLD.cleanup_required AND NEW.cancel_requested_at IS NULL)
        OR (OLD.lifecycle_state = 'RECEIVING' AND
            NEW.lifecycle_state IN ('SCANNING','RETRYABLE','REJECTED'))
        OR (OLD.lifecycle_state = 'SCANNING' AND
            NEW.lifecycle_state IN ('READY','RETRYABLE','REJECTED','CONFLICTED'))
        OR (OLD.lifecycle_state IN ('RECEIVING','SCANNING') AND
            NEW.lifecycle_state = OLD.lifecycle_state AND
            OLD.cancel_requested_at IS NULL AND NEW.cancel_requested_at IS NOT NULL)
        OR (OLD.lifecycle_state IN ('RECEIVING','SCANNING') AND
            NEW.lifecycle_state = 'CANCELLED' AND
            NEW.cancel_requested_at IS NOT NULL AND NEW.cleanup_required)
        OR (OLD.lifecycle_state IN ('CREATED','RETRYABLE') AND
            NEW.lifecycle_state IN ('CANCELLED','EXPIRED') AND NOT OLD.cleanup_required)
        OR (OLD.lifecycle_state IN ('RETRYABLE','REJECTED','CONFLICTED','CANCELLED') AND
            NEW.lifecycle_state = OLD.lifecycle_state AND OLD.cleanup_required AND
            NOT NEW.cleanup_required AND NEW.cleanup_receipt_sha256 IS NOT NULL);
    IF NOT allowed OR
       (NEW.id, NEW.actor_id, NEW.document_id, NEW.candidate_version_id,
        NEW.expected_current_version_id, NEW.folder_id, NEW.request_sha256,
        NEW.expected_content_sha256, NEW.expected_plaintext_bytes,
        NEW.ciphertext_upper_bound, NEW.original_name, NEW.declared_mime,
        NEW.description, NEW.tags, NEW.billing_reservation_id,
        NEW.expires_at, NEW.created_at)
       IS DISTINCT FROM
       (OLD.id, OLD.actor_id, OLD.document_id, OLD.candidate_version_id,
        OLD.expected_current_version_id, OLD.folder_id, OLD.request_sha256,
        OLD.expected_content_sha256, OLD.expected_plaintext_bytes,
        OLD.ciphertext_upper_bound, OLD.original_name, OLD.declared_mime,
        OLD.description, OLD.tags, OLD.billing_reservation_id,
        OLD.expires_at, OLD.created_at) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'DOCUMENT_UPLOAD_IMMUTABLE';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE OR REPLACE TRIGGER documents_uploads_transition_guard
BEFORE UPDATE OR DELETE ON documents_uploads
FOR EACH ROW EXECUTE FUNCTION documents_guard_upload_transition_restore_v1();
ALTER TABLE documents_uploads ENABLE ALWAYS TRIGGER documents_uploads_transition_guard;

CREATE FUNCTION documents_recovery_terminalize_transients_v1(
    p_restore_operation_id uuid
)
RETURNS TABLE(cancelled_uploads bigint, expired_archives bigint, cancelled_archives bigint)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_restore'
       OR p_restore_operation_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'documents transient restore policy denied';
    END IF;
    WITH changed AS (
        UPDATE documents_uploads SET lifecycle_state='CANCELLED',
            lease_owner_id=NULL,lease_expires_at=NULL,cancel_requested_at=v_now,
            cleanup_required=false,updated_at=v_now
        WHERE lifecycle_state NOT IN ('READY','CANCELLED','EXPIRED')
        RETURNING id
    ), disposed AS (
        INSERT INTO documents_restore_transient_dispositions
            (restore_operation_id,target_kind,target_id,disposition,disposed_at)
        SELECT p_restore_operation_id,'UPLOAD',id,'CANCELLED',v_now FROM changed
        ON CONFLICT DO NOTHING
    )
    SELECT pg_catalog.count(*) INTO cancelled_uploads FROM changed;
    WITH changed AS (
        UPDATE documents_archives SET lifecycle_state='EXPIRED',updated_at=v_now
        WHERE lifecycle_state='READY'
        RETURNING id
    ), disposed AS (
        INSERT INTO documents_restore_transient_dispositions
            (restore_operation_id,target_kind,target_id,disposition,disposed_at)
        SELECT p_restore_operation_id,'ARCHIVE',id,'EXPIRED',v_now FROM changed
        ON CONFLICT DO NOTHING
    )
    SELECT pg_catalog.count(*) INTO expired_archives FROM changed;
    WITH changed AS (
        UPDATE documents_archives SET lifecycle_state='CANCELLED',updated_at=v_now
        WHERE lifecycle_state='PENDING'
        RETURNING id
    ), disposed AS (
        INSERT INTO documents_restore_transient_dispositions
            (restore_operation_id,target_kind,target_id,disposition,disposed_at)
        SELECT p_restore_operation_id,'ARCHIVE',id,'CANCELLED',v_now FROM changed
        ON CONFLICT DO NOTHING
    )
    SELECT pg_catalog.count(*) INTO cancelled_archives FROM changed;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd
