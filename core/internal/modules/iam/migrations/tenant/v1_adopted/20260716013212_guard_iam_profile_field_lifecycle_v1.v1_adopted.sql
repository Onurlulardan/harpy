-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_profile_field_lifecycle_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'INSERT' AND TG_OP IS DISTINCT FROM 'UPDATE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM profile field lifecycle invocation is invalid';
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        IF NEW.state IS DISTINCT FROM 'active' OR NEW.retired_at IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                CONSTRAINT = 'iam_profile_fields_lifecycle_guard',
                MESSAGE = 'IAM profile field must begin as active';
        END IF;
        RETURN NEW;
    END IF;
    IF NEW.id IS DISTINCT FROM OLD.id OR
       NEW.resource_kind IS DISTINCT FROM OLD.resource_kind OR
       NEW.field_key IS DISTINCT FROM OLD.field_key OR
       NEW.value_kind IS DISTINCT FROM OLD.value_kind OR
       NEW.cardinality IS DISTINCT FROM OLD.cardinality OR
       NEW.required IS DISTINCT FROM OLD.required OR
       NEW.classification IS DISTINCT FROM OLD.classification OR
       NEW.normalization_profile IS DISTINCT FROM OLD.normalization_profile OR
       NEW.created_at IS DISTINCT FROM OLD.created_at OR
       NEW.revision IS DISTINCT FROM OLD.revision + 1 OR
       NEW.updated_at < OLD.updated_at OR
       OLD.state IS DISTINCT FROM 'active' OR
       NEW.state IS DISTINCT FROM 'retired' OR NEW.retired_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_profile_fields_lifecycle_guard',
            MESSAGE = 'IAM profile field semantic or lifecycle is immutable';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_profile_field_lifecycle_v1() FROM PUBLIC;

CREATE TRIGGER iam_profile_fields_lifecycle_guard
BEFORE INSERT OR UPDATE ON iam_profile_fields
FOR EACH ROW EXECUTE FUNCTION iam_guard_profile_field_lifecycle_v1();
ALTER TABLE iam_profile_fields
    ENABLE ALWAYS TRIGGER iam_profile_fields_lifecycle_guard;
