-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_guard_reference_target_v1() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE item_state text;
BEGIN
    IF NEW.lifecycle_state <> 'ACTIVE' THEN
        RETURN NEW;
    END IF;
    SELECT lifecycle_state INTO item_state
    FROM documents_items WHERE id = NEW.item_id FOR SHARE;
    IF item_state IS DISTINCT FROM 'ACTIVE' THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'DOCUMENT_REFERENCE_TARGET_CLOSED';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER documents_links_target_guard
BEFORE INSERT OR UPDATE OF lifecycle_state ON documents_links
FOR EACH ROW EXECUTE FUNCTION documents_guard_reference_target_v1();
ALTER TABLE documents_links ENABLE ALWAYS TRIGGER documents_links_target_guard;

CREATE TRIGGER documents_legal_holds_target_guard
BEFORE INSERT OR UPDATE OF lifecycle_state ON documents_legal_holds
FOR EACH ROW EXECUTE FUNCTION documents_guard_reference_target_v1();
ALTER TABLE documents_legal_holds ENABLE ALWAYS TRIGGER documents_legal_holds_target_guard;

GRANT EXECUTE ON FUNCTION documents_guard_reference_target_v1()
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION documents_guard_reference_target_v1()
TO forgecrud_adopter;
