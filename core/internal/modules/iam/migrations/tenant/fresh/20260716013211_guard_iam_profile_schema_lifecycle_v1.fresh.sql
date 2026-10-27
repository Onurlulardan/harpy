-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_profile_schema_lifecycle_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'INSERT' AND TG_OP IS DISTINCT FROM 'UPDATE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM profile schema lifecycle invocation is invalid';
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        IF NEW.state IS DISTINCT FROM 'draft' OR NEW.activated_at IS NOT NULL OR
           NEW.retired_at IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                CONSTRAINT = 'iam_profile_schema_versions_lifecycle_guard',
                MESSAGE = 'IAM profile schema must begin as draft';
        END IF;
        RETURN NEW;
    END IF;
    IF NEW.id IS DISTINCT FROM OLD.id OR
       NEW.resource_kind IS DISTINCT FROM OLD.resource_kind OR
       NEW.version IS DISTINCT FROM OLD.version OR
       NEW.created_at IS DISTINCT FROM OLD.created_at OR
       NEW.revision IS DISTINCT FROM OLD.revision + 1 THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_profile_schema_versions_lifecycle_guard',
            MESSAGE = 'IAM profile schema identity or revision is immutable';
    END IF;
    IF OLD.state IS NOT DISTINCT FROM 'draft' AND
       NEW.state IS NOT DISTINCT FROM 'draft' AND
       NEW.activated_at IS NULL AND NEW.retired_at IS NULL THEN
        RETURN NEW;
    END IF;
    IF OLD.state IS NOT DISTINCT FROM 'draft' AND
       NEW.state IS NOT DISTINCT FROM 'active' AND
       NEW.activated_at IS NOT NULL AND NEW.retired_at IS NULL AND
       NEW.definition_sha256 IS NOT DISTINCT FROM
           iam_profile_definition_sha256_v1(NEW.id) THEN
        RETURN NEW;
    END IF;
    IF OLD.state IS NOT DISTINCT FROM 'active' AND
       NEW.state IS NOT DISTINCT FROM 'retired' AND
       NEW.definition_sha256 IS NOT DISTINCT FROM OLD.definition_sha256 AND
       NEW.activated_at IS NOT DISTINCT FROM OLD.activated_at AND
       NEW.retired_at IS NOT NULL THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION USING ERRCODE = '23514',
        CONSTRAINT = 'iam_profile_schema_versions_lifecycle_guard',
        MESSAGE = 'IAM profile schema lifecycle transition is forbidden';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_profile_schema_lifecycle_v1() FROM PUBLIC;

CREATE TRIGGER iam_profile_schema_versions_lifecycle_guard
BEFORE INSERT OR UPDATE ON iam_profile_schema_versions
FOR EACH ROW EXECUTE FUNCTION iam_guard_profile_schema_lifecycle_v1();
ALTER TABLE iam_profile_schema_versions
    ENABLE ALWAYS TRIGGER iam_profile_schema_versions_lifecycle_guard;
