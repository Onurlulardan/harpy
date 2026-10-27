-- +goose Up
CREATE TABLE documents_forms_cleanup_operations (
    operation_id uuid PRIMARY KEY,
    definition_id uuid NOT NULL,
    reference_epoch bigint NOT NULL CHECK (reference_epoch > 0),
    lifecycle_state text NOT NULL
        CHECK (lifecycle_state IN ('REQUESTED','RUNNING','SUCCEEDED')),
    catalog_version text NOT NULL
        CHECK (catalog_version = 'documents.forms-cleanup.v1'),
    source_owner text NOT NULL CHECK (source_owner = 'FORMS'),
    source_kind text NOT NULL CHECK (source_kind = 'form_record'),
    usage text NOT NULL CHECK (usage = 'LIFECYCLE'),
    link_count bigint NOT NULL DEFAULT 0 CHECK (link_count >= 0),
    item_count bigint NOT NULL DEFAULT 0 CHECK (item_count >= 0),
    version_count bigint NOT NULL DEFAULT 0 CHECK (version_count >= 0),
    ciphertext_bytes bigint NOT NULL DEFAULT 0 CHECK (ciphertext_bytes >= 0),
    link_catalog_sha256 bytea CHECK (link_catalog_sha256 IS NULL OR
        pg_catalog.octet_length(link_catalog_sha256) = 32),
    zero_blocker_sha256 bytea CHECK (zero_blocker_sha256 IS NULL OR
        pg_catalog.octet_length(zero_blocker_sha256) = 32),
    receipt_sha256 bytea CHECK (receipt_sha256 IS NULL OR
        pg_catalog.octet_length(receipt_sha256) = 32),
    requested_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    completed_at timestamptz,
    UNIQUE (definition_id, reference_epoch),
    CHECK (pg_catalog.isfinite(requested_at) AND pg_catalog.isfinite(updated_at)),
    CHECK (
        (lifecycle_state = 'REQUESTED' AND link_catalog_sha256 IS NULL AND
         zero_blocker_sha256 IS NULL AND receipt_sha256 IS NULL AND
         completed_at IS NULL) OR
        (lifecycle_state = 'RUNNING' AND link_catalog_sha256 IS NOT NULL AND
         zero_blocker_sha256 IS NULL AND receipt_sha256 IS NULL AND
         completed_at IS NULL) OR
        (lifecycle_state = 'SUCCEEDED' AND link_catalog_sha256 IS NOT NULL AND
         zero_blocker_sha256 IS NOT NULL AND receipt_sha256 IS NOT NULL AND
         completed_at IS NOT NULL AND pg_catalog.isfinite(completed_at))
    )
);
CREATE TABLE documents_forms_cleanup_targets (
    operation_id uuid NOT NULL,
    item_id uuid NOT NULL,
    purge_job_id uuid,
    purge_run_at timestamptz,
    enqueued_at timestamptz,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (operation_id, item_id),
    CONSTRAINT documents_forms_cleanup_targets_operation_fk
        FOREIGN KEY (operation_id)
        REFERENCES documents_forms_cleanup_operations(operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT documents_forms_cleanup_targets_item_fk
        FOREIGN KEY (item_id) REFERENCES documents_items(id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK (
        (purge_job_id IS NULL AND purge_run_at IS NULL AND enqueued_at IS NULL) OR
        (purge_job_id IS NOT NULL AND purge_run_at IS NOT NULL AND
         pg_catalog.isfinite(purge_run_at) AND
         (enqueued_at IS NULL OR pg_catalog.isfinite(enqueued_at)))
    )
);
CREATE INDEX documents_forms_cleanup_pending_idx
    ON documents_forms_cleanup_operations(updated_at, operation_id)
    WHERE lifecycle_state <> 'SUCCEEDED';
CREATE INDEX documents_forms_cleanup_target_job_idx
    ON documents_forms_cleanup_targets(operation_id, item_id)
    WHERE enqueued_at IS NULL;
-- +goose StatementBegin
CREATE FUNCTION documents_guard_forms_cleanup_v1() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $function$
DECLARE
    v_version_count bigint;
    v_ciphertext_bytes bigint;
    v_version_sha256 bytea;
    v_zero_sha256 bytea;
    v_receipt_sha256 bytea;
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'DOCUMENTS_FORMS_CLEANUP_IMMUTABLE';
    END IF;
    IF TG_TABLE_NAME = 'documents_forms_cleanup_targets' THEN
        IF (NEW.operation_id, NEW.item_id, NEW.created_at) IS DISTINCT FROM
           (OLD.operation_id, OLD.item_id, OLD.created_at) THEN
            RAISE EXCEPTION USING ERRCODE = '42501',
                MESSAGE = 'DOCUMENTS_FORMS_CLEANUP_TARGET_IMMUTABLE';
        END IF;
        IF OLD.purge_job_id IS NULL AND NEW.purge_job_id IS NOT NULL AND
           NEW.purge_run_at IS NOT NULL AND NEW.enqueued_at IS NULL THEN
            RETURN NEW;
        END IF;
        IF OLD.purge_job_id IS NOT NULL AND OLD.enqueued_at IS NULL AND
           NEW.enqueued_at IS NOT NULL AND
           (NEW.purge_job_id, NEW.purge_run_at) IS NOT DISTINCT FROM
           (OLD.purge_job_id, OLD.purge_run_at) THEN
            RETURN NEW;
        END IF;
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'DOCUMENTS_FORMS_CLEANUP_TARGET_IMMUTABLE';
    END IF;
    IF (NEW.operation_id, NEW.definition_id, NEW.reference_epoch,
        NEW.catalog_version, NEW.source_owner, NEW.source_kind, NEW.usage,
        NEW.requested_at) IS DISTINCT FROM
       (OLD.operation_id, OLD.definition_id, OLD.reference_epoch,
        OLD.catalog_version, OLD.source_owner, OLD.source_kind, OLD.usage,
        OLD.requested_at) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'DOCUMENTS_FORMS_CLEANUP_AUTHORITY_IMMUTABLE';
    END IF;
    IF OLD.lifecycle_state = 'REQUESTED' AND NEW.lifecycle_state = 'RUNNING' THEN
        RETURN NEW;
    END IF;
    IF OLD.lifecycle_state <> 'RUNNING' OR NEW.lifecycle_state <> 'SUCCEEDED' OR
       EXISTS (
           SELECT 1 FROM documents_links AS link
           WHERE link.source_owner = 'FORMS' AND link.source_kind = 'form_record'
             AND link.source_id = NEW.definition_id AND link.usage = 'LIFECYCLE'
             AND link.lifecycle_state = 'ACTIVE'
       ) OR EXISTS (
           SELECT 1 FROM documents_forms_cleanup_targets AS target
           JOIN documents_items AS item ON item.id = target.item_id
           WHERE target.operation_id = NEW.operation_id
             AND item.lifecycle_state <> 'PURGED'
       ) OR EXISTS (
           SELECT 1 FROM documents_forms_cleanup_targets AS target
           JOIN documents_versions AS version ON version.item_id = target.item_id
           WHERE target.operation_id = NEW.operation_id
             AND (version.lifecycle_state <> 'PURGED' OR NOT EXISTS (
                 SELECT 1 FROM documents_blob_deletions AS deletion
                 WHERE deletion.target_kind = 'VERSION'
                   AND deletion.target_id = version.id
                   AND deletion.lifecycle_state = 'CONFIRMED'
             ))
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'DOCUMENTS_FORMS_CLEANUP_RECEIPT_BLOCKED';
    END IF;
    SELECT count(version.id), COALESCE(pg_catalog.sum(version.ciphertext_bytes), 0),
           pg_catalog.sha256(pg_catalog.convert_to(COALESCE(pg_catalog.string_agg(
               version.id::text || '|' ||
               pg_catalog.encode(version.plaintext_sha256, 'hex') || '|' ||
               pg_catalog.encode(version.delete_receipt_sha256, 'hex'),
               '|' ORDER BY version.id), ''), 'UTF8'))
    INTO v_version_count, v_ciphertext_bytes, v_version_sha256
    FROM documents_forms_cleanup_targets AS target
    JOIN documents_versions AS version ON version.item_id = target.item_id
    WHERE target.operation_id = NEW.operation_id
      AND version.lifecycle_state = 'PURGED';
    v_zero_sha256 := pg_catalog.sha256(pg_catalog.convert_to(
        NEW.operation_id::text || '|' || NEW.definition_id::text || '|' ||
        NEW.reference_epoch::text || '|' || NEW.link_count::text || '|' ||
        NEW.item_count::text || '|' || v_version_count::text || '|' ||
        v_ciphertext_bytes::text || '|' ||
        pg_catalog.encode(NEW.link_catalog_sha256, 'hex') || '|' ||
        pg_catalog.encode(v_version_sha256, 'hex'), 'UTF8'));
    v_receipt_sha256 := pg_catalog.sha256(pg_catalog.convert_to(
        NEW.operation_id::text || '|' || NEW.definition_id::text || '|' ||
        NEW.reference_epoch::text || '|documents.forms-cleanup.v1|' ||
        pg_catalog.encode(v_zero_sha256, 'hex') || '|' ||
        NEW.completed_at::text, 'UTF8'));
    IF (NEW.version_count, NEW.ciphertext_bytes, NEW.zero_blocker_sha256,
        NEW.receipt_sha256) IS DISTINCT FROM
       (v_version_count, v_ciphertext_bytes, v_zero_sha256, v_receipt_sha256) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'DOCUMENTS_FORMS_CLEANUP_RECEIPT_TAMPERED';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd
CREATE TRIGGER documents_forms_cleanup_operation_guard
BEFORE UPDATE OR DELETE ON documents_forms_cleanup_operations
FOR EACH ROW EXECUTE FUNCTION documents_guard_forms_cleanup_v1();
ALTER TABLE documents_forms_cleanup_operations
    ENABLE ALWAYS TRIGGER documents_forms_cleanup_operation_guard;

CREATE TRIGGER documents_forms_cleanup_target_guard
BEFORE UPDATE OR DELETE ON documents_forms_cleanup_targets
FOR EACH ROW EXECUTE FUNCTION documents_guard_forms_cleanup_v1();
ALTER TABLE documents_forms_cleanup_targets
    ENABLE ALWAYS TRIGGER documents_forms_cleanup_target_guard;

GRANT SELECT ON documents_forms_cleanup_operations TO harpy_runtime;
GRANT INSERT (operation_id,definition_id,reference_epoch,lifecycle_state,
 catalog_version,source_owner,source_kind,usage,requested_at,updated_at)
ON documents_forms_cleanup_operations TO harpy_runtime;
GRANT UPDATE (
    lifecycle_state, link_count, item_count, version_count, ciphertext_bytes,
    link_catalog_sha256, zero_blocker_sha256, receipt_sha256,
    updated_at, completed_at
) ON documents_forms_cleanup_operations TO harpy_runtime;
GRANT SELECT ON documents_forms_cleanup_targets TO harpy_runtime;
GRANT INSERT (operation_id, item_id, created_at)
ON documents_forms_cleanup_targets TO harpy_runtime;
GRANT UPDATE (purge_job_id, purge_run_at, enqueued_at)
ON documents_forms_cleanup_targets TO harpy_runtime;
GRANT EXECUTE ON FUNCTION documents_guard_forms_cleanup_v1()
TO harpy_runtime;
