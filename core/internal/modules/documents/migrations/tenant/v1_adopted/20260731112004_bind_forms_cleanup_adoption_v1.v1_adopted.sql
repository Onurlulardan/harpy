-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_guard_cleanup_adoption_v1() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $function$
DECLARE
    v_adoption bytea;
    v_catalog bytea;
BEGIN
    IF OLD.lifecycle_state<>'RUNNING' OR NEW.lifecycle_state<>'SUCCEEDED' THEN
        RETURN NEW;
    END IF;
    SELECT receipt.receipt_sha256 INTO v_adoption
    FROM documents_form_record_adoption_receipts AS receipt
    WHERE receipt.definition_id=NEW.definition_id
      AND receipt.reference_epoch=NEW.reference_epoch
      AND receipt.lifecycle_state='SUCCEEDED';
    IF v_adoption IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='DOCUMENTS_FORMS_CLEANUP_ADOPTION_INCOMPLETE';
    END IF;
    SELECT pg_catalog.sha256(pg_catalog.convert_to(
        pg_catalog.encode(v_adoption,'hex') || '|' ||
        COALESCE(pg_catalog.string_agg(
            link.id::text || '|' || link.item_id::text || '|' ||
            link.source_id::text || '|' || link.source_version_id::text || '|' ||
            link.reference_epoch::text,'|' ORDER BY link.id),''),'UTF8'))
    INTO v_catalog
    FROM documents_links AS link
    WHERE link.source_owner='FORMS' AND link.source_kind='form_record'
      AND link.source_id=NEW.definition_id AND link.usage='LIFECYCLE'
      AND link.lifecycle_state='RELEASED' AND link.released_at=OLD.updated_at;
    IF OLD.link_catalog_sha256 IS DISTINCT FROM v_catalog THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='DOCUMENTS_FORMS_CLEANUP_ADOPTION_UNBOUND';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER documents_forms_cleanup_adoption_guard
BEFORE UPDATE ON documents_forms_cleanup_operations
FOR EACH ROW EXECUTE FUNCTION documents_guard_cleanup_adoption_v1();
ALTER TABLE documents_forms_cleanup_operations
    ENABLE ALWAYS TRIGGER documents_forms_cleanup_adoption_guard;

REVOKE ALL ON FUNCTION documents_guard_cleanup_adoption_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_guard_cleanup_adoption_v1()
TO harpy_runtime;
