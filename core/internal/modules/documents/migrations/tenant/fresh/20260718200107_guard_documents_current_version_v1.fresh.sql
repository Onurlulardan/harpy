-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_guard_current_version_v1() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
BEGIN
    IF NEW.lifecycle_state IN ('PURGING','PURGED') THEN
        IF NEW.current_version_id IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'DOCUMENT_DELETING_CURRENT_VERSION';
        END IF;
        RETURN NEW;
    END IF;
    IF NEW.current_version_id IS NULL OR NOT EXISTS (
        SELECT 1 FROM documents_versions version
        WHERE version.id = NEW.current_version_id
          AND version.item_id = NEW.id
          AND version.lifecycle_state = 'READY'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'DOCUMENT_CURRENT_VERSION_INVALID';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE CONSTRAINT TRIGGER documents_items_current_version_guard
AFTER INSERT OR UPDATE OF current_version_id, lifecycle_state ON documents_items
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION documents_guard_current_version_v1();
