-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_guard_revision_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'TRUNCATE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'system resource layout revisions cannot be truncated';
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        IF NEW.state IS DISTINCT FROM 'draft'
           OR NEW.published_at IS NOT NULL
           OR NEW.row_revision IS DISTINCT FROM 1
           OR NEW.updated_at IS DISTINCT FROM NEW.created_at THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                CONSTRAINT = 'system_resource_layout_revisions_lifecycle_guard',
                MESSAGE = 'system resource layout revision must begin as draft';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN
        IF OLD.state IS DISTINCT FROM 'draft' THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                CONSTRAINT = 'system_resource_layout_revisions_lifecycle_guard',
                MESSAGE = 'published system resource layout is immutable';
        END IF;
        RETURN OLD;
    END IF;
    IF TG_OP IS DISTINCT FROM 'UPDATE'
       OR NEW.id IS DISTINCT FROM OLD.id
       OR NEW.resource_kind IS DISTINCT FROM OLD.resource_kind
       OR NEW.version IS DISTINCT FROM OLD.version
       OR NEW.created_at IS DISTINCT FROM OLD.created_at
       OR NEW.row_revision IS DISTINCT FROM OLD.row_revision + 1
       OR NEW.updated_at < OLD.updated_at
       OR OLD.state IS DISTINCT FROM 'draft' THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_revisions_lifecycle_guard',
            MESSAGE = 'system resource layout revision transition is forbidden';
    END IF;
    IF NEW.state IS NOT DISTINCT FROM 'draft'
       AND NEW.published_at IS NULL THEN
        RETURN NEW;
    END IF;
    IF NEW.state IS NOT DISTINCT FROM 'published'
       AND NEW.published_at IS NOT NULL
       AND NEW.definition_json IS NOT DISTINCT FROM OLD.definition_json
       AND NEW.definition_sha256 IS NOT DISTINCT FROM OLD.definition_sha256 THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION USING ERRCODE = '23514',
        CONSTRAINT = 'system_resource_layout_revisions_lifecycle_guard',
        MESSAGE = 'system resource layout publish transition is invalid';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_guard_revision_v1()
    FROM PUBLIC;

CREATE TRIGGER system_resource_layout_revisions_insert_guard
BEFORE INSERT ON system_resource_layout_revisions
FOR EACH ROW EXECUTE FUNCTION system_resource_layout_guard_revision_v1();
ALTER TABLE system_resource_layout_revisions
    ENABLE ALWAYS TRIGGER system_resource_layout_revisions_insert_guard;

CREATE TRIGGER system_resource_layout_revisions_update_guard
BEFORE UPDATE ON system_resource_layout_revisions
FOR EACH ROW EXECUTE FUNCTION system_resource_layout_guard_revision_v1();
ALTER TABLE system_resource_layout_revisions
    ENABLE ALWAYS TRIGGER system_resource_layout_revisions_update_guard;

CREATE TRIGGER system_resource_layout_revisions_delete_guard
BEFORE DELETE ON system_resource_layout_revisions
FOR EACH ROW EXECUTE FUNCTION system_resource_layout_guard_revision_v1();
ALTER TABLE system_resource_layout_revisions
    ENABLE ALWAYS TRIGGER system_resource_layout_revisions_delete_guard;

CREATE TRIGGER system_resource_layout_revisions_truncate_guard
BEFORE TRUNCATE ON system_resource_layout_revisions
FOR EACH STATEMENT EXECUTE FUNCTION system_resource_layout_guard_revision_v1();
ALTER TABLE system_resource_layout_revisions
    ENABLE ALWAYS TRIGGER system_resource_layout_revisions_truncate_guard;

